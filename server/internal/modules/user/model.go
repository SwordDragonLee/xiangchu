package user

import "time"

// Status values shared by accounts
const (
	StatusEnabled  = 1
	StatusDisabled = 0
)

// User 对应 users 表
type User struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	Phone        string    `gorm:"size:20;uniqueIndex" json:"phone"`
	Nickname     string    `gorm:"size:64" json:"nickname"`
	Avatar       string    `gorm:"size:255" json:"avatar"`
	OpenID       string    `gorm:"size:64;index" json:"open_id"`
	PasswordHash string    `json:"-"`
	Status       int8      `gorm:"default:1" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserAddress 对应 user_addresses 表
type UserAddress struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	UserID       int64     `gorm:"index" json:"user_id"`
	ContactName  string    `gorm:"size:32" json:"contact_name"`
	ContactPhone string    `gorm:"size:20" json:"contact_phone"`
	Region       string    `gorm:"size:64"  json:"region"`
	Address      string    `gorm:"size:255" json:"address"`
	IsDefault    bool      `json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
}

// Coupon 平台发放的优惠券模板
type Coupon struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:64" json:"name"`
	Type          int8      `json:"type"`
	Threshold     int64     `json:"threshold"`
	Amount        int64     `json:"amount"`
	ValidFrom     time.Time `json:"valid_from"`
	ValidTo       time.Time `json:"valid_to"`
	TotalCount    int       `json:"total_count"`
	ReceivedCount int       `json:"received_count"`
	Status        int8      `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// UserCoupon 用户持有的优惠券实例
type UserCoupon struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	UserID    int64     `gorm:"index" json:"user_id"`
	CouponID  int64     `json:"coupon_id"`
	Status    int8      `json:"status"`
	OrderID   int64     `gorm:"index" json:"order_id"`
	ExpireAt  time.Time `json:"expire_at"`
	CreatedAt time.Time `json:"created_at"`
}
