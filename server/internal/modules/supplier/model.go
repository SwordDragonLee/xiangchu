package supplier

import "time"

// 入驻审核状态：0 待审核 1 已通过 2 已拒绝 3 已禁用（同 chefs 四态）
const (
	StatusPending  int8 = 0
	StatusApproved int8 = 1
	StatusRejected int8 = 2
	StatusDisabled int8 = 3
)

// Supplier 对应 suppliers 表（一用户一供应商档案，user_id 唯一定位）
type Supplier struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	UserID      int64     `gorm:"index" json:"user_id"`
	Name        string    `gorm:"size:64" json:"name"`
	License     string    `gorm:"size:255" json:"license"`
	Contact     string    `gorm:"size:32" json:"contact"`
	Phone       string    `gorm:"size:20" json:"phone"`
	Region      string    `gorm:"size:64" json:"region"`
	Address     string    `gorm:"size:255" json:"address"`
	Categories  string    `gorm:"size:128" json:"categories"` // 逗号分隔，如 "蔬菜,肉类"
	Rating      float64   `gorm:"default:0" json:"rating"`
	Status      int8      `gorm:"default:0" json:"status"`
	AuditRemark string    `gorm:"size:255" json:"audit_remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 显式表名
func (Supplier) TableName() string { return "suppliers" }
