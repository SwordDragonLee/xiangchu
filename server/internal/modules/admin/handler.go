package admin

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/xiangchu/server/internal/infra/logger"
	"github.com/xiangchu/server/internal/middleware"
	"github.com/xiangchu/server/internal/pkg/errs"
	"github.com/xiangchu/server/internal/pkg/jwtx"
	"github.com/xiangchu/server/internal/pkg/response"
)

// Handler 管理端模块自有接口：登录 / 我的信息 / 字典 / 角色 / 操作日志
type Handler struct {
	svc  Service
	repo *Repo
}

func NewHandler(svc Service, repo *Repo) *Handler {
	return &Handler{svc: svc, repo: repo}
}

// Routes 注册 admin 模块自有路由；业务模块的 /admin/* 接口由各自 Routes
// 注册（自带 Auth(mk, "admin") 守卫），bootstrap 应在本模块之后挂载
func (h *Handler) Routes(r gin.IRouter, mk *jwtx.Maker) {
	g := r.Group("/admin")
	g.Use(NewOperLogMiddleware(h.repo))

	g.POST("/auth/login", h.login)

	auth := g.Group("")
	auth.Use(middleware.Auth(mk, "admin"))
	auth.GET("/me", h.me)
	auth.GET("/dicts", h.dicts)
	auth.GET("/roles", h.roles)
	auth.GET("/oper-logs", h.operLogs)
}

// login 管理员登录
//
//	@Tags		管理端
//	@Summary	后台登录
//	@Description	仅平台管理员（admins 表），用户名+密码
//	@Accept		json
//	@Produce	json
//	@Param		req	body		admin.LoginReq	true	"账号密码"
//	@Success	200	{object}	response.R{data=admin.LoginResult}
//	@Router		/admin/auth/login [post]
func (h *Handler) login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errs.BadRequest("参数错误"))
		return
	}
	token, role, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, LoginResult{Token: token, Role: role})
}

// me 当前登录管理员信息
//
//	@Tags		管理端
//	@Security	BearerAuth
//	@Summary	我的信息
//	@Description	admin 角色
//	@Produce	json
//	@Success	200	{object}	response.R{data=admin.MeVO}
//	@Router		/admin/me [get]
func (h *Handler) me(c *gin.Context) {
	a, err := h.repo.Get(c.Request.Context(), middleware.UID(c))
	if err != nil {
		response.Fail(c, errs.ErrNotFound)
		return
	}
	response.OK(c, MeVO{ID: a.ID, Username: a.Username, RealName: a.RealName, Role: a.Role})
}

// dicts 平台字典（宴席类型 / 菜系 / 订单与采购单状态）
//
//	@Tags		管理端
//	@Security	BearerAuth
//	@Summary	平台字典
//	@Description	admin 角色
//	@Produce	json
//	@Success	200	{object}	response.R{data=admin.DictsVO}
//	@Router		/admin/dicts [get]
func (h *Handler) dicts(c *gin.Context) {
	vo, err := h.svc.Dicts(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, vo)
}

// roles 角色列表（一期预留）
//
//	@Tags		管理端
//	@Security	BearerAuth
//	@Summary	角色列表
//	@Description	admin 角色；一期权限硬编码，仅展示预留
//	@Produce	json
//	@Success	200	{object}	response.R{data=[]admin.Role}
//	@Router		/admin/roles [get]
func (h *Handler) roles(c *gin.Context) {
	list, err := h.repo.ListRoles(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, list)
}

// operLogs 操作日志分页查询，支持 admin_id / action 过滤
//
//	@Tags		管理端
//	@Security	BearerAuth
//	@Summary	操作日志
//	@Description	admin 角色；仅覆盖 admin 模块自身路由的变更操作
//	@Produce	json
//	@Param		admin_id	query		int		false	"管理员ID"
//	@Param		action		query		string	false	"动作，如 POST /admin/chefs/12/audit"
//	@Param		page		query		int		false	"页码"	Default(1)
//	@Param		size		query		int		false	"每页条数"	Default(10)	Maximum(100)
//	@Success	200	{object}	response.R{data=response.PageList{list=[]admin.OperationLog}}
//	@Router		/admin/oper-logs [get]
func (h *Handler) operLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	f := &LogFilter{Action: c.Query("action"), Page: page, Size: size}
	if v := c.Query("admin_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.AdminID = &id
		}
	}
	list, total, err := h.svc.OperLogs(c.Request.Context(), f)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, response.PageData(list, total, f.Page, f.Size))
}

// ---------- 操作日志中间件 ----------

// NewOperLogMiddleware 管理端操作日志中间件：对变更类请求（POST/PUT/DELETE/PATCH）
// 异步写入 operation_logs。v1 仅应用于 admin 模块自有路由；业务模块的 /admin/*
// 接口已由各自 Auth 守卫，暂不自动记录（见 README v1 说明）
func NewOperLogMiddleware(repo *Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先缓存请求体并回填，保证后续 handler 仍可正常绑定参数
		var body []byte
		if c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
		}
		if len(body) > 0 {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		c.Next()

		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		default:
			return
		}
		entry := &OperationLog{
			AdminID:    middleware.UID(c),
			Action:     truncate(c.Request.Method+" "+c.Request.URL.Path, 64),
			TargetType: targetTypeOf(c.Request.URL.Path),
			TargetID:   targetIDOf(c),
			Detail:     truncate(string(body), 255),
			IP:         c.ClientIP(),
		}
		go func(e *OperationLog) {
			defer func() { _ = recover() }()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := repo.CreateLog(ctx, e); err != nil {
				logger.L().Error("写入操作日志失败", zap.Error(err), zap.Int64("admin_id", e.AdminID))
			}
		}(entry)
	}
}

// targetTypeOf 取 /admin/ 后的第一段作为目标类型，如 /admin/chefs/12 -> chefs
func targetTypeOf(path string) string {
	seg := strings.Split(strings.TrimPrefix(strings.Trim(path, "/"), "admin/"), "/")
	if len(seg) == 0 {
		return ""
	}
	return truncate(seg[0], 32)
}

// targetIDOf 取路由参数 :id（无则 0）
func targetIDOf(c *gin.Context) int64 {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	return id
}

// truncate 按字节截断，并保证不破坏 UTF-8 多字节字符
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	cut := s[:max]
	for len(cut) > 0 && !utf8.ValidString(cut) {
		cut = cut[:len(cut)-1]
	}
	return cut
}
