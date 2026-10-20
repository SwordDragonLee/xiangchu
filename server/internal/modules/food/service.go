package food

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/xiangchu/server/internal/infra/logger"
	"github.com/xiangchu/server/internal/infra/mq"
	"github.com/xiangchu/server/internal/modules/notify"
	"github.com/xiangchu/server/internal/modules/pay"
	"github.com/xiangchu/server/internal/modules/supplier"
	"github.com/xiangchu/server/internal/pkg/errs"
	"github.com/xiangchu/server/internal/pkg/orderid"
	"go.uber.org/zap"
)

// PayService 支付模块能力（结构性接口，与 pay.Service 方法子集一致，可直接注入）
type PayService interface {
	// CreatePayment 创建支付单：bizType 见 pay.BizTypePurchase(2)；返回支付单号与 JSAPI 调起参数
	CreatePayment(ctx context.Context, bizType int, bizID, userID, amount int64) (*pay.PayResult, error)
	// Refund 按支付单号原路退款
	Refund(ctx context.Context, paymentNo string, refundFen int64) error
	// CreateSettlement 生成结算单：role 1 乡厨 / 2 供应商；net=gross−佣金由 pay 按费率计算
	CreateSettlement(ctx context.Context, role int8, targetID int64, bizNo string, gross int64) error
}

// SupplierService 供应商模块能力（结构性接口，注入 supplier.Service）
type SupplierService interface {
	GetSupplier(ctx context.Context, id int64) (*supplier.Supplier, error)
	GetByUser(ctx context.Context, uid int64) (*supplier.Supplier, error)
}

// Service 食材采购模块服务接口（跨模块只允许通过此接口调用）
type Service interface {
	// 公共市场
	ListIngredients(ctx context.Context, f *IngredientFilter) ([]*IngredientVO, int64, error)
	// 供应商商品管理（uid 为登录用户，内部换算 supplier_id）
	ListOwnIngredients(ctx context.Context, uid int64, f *IngredientFilter) ([]*IngredientVO, int64, error)
	SaveIngredient(ctx context.Context, uid, id int64, req *IngredientReq) (int64, error)
	OffShelfIngredient(ctx context.Context, uid, id int64) error
	// 采购单
	CreatePurchase(ctx context.Context, chefUID int64, req *CreatePurchaseReq) (*CreatePurchaseResp, error)
	MyPurchases(ctx context.Context, chefUID int64, status *int8, page, size int) ([]*PurchaseOrderVO, int64, error)
	ChefPurchase(ctx context.Context, chefUID, id int64) (*PurchaseOrderVO, error)
	ChefCancel(ctx context.Context, chefUID, id int64) error
	// 供应商履约
	SupplierPurchases(ctx context.Context, supUID int64, status *int8, page, size int) ([]*PurchaseOrderVO, int64, error)
	SupplierPurchase(ctx context.Context, supUID, id int64) (*PurchaseOrderVO, error)
	SupplierAccept(ctx context.Context, supUID, id int64) error
	SupplierReject(ctx context.Context, supUID, id int64, reason string) error
	SupplierReady(ctx context.Context, supUID, id int64) error
	SupplierComplete(ctx context.Context, supUID, id int64) error
	// 管理端
	AdminPurchases(ctx context.Context, f *PurchaseFilter) ([]*PurchaseOrderVO, int64, error)
	AdminStats(ctx context.Context) (*PurchaseStatsVO, error)
	// 支付回调（MQ 消费入口）
	HandlePaySuccess(bizID int64) error
}

type serviceImpl struct {
	repo *Repo
	sup  SupplierService
	pay  PayService
	bus  *mq.MQ
}

// NewService 构造食材采购服务
func NewService(repo *Repo, sup SupplierService, pay PayService, bus *mq.MQ) Service {
	return &serviceImpl{repo: repo, sup: sup, pay: pay, bus: bus}
}

// supIDByUser 登录用户 → 供应商档案（校验已通过审核）
func (s *serviceImpl) supIDByUser(ctx context.Context, uid int64) (*supplier.Supplier, error) {
	sup, err := s.sup.GetByUser(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrForbidden
		}
		return nil, err
	}
	if sup.Status != supplier.StatusApproved {
		return nil, errs.ErrStatusDenied
	}
	return sup, nil
}

// ---------- 食材 ----------

func (s *serviceImpl) ListIngredients(ctx context.Context, f *IngredientFilter) ([]*IngredientVO, int64, error) {
	list, total, err := s.repo.ListIngredients(ctx, f, true)
	if err != nil {
		return nil, 0, err
	}
	return toIngredientVOs(list), total, nil
}

func (s *serviceImpl) ListOwnIngredients(ctx context.Context, uid int64, f *IngredientFilter) ([]*IngredientVO, int64, error) {
	sup, err := s.supIDByUser(ctx, uid)
	if err != nil {
		return nil, 0, err
	}
	f.SupplierID = sup.ID
	list, total, err := s.repo.ListIngredients(ctx, f, false)
	if err != nil {
		return nil, 0, err
	}
	return toIngredientVOs(list), total, nil
}

// SaveIngredient 新增/编辑食材（id>0 时校验归属；编辑可直接盘点库存）
func (s *serviceImpl) SaveIngredient(ctx context.Context, uid, id int64, req *IngredientReq) (int64, error) {
	sup, err := s.supIDByUser(ctx, uid)
	if err != nil {
		return 0, err
	}
	images, _ := json.Marshal(req.Images)
	if id > 0 {
		ing, err := s.repo.GetOwnIngredient(ctx, sup.ID, id)
		if err != nil {
			return 0, err
		}
		ing.Name, ing.Category, ing.Unit = req.Name, req.Category, req.Unit
		ing.Price, ing.Origin, ing.Stock = req.Price, req.Origin, req.Stock
		ing.Images = string(images)
		if err := s.repo.UpdateIngredient(ctx, ing); err != nil {
			return 0, err
		}
		return ing.ID, nil
	}
	ing := &Ingredient{
		SupplierID: sup.ID, Name: req.Name, Category: req.Category,
		Unit: req.Unit, Price: req.Price, Stock: req.Stock,
		Origin: req.Origin, Images: string(images), Status: IngStatusOn,
	}
	if err := s.repo.CreateIngredient(ctx, ing); err != nil {
		return 0, err
	}
	return ing.ID, nil
}

// OffShelfIngredient 下架（逻辑删除，不物理删）
func (s *serviceImpl) OffShelfIngredient(ctx context.Context, uid, id int64) error {
	sup, err := s.supIDByUser(ctx, uid)
	if err != nil {
		return err
	}
	if _, err := s.repo.GetOwnIngredient(ctx, sup.ID, id); err != nil {
		return err
	}
	return s.repo.SetIngredientStatus(ctx, sup.ID, id, IngStatusOff)
}

// toIngredientVOs 模型转视图（images JSON 解析失败降级为空数组）
func toIngredientVOs(list []*Ingredient) []*IngredientVO {
	vos := make([]*IngredientVO, 0, len(list))
	for _, ing := range list {
		var imgs []string
		_ = json.Unmarshal([]byte(ing.Images), &imgs)
		vos = append(vos, &IngredientVO{
			ID: ing.ID, SupplierID: ing.SupplierID, Name: ing.Name,
			Category: ing.Category, Unit: ing.Unit, Price: ing.Price,
			Stock: ing.Stock, Origin: ing.Origin, Images: imgs,
			Status: ing.Status, CreatedAt: ing.CreatedAt,
		})
	}
	return vos
}

// ---------- 采购下单 ----------

// CreatePurchase 乡厨采购下单：校验供应商/食材/履约信息，落单后创建支付单
func (s *serviceImpl) CreatePurchase(ctx context.Context, chefUID int64, req *CreatePurchaseReq) (*CreatePurchaseResp, error) {
	chefID, err := s.repo.ChefIDByUser(ctx, chefUID)
	if err != nil {
		return nil, err
	}
	if req.OrderID > 0 {
		owner, err := s.repo.OrderOwner(ctx, req.OrderID)
		if err != nil {
			return nil, err
		}
		if owner != chefUID {
			return nil, errs.ErrForbidden
		}
	}
	sup, err := s.sup.GetSupplier(ctx, req.SupplierID)
	if err != nil {
		return nil, errs.ErrNotFound
	}
	if sup.Status != supplier.StatusApproved {
		return nil, errs.ErrStatusDenied
	}
	if req.FulfillMethod == FulfillDeliver {
		if req.DeliverAddress == "" {
			return nil, errs.BadRequest("送货需填写收货地址")
		}
		if req.DeliverTime == nil || !req.DeliverTime.After(time.Now()) {
			return nil, errs.BadRequest("送达时间必须晚于当前时间")
		}
	}
	po, items, err := s.buildOrder(ctx, req, chefID, sup)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreatePurchase(ctx, po, items); err != nil {
		return nil, err
	}
	return s.payFor(ctx, po, chefUID)
}

// buildOrder 校验食材并生成采购单与明细快照（接单时才锁库存）
func (s *serviceImpl) buildOrder(ctx context.Context, req *CreatePurchaseReq, chefID int64, sup *supplier.Supplier) (*PurchaseOrder, []*PurchaseItem, error) {
	ids := make([]int64, 0, len(req.Items))
	for _, it := range req.Items {
		ids = append(ids, it.IngredientID)
	}
	ings, err := s.repo.ListIngredientsByIDs(ctx, ids)
	if err != nil {
		return nil, nil, err
	}
	byID := make(map[int64]*Ingredient, len(ings))
	for _, ing := range ings {
		byID[ing.ID] = ing
	}
	items := make([]*PurchaseItem, 0, len(req.Items))
	var total int64
	for _, it := range req.Items {
		ing, ok := byID[it.IngredientID]
		if !ok || ing.SupplierID != sup.ID || ing.Status != IngStatusOn {
			return nil, nil, errs.BadRequest("食材不可用，请刷新后重试")
		}
		if ing.Stock <= 0 {
			return nil, nil, errs.ErrStockShort // 下单仅软校验，接单时才锁库存
		}
		sub := ing.Price * int64(it.Quantity)
		total += sub
		items = append(items, &PurchaseItem{
			IngredientID: ing.ID, Name: ing.Name, Unit: ing.Unit,
			Price: ing.Price, Quantity: it.Quantity, Subtotal: sub,
		})
	}
	po := &PurchaseOrder{
		PurchaseNo: orderid.GenNo("CG"), ChefID: chefID, SupplierID: sup.ID,
		OrderID: req.OrderID, TotalAmount: total, FulfillMethod: req.FulfillMethod,
		DeliverAddress: req.DeliverAddress, DeliverTime: req.DeliverTime,
		Remark: req.Remark, Status: POStatusPending,
	}
	if req.FulfillMethod == FulfillPickup {
		po.PickupAddress = sup.Address // 自取：供应商地址快照
	}
	return po, items, nil
}

// payFor 创建支付单（bizType=2）；pay 模块未接入时降级为仅返回采购单信息
func (s *serviceImpl) payFor(ctx context.Context, po *PurchaseOrder, chefUID int64) (*CreatePurchaseResp, error) {
	resp := &CreatePurchaseResp{
		PurchaseID: po.ID, PurchaseNo: po.PurchaseNo, TotalAmount: po.TotalAmount,
	}
	if s.pay == nil {
		return resp, nil
	}
	res, err := s.pay.CreatePayment(ctx, int(pay.BizTypePurchase), po.ID, chefUID, po.TotalAmount)
	if err != nil {
		return nil, err
	}
	resp.PaymentNo, resp.Params = res.PaymentNo, res.Params
	return resp, nil
}

// ---------- 支付回调与查询 ----------

// HandlePaySuccess 采购支付成功：标记实付并通知供应商接单（幂等）
func (s *serviceImpl) HandlePaySuccess(bizID int64) error {
	ctx := context.Background()
	po, err := s.repo.GetPurchase(ctx, bizID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if po.PaidAmount > 0 {
		return nil
	}
	if err := s.repo.MarkPaid(ctx, po.ID, po.TotalAmount); err != nil {
		return err
	}
	uid, err := s.repo.SupplierUIDByID(ctx, po.SupplierID)
	if err != nil || uid == 0 {
		logger.L().Warn("采购支付回调：供应商用户缺失", zap.Int64("supplier_id", po.SupplierID))
		return nil
	}
	notify.Send(s.bus, &notify.Event{
		Type:    "purchase.paid",
		Title:   "新采购订单",
		Content: fmt.Sprintf("采购单 %s 已支付，请及时接单", po.PurchaseNo),
		UserIDs: []int64{uid},
	})
	return nil
}

// MyPurchases 乡厨采购单列表
func (s *serviceImpl) MyPurchases(ctx context.Context, chefUID int64, status *int8, page, size int) ([]*PurchaseOrderVO, int64, error) {
	chefID, err := s.repo.ChefIDByUser(ctx, chefUID)
	if err != nil {
		return nil, 0, err
	}
	list, total, err := s.repo.ListPurchases(ctx, &PurchaseFilter{ChefID: chefID, Status: status, Page: page, Size: size})
	if err != nil {
		return nil, 0, err
	}
	return s.toPurchaseVOs(ctx, list), total, nil
}

// ChefPurchase 乡厨采购单详情（仅本人）
func (s *serviceImpl) ChefPurchase(ctx context.Context, chefUID, id int64) (*PurchaseOrderVO, error) {
	chefID, err := s.repo.ChefIDByUser(ctx, chefUID)
	if err != nil {
		return nil, err
	}
	po, err := s.repo.GetPurchase(ctx, id)
	if err != nil {
		return nil, err
	}
	if po.ChefID != chefID {
		return nil, errs.ErrForbidden
	}
	return s.toPurchaseVO(ctx, po)
}

// SupplierPurchase 供应商采购单详情（仅本人）
func (s *serviceImpl) SupplierPurchase(ctx context.Context, supUID, id int64) (*PurchaseOrderVO, error) {
	sup, err := s.supIDByUser(ctx, supUID)
	if err != nil {
		return nil, err
	}
	po, err := s.repo.GetPurchase(ctx, id)
	if err != nil {
		return nil, err
	}
	if po.SupplierID != sup.ID {
		return nil, errs.ErrForbidden
	}
	return s.toPurchaseVO(ctx, po)
}

// toPurchaseVOs 列表转视图（不带明细，避免 N+1）
func (s *serviceImpl) toPurchaseVOs(ctx context.Context, list []*PurchaseOrder) []*PurchaseOrderVO {
	vos := make([]*PurchaseOrderVO, 0, len(list))
	for _, po := range list {
		vo, err := s.toPurchaseVO(ctx, po)
		if err != nil {
			continue
		}
		vos = append(vos, vo)
	}
	return vos
}

// toPurchaseVO 单个采购单转视图（带明细与供应商名称）
func (s *serviceImpl) toPurchaseVO(ctx context.Context, po *PurchaseOrder) (*PurchaseOrderVO, error) {
	items, err := s.repo.ListPurchaseItems(ctx, po.ID)
	if err != nil {
		return nil, err
	}
	vo := &PurchaseOrderVO{
		ID: po.ID, PurchaseNo: po.PurchaseNo, ChefID: po.ChefID,
		SupplierID: po.SupplierID, OrderID: po.OrderID,
		TotalAmount: po.TotalAmount, PaidAmount: po.PaidAmount,
		FulfillMethod: po.FulfillMethod, PickupAddress: po.PickupAddress,
		DeliverAddress: po.DeliverAddress, DeliverTime: po.DeliverTime,
		Status: po.Status, Remark: po.Remark, CreatedAt: po.CreatedAt,
		Items: make([]*PurchaseItemVO, 0, len(items)),
	}
	for _, it := range items {
		vo.Items = append(vo.Items, &PurchaseItemVO{
			ID: it.ID, IngredientID: it.IngredientID, Name: it.Name,
			Unit: it.Unit, Price: it.Price, Quantity: it.Quantity, Subtotal: it.Subtotal,
		})
	}
	if sup, err := s.sup.GetSupplier(ctx, po.SupplierID); err == nil {
		vo.SupplierName = sup.Name
	}
	return vo, nil
}

// ---------- 供应商履约 ----------

// SupplierPurchases 供应商采购单列表（status 可选）
func (s *serviceImpl) SupplierPurchases(ctx context.Context, supUID int64, status *int8, page, size int) ([]*PurchaseOrderVO, int64, error) {
	sup, err := s.supIDByUser(ctx, supUID)
	if err != nil {
		return nil, 0, err
	}
	list, total, err := s.repo.ListPurchases(ctx, &PurchaseFilter{
		SupplierID: sup.ID, Status: status, Page: page, Size: size,
	})
	if err != nil {
		return nil, 0, err
	}
	return s.toPurchaseVOs(ctx, list), total, nil
}

// SupplierAccept 接单：0→1 并锁定库存，任一行库存不足即整体失败
func (s *serviceImpl) SupplierAccept(ctx context.Context, supUID, id int64) error {
	po, err := s.ownPendingPO(ctx, supUID, id)
	if err != nil {
		return err
	}
	items, err := s.repo.ListPurchaseItems(ctx, po.ID)
	if err != nil {
		return err
	}
	n, err := s.repo.Transition(ctx, po.ID, POStatusPending, POStatusPrepping)
	if err != nil {
		return err
	}
	if n == 0 {
		return errs.ErrStatusDenied
	}
	// 锁定库存；不足则回滚状态
	if err := s.repo.LockStock(ctx, items); err != nil {
		_, _ = s.repo.Transition(ctx, po.ID, POStatusPrepping, POStatusPending)
		return err
	}
	return s.notifyChef(ctx, po, "purchase.accepted", "采购单已接单",
		fmt.Sprintf("采购单 %s 供应商已接单，开始备货", po.PurchaseNo))
}

// SupplierReject 拒单：0→4 已取消；已支付则原路退款
func (s *serviceImpl) SupplierReject(ctx context.Context, supUID, id int64, reason string) error {
	po, err := s.ownPendingPO(ctx, supUID, id)
	if err != nil {
		return err
	}
	if err := s.transitionOrRefund(ctx, po, reason); err != nil {
		return err
	}
	return s.notifyChef(ctx, po, "purchase.rejected", "采购单被拒单",
		fmt.Sprintf("采购单 %s 供应商拒单：%s", po.PurchaseNo, reason))
}

// ChefCancel 乡厨取消：仅待接单可取消；已支付则原路退款
func (s *serviceImpl) ChefCancel(ctx context.Context, chefUID, id int64) error {
	chefID, err := s.repo.ChefIDByUser(ctx, chefUID)
	if err != nil {
		return err
	}
	po, err := s.repo.GetPurchase(ctx, id)
	if err != nil {
		return err
	}
	if po.ChefID != chefID {
		return errs.ErrForbidden
	}
	if po.Status != POStatusPending {
		return errs.ErrStatusDenied // 备货中不可取消，请线下协商
	}
	if err := s.transitionOrRefund(ctx, po, "乡厨取消"); err != nil {
		return err
	}
	return s.notifySupplier(ctx, po, "purchase.cancelled", "采购单已取消",
		fmt.Sprintf("采购单 %s 乡厨已取消", po.PurchaseNo))
}

// SupplierReady 备货完成：1→2（自取=待自取，送货=待送达）
func (s *serviceImpl) SupplierReady(ctx context.Context, supUID, id int64) error {
	sup, err := s.supIDByUser(ctx, supUID)
	if err != nil {
		return err
	}
	po, err := s.repo.GetPurchase(ctx, id)
	if err != nil {
		return err
	}
	if po.SupplierID != sup.ID {
		return errs.ErrForbidden
	}
	n, err := s.repo.Transition(ctx, po.ID, POStatusPrepping, POStatusAwaiting)
	if err != nil {
		return err
	}
	if n == 0 {
		return errs.ErrStatusDenied
	}
	action := "已备货完成，请前往自取"
	if po.FulfillMethod == FulfillDeliver {
		action = "已备货完成，将按约送达"
	}
	return s.notifyChef(ctx, po, "purchase.ready", "采购单备货完成",
		fmt.Sprintf("采购单 %s %s", po.PurchaseNo, action))
}

// SupplierComplete 完成履约：2→3 并生成供应商结算（自取=已取走，送货=已送达）
func (s *serviceImpl) SupplierComplete(ctx context.Context, supUID, id int64) error {
	sup, err := s.supIDByUser(ctx, supUID)
	if err != nil {
		return err
	}
	po, err := s.repo.GetPurchase(ctx, id)
	if err != nil {
		return err
	}
	if po.SupplierID != sup.ID {
		return errs.ErrForbidden
	}
	n, err := s.repo.Transition(ctx, po.ID, POStatusAwaiting, POStatusCompleted)
	if err != nil {
		return err
	}
	if n == 0 {
		return errs.ErrStatusDenied
	}
	if s.pay != nil {
		if err := s.pay.CreateSettlement(ctx, 2, sup.ID, po.PurchaseNo, po.TotalAmount); err != nil {
			logger.L().Error("采购结算生成失败",
				zap.Int64("purchase_id", po.ID), zap.String("purchase_no", po.PurchaseNo), zap.Error(err))
		}
	}
	return s.notifyChef(ctx, po, "purchase.completed", "采购单已完成",
		fmt.Sprintf("采购单 %s 已完成履约", po.PurchaseNo))
}

// ---------- 内部辅助 ----------

// ownPendingPO 校验供应商归属且处于待接单
func (s *serviceImpl) ownPendingPO(ctx context.Context, supUID, id int64) (*PurchaseOrder, error) {
	sup, err := s.supIDByUser(ctx, supUID)
	if err != nil {
		return nil, err
	}
	po, err := s.repo.GetPurchase(ctx, id)
	if err != nil {
		return nil, err
	}
	if po.SupplierID != sup.ID {
		return nil, errs.ErrForbidden
	}
	if po.Status != POStatusPending {
		return nil, errs.ErrStatusDenied
	}
	return po, nil
}

// transitionOrRefund 取消/拒单：0→4，已支付则原路退款
func (s *serviceImpl) transitionOrRefund(ctx context.Context, po *PurchaseOrder, reason string) error {
	n, err := s.repo.Transition(ctx, po.ID, POStatusPending, POStatusCancelled)
	if err != nil {
		return err
	}
	if n == 0 {
		return errs.ErrStatusDenied
	}
	if po.PaidAmount > 0 {
		s.refund(ctx, po, reason)
	}
	return nil
}

// refund 原路退款（失败仅记日志，取消状态不回滚，由人工对账）
func (s *serviceImpl) refund(ctx context.Context, po *PurchaseOrder, reason string) {
	if s.pay == nil {
		return
	}
	payNo, err := s.repo.LatestPaidPaymentNo(ctx, po.ID)
	if err != nil || payNo == "" {
		logger.L().Warn("采购退款：未找到有效支付单",
			zap.Int64("purchase_id", po.ID), zap.Error(err))
		return
	}
	if err := s.pay.Refund(ctx, payNo, po.PaidAmount); err != nil {
		logger.L().Error("采购退款失败",
			zap.Int64("purchase_id", po.ID), zap.String("payment_no", payNo), zap.Error(err))
	}
}

// notifyChef 通知乡厨（chefs.id → user_id 只读换算）
func (s *serviceImpl) notifyChef(ctx context.Context, po *PurchaseOrder, typ, title, content string) error {
	uid, err := s.repo.ChefUIDByID(ctx, po.ChefID)
	if err != nil || uid == 0 {
		logger.L().Warn("采购通知：乡厨用户缺失", zap.Int64("chef_id", po.ChefID))
		return nil
	}
	notify.Send(s.bus, &notify.Event{Type: typ, Title: title, Content: content, UserIDs: []int64{uid}})
	return nil
}

// notifySupplier 通知供应商（suppliers.id → user_id 只读换算）
func (s *serviceImpl) notifySupplier(ctx context.Context, po *PurchaseOrder, typ, title, content string) error {
	uid, err := s.repo.SupplierUIDByID(ctx, po.SupplierID)
	if err != nil || uid == 0 {
		logger.L().Warn("采购通知：供应商用户缺失", zap.Int64("supplier_id", po.SupplierID))
		return nil
	}
	notify.Send(s.bus, &notify.Event{Type: typ, Title: title, Content: content, UserIDs: []int64{uid}})
	return nil
}

// ---------- 管理端 ----------

func (s *serviceImpl) AdminPurchases(ctx context.Context, f *PurchaseFilter) ([]*PurchaseOrderVO, int64, error) {
	list, total, err := s.repo.ListPurchases(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	return s.toPurchaseVOs(ctx, list), total, nil
}

// AdminStats 采购监控统计
func (s *serviceImpl) AdminStats(ctx context.Context) (*PurchaseStatsVO, error) {
	return s.repo.PurchaseStats(ctx)
}

// ---------- MQ 消费 ----------

// PaySuccessEvent pay.success 事件体（与 pay 模块发布结构对齐，biz_type 2 为采购）
type PaySuccessEvent struct {
	PaymentNo string `json:"payment_no"`
	BizType   int8   `json:"biz_type"`
	BizID     int64  `json:"biz_id"`
	UserID    int64  `json:"user_id"`
	Amount    int64  `json:"amount"`
}

// RegisterConsumers 注册采购模块消费者：过滤 biz_type==2 的支付成功事件
func RegisterConsumers(bus *mq.MQ, svc Service) {
	bus.Consume(mq.QueueFoodPay, func(ctx context.Context, body []byte) error {
		var ev PaySuccessEvent
		if err := json.Unmarshal(body, &ev); err != nil {
			return err
		}
		if ev.BizType != 2 {
			return nil // 非采购支付，交给其他模块的消费者
		}
		return svc.HandlePaySuccess(ev.BizID)
	})
}
