package admin

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

// ---------- admin ----------

func (r *Repo) GetByUsername(ctx context.Context, username string) (*Admin, error) {
	var a Admin
	err := r.q(ctx).Where("username = ?", username).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repo) Get(ctx context.Context, id int64) (*Admin, error) {
	var a Admin
	err := r.q(ctx).First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// ---------- operation log ----------

func (r *Repo) ListLogs(ctx context.Context, f *LogFilter) ([]*OperationLog, int64, error) {
	q := r.q(ctx).Model(&OperationLog{})
	if f.AdminID != nil {
		q = q.Where("admin_id = ?", *f.AdminID)
	}
	if f.Action != "" {
		q = q.Where("action LIKE ?", "%"+f.Action+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*OperationLog
	err := q.Order("id DESC").Offset((f.Page - 1) * f.Size).Limit(f.Size).Find(&list).Error
	return list, total, err
}

func (r *Repo) CreateLog(ctx context.Context, l *OperationLog) error {
	return r.q(ctx).Create(l).Error
}

// ---------- role ----------

func (r *Repo) ListRoles(ctx context.Context) ([]*Role, error) {
	var list []*Role
	err := r.q(ctx).Order("id ASC").Find(&list).Error
	return list, err
}
