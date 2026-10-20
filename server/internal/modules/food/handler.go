package food

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xiangchu/server/internal/middleware"
	"github.com/xiangchu/server/internal/pkg/errs"
	"github.com/xiangchu/server/internal/pkg/jwtx"
	"github.com/xiangchu/server/internal/pkg/pagination"
	"github.com/xiangchu/server/internal/pkg/response"
	"gorm.io/gorm"
)

type Handler struct {
	svc Service
}

// NewHandler 构造食材采购模块 Handler
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Routes 注册食材采购模块路由
func (h *Handler) Routes(r gin.IRouter, mk *jwtx.Maker) {
	r.GET("/ingredients", h.listIngredients)

	sup := r.Group("", middleware.Auth(mk, "supplier"))
	{
		sup.GET("/supplier/ingredients", h.myIngredients)
		sup.POST("/supplier/ingredients", h.createIngredient)
		sup.PUT("/supplier/ingredients/:id", h.updateIngredient)
		sup.DELETE("/supplier/ingredients/:id", h.offShelfIngredient)
		sup.GET("/supplier/purchases", h.supplierPurchases)
		sup.GET("/supplier/purchases/:id", h.supplierPurchase)
		sup.POST("/supplier/purchases/:id/accept", h.accept)
		sup.POST("/supplier/purchases/:id/reject", h.reject)
		sup.POST("/supplier/purchases/:id/ready", h.ready)
		sup.POST("/supplier/purchases/:id/complete", h.complete)
	}
	chef := r.Group("", middleware.Auth(mk, "chef"))
	{
		chef.POST("/purchases", h.createPurchase)
		chef.GET("/purchases", h.myPurchases)
		chef.GET("/purchases/:id", h.myPurchase)
		chef.POST("/purchases/:id/cancel", h.cancelPurchase)
	}
	admin := r.Group("", middleware.Auth(mk, "admin"))
	{
		admin.GET("/admin/purchases", h.adminPurchases)
		admin.GET("/admin/purchases/fulfill-stats", h.adminStats)
	}
}

// ---------- 食材 ----------

// listIngredients 公共市场浏览（仅 status=1）
//
//	@Tags		采购供应域
//	@Summary	食材市场列表（公开）
//	@Produce	json
//	@Param		keyword		query		string	false	"关键词（名称模糊）"
//	@Param		category	query		string	false	"分类"
//	@Param		supplier_id	query		int		false	"供应商ID"
//	@Param		page		query		int		false	"页码"	Default(1)
//	@Param		size		query		int		false	"每页条数"	Default(10)	Maximum(100)
//	@Success	200	{object}	response.R{data=response.PageList{list=[]food.IngredientVO}}
//	@Router		/ingredients [get]
func (h *Handler) listIngredients(c *gin.Context) {
	page, size, err := pageParams(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	f := &IngredientFilter{
		Keyword:  c.Query("keyword"),
		Category: c.Query("category"),
		Page:     page, Size: size,
	}
	if raw := c.Query("supplier_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			response.Fail(c, errs.BadRequest("supplier_id 参数不合法"))
			return
		}
		f.SupplierID = id
	}
	list, total, err := h.svc.ListIngredients(c.Request.Context(), f)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, response.PageData(list, total, page, size))
}

// myIngredients 供应商自己的食材（含下架）
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	我的食材列表
//	@Description	supplier 角色；含已下架
//	@Produce	json
//	@Param		keyword		query		string	false	"关键词"
//	@Param		category	query		string	false	"分类"
//	@Param		page		query		int		false	"页码"	Default(1)
//	@Param		size		query		int		false	"每页条数"	Default(10)	Maximum(100)
//	@Success	200	{object}	response.R{data=response.PageList{list=[]food.IngredientVO}}
//	@Router		/supplier/ingredients [get]
func (h *Handler) myIngredients(c *gin.Context) {
	page, size, err := pageParams(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	f := &IngredientFilter{
		Keyword:  c.Query("keyword"),
		Category: c.Query("category"),
		Page:     page, Size: size,
	}
	list, total, err := h.svc.ListOwnIngredients(c.Request.Context(), middleware.UID(c), f)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, response.PageData(list, total, page, size))
}

// createIngredient 新增食材
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	新增食材
//	@Description	supplier 角色；金额单位分
//	@Accept		json
//	@Produce	json
//	@Param		req	body		food.IngredientReq	true	"食材参数"
//	@Success	200	{object}	response.R{data=map[string]int64}	"data.id=食材ID"
//	@Router		/supplier/ingredients [post]
func (h *Handler) createIngredient(c *gin.Context) {
	var req IngredientReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errs.BadRequest("参数错误："+err.Error()))
		return
	}
	id, err := h.svc.SaveIngredient(c.Request.Context(), middleware.UID(c), 0, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

// updateIngredient 编辑食材
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	编辑食材
//	@Description	supplier 角色；覆盖更新，金额单位分
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int					true	"食材ID"
//	@Param		req	body		food.IngredientReq	true	"食材参数"
//	@Success	200	{object}	response.R{data=map[string]int64}	"data.id=食材ID"
//	@Router		/supplier/ingredients/{id} [put]
func (h *Handler) updateIngredient(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, errs.ErrNotFound)
		return
	}
	var req IngredientReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errs.BadRequest("参数错误："+err.Error()))
		return
	}
	id, err = h.svc.SaveIngredient(c.Request.Context(), middleware.UID(c), id, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

// offShelfIngredient 下架食材（删除即下架，不物理删除）
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	下架食材
//	@Description	supplier 角色
//	@Produce	json
//	@Param		id	path		int	true	"食材ID"
//	@Success	200	{object}	response.R{data=map[string]bool}
//	@Router		/supplier/ingredients/{id} [delete]
func (h *Handler) offShelfIngredient(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, errs.ErrNotFound)
		return
	}
	if err := h.svc.OffShelfIngredient(c.Request.Context(), middleware.UID(c), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"off_shelf": true})
}

// ---------- 乡厨采购 ----------

// createPurchase 采购下单（body 含 order_id，0=手动采购）
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	采购下单
//	@Description	chef 角色；fulfill_method 1自取 2送货（送货必填送货地址与期望送达时间）；order_id>0 关联宴席订单；返回支付参数
//	@Accept		json
//	@Produce	json
//	@Param		req	body		food.CreatePurchaseReq	true	"采购参数"
//	@Success	200	{object}	response.R{data=food.CreatePurchaseResp}
//	@Router		/purchases [post]
func (h *Handler) createPurchase(c *gin.Context) {
	var req CreatePurchaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errs.BadRequest("参数错误："+err.Error()))
		return
	}
	resp, err := h.svc.CreatePurchase(c.Request.Context(), middleware.UID(c), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

// myPurchases 我的采购单列表
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	我的采购单列表
//	@Description	chef 角色；状态 0待接单 1备货中 2待自取/待送达 3已完成 4已取消
//	@Produce	json
//	@Param		status	query		int		false	"状态筛选"
//	@Param		page	query		int		false	"页码"	Default(1)
//	@Param		size	query		int		false	"每页条数"	Default(10)	Maximum(100)
//	@Success	200	{object}	response.R{data=response.PageList{list=[]food.PurchaseOrderVO}}
//	@Router		/purchases [get]
func (h *Handler) myPurchases(c *gin.Context) {
	page, size, err := pageParams(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var status *int8
	if raw := c.Query("status"); raw != "" {
		st, err := strconv.Atoi(raw)
		if err != nil {
			response.Fail(c, errs.BadRequest("status 参数不合法"))
			return
		}
		s := int8(st)
		status = &s
	}
	list, total, err := h.svc.MyPurchases(c.Request.Context(), middleware.UID(c), status, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, response.PageData(list, total, page, size))
}

// myPurchase 采购单详情（仅本人）
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	采购单详情
//	@Description	chef 角色；含明细
//	@Produce	json
//	@Param		id	path		int	true	"采购单ID"
//	@Success	200	{object}	response.R{data=food.PurchaseOrderVO}
//	@Router		/purchases/{id} [get]
func (h *Handler) myPurchase(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, errs.ErrNotFound)
		return
	}
	vo, err := h.svc.ChefPurchase(c.Request.Context(), middleware.UID(c), id)
	if err != nil {
		response.Fail(c, notFound(err))
		return
	}
	response.OK(c, vo)
}

// cancelPurchase 取消采购单（仅待接单）
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	取消采购单
//	@Description	chef 角色；仅待接单可取消，供应商已接单锁库需双方沟通后由供应商处理
//	@Produce	json
//	@Param		id	path		int	true	"采购单ID"
//	@Success	200	{object}	response.R{data=map[string]bool}
//	@Router		/purchases/{id}/cancel [post]
func (h *Handler) cancelPurchase(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, errs.ErrNotFound)
		return
	}
	if err := h.svc.ChefCancel(c.Request.Context(), middleware.UID(c), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"cancelled": true})
}

// ---------- 供应商履约 ----------

// supplierPurchases 供应商采购单列表（status 可选）
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	待履约采购单列表
//	@Description	supplier 角色
//	@Produce	json
//	@Param		status	query		int		false	"状态筛选：0待接单 1备货中 2待自取/待送达 3已完成 4已取消"
//	@Param		page	query		int		false	"页码"	Default(1)
//	@Param		size	query		int		false	"每页条数"	Default(10)	Maximum(100)
//	@Success	200	{object}	response.R{data=response.PageList{list=[]food.PurchaseOrderVO}}
//	@Router		/supplier/purchases [get]
func (h *Handler) supplierPurchases(c *gin.Context) {
	page, size, err := pageParams(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var status *int8
	if raw := c.Query("status"); raw != "" {
		st, err := strconv.Atoi(raw)
		if err != nil {
			response.Fail(c, errs.BadRequest("status 参数不合法"))
			return
		}
		s := int8(st)
		status = &s
	}
	list, total, err := h.svc.SupplierPurchases(c.Request.Context(), middleware.UID(c), status, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, response.PageData(list, total, page, size))
}

// supplierPurchase 供应商采购单详情（仅本人）
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	采购单详情
//	@Description	supplier 角色；含明细
//	@Produce	json
//	@Param		id	path		int	true	"采购单ID"
//	@Success	200	{object}	response.R{data=food.PurchaseOrderVO}
//	@Router		/supplier/purchases/{id} [get]
func (h *Handler) supplierPurchase(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, errs.ErrNotFound)
		return
	}
	vo, err := h.svc.SupplierPurchase(c.Request.Context(), middleware.UID(c), id)
	if err != nil {
		response.Fail(c, notFound(err))
		return
	}
	response.OK(c, vo)
}

// accept 接单并锁定库存
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	接单
//	@Description	supplier 角色；待接单(0)→备货中(1)，锁定 locked_stock，库存不足拒单
//	@Produce	json
//	@Param		id	path		int	true	"采购单ID"
//	@Success	200	{object}	response.R{data=map[string]bool}
//	@Router		/supplier/purchases/{id}/accept [post]
func (h *Handler) accept(c *gin.Context) {
	h.supAction(c, func(ctx *actionCtx) error {
		return h.svc.SupplierAccept(ctx.c, ctx.uid, ctx.id)
	})
}

// reject 拒单（reason 必填）
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	拒单
//	@Description	supplier 角色；仅接单前可拒，库存不足时常用
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int				true	"采购单ID"
//	@Param		req	body		food.RejectReq	true	"拒单原因"
//	@Success	200	{object}	response.R{data=map[string]bool}
//	@Router		/supplier/purchases/{id}/reject [post]
func (h *Handler) reject(c *gin.Context) {
	var req RejectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errs.BadRequest("参数错误："+err.Error()))
		return
	}
	h.supAction(c, func(ctx *actionCtx) error {
		return h.svc.SupplierReject(ctx.c, ctx.uid, ctx.id, req.Reason)
	})
}

// ready 备货完成（自取→待自取，送货→待送达）
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	备货完成
//	@Description	supplier 角色；备货中(1)→待自取/待送达(2)
//	@Produce	json
//	@Param		id	path		int	true	"采购单ID"
//	@Success	200	{object}	response.R{data=map[string]bool}
//	@Router		/supplier/purchases/{id}/ready [post]
func (h *Handler) ready(c *gin.Context) {
	h.supAction(c, func(ctx *actionCtx) error {
		return h.svc.SupplierReady(ctx.c, ctx.uid, ctx.id)
	})
}

// complete 完成履约并触发结算（自取=已取走，送货=已送达）
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	完成履约
//	@Description	supplier 角色；→已完成(3)，按佣金生成供应商结算单
//	@Produce	json
//	@Param		id	path		int	true	"采购单ID"
//	@Success	200	{object}	response.R{data=map[string]bool}
//	@Router		/supplier/purchases/{id}/complete [post]
func (h *Handler) complete(c *gin.Context) {
	h.supAction(c, func(ctx *actionCtx) error {
		return h.svc.SupplierComplete(ctx.c, ctx.uid, ctx.id)
	})
}

// actionCtx 供应商履约动作上下文
type actionCtx struct {
	c   *gin.Context
	uid int64
	id  int64
}

// supAction 解析 :id 并执行供应商动作
func (h *Handler) supAction(c *gin.Context, act func(*actionCtx) error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, errs.ErrNotFound)
		return
	}
	if err := act(&actionCtx{c: c, uid: middleware.UID(c), id: id}); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"done": true})
}

// ---------- 管理端 ----------

// adminPurchases 采购单监控列表（status/fulfill_method/page/size）
//
//	@Tags		管理端
//	@Security	BearerAuth
//	@Summary	采购单监控列表
//	@Description	admin 角色
//	@Produce	json
//	@Param		status			query		int		false	"状态筛选：0待接单 1备货中 2待自取/待送达 3已完成 4已取消"
//	@Param		fulfill_method	query		int		false	"履约方式"	Enums(1,2)
//	@Param		page			query		int		false	"页码"	Default(1)
//	@Param		size			query		int		false	"每页条数"	Default(10)	Maximum(100)
//	@Success	200	{object}	response.R{data=response.PageList{list=[]food.PurchaseOrderVO}}
//	@Router		/admin/purchases [get]
func (h *Handler) adminPurchases(c *gin.Context) {
	page, size, err := pageParams(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	f := &PurchaseFilter{Page: page, Size: size}
	if raw := c.Query("status"); raw != "" {
		st, err := strconv.Atoi(raw)
		if err != nil {
			response.Fail(c, errs.BadRequest("status 参数不合法"))
			return
		}
		s := int8(st)
		f.Status = &s
	}
	if raw := c.Query("fulfill_method"); raw != "" {
		fm, err := strconv.Atoi(raw)
		if err != nil || (fm != int(FulfillPickup) && fm != int(FulfillDeliver)) {
			response.Fail(c, errs.BadRequest("fulfill_method 参数不合法"))
			return
		}
		f.FulfillMethod = int8(fm)
	}
	list, total, err := h.svc.AdminPurchases(c.Request.Context(), f)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, response.PageData(list, total, page, size))
}

// adminStats 履约方式统计
//
//	@Tags		管理端
//	@Security	BearerAuth
//	@Summary	采购履约统计
//	@Description	admin 角色
//	@Produce	json
//	@Success	200	{object}	response.R{data=food.PurchaseStatsVO}
//	@Router		/admin/purchases/fulfill-stats [get]
func (h *Handler) adminStats(c *gin.Context) {
	vo, err := h.svc.AdminStats(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, vo)
}

// pageParams 解析并规整分页参数（非法值回落默认页）
func pageParams(c *gin.Context) (int, int, error) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	return pagination.Normalize(page, size)
}

// notFound 记录不存在 → 业务 40400
func notFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.ErrNotFound
	}
	return err
}
