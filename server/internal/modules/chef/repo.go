package chef

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repo struct {
	store *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{store: db} }

func (r *Repo) q(ctx context.Context) *gorm.DB { return r.store.WithContext(ctx) }

// ---------- 乡厨档案 ----------

func (r *Repo) GetChef(ctx context.Context, id int64) (*Chefs, error) {
	var c Chefs
	if err := r.q(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repo) GetByUser(ctx context.Context, uid int64) (*Chefs, error) {
	var c Chefs
	if err := r.q(ctx).Where("user_id = ?", uid).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repo) CountApprovedByUser(ctx context.Context, uid int64) (int64, error) {
	var n int64
	err := r.q(ctx).Model(&Chefs{}).
		Where("user_id = ? AND status = ?", uid, ChefStatusApproved).Count(&n).Error
	return n, err
}

// FilterApprovedUsers 批量查询有审核通过档案的 uid 子集（管理端用户列表拼角色用）
func (r *Repo) FilterApprovedUsers(ctx context.Context, uids []int64) (map[int64]bool, error) {
	out := make(map[int64]bool, len(uids))
	if len(uids) == 0 {
		return out, nil
	}
	var ids []int64
	err := r.q(ctx).Model(&Chefs{}).
		Where("user_id IN ? AND status = ?", uids, ChefStatusApproved).
		Pluck("user_id", &ids).Error
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

func (r *Repo) CreateChef(ctx context.Context, c *Chefs) error {
	return r.q(ctx).Create(c).Error
}

func (r *Repo) UpdateChef(ctx context.Context, c *Chefs, fields ...string) error {
	m := r.q(ctx).Model(c)
	if len(fields) > 0 {
		m = m.Select(fields)
	}
	return m.Updates(c).Error
}

// ListApproved 已通过审核的乡厨分页列表（公开浏览）
func (r *Repo) ListApproved(ctx context.Context, f *ListFilter) ([]*Chefs, int64, error) {
	q := r.q(ctx).Model(&Chefs{}).Where("status = ?", ChefStatusApproved)
	if f.Region != "" {
		q = q.Where("region LIKE ?", "%"+f.Region+"%")
	}
	if f.Cuisine != "" {
		q = q.Where("FIND_IN_SET(?, cuisines) > 0", f.Cuisine)
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		q = q.Where("real_name LIKE ? OR phone LIKE ?", kw, kw)
	}
	if f.MinRating > 0 {
		q = q.Where("rating >= ?", f.MinRating)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	switch f.OrderBy {
	case "rating_desc":
		q = q.Order("rating DESC, id DESC")
	case "order_count_desc":
		q = q.Order("order_count DESC, id DESC")
	default:
		q = q.Order("id DESC")
	}
	var list []*Chefs
	err := q.Offset((f.Page - 1) * f.Size).Limit(f.Size).Find(&list).Error
	return list, total, err
}

func (r *Repo) IncOrderCount(ctx context.Context, chefID int64) error {
	return r.q(ctx).Model(&Chefs{}).Where("id = ?", chefID).
		UpdateColumn("order_count", gorm.Expr("order_count + 1")).Error
}

func (r *Repo) UpdateRating(ctx context.Context, chefID int64, rating float64) error {
	return r.q(ctx).Model(&Chefs{}).Where("id = ?", chefID).Update("rating", rating).Error
}

func (r *Repo) CountAll(ctx context.Context) (int64, error) {
	var n int64
	err := r.q(ctx).Model(&Chefs{}).Count(&n).Error
	return n, err
}

func (r *Repo) CountByStatus(ctx context.Context, status int8) (int64, error) {
	var n int64
	err := r.q(ctx).Model(&Chefs{}).Where("status = ?", status).Count(&n).Error
	return n, err
}

// ---------- 档期 ----------

func (r *Repo) GetSchedule(ctx context.Context, chefID int64, date time.Time) (*ChefSchedule, error) {
	var s ChefSchedule
	err := r.q(ctx).Where("chef_id = ? AND date = ?", chefID, date).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpsertSchedule 按 (chef_id, date) 唯一键 upsert
func (r *Repo) UpsertSchedule(ctx context.Context, s *ChefSchedule) error {
	return r.q(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "chef_id"}, {Name: "date"}},
		DoUpdates: clause.Assignments(map[string]any{
			"status":   s.Status,
			"order_id": s.OrderID,
		}),
	}).Create(s).Error
}

func (r *Repo) ListSchedules(ctx context.Context, chefID int64, from, to time.Time) ([]*ChefSchedule, error) {
	var list []*ChefSchedule
	err := r.q(ctx).
		Where("chef_id = ? AND date >= ? AND date < ?", chefID, from, to).
		Order("date ASC").Find(&list).Error
	return list, err
}

func (r *Repo) ListSchedulesByDates(ctx context.Context, chefID int64, dates []time.Time) ([]*ChefSchedule, error) {
	if len(dates) == 0 {
		return nil, nil
	}
	var list []*ChefSchedule
	err := r.q(ctx).Where("chef_id = ? AND date IN ?", chefID, dates).Find(&list).Error
	return list, err
}

// ListLockedByOrder 查某订单锁定的档期（用于释放时回删 Redis 键）
func (r *Repo) ListLockedByOrder(ctx context.Context, chefID, orderID int64) ([]*ChefSchedule, error) {
	var list []*ChefSchedule
	err := r.q(ctx).
		Where("chef_id = ? AND order_id = ? AND status = ?", chefID, orderID, ScheduleLocked).
		Find(&list).Error
	return list, err
}

// ReleaseSchedules 将某订单锁定的档期还原为可约
func (r *Repo) ReleaseSchedules(ctx context.Context, chefID, orderID int64) error {
	return r.q(ctx).Model(&ChefSchedule{}).
		Where("chef_id = ? AND order_id = ? AND status = ?", chefID, orderID, ScheduleLocked).
		Updates(map[string]any{"status": ScheduleFree, "order_id": 0}).Error
}

func (r *Repo) SetScheduleStatus(ctx context.Context, id int64, status int8) error {
	return r.q(ctx).Model(&ChefSchedule{}).Where("id = ?", id).Update("status", status).Error
}

// ---------- 套餐 ----------

func (r *Repo) ListPackages(ctx context.Context, chefID int64, onlyOn bool) ([]*Package, error) {
	q := r.q(ctx).Where("chef_id = ?", chefID)
	if onlyOn {
		q = q.Where("status = ?", PackageOn)
	}
	var list []*Package
	err := q.Order("sort DESC, id ASC").Find(&list).Error
	return list, err
}

func (r *Repo) GetPackage(ctx context.Context, id int64) (*Package, error) {
	var p Package
	if err := r.q(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) CreatePackage(ctx context.Context, p *Package) error {
	return r.q(ctx).Create(p).Error
}

func (r *Repo) UpdatePackage(ctx context.Context, p *Package, fields ...string) error {
	m := r.q(ctx).Model(p)
	if len(fields) > 0 {
		m = m.Select(fields)
	}
	return m.Updates(p).Error
}

// DeletePackage 仅删除本人套餐；未命中视为不存在
func (r *Repo) DeletePackage(ctx context.Context, chefID, id int64) error {
	res := r.q(ctx).Where("chef_id = ? AND id = ?", chefID, id).Delete(&Package{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ---------- 案例 ----------

func (r *Repo) ListCases(ctx context.Context, chefID int64) ([]*ChefCase, error) {
	var list []*ChefCase
	err := r.q(ctx).Where("chef_id = ?", chefID).Order("sort DESC, id ASC").Find(&list).Error
	return list, err
}

// ---------- 跨模块只读聚合（经批准的 raw SQL 特例） ----------
// chef 模块不拥有 reviews / orders 表（归属 order 模块），禁止 import order 包；
// 仅允许以下三条只读聚合 SQL，不得增删改、不得扩展读取范围。

// ReviewAgg 聚合 reviews 表评分：AVG(rating), COUNT(*)
type reviewAgg struct {
	Avg float64 `gorm:"column:avg_rating"`
	Cnt int64   `gorm:"column:cnt"`
}

func (r *Repo) ReviewAgg(ctx context.Context, chefID int64) (avg float64, cnt int64, err error) {
	var out reviewAgg
	err = r.q(ctx).Raw(
		"SELECT IFNULL(AVG(rating), 0) AS avg_rating, COUNT(*) AS cnt FROM reviews WHERE chef_id = ? AND status = 1",
		chefID,
	).Scan(&out).Error
	return out.Avg, out.Cnt, err
}

// CountOrdersBetween 统计宴期落在 [from, to) 的订单数（工作台“今日订单”）
func (r *Repo) CountOrdersBetween(ctx context.Context, chefID int64, from, to time.Time) (int64, error) {
	var n int64
	err := r.q(ctx).Raw(
		"SELECT COUNT(*) FROM orders WHERE chef_id = ? AND banquet_time >= ? AND banquet_time < ?",
		chefID, from, to,
	).Scan(&n).Error
	return n, err
}

// CountOrdersByStatus 按状态统计订单数（工作台“待处理”，orders.status=1）
func (r *Repo) CountOrdersByStatus(ctx context.Context, chefID int64, status int) (int64, error) {
	var n int64
	err := r.q(ctx).Raw(
		"SELECT COUNT(*) FROM orders WHERE chef_id = ? AND status = ?",
		chefID, status,
	).Scan(&n).Error
	return n, err
}

// AdminList 管理端列表：关键词匹配姓名/手机号，可选状态过滤
func (r *Repo) AdminList(ctx context.Context, keyword string, status *int8, page, size int) ([]*Chefs, int64, error) {
	q := r.q(ctx).Model(&Chefs{})
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("real_name LIKE ? OR phone LIKE ?", kw, kw)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*Chefs
	err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// Audit 更新审核状态与备注
func (r *Repo) Audit(ctx context.Context, id int64, status int8, remark string) error {
	return r.q(ctx).Model(&Chefs{}).Where("id = ?", id).
		Updates(map[string]any{"status": status, "audit_remark": remark}).Error
}
