package user

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

// ---------- user ----------

func (r *Repo) q(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *Repo) GetByPhone(ctx context.Context, phone string) (*User, error) {
	var u User
	err := r.q(ctx).Where("phone = ?", phone).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) Get(ctx context.Context, id int64) (*User, error) {
	var u User
	err := r.q(ctx).First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) Create(ctx context.Context, u *User) error {
	return r.q(ctx).Create(u).Error
}

func (r *Repo) Update(ctx context.Context, u *User, fields ...string) error {
	m := r.q(ctx).Model(u)
	if len(fields) > 0 {
		m = m.Select(fields)
	}
	return m.Updates(u).Error
}

func (r *Repo) SetStatus(ctx context.Context, id int64, status int8) error {
	return r.q(ctx).Model(&User{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *Repo) List(ctx context.Context, keyword string, page, size int) ([]*User, int64, error) {
	q := r.q(ctx).Model(&User{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("phone LIKE ? OR nickname LIKE ?", kw, kw)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*User
	err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// ---------- address ----------

func (r *Repo) ListAddress(ctx context.Context, uid int64) ([]*UserAddress, error) {
	var list []*UserAddress
	err := r.q(ctx).Where("user_id = ?", uid).Order("is_default DESC, id DESC").Find(&list).Error
	return list, err
}

func (r *Repo) GetAddress(ctx context.Context, uid, id int64) (*UserAddress, error) {
	var a UserAddress
	err := r.q(ctx).Where("user_id = ?", uid).First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repo) CreateAddress(ctx context.Context, a *UserAddress) error {
	return r.q(ctx).Create(a).Error
}

func (r *Repo) UpdateAddress(ctx context.Context, a *UserAddress, fields ...string) error {
	m := r.q(ctx).Model(a)
	if len(fields) > 4 {
		m = m.Select(fields)
	}
	return m.Updates(a).Error
}

func (r *Repo) DeleteAddress(ctx context.Context, uid, id int64) error {
	return r.q(ctx).Where("user_id = ?", uid).Delete(&UserAddress{}, id).Error
}

// ---------- coupon ----------

func (r *Repo) ListCoupon(ctx context.Context, onlyActive bool, page, size int) ([]*Coupon, int64, error) {
	q := r.q(ctx).Model(&Coupon{})
	if onlyActive {
		now := time.Now()
		q = q.Where("status = 1 AND valid_to > ?", now)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*Coupon
	err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (r *Repo) GetCoupon(ctx context.Context, id int) (*Coupon, error) {
	var c Coupon
	err := r.q(ctx).First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repo) CreateUserCoupon(ctx context.Context, uc *UserCoupon) error {
	return r.q(ctx).Create(uc).Error
}

func (r *Repo) MyCoupons(ctx context.Context, uid int64) ([]*UserCoupon, error) {
	var list []*UserCoupon
	err := r.q(ctx).Where("user_id = ?", uid).Order("id DESC").Find(&list).Error
	return list, err
}

func (r *Repo) GetUserCoupon(ctx context.Context, uid, id int64) (*UserCoupon, error) {
	var uc UserCoupon
	err := r.q(ctx).Where("user_id = ?", uid).First(&uc, id).Error
	if err != nil {
		// not found or not mine
		return nil, err
	}
	return &uc, nil
}

func (r *Repo) UseCoupon(ctx context.Context, id int64, orderID int64) error {
	return r.q(ctx).Model(&UserCoupon{}).Where("id = ?", id).
		Updates(map[string]any{"status": 1, "order_id": orderID}).Error
}

// ---------- repo 补充方法 ----------

func (r *Repo) GetByOpenID(ctx context.Context, openID string) (*User, error) {
	var u User
	err := r.q(ctx).Where("open_id = ?", openID).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) ClearDefault(ctx context.Context, uid int64) error {
	return r.q(ctx).Model(&UserAddress{}).Where("user_id = ?", uid).
		Update("is_default", false).Error
}

func (r *Repo) IncrCouponReceived(ctx context.Context, id int64) error {
	return r.q(ctx).Model(&Coupon{}).Where("id = ?", id).
		UpdateColumn("received_count", gorm.Expr("received_count + 1")).Error
}

func (r *Repo) ReleaseCoupon(ctx context.Context, ucID int64) error {
	return r.q(ctx).Model(&UserCoupon{}).Where("id = ? AND status = 1", ucID).
		Update("status", 0).Error
}
