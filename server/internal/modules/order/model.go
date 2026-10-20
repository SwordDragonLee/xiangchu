package order

import "time"

// 订单状态
const (
	StatusUnpaid    int8 = 0 // 待支付
	StatusPending   int8 = 1 // 待接单
	StatusAccepted  int8 = 2 // 已接单
	StatusPrepping  int8 = 3 // 备菜中
	StatusServing   int8 = 4 // 服务中
	StatusDone      int8 = 5 // 已完成
	StatusCancelled int8 = 6 // 已取消
	StatusRejected  int8 = 7 // 已拒绝
)

// 退款状态
const (
	RefundNone     int8 = 0 // 无
	RefundApplying int8 = 1 // 申请中
	RefundApproved int8 = 2 // 已同意
	RefundRejected int8 = 3 // 已拒绝
	RefundDone     int8 = 4 // 已退款
)

// 支付方式
const (
	PayTypeDeposit int8 = 1 // 定金
	PayTypeFull    int8 = 2 // 全款
)

// BizTypeOrder payments.biz_type：宴席订单
const BizTypeOrder = 1

// 乡厨档期锁定状态（chef_schedules.status）
const scheduleLocked = 1

// Order 宴席订单，对应 orders 表
type Order struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	OrderNo      string    `gorm:"size:32;uniqueIndex" json:"order_no"`
	UserID       int64     `gorm:"index" json:"user_id"`
	ChefID       int64     `gorm:"index" json:"chef_id"`
	BanquetType  string    `gorm:"size:16" json:"banquet_type"`
	BanquetTime  time.Time `json:"banquet_time"`
	Region       string    `gorm:"size:64" json:"region"`
	Address      string    `gorm:"size:255" json:"address"` // 地址快照
	ContactName  string    `gorm:"size:32" json:"contact_name"`
	ContactPhone string    `gorm:"size:20" json:"contact_phone"`
	TableCount   int       `json:"table_count"`
	PeopleCount  int       `json:"people_count"`
	PackageID    int64     `json:"package_id"`
	MenuRemark   string    `gorm:"size:255" json:"menu_remark"`
	Amount       int64     `json:"amount"`      // 应付总额 = 套餐单价×桌数 - 券
	PaidAmount   int64     `json:"paid_amount"` // 实付累计
	PayType      int8      `json:"pay_type"`    // 1 定金 2 全款
	Deposit      int64     `json:"deposit"`     // 下单时应付定金/全款
	Status       int8      `gorm:"index" json:"status"`
	RefundStatus int8      `json:"refund_status"`
	CancelReason string    `gorm:"size:255" json:"cancel_reason"`
	UserCouponID int64     `json:"user_coupon_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Review 宴席评价，对应 reviews 表
type Review struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	OrderID   int64      `gorm:"uniqueIndex" json:"order_id"`
	UserID    int64      `gorm:"index" json:"user_id"`
	ChefID    int64      `gorm:"index" json:"chef_id"`
	Rating    int        `json:"rating"` // 1-5
	Tags      string     `gorm:"size:255" json:"tags"`
	Content   string     `gorm:"type:text" json:"content"`
	Images    string     `gorm:"type:text" json:"images"` // JSON 数组字符串
	Reply     string     `gorm:"size:255" json:"reply"`
	ReplyAt   *time.Time `json:"reply_at"`
	Status    int8       `gorm:"default:1" json:"status"` // 1 显示 0 隐藏（管理端隐藏用）
	CreatedAt time.Time  `json:"created_at"`
}

// AfterSale 售后单，对应 after_sales 表
type AfterSale struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	AfterNo     string    `gorm:"size:32" json:"after_no"`
	OrderID     int64     `gorm:"index" json:"order_id"`
	UserID      int64     `json:"user_id"`
	ChefID      int64     `json:"chef_id"`
	Amount      int64     `json:"amount"`
	Reason      string    `gorm:"size:255" json:"reason"`
	Images      string    `gorm:"type:text" json:"images"`
	Status      int8      `json:"status"` // 0 待审核 1 已同意待退款 2 已拒绝 3 已退款
	AuditRemark string    `gorm:"size:255" json:"audit_remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 售后状态
const (
	AfterSalePending  int8 = 0 // 待审核
	AfterSaleApproved int8 = 1 // 已同意待退款
	AfterSaleRefused  int8 = 2 // 已拒绝
	AfterSaleRefunded int8 = 3 // 已退款
)
