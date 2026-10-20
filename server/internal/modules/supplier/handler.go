package supplier

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/xiangchu/server/internal/middleware"
	"github.com/xiangchu/server/internal/pkg/errs"
	"gorm.io/gorm"

	"github.com/xiangchu/server/internal/pkg/jwtx"
	"github.com/xiangchu/server/internal/pkg/pagination"
	"github.com/xiangchu/server/internal/pkg/response"
)

type Handler struct {
	svc Service
}

// NewHandler 构造供应商模块 Handler
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Routes 注册供应商模块路由
func (h *Handler) Routes(r gin.IRouter, mk *jwtx.Maker) {
	// 入驻申请：任何登录用户可提交（含被驳回后重新提交），审核通过后才签发 supplier 角色
	usr := r.Group("", middleware.Auth(mk, "user"))
	{
		usr.POST("/supplier/apply", h.apply)
		usr.GET("/supplier/apply-status", h.profile) // 待审核/被拒用户查申请状态（同 /chef/apply-status 复用模式）
	}
	sup := r.Group("", middleware.Auth(mk, "supplier"))
	{
		sup.GET("/supplier/profile", h.profile)
		sup.PUT("/supplier/profile", h.updateProfile)
	}
	admin := r.Group("", middleware.Auth(mk, "admin"))
	{
		admin.GET("/admin/suppliers", h.listForAdmin)
		admin.POST("/admin/suppliers/:id/audit", h.audit)
		admin.GET("/admin/suppliers/stats", h.stats)
	}
}

// profile 我的供应商档案
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	我的档案 / 申请进度
//	@Description	GET /supplier/apply-status（user 角色，查申请审核状态）与 GET /supplier/profile（supplier 角色）共用
//	@Produce	json
//	@Success	200	{object}	response.R{data=supplier.Supplier}
//	@Router		/supplier/profile [get]
//	@Router		/supplier/apply-status [get]
func (h *Handler) profile(c *gin.Context) {
	sup, err := h.svc.GetByUser(c.Request.Context(), middleware.UID(c))
	if err != nil {
		response.Fail(c, notFound(err))
		return
	}
	response.OK(c, sup)
}

// apply 提交入驻申请
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	入驻申请/重新提交
//	@Description	任何登录用户可提交（含被驳回后重新提交），审核通过后签发 supplier 角色
//	@Accept		json
//	@Produce	json
//	@Param		req	body		supplier.ApplyReq	true	"入驻材料"
//	@Success	200	{object}	response.R{data=map[string]bool}
//	@Router		/supplier/apply [post]
func (h *Handler) apply(c *gin.Context) {
	var req ApplyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errs.BadRequest("参数错误："+err.Error()))
		return
	}
	if err := h.svc.Apply(c.Request.Context(), middleware.UID(c), &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"submitted": true})
}

// updateProfile 修改自己的资料
//
//	@Tags		采购供应域
//	@Security	BearerAuth
//	@Summary	维护我的资料
//	@Description	supplier 角色
//	@Accept		json
//	@Produce	json
//	@Param		req	body		supplier.ApplyReq	true	"资料"
//	@Success	200	{object}	response.R{data=map[string]bool}
//	@Router		/supplier/profile [put]
func (h *Handler) updateProfile(c *gin.Context) {
	var req ApplyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errs.BadRequest("参数错误："+err.Error()))
		return
	}
	if err := h.svc.UpdateOwnProfile(c.Request.Context(), middleware.UID(c), &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"updated": true})
}

// listForAdmin 管理端供应商列表（keyword/status/page/size）
//
//	@Tags		管理端
//	@Security	BearerAuth
//	@Summary	供应商列表
//	@Description	admin 角色；含全部审核状态
//	@Produce	json
//	@Param		keyword	query		string	false	"关键词"
//	@Param		status	query		int		false	"0待审核 1已通过 2已拒绝 3已禁用"
//	@Param		page	query		int		false	"页码"	Default(1)
//	@Param		size	query		int		false	"每页条数"	Default(10)	Maximum(100)
//	@Success	200	{object}	response.R{data=response.PageList{list=[]supplier.Supplier}}
//	@Router		/admin/suppliers [get]
func (h *Handler) listForAdmin(c *gin.Context) {
	page, size, err := pageParams(c)
	if err != nil {
		response.Fail(c, err)
		return
	}
	f := &ListFilter{Keyword: c.Query("keyword"), Page: page, Size: size}
	if raw := c.Query("status"); raw != "" {
		st, err := strconv.Atoi(raw)
		if err != nil {
			response.Fail(c, errs.BadRequest("status 参数不合法"))
			return
		}
		s := int8(st)
		f.Status = &s
	}
	list, total, err := h.svc.ListForAdmin(c.Request.Context(), f)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, response.PageData(list, total, page, size))
}

// audit 管理端审核（approve/remark）
//
//	@Tags		管理端
//	@Security	BearerAuth
//	@Summary	供应商入驻审核
//	@Description	admin 角色；驳回需附 remark
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int					true	"供应商ID"
//	@Param		req	body		supplier.AuditReq	true	"审核结果"
//	@Success	200	{object}	response.R{data=map[string]bool}
//	@Router		/admin/suppliers/{id}/audit [post]
func (h *Handler) audit(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Fail(c, errs.ErrNotFound)
		return
	}
	var req AuditReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errs.BadRequest("参数错误："+err.Error()))
		return
	}
	if err := h.svc.Audit(c.Request.Context(), id, req.Approve, req.Remark); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"audited": true})
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

// stats 供应商统计（总数/待审核数，管理端仪表盘）
//
//	@Tags		管理端
//	@Security	BearerAuth
//	@Summary	供应商统计
//	@Description	admin 角色
//	@Produce	json
//	@Success	200	{object}	response.R{data=supplier.SupplierStatsVO}
//	@Router		/admin/suppliers/stats [get]
func (h *Handler) stats(c *gin.Context) {
	res, err := h.svc.Stats(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, res)
}
