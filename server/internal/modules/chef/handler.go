package chef

import (
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/xiangchu/server/internal/middleware"
	"github.com/xiangchu/server/internal/pkg/errs"
	"github.com/xiangchu/server/internal/pkg/jwtx"
	"github.com/xiangchu/server/internal/pkg/pagination"
	"github.com/xiangchu/server/internal/pkg/response"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Routes 注册路由：
//   - /chefs 公开浏览（列表/详情/档期日历）；GET /chefs/:id/reviews 由 order 模块挂载，此处不注册
//   - /chef  乡厨端（JWT，角色 chef）
func (h *Handler) Routes(r gin.IRouter, mk *jwtx.Maker) {
	pub := r.Group("/chefs")
	{
		pub.GET("", h.List)
		pub.GET("/:id", h.Detail)
		pub.GET("/:id/schedules", h.PublicSchedules)
	}
	// 入驻申请：任何登录用户可提交（含被驳回后重新提交），审核通过后才签发 chef 角色
	// apply-status：申请者（含待审核/被驳回）查询本人申请记录，页面据此展示审核状态
	usr := r.Group("/chef", middleware.Auth(mk, "user"))
	{
		usr.POST("/apply", h.Apply)
		usr.GET("/apply-status", h.MyProfile)
	}
	auth := r.Group("/chef", middleware.Auth(mk, "chef"))
	{
		auth.GET("/workbench", h.Workbench)
		auth.GET("/profile", h.MyProfile)
		auth.PUT("/profile", h.UpdateMyProfile)
		auth.GET("/schedules", h.MySchedules)
		auth.PUT("/schedules/rest-days", h.SetRestDays)
		auth.GET("/packages", h.MyPackages)
		auth.POST("/packages", h.CreatePackage)
		auth.PUT("/packages/:id", h.UpdatePackage)
		auth.DELETE("/packages/:id", h.DeletePackage)
	}
	adm := r.Group("/admin", middleware.Auth(mk, "admin"))
	{
		adm.GET("/chefs", h.AdminList)
		adm.POST("/chefs/:id/audit", h.AuditChef)
		adm.GET("/chefs/stats", h.AdminStats)
	}
}

// svcFail 业务错误出口：gorm 未找到统一映射为 40400
func svcFail(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Fail(c, errs.ErrNotFound)
		return
	}
	response.Fail(c, err)
}

// ---------- 参数解析 ----------

func paramInt64(c *gin.Context, name string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, errs.BadRequest("参数错误")
	}
	return id, nil
}

func queryInt(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// queryDay 解析 yyyy-MM-dd 查询参数，空值取默认
func queryDay(c *gin.Context, key string, def time.Time) (time.Time, error) {
	v := c.Query(key)
	if v == "" {
		return def, nil
	}
	t, err := time.ParseInLocation("2006-01-02", v, time.Local)
	if err != nil {
		return time.Time{}, errs.BadRequest("日期格式应为 yyyy-MM-dd")
	}
	return t, nil
}

// ---------- 公开浏览 /chefs ----------

// List 乡厨公开列表
//
//	@Tags		乡厨域
//	@Summary	乡厨列表（公开）
//	@Produce	json
//	@Param		region		query		string	false	"区域，模糊匹配"
//	@Param		cuisine		query		string	false	"菜系，FIND_IN_SET 命中"
//	@Param		keyword		query		string	false	"关键词（姓名/手机号）"
//	@Param		min_rating	query		number	false	"最低评分"
//	@Param		order_by	query		string	false	"排序"	Enums(rating_desc,order_count_desc)
//	@Param		page		query		int		false	"页码"	Default(1)
//	@Param		size		query		int		false	"每页条数"	Default(10)	Maximum(100)
//	@Success	200	{object}	response.R{data=response.PageList{list=[]chef.Chefs}}
//	@Router		/chefs [get]
func (h *Handler) List(c *gin.Context) {
	page, size, err := pagination.Normalize(
		queryInt(c, "page", pagination.DefaultPage),
		queryInt(c, "size", pagination.DefaultSize))
	if err != nil {
		response.Fail(c, err)
		return
	}
	minRating, _ := strconv.ParseFloat(c.Query("min_rating"), 64)
	list, total, err := h.svc.List(c.Request.Context(), &ListFilter{
		Region:    c.Query("region"),
		Cuisine:   c.Query("cuisine"),
		Keyword:   c.Query("keyword"),
		MinRating: minRating,
		OrderBy:   c.Query("order_by"), // rating_desc | order_count_desc
		Page:      page,
		Size:      size,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, response.PageData(list, total, page, size))
}

// Detail 乡厨详情
//
//	@Tags		乡厨域
//	@Summary	乡厨详情（公开）
//	@Description	含套餐与案例；评价走 GET /chefs/{id}/reviews
//	@Produce	json
//	@Param		id	path		int	true	"乡厨ID"
//	@Success	200	{object}	response.R{data=chef.ChefDetailVO}
//	@Router		/chefs/{id} [get]
func (h *Handler) Detail(c *gin.Context) {
	id, err := paramInt64(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}
	vo, err := h.svc.Detail(c.Request.Context(), id)
	if err != nil {
		svcFail(c, err)
		return
	}
	response.OK(c, vo)
}

// PublicSchedules GET /chefs/:id/schedules?from=&to=
//
//	@Tags		乡厨域
//	@Summary	档期日历（公开）
//	@Description	0可约 1已锁 2休息；from 默认今天，to 默认 +30 天
//	@Produce	json
//	@Param		id	path		int		true	"乡厨ID"
//	@Param		from	query		string	false	"起始日 yyyy-MM-dd"
//	@Param		to		query		string	false	"结束日 yyyy-MM-dd"
//	@Success	200	{object}	response.R{data=[]chef.ChefSchedule}
//	@Router		/chefs/{id}/schedules [get]
func (h *Handler) PublicSchedules(c *gin.Context) {
	id, err := paramInt64(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}
	h.schedules(c, id)
}

// MySchedules GET /chef/schedules?from=&to= （本人档期）
//
//	@Tags		乡厨域
//	@Security	BearerAuth
//	@Summary	我的档期日历
//	@Description	chef 角色；from 默认今天，to 默认 +30 天
//	@Produce	json
//	@Param		from	query		string	false	"起始日 yyyy-MM-dd"
//	@Param		to		query		string	false	"结束日 yyyy-MM-dd"
//	@Success	200	{object}	response.R{data=[]chef.ChefSchedule}
//	@Router		/chef/schedules [get]
func (h *Handler) MySchedules(c *gin.Context) {
	chef, err := h.svc.GetChefByUser(c.Request.Context(), middleware.UID(c))
	if err != nil {
		svcFail(c, err)
		return
	}
	h.schedules(c, chef.ID)
}

// schedules 档期日历共用实现，from 默认今天，跨度默认 30 天
func (h *Handler) schedules(c *gin.Context, chefID int64) {
	from := dayStart(time.Now())
	to := from.AddDate(0, 1, 0)
	from, err := queryDay(c, "from", from)
	if err != nil {
		response.Fail(c, err)
		return
	}
	to, err = queryDay(c, "to", to)
	if err != nil {
		response.Fail(c, err)
		return
	}
	list, err := h.svc.Schedules(c.Request.Context(), chefID, from, to)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, list)
}

// ---------- 乡厨端 /chef ----------

// Apply POST /chef/apply 入驻/重新提交审核
//
//	@Tags		乡厨域
//	@Security	BearerAuth
//	@Summary	入驻申请/重新提交
//	@Description	任何登录用户可提交（含被驳回后重新提交），审核通过后签发 chef 角色
//	@Accept		json
//	@Produce	json
//	@Param		req	body		chef.ApplyReq	true	"入驻材料"
//	@Success	200	{object}	response.R
//	@Router		/chef/apply [post]
func (h *Handler) Apply(c *gin.Context) {
	var req ApplyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errs.BadRequest("参数错误"))
		return
	}
	if err := h.svc.Apply(c.Request.Context(), middleware.UID(c), &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// MyProfile 我的档案
//
//	@Tags		乡厨域
//	@Security	BearerAuth
//	@Summary	我的档案 / 申请进度
//	@Description	GET /chef/apply-status（user 角色，查申请审核状态）与 GET /chef/profile（chef 角色）共用
//	@Produce	json
//	@Success	200	{object}	response.R{data=chef.Chefs}
//	@Router		/chef/profile [get]
//	@Router		/chef/apply-status [get]
func (h *Handler) MyProfile(c *gin.Context) {
	chef, err := h.svc.GetChefByUser(c.Request.Context(), middleware.UID(c))
	if err != nil {
		svcFail(c, err)
		return
	}
	response.OK(c, chef)
}

// UpdateMyProfile PUT /chef/profile 维护资料
//
//	@Tags		乡厨域
//	@Security	BearerAuth
//	@Summary	维护我的资料
//	@Description	chef 角色；空字段不更新
//	@Accept		json
//	@Produce	json
//	@Param		req	body		chef.UpdateReq	true	"资料"
//	@Success	200	{object}	response.R
//	@Router		/chef/profile [put]
func (h *Handler) UpdateMyProfile(c *gin.Context) {
	var req UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errs.BadRequest("参数错误"))
		return
	}
	if err := h.svc.UpdateOwnProfile(c.Request.Context(), middleware.UID(c), &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// Workbench GET /chef/workbench 乡厨工作台
//
//	@Tags		乡厨域
//	@Security	BearerAuth
//	@Summary	工作台
//	@Description	chef 角色；今日订单数/待接单数/本月净结算收入（分）
//	@Produce	json
//	@Success	200	{object}	response.R{data=chef.WorkbenchVO}
//	@Router		/chef/workbench [get]
func (h *Handler) Workbench(c *gin.Context) {
	vo, err := h.svc.Workbench(c.Request.Context(), middleware.UID(c))
	if err != nil {
		svcFail(c, err)
		return
	}
	response.OK(c, vo)
}

// SetRestDays PUT /chef/schedules/rest-days 标记/取消休息日
//
//	@Tags		乡厨域
//	@Security	BearerAuth
//	@Summary	标记/取消休息日
//	@Description	chef 角色；已锁档期不可标休
//	@Accept		json
//	@Produce	json
//	@Param		req	body		chef.RestDaysReq	true	"日期列表 + 是否休息"
//	@Success	200	{object}	response.R
//	@Router		/chef/schedules/rest-days [put]
func (h *Handler) SetRestDays(c *gin.Context) {
	var req RestDaysReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errs.BadRequest("参数错误"))
		return
	}
	dates := make([]time.Time, 0, len(req.Dates))
	for _, s := range req.Dates {
		t, err := time.ParseInLocation("2006-01-02", s, time.Local)
		if err != nil {
			response.Fail(c, errs.BadRequest("日期格式应为 yyyy-MM-dd"))
			return
		}
		dates = append(dates, t)
	}
	if err := h.svc.SetRestDays(c.Request.Context(), middleware.UID(c), dates, req.Rest); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// ---------- 套餐管理 /chef/packages ----------

// MyPackages 我的套餐列表
//
//	@Tags		乡厨域
//	@Security	BearerAuth
//	@Summary	我的套餐列表
//	@Description	chef 角色
//	@Produce	json
//	@Success	200	{object}	response.R{data=[]chef.Package}
//	@Router		/chef/packages [get]
func (h *Handler) MyPackages(c *gin.Context) {
	chef, err := h.svc.GetChefByUser(c.Request.Context(), middleware.UID(c))
	if err != nil {
		svcFail(c, err)
		return
	}
	list, err := h.svc.Packages(c.Request.Context(), chef.ID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, list)
}

// CreatePackage 新建套餐
//
//	@Tags		乡厨域
//	@Security	BearerAuth
//	@Summary	新建套餐
//	@Description	chef 角色；金额单位分
//	@Accept		json
//	@Produce	json
//	@Param		req	body		chef.PackageReq	true	"套餐"
//	@Success	200	{object}	response.R{data=map[string]int64}	"data.id=套餐ID"
//	@Router		/chef/packages [post]
func (h *Handler) CreatePackage(c *gin.Context) {
	var req PackageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errs.BadRequest("参数错误"))
		return
	}
	req.ID = 0
	id, err := h.svc.SavePackage(c.Request.Context(), middleware.UID(c), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

// UpdatePackage PUT /chef/packages/:id
//
//	@Tags		乡厨域
//	@Security	BearerAuth
//	@Summary	更新套餐
//	@Description	chef 角色
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int				true	"套餐ID"
//	@Param		req	body		chef.PackageReq	true	"套餐"
//	@Success	200	{object}	response.R
//	@Router		/chef/packages/{id} [put]
func (h *Handler) UpdatePackage(c *gin.Context) {
	id, err := paramInt64(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req PackageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errs.BadRequest("参数错误"))
		return
	}
	req.ID = id
	if _, err := h.svc.SavePackage(c.Request.Context(), middleware.UID(c), &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// DeletePackage DELETE /chef/packages/:id
//
//	@Tags		乡厨域
//	@Security	BearerAuth
//	@Summary	删除套餐
//	@Description	chef 角色
//	@Produce	json
//	@Param		id	path		int	true	"套餐ID"
//	@Success	200	{object}	response.R
//	@Router		/chef/packages/{id} [delete]
func (h *Handler) DeletePackage(c *gin.Context) {
	id, err := paramInt64(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.svc.DeletePackage(c.Request.Context(), middleware.UID(c), id); err != nil {
		svcFail(c, err)
		return
	}
	response.OK(c, nil)
}

// ---------- 管理端 ----------

// AdminList 乡厨列表（管理端，含全部状态）
//
//	@Tags		管理端
//	@Security	BearerAuth
//	@Summary	乡厨列表
//	@Description	admin 角色；含全部审核状态
//	@Produce	json
//	@Param		keyword	query		string	false	"关键词"
//	@Param		status	query		int		false	"0待审核 1已通过 2已拒绝 3已禁用"
//	@Param		page	query		int		false	"页码"	Default(1)
//	@Param		size	query		int		false	"每页条数"	Default(10)	Maximum(100)
//	@Success	200	{object}	response.R{data=response.PageList{list=[]chef.Chefs}}
//	@Router		/admin/chefs [get]
func (h *Handler) AdminList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	page, size, err := pagination.Normalize(page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var status *int8
	if v := c.Query("status"); v != "" {
		sv, err := strconv.Atoi(v)
		if err != nil {
			response.Fail(c, errs.BadRequest("status 参数不合法"))
			return
		}
		s8 := int8(sv)
		status = &s8
	}
	list, total, err := h.svc.AdminList(c.Request.Context(), c.Query("keyword"), status, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, response.PageData(list, total, page, size))
}

// AuditChef 乡厨入驻审核
//
//	@Tags		管理端
//	@Security	BearerAuth
//	@Summary	乡厨入驻审核
//	@Description	admin 角色；驳回需附 remark
//	@Accept		json
//	@Produce	json
//	@Param		id	path		int				true	"乡厨ID"
//	@Param		req	body		chef.AuditReq	true	"审核结果"
//	@Success	200	{object}	response.R
//	@Router		/admin/chefs/{id}/audit [post]
func (h *Handler) AuditChef(c *gin.Context) {
	var req AuditReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.svc.Audit(c.Request.Context(), id, req.Approve, req.Remark); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// AdminStats 乡厨统计（总数/待审核数）
//
//	@Tags		管理端
//	@Security	BearerAuth
//	@Summary	乡厨统计
//	@Description	admin 角色
//	@Produce	json
//	@Success	200	{object}	response.R{data=chef.ChefStatsVO}
//	@Router		/admin/chefs/stats [get]
func (h *Handler) AdminStats(c *gin.Context) {
	res, err := h.svc.Stats(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, res)
}
