package supplier

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/xiangchu/server/internal/pkg/errs"
)

// Service 供应商模块服务接口（跨模块只允许通过此接口调用）
type Service interface {
	// HasApprovedProfile 登录角色判定：user_id 下存在 status=1 的档案
	HasApprovedProfile(ctx context.Context, uid int64) bool
	// FilterApproved 批量查询：uids 中有已通过审核档案的子集（管理端用户列表拼角色用）
	FilterApproved(ctx context.Context, uids []int64) (map[int64]bool, error)
	GetSupplier(ctx context.Context, id int64) (*Supplier, error)
	GetByUser(ctx context.Context, uid int64) (*Supplier, error)
	Apply(ctx context.Context, uid int64, req *ApplyReq) error
	UpdateOwnProfile(ctx context.Context, uid int64, req *ApplyReq) error
	ListForAdmin(ctx context.Context, f *ListFilter) ([]*Supplier, int64, error)
	Audit(ctx context.Context, id int64, approve bool, remark string) error
	Stats(ctx context.Context) (*SupplierStatsVO, error)
}

type serviceImpl struct {
	repo *Repo
}

// NewService 构造供应商服务
func NewService(repo *Repo) Service { return &serviceImpl{repo: repo} }

// HasApprovedProfile 实现 user.ProfileChecker（bootstrap 结构性注入）
func (s *serviceImpl) HasApprovedProfile(ctx context.Context, uid int64) bool {
	n, err := s.repo.CountApproved(ctx, uid)
	return err == nil && n > 0
}

// FilterApproved 批量版 HasApprovedProfile，整页用户只查一次
func (s *serviceImpl) FilterApproved(ctx context.Context, uids []int64) (map[int64]bool, error) {
	return s.repo.FilterApprovedUsers(ctx, uids)
}

func (s *serviceImpl) GetSupplier(ctx context.Context, id int64) (*Supplier, error) {
	return s.repo.Get(ctx, id)
}

// GetByUser 查自己的档案；无档案返回 gorm.ErrRecordNotFound
func (s *serviceImpl) GetByUser(ctx context.Context, uid int64) (*Supplier, error) {
	return s.repo.GetByUser(ctx, uid)
}

// Apply 提交入驻申请：按 user_id upsert，状态重置为待审核
func (s *serviceImpl) Apply(ctx context.Context, uid int64, req *ApplyReq) error {
	exist, err := s.repo.GetByUser(ctx, uid)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.repo.Create(ctx, &Supplier{
			UserID: uid, Name: req.Name, License: req.License,
			Contact: req.Contact, Phone: req.Phone, Region: req.Region,
			Address: req.Address, Categories: req.Categories, Status: StatusPending,
		})
	}
	if err != nil {
		return err
	}
	if exist.Status == StatusApproved {
		return errs.BadRequest("已通过审核，请勿重复提交")
	}
	if exist.Status == StatusDisabled {
		return errs.ErrStatusDenied
	}
	exist.Name, exist.License = req.Name, req.License
	exist.Contact, exist.Phone = req.Contact, req.Phone
	exist.Region, exist.Address = req.Region, req.Address
	exist.Categories, exist.Status, exist.AuditRemark = req.Categories, StatusPending, ""
	return s.repo.UpdateAll(ctx, exist)
}

// UpdateOwnProfile 已驳回的资料修改后自动重新进入待审核，其余状态不变
func (s *serviceImpl) UpdateOwnProfile(ctx context.Context, uid int64, req *ApplyReq) error {
	sup, err := s.repo.GetByUser(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.ErrNotFound
		}
		return err
	}
	if sup.Status == StatusDisabled {
		return errs.ErrStatusDenied
	}
	sup.Name, sup.License = req.Name, req.License
	sup.Contact, sup.Phone = req.Contact, req.Phone
	sup.Region, sup.Address = req.Region, req.Address
	sup.Categories = req.Categories
	return s.repo.UpdateOwn(ctx, sup, sup.Status == StatusRejected)
}

// ListForAdmin 管理端列表
func (s *serviceImpl) ListForAdmin(ctx context.Context, f *ListFilter) ([]*Supplier, int64, error) {
	return s.repo.List(ctx, f)
}

// Audit 审核通过/拒绝；拒绝必须填备注
func (s *serviceImpl) Audit(ctx context.Context, id int64, approve bool, remark string) error {
	status := StatusRejected
	if approve {
		status = StatusApproved
		if remark == "" {
			remark = "审核通过"
		}
	} else if remark == "" {
		return errs.BadRequest("拒绝时必须填写审核备注")
	}
	return s.repo.UpdateStatus(ctx, id, status, remark)
}

// Stats 管理端统计：档案总数与待审核数
func (s *serviceImpl) Stats(ctx context.Context) (*SupplierStatsVO, error) {
	total, pending, err := s.repo.CountGroup(ctx)
	if err != nil {
		return nil, err
	}
	return &SupplierStatsVO{Total: total, Pending: pending}, nil
}
