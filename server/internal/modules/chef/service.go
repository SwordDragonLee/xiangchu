package chef

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xiangchu/server/internal/infra/logger"
	"github.com/xiangchu/server/internal/infra/mq"
	"github.com/xiangchu/server/internal/modules/pay"
	"github.com/xiangchu/server/internal/pkg/errs"
)

// Service 乡厨模块服务接口（跨模块只允许通过此接口调用）
type Service interface {
	HasApprovedProfile(ctx context.Context, uid int64) bool
	// FilterApproved 批量查询：uids 中有已通过审核档案的子集（管理端用户列表拼角色用）
	FilterApproved(ctx context.Context, uids []int64) (map[int64]bool, error)
	GetChef(ctx context.Context, id int64) (*Chefs, error)
	GetChefByUser(ctx context.Context, uid int64) (*Chefs, error) // uid 无档案 → gorm.ErrRecordNotFound
	LockSchedule(ctx context.Context, chefID int64, date time.Time, orderID int64) error
	ReleaseSchedule(ctx context.Context, chefID, orderID int64) error
	RecalcRating(ctx context.Context, chefID int64) (float64, error)
	List(ctx context.Context, f *ListFilter) ([]*Chefs, int64, error)
	Detail(ctx context.Context, id int64) (*ChefDetailVO, error)
	Apply(ctx context.Context, uid int64, req *ApplyReq) error
	UpdateOwnProfile(ctx context.Context, uid int64, req *UpdateReq) error
	Schedules(ctx context.Context, chefID int64, from, to time.Time) ([]*ChefSchedule, error)
	SetRestDays(ctx context.Context, chefUID int64, dates []time.Time, rest bool) error
	Packages(ctx context.Context, chefID int64) ([]*Package, error)
	SavePackage(ctx context.Context, chefUID int64, req *PackageReq) (int64, error)
	DeletePackage(ctx context.Context, chefUID, packageID int64) error
	Workbench(ctx context.Context, chefUID int64) (*WorkbenchVO, error)
	IncOrderCount(ctx context.Context, chefID int64) error
	Stats(ctx context.Context) (*ChefStatsVO, error)
	AdminList(ctx context.Context, keyword string, status *int8, page, size int) ([]*Chefs, int64, error)
	Audit(ctx context.Context, id int64, approve bool, remark string) error
}

// ---------- impl ----------

const (
	scheduleLockTTL    = 48 * time.Hour // Redis 档期锁 TTL
	orderStatusPending = 1              // orders 表“已支付/待处理”状态值（order 模块定义）
	ratingDecimalKeep  = 100.0          // 评分保留两位小数
)

type serviceImpl struct {
	repo   *Repo
	rdb    *redis.Client
	paySvc pay.Service // 结算收入查询（bootstrap 注入）
}

func NewService(repo *Repo, rdb *redis.Client, paySvc pay.Service) Service {
	return &serviceImpl{repo: repo, rdb: rdb, paySvc: paySvc}
}

var _ Service = (*serviceImpl)(nil)

// dayStart 截断到当日 0 点（本地时区）
func dayStart(t time.Time) time.Time {
	t = t.Local()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// scheduleKey Redis 档期锁键：chef:schedule:{chefID}:{yyyyMMdd}
func scheduleKey(chefID int64, date time.Time) string {
	return fmt.Sprintf("chef:schedule:%d:%s", chefID, date.Format("20060102"))
}

// HasApprovedProfile 登录角色判定用：是否有已通过审核的乡厨档案
func (s *serviceImpl) HasApprovedProfile(ctx context.Context, uid int64) bool {
	n, err := s.repo.CountApprovedByUser(ctx, uid)
	if err != nil {
		logger.L().Warn("查询乡厨审核状态失败", zap.Int64("uid", uid), zap.Error(err))
		return false
	}
	return n > 0
}

// FilterApproved 批量版 HasApprovedProfile，整页用户只查一次
func (s *serviceImpl) FilterApproved(ctx context.Context, uids []int64) (map[int64]bool, error) {
	return s.repo.FilterApprovedUsers(ctx, uids)
}

func (s *serviceImpl) GetChef(ctx context.Context, id int64) (*Chefs, error) {
	return s.repo.GetChef(ctx, id)
}

func (s *serviceImpl) GetChefByUser(ctx context.Context, uid int64) (*Chefs, error) {
	return s.repo.GetByUser(ctx, uid)
}

// LockSchedule 支付成功锁档期。
// Redis SETNX 抢锁（TTL 48h）→ DB upsert；SETNX 失败时若为同一订单则幂等成功，
// 否则视为档期冲突；DB 落库失败时回删 Redis 锁。
func (s *serviceImpl) LockSchedule(ctx context.Context, chefID int64, date time.Time, orderID int64) error {
	d := dayStart(date)
	key := scheduleKey(chefID, d)
	ok, err := s.rdb.SetNX(ctx, key, orderID, scheduleLockTTL).Result()
	if err != nil {
		return err
	}
	if !ok {
		// 锁已存在：同一订单重复加锁 → 幂等成功
		row, gerr := s.repo.GetSchedule(ctx, chefID, d)
		if gerr == nil && row.Status == ScheduleLocked && row.OrderID == orderID {
			return nil
		}
		return errs.ErrScheduleTaken
	}
	// 兜底：Redis 锁过期但 DB 行仍被其他订单锁定
	row, gerr := s.repo.GetSchedule(ctx, chefID, d)
	switch {
	case gerr == nil && row.Status == ScheduleLocked && row.OrderID != orderID:
		_ = s.rdb.Del(ctx, key).Err()
		return errs.ErrScheduleTaken
	case gerr != nil && !errors.Is(gerr, gorm.ErrRecordNotFound):
		_ = s.rdb.Del(ctx, key).Err()
		return gerr
	}
	if err := s.repo.UpsertSchedule(ctx, &ChefSchedule{
		ChefID:  chefID,
		Date:    d,
		Status:  ScheduleLocked,
		OrderID: orderID,
	}); err != nil {
		_ = s.rdb.Del(ctx, key).Err() // 回滚 Redis 锁
		return err
	}
	return nil
}

// ReleaseSchedule 取消/退款时释放档期：还原 DB 行并回删 Redis 键
func (s *serviceImpl) ReleaseSchedule(ctx context.Context, chefID, orderID int64) error {
	rows, err := s.repo.ListLockedByOrder(ctx, chefID, orderID)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	if err := s.repo.ReleaseSchedules(ctx, chefID, orderID); err != nil {
		return err
	}
	for _, row := range rows {
		_ = s.rdb.Del(ctx, scheduleKey(chefID, row.Date)).Err()
	}
	return nil
}

// RecalcRating 重算乡厨评分（评价事件触发）。
// reviews 表归属 order 模块，此处为经批准的只读聚合特例（见 repo.ReviewAgg）。
func (s *serviceImpl) RecalcRating(ctx context.Context, chefID int64) (float64, error) {
	avg, cnt, err := s.repo.ReviewAgg(ctx, chefID)
	if err != nil {
		return 0, err
	}
	if cnt == 0 {
		avg = 0
	}
	avg = math.Round(avg*ratingDecimalKeep) / ratingDecimalKeep
	if err := s.repo.UpdateRating(ctx, chefID, avg); err != nil {
		return 0, err
	}
	return avg, nil
}

// List 公开浏览列表：仅返回已通过审核（status=1）的乡厨
func (s *serviceImpl) List(ctx context.Context, f *ListFilter) ([]*Chefs, int64, error) {
	if f == nil {
		f = &ListFilter{}
	}
	return s.repo.ListApproved(ctx, f)
}

// Detail 详情：档案 + 上架套餐 + 案例；评价由 order 模块端点返回
func (s *serviceImpl) Detail(ctx context.Context, id int64) (*ChefDetailVO, error) {
	c, err := s.repo.GetChef(ctx, id)
	if err != nil {
		return nil, err
	}
	pkgs, err := s.repo.ListPackages(ctx, c.ID, true)
	if err != nil {
		return nil, err
	}
	cases, err := s.repo.ListCases(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	return &ChefDetailVO{Chef: c, Packages: pkgs, Cases: cases}, nil
}

// Apply 入驻/重新提交审核：按 user_id upsert，状态置为待审核
func (s *serviceImpl) Apply(ctx context.Context, uid int64, req *ApplyReq) error {
	if req.RealName == "" || req.Phone == "" {
		return errs.BadRequest("姓名和手机号不能为空")
	}
	c, err := s.repo.GetByUser(ctx, uid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.repo.CreateChef(ctx, &Chefs{
			UserID:     uid,
			RealName:   req.RealName,
			Phone:      req.Phone,
			Region:     req.Region,
			Cuisines:   req.Cuisines,
			Experience: req.Experience,
			Status:     ChefStatusPending,
		})
	}
	if err != nil {
		return err
	}
	c.RealName = req.RealName
	c.Phone = req.Phone
	c.Region = req.Region
	c.Cuisines = req.Cuisines
	c.Experience = req.Experience
	c.Status = ChefStatusPending
	c.AuditRemark = ""
	return s.repo.UpdateChef(ctx, c,
		"real_name", "phone", "region", "cuisines", "experience", "status", "audit_remark")
}

// UpdateOwnProfile 乡厨维护资料（空字段不更新，不改审核状态）
func (s *serviceImpl) UpdateOwnProfile(ctx context.Context, uid int64, req *UpdateReq) error {
	c, err := s.repo.GetByUser(ctx, uid)
	if err != nil {
		return err
	}
	fields := make([]string, 0, 5)
	if req.RealName != "" {
		c.RealName = req.RealName
		fields = append(fields, "real_name")
	}
	if req.Phone != "" {
		c.Phone = req.Phone
		fields = append(fields, "phone")
	}
	if req.Region != "" {
		c.Region = req.Region
		fields = append(fields, "region")
	}
	if req.Cuisines != "" {
		c.Cuisines = req.Cuisines
		fields = append(fields, "cuisines")
	}
	if req.Experience != "" {
		c.Experience = req.Experience
		fields = append(fields, "experience")
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.UpdateChef(ctx, c, fields...)
}

// Schedules 查询 [from, to) 内的档期（公开日历与个人档期共用）
func (s *serviceImpl) Schedules(ctx context.Context, chefID int64, from, to time.Time) ([]*ChefSchedule, error) {
	return s.repo.ListSchedules(ctx, chefID, dayStart(from), dayStart(to).AddDate(0, 0, 1))
}

// SetRestDays 乡厨标记/取消休息日；已锁定的档期不可改动
func (s *serviceImpl) SetRestDays(ctx context.Context, chefUID int64, dates []time.Time, rest bool) error {
	c, err := s.repo.GetByUser(ctx, chefUID)
	if err != nil {
		return err
	}
	days := make([]time.Time, 0, len(dates))
	for _, d := range dates {
		days = append(days, dayStart(d))
	}
	rows, err := s.repo.ListSchedulesByDates(ctx, c.ID, days)
	if err != nil {
		return err
	}
	exist := make(map[int64]*ChefSchedule, len(rows))
	for _, row := range rows {
		exist[row.Date.Local().Unix()] = row
	}
	for _, d := range days {
		row, has := exist[d.Unix()]
		if has && row.Status == ScheduleLocked {
			continue // 已锁定档期不可改
		}
		if !rest {
			// 取消休息日 → 恢复可约（仅对休息日行生效）
			if has && row.Status == ScheduleRest {
				if err := s.repo.SetScheduleStatus(ctx, row.ID, ScheduleFree); err != nil {
					return err
				}
			}
			continue
		}
		if has && row.Status == ScheduleRest {
			continue
		}
		if err := s.repo.UpsertSchedule(ctx, &ChefSchedule{
			ChefID: c.ID, Date: d, Status: ScheduleRest, OrderID: 0,
		}); err != nil {
			return err
		}
	}
	return nil
}

// Packages 套餐列表（含未上架，供乡厨管理端使用；公开详情走 Detail 内部查询）
func (s *serviceImpl) Packages(ctx context.Context, chefID int64) ([]*Package, error) {
	return s.repo.ListPackages(ctx, chefID, false)
}

// SavePackage 创建/更新本人套餐；金额单位分。返回套餐 ID。
func (s *serviceImpl) SavePackage(ctx context.Context, chefUID int64, req *PackageReq) (int64, error) {
	c, err := s.repo.GetByUser(ctx, chefUID)
	if err != nil {
		return 0, err
	}
	if req.Dishes != "" && !json.Valid([]byte(req.Dishes)) {
		return 0, errs.BadRequest("dishes 须为 JSON 数组字符串")
	}
	if req.ID > 0 {
		p, err := s.repo.GetPackage(ctx, req.ID)
		if err != nil {
			return 0, err
		}
		if p.ChefID != c.ID {
			return 0, errs.ErrForbidden
		}
		p.Name = req.Name
		p.BanquetType = req.BanquetType
		p.PricePerTable = req.PricePerTable
		p.Dishes = req.Dishes
		p.MinTables = req.MinTables
		p.Status = req.Status
		p.Sort = req.Sort
		if err := s.repo.UpdatePackage(ctx, p,
			"name", "banquet_type", "price_per_table", "dishes", "min_tables", "status", "sort"); err != nil {
			return 0, err
		}
		return p.ID, nil
	}
	status := req.Status
	if status == PackageOff {
		status = PackageOn // 创建未指定状态默认上架
	}
	p := &Package{
		ChefID:        c.ID,
		Name:          req.Name,
		BanquetType:   req.BanquetType,
		PricePerTable: req.PricePerTable,
		Dishes:        req.Dishes,
		MinTables:     req.MinTables,
		Status:        status,
		Sort:          req.Sort,
	}
	if err := s.repo.CreatePackage(ctx, p); err != nil {
		return 0, err
	}
	return p.ID, nil
}

// DeletePackage 删除本人套餐
func (s *serviceImpl) DeletePackage(ctx context.Context, chefUID, packageID int64) error {
	c, err := s.repo.GetByUser(ctx, chefUID)
	if err != nil {
		return err
	}
	return s.repo.DeletePackage(ctx, c.ID, packageID)
}

// Workbench 乡厨工作台。
// orders 表归属 order 模块，此处为经批准的只读聚合特例（见 repo.CountOrdersBetween）；
// MonthIncome 来自 pay 模块结算单汇总（RoleChef + 本月，已结算净额）。
func (s *serviceImpl) Workbench(ctx context.Context, chefUID int64) (*WorkbenchVO, error) {
	c, err := s.repo.GetByUser(ctx, chefUID)
	if err != nil {
		return nil, err
	}
	today := dayStart(time.Now())
	tomorrow := today.AddDate(0, 0, 1)
	todayOrders, err := s.repo.CountOrdersBetween(ctx, c.ID, today, tomorrow)
	if err != nil {
		return nil, err
	}
	pending, err := s.repo.CountOrdersByStatus(ctx, c.ID, orderStatusPending)
	if err != nil {
		return nil, err
	}
	income, err := s.paySvc.Income(ctx, chefUID, pay.RoleChef)
	if err != nil {
		return nil, err
	}
	return &WorkbenchVO{
		TodayOrders:   todayOrders,
		PendingOrders: pending,
		MonthIncome:   income.MonthNet,
	}, nil
}

// IncOrderCount 订单完成时累加接单数（order 模块回调）
func (s *serviceImpl) IncOrderCount(ctx context.Context, chefID int64) error {
	return s.repo.IncOrderCount(ctx, chefID)
}

// Stats 管理端仪表盘：乡厨总数与待审核数
func (s *serviceImpl) Stats(ctx context.Context) (*ChefStatsVO, error) {
	total, err := s.repo.CountAll(ctx)
	if err != nil {
		return nil, err
	}
	pending, err := s.repo.CountByStatus(ctx, ChefStatusPending)
	if err != nil {
		return nil, err
	}
	return &ChefStatsVO{Total: total, Pending: pending}, nil
}

// RegisterConsumers 注册 chef 模块 MQ 消费者（由 bootstrap 调用）：
// 评价创建事件（review.created → q.chef.review）触发评分重算
func RegisterConsumers(bus *mq.MQ, svc Service) {
	if bus == nil || svc == nil {
		logger.L().Warn("MQ 或 chef 服务未就绪，跳过 chef 消费者注册")
		return
	}
	bus.Consume(mq.QueueChefReview, func(ctx context.Context, body []byte) error {
		var e ReviewEvent
		if err := json.Unmarshal(body, &e); err != nil {
			return err
		}
		avg, err := svc.RecalcRating(ctx, e.ChefID)
		if err != nil {
			return err
		}
		logger.L().Info("乡厨评分已重算",
			zap.Int64("chef_id", e.ChefID), zap.Float64("rating", avg))
		return nil
	})
}

// AdminList 管理端乡厨列表（含全部状态）
func (s *serviceImpl) AdminList(ctx context.Context, keyword string, status *int8, page, size int) ([]*Chefs, int64, error) {
	return s.repo.AdminList(ctx, keyword, status, page, size)
}

// Audit 管理端审核乡厨入驻（approve→1已通过 / 拒绝→2已拒绝）
func (s *serviceImpl) Audit(ctx context.Context, id int64, approve bool, remark string) error {
	status := int8(2)
	if approve {
		status = ChefStatusApproved
	}
	return s.repo.Audit(ctx, id, status, remark)
}
