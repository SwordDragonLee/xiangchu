package admin

import (
	"context"

	"github.com/xiangchu/server/internal/pkg/errs"
	"github.com/xiangchu/server/internal/pkg/jwtx"
	"github.com/xiangchu/server/internal/pkg/pagination"
	"github.com/xiangchu/server/internal/pkg/password"
)

// Service 管理端模块服务接口：仅登录 + 操作日志查询 + 字典；
// 各业务操作（审核、提现等）由各业务模块的 /admin/* 路由承担
type Service interface {
	Login(ctx context.Context, username, password string) (token string, role string, err error)
	OperLogs(ctx context.Context, f *LogFilter) ([]*OperationLog, int64, error)
	Dicts(ctx context.Context) (*DictsVO, error)
}

// ---------- impl ----------

type serviceImpl struct {
	repo *Repo
	mk   *jwtx.Maker
}

func NewService(repo *Repo, mk *jwtx.Maker) Service {
	return &serviceImpl{repo: repo, mk: mk}
}

// Login 管理员账号密码登录，JWT role 取 admins.role（中间件 Auth("admin") 依赖该 claim）
func (s *serviceImpl) Login(ctx context.Context, username, pwd string) (string, string, error) {
	a, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return "", "", errs.BadRequest("用户名或密码错误")
	}
	if a.Status == StatusDisabled {
		return "", "", errs.ErrUserDisabled
	}
	if !password.Check(a.PasswordHash, pwd) {
		return "", "", errs.BadRequest("用户名或密码错误")
	}
	token, err := s.mk.Sign(a.ID, a.Role)
	if err != nil {
		return "", "", err
	}
	return token, a.Role, nil
}

// OperLogs 操作日志分页查询
func (s *serviceImpl) OperLogs(ctx context.Context, f *LogFilter) ([]*OperationLog, int64, error) {
	if f == nil {
		f = &LogFilter{}
	}
	page, size, err := pagination.Normalize(f.Page, f.Size)
	if err != nil {
		return nil, 0, err
	}
	f.Page, f.Size = page, size
	return s.repo.ListLogs(ctx, f)
}

// Dicts 平台字典（代码内常量，返回副本防止调用方修改全局值）
func (s *serviceImpl) Dicts(ctx context.Context) (*DictsVO, error) {
	return &DictsVO{
		BanquetTypes:   append([]string(nil), BanquetTypes...),
		Cuisines:       append([]string(nil), Cuisines...),
		OrderStatus:    copyStatusMap(OrderStatusName),
		PurchaseStatus: copyStatusMap(PurchaseStatusName),
	}, nil
}

func copyStatusMap(m map[int8]string) map[int8]string {
	out := make(map[int8]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
