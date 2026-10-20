package admin

import "time"

// Status values shared by accounts
const (
	StatusEnabled  = 1
	StatusDisabled = 0
)

// Admin 对应 admins 表
type Admin struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:32;uniqueIndex" json:"username"`
	PasswordHash string    `gorm:"size:255" json:"-"`
	RealName     string    `gorm:"size:32" json:"real_name"`
	Role         string    `gorm:"size:16;default:admin" json:"role"`
	Status       int8      `gorm:"default:1" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OperationLog 对应 operation_logs 表
type OperationLog struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	AdminID    int64     `gorm:"index" json:"admin_id"`
	Action     string    `gorm:"size:64" json:"action"`
	TargetType string    `gorm:"size:32" json:"target_type"`
	TargetID   int64     `json:"target_id"`
	Detail     string    `gorm:"size:255" json:"detail"`
	IP         string    `gorm:"size:45" json:"ip"`
	CreatedAt  time.Time `json:"created_at"`
}

// Role 对应 roles 表（一期仅模型预留；种子数据由 bootstrap 迁移写入）
type Role struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	Code        string    `gorm:"size:16;uniqueIndex" json:"code"`
	Name        string    `gorm:"size:32" json:"name"`
	Permissions string    `gorm:"type:text" json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
}
