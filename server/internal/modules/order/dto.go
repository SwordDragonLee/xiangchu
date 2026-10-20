package order

import (
	"github.com/xiangchu/server/internal/infra/wxpay"
)

// CreateReq 创建宴席订单
type CreateReq struct {
	ChefID       int64  `json:"chef_id" binding:"required"`
	PackageID    int64  `json:"package_id" binding:"required"`
	BanquetType  string `json:"banquet_type" binding:"required,max=16"`
	BanquetTime  string `json:"banquet_time" binding:"required"` // 2006-01-02 15:04:05 / RFC3339 / 2006-01-02
	Region       string `json:"region" binding:"required,max=64"`
	Address      string `json:"address" binding:"required,max=255"`
	ContactName  string `json:"contact_name" binding:"required,max=32"`
	ContactPhone string `json:"contact_phone" binding:"required,len=11"`
	TableCount   int    `json:"table_count" binding:"required,min=1,max=999"`
	PeopleCount  int    `json:"people_count" binding:"required,min=1,max=9999"`
	MenuRemark   string `json:"menu_remark" binding:"max=255"`
	UserCouponID int64  `json:"user_coupon_id"`
	PayType      int8   `json:"pay_type" binding:"required,oneof=1 2"` // 1 定金 2 全款
}

// CreateResp 下单/付尾款返回：本次支付参数
type CreateResp struct {
	OrderID   int64            `json:"order_id"`
	OrderNo   string           `json:"order_no"`
	Deposit   int64            `json:"deposit"` // 本次支付金额（分）
	PaymentNo string           `json:"payment_no"`
	Params    *wxpay.PayParams `json:"params"`
}

// ReviewReq 发表评价
type ReviewReq struct {
	Rating  int      `json:"rating" binding:"required,min=1,max=5"`
	Tags    []string `json:"tags" binding:"max=10"`
	Content string   `json:"content" binding:"max=500"`
	Images  []string `json:"images" binding:"max=9"`
}

// AfterSaleReq 申请售后退款
type AfterSaleReq struct {
	Amount int64    `json:"amount" binding:"required,min=1"`
	Reason string   `json:"reason" binding:"required,max=255"`
	Images []string `json:"images" binding:"max=9"`
}

// RejectReq 乡厨拒单
type RejectReq struct {
	Reason string `json:"reason" binding:"required,max=255"`
}

// AuditReq 管理端售后审核
type AuditReq struct {
	Approve bool   `json:"approve"`
	Remark  string `json:"remark" binding:"max=255"`
}

// ReviewStatusReq 评价状态（1显示 0隐藏）
type ReviewStatusReq struct {
	Status *int8 `json:"status" binding:"required"`
}

// ListQuery 列表通用查询（用户端不含 status 筛选，按需取用）
type ListQuery struct {
	Status  *int8  `form:"status"`
	Keyword string `form:"keyword"`
	Page    int    `form:"page,default=1"`
	Size    int    `form:"size,default=10"`
}

// OrderVO 订单列表/详情视图
type OrderVO struct {
	*Order
	PackageName string `json:"package_name"`
}

// OrderDetail 订单详情：含套餐名、评价与售后记录
type OrderDetail struct {
	*Order
	PackageName string       `json:"package_name"`
	Review      *Review      `json:"review,omitempty"`
	AfterSales  []*AfterSale `json:"after_sales"`
}

// OrderStatsVO 管理端订单统计
type OrderStatsVO struct {
	GMV            int64         `json:"gmv"` // 已支付总额（payments 口径）
	Total          int64         `json:"total"`
	Pending        int64         `json:"pending"`   // 待接单
	Completed      int64         `json:"completed"` // 已完成
	RepurchaseRate float64       `json:"repurchase_rate"`
	Regions        []RegionCount `json:"regions"`
}

// RegionCount 已完成订单区域分布
type RegionCount struct {
	Region string `json:"region"`
	Count  int64  `json:"count" gorm:"column:cnt"`
}

// MQ 消息体
type PaySuccessMsg struct {
	BizType   int    `json:"biz_type"`
	BizID     int64  `json:"biz_id"`
	PaymentNo string `json:"payment_no"`
}

type BizMsg struct {
	BizType int   `json:"biz_type"`
	BizID   int64 `json:"biz_id"`
}

type OrderTimeoutMsg struct {
	OrderID int64 `json:"order_id"`
}

// notify 事件类型
const (
	NotifyOrderCreated  = "order.created"
	NotifyOrderAccepted = "order.accepted"
	NotifyOrderRejected = "order.rejected"
	NotifyOrderDone     = "order.completed"
)

// banquetTimeLayouts 支持的宴席时间格式
var banquetTimeLayouts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02 15:04",
	"2006-01-02",
}
