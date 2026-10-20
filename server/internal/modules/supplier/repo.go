package supplier

import (
	"context"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) q(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

// Get 主键查询
func (r *Repo) Get(ctx context.Context, id int64) (*Supplier, error) {
	var s Supplier
	if err := r.q(ctx).First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// GetByUser 按登录用户查档案；无档案返回 gorm.ErrRecordNotFound
func (r *Repo) GetByUser(ctx context.Context, uid int64) (*Supplier, error) {
	var s Supplier
	if err := r.q(ctx).Where("user_id = ?", uid).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// CountApproved 已通过审核的档案数（登录角色判定用）
func (r *Repo) CountApproved(ctx context.Context, uid int64) (int64, error) {
	var n int64
	err := r.q(ctx).Model(&Supplier{}).
		Where("user_id = ? AND status = ?", uid, StatusApproved).
		Count(&n).Error
	return n, err
}

// FilterApprovedUsers 批量查询有审核通过档案的 uid 子集（管理端用户列表拼角色用）
func (r *Repo) FilterApprovedUsers(ctx context.Context, uids []int64) (map[int64]bool, error) {
	out := make(map[int64]bool, len(uids))
	if len(uids) == 0 {
		return out, nil
	}
	var ids []int64
	err := r.q(ctx).Model(&Supplier{}).
		Where("user_id IN ? AND status = ?", uids, StatusApproved).
		Pluck("user_id", &ids).Error
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

// Create 新建入驻申请（status=0 待审核）
func (r *Repo) Create(ctx context.Context, s *Supplier) error {
	return r.q(ctx).Create(s).Error
}

// UpdateAll 全量更新档案字段并重置审核态（重新提交申请）
func (r *Repo) UpdateAll(ctx context.Context, s *Supplier) error {
	return r.q(ctx).Model(s).Select(
		"name", "license", "contact", "phone",
		"region", "address", "categories", "status", "audit_remark",
	).Updates(s).Error
}

// UpdateOwn 更新自己的资料；rejected=true 时同时重置为待审核
func (r *Repo) UpdateOwn(ctx context.Context, s *Supplier, rejected bool) error {
	fields := []string{"name", "license", "contact", "phone",
		"region", "address", "categories"}
	if rejected {
		fields = append(fields, "status")
	}
	return r.q(ctx).Model(s).Select(fields).Updates(s).Error
}

// UpdateStatus 审核/禁用：仅更新状态与审核备注
func (r *Repo) UpdateStatus(ctx context.Context, id int64, status int8, remark string) error {
	return r.q(ctx).Model(&Supplier{}).Where("id = ?", id).
		Updates(map[string]any{"status": status, "audit_remark": remark}).Error
}

// List 管理端分页列表
func (r *Repo) List(ctx context.Context, f *ListFilter) ([]*Supplier, int64, error) {
	q := r.q(ctx).Model(&Supplier{})
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("name LIKE ? OR contact LIKE ? OR phone LIKE ?", kw, kw, kw)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*Supplier
	err := q.Order("id DESC").Offset((f.Page - 1) * f.Size).Limit(f.Size).Find(&list).Error
	return list, total, err
}

// CountGroup 按状态计数（Stats 用）
func (r *Repo) CountGroup(ctx context.Context) (total, pending int64, err error) {
	if err = r.q(ctx).Model(&Supplier{}).Count(&total).Error; err != nil {
		return 0, 0, err
	}
	err = r.q(ctx).Model(&Supplier{}).
		Where("status = ?", StatusPending).Count(&pending).Error
	return total, pending, err
}
