package chef

import "time"

// 乡厨审核状态
const (
	ChefStatusPending  int8 = 0 // 待审核
	ChefStatusApproved int8 = 1 // 已通过
	ChefStatusRejected int8 = 2 // 已拒绝
	ChefStatusDisabled int8 = 3 // 已禁用
)

// 档期状态
const (
	ScheduleFree   int8 = 0 // 可约
	ScheduleLocked int8 = 1 // 已锁定
	ScheduleRest   int8 = 2 // 休息日
)

// 套餐上下架
const (
	PackageOff int8 = 0 // 下架
	PackageOn  int8 = 1 // 上架
)

// Chefs 对应 chefs 表；金额字段无，评分由评价事件异步重算
type Chefs struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	UserID      int64     `gorm:"index" json:"user_id"`
	RealName    string    `gorm:"size:32" json:"real_name"`
	Phone       string    `gorm:"size:20" json:"phone"`
	Region      string    `gorm:"size:64" json:"region"`
	Cuisines    string    `gorm:"size:128" json:"cuisines"` // CSV，如 "川菜,家常菜"
	Experience  string    `gorm:"type:text" json:"experience"`
	Rating      float64   `gorm:"default:0" json:"rating"` // 0-5，MQ 异步重算
	OrderCount  int       `json:"order_count"`
	Status      int8      `gorm:"default:0" json:"status"`
	AuditRemark string    `gorm:"size:255" json:"audit_remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ChefSchedule 对应 chef_schedules 表；Date 按天存储（当日 0 点）
type ChefSchedule struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	ChefID    int64     `gorm:"index;uniqueIndex:uk_chef_date" json:"chef_id"`
	Date      time.Time `gorm:"type:datetime;uniqueIndex:uk_chef_date" json:"date"`
	Status    int8      `json:"status"`
	OrderID   int64     `json:"order_id"` // 0=未关联订单
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ChefCase 对应 chef_cases 表（乡厨案例）
type ChefCase struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	ChefID      int64     `gorm:"index" json:"chef_id"`
	Title       string    `gorm:"size:64" json:"title"`
	Cover       string    `gorm:"size:255" json:"cover"`
	Images      string    `gorm:"type:text" json:"images"` // JSON 数组字符串
	Description string    `gorm:"type:text" json:"description"`
	Sort        int       `json:"sort"`
	CreatedAt   time.Time `json:"created_at"`
}

// Package 对应 packages 表（宴席套餐）；金额单位为分
type Package struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	ChefID        int64     `gorm:"index" json:"chef_id"`
	Name          string    `gorm:"size:64" json:"name"`
	BanquetType   string    `gorm:"size:16" json:"banquet_type"` // 婚宴/寿宴/白事/满月/节庆
	PricePerTable int64     `json:"price_per_table"`             // 分/桌
	Dishes        string    `gorm:"type:text" json:"dishes"`     // JSON 数组 [{name,category}]
	MinTables     int       `json:"min_tables"`
	Status        int8      `gorm:"default:1" json:"status"`
	Sort          int       `json:"sort"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
