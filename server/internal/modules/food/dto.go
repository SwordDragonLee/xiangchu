package food

import (
	"time"

	"github.com/xiangchu/server/internal/infra/wxpay"
)

// IngredientReq 供应商新增/编辑食材
type IngredientReq struct {
	Name     string   `json:"name" binding:"required,max=64"`
	Category string   `json:"category" binding:"required,max=32"`
	Unit     string   `json:"unit" binding:"required,max=16"`
	Price    int64    `json:"price" binding:"required,min=1"` // 分
	Stock    int      `json:"stock" binding:"min=0"`
	Origin   string   `json:"origin" binding:"max=64"`
	Images   []string `json:"images"`
}

// IngredientFilter 食材列表筛选（公共市场/供应商自有/管理端共用）
type IngredientFilter struct {
	Keyword    string
	Category   string
	SupplierID int64
	Page       int
	Size       int
}

// IngredientVO 食材视图（images 解为字符串数组）
type IngredientVO struct {
	ID         int64     `json:"id"`
	SupplierID int64     `json:"supplier_id"`
	Name       string    `json:"name"`
	Category   string    `json:"category"`
	Unit       string    `json:"unit"`
	Price      int64     `json:"price"`
	Stock      int       `json:"stock"`
	Origin     string    `json:"origin"`
	Images     []string  `json:"images"`
	Status     int8      `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// PurchaseItemReq 采购明细行
type PurchaseItemReq struct {
	IngredientID int64 `json:"ingredient_id" binding:"required,min=1"`
	Quantity     int   `json:"quantity" binding:"required,min=1"`
}

// CreatePurchaseReq 采购下单（order_id>0 表示关联宴席订单采购）
type CreatePurchaseReq struct {
	SupplierID     int64             `json:"supplier_id" binding:"required,min=1"`
	OrderID        int64             `json:"order_id"`
	Items          []PurchaseItemReq `json:"items" binding:"required,min=1,dive"`
	FulfillMethod  int8              `json:"fulfill_method" binding:"required,oneof=1 2"`
	DeliverAddress string            `json:"deliver_address" binding:"max=255"`
	DeliverTime    *time.Time        `json:"deliver_time"`
	Remark         string            `json:"remark" binding:"max=255"`
}

// CreatePurchaseResp 下单返回：含拉起支付所需参数
type CreatePurchaseResp struct {
	PurchaseID  int64            `json:"purchase_id"`
	PurchaseNo  string           `json:"purchase_no"`
	TotalAmount int64            `json:"total_amount"`
	PaymentNo   string           `json:"payment_no"`
	Params      *wxpay.PayParams `json:"params"`
}

// PurchaseItemVO 采购明细视图
type PurchaseItemVO struct {
	ID           int64  `json:"id"`
	IngredientID int64  `json:"ingredient_id"`
	Name         string `json:"name"`
	Unit         string `json:"unit"`
	Price        int64  `json:"price"`
	Quantity     int    `json:"quantity"`
	Subtotal     int64  `json:"subtotal"`
}

// PurchaseOrderVO 采购单视图（含明细）
type PurchaseOrderVO struct {
	ID             int64             `json:"id"`
	PurchaseNo     string            `json:"purchase_no"`
	ChefID         int64             `json:"chef_id"`
	SupplierID     int64             `json:"supplier_id"`
	SupplierName   string            `json:"supplier_name"`
	OrderID        int64             `json:"order_id"`
	TotalAmount    int64             `json:"total_amount"`
	PaidAmount     int64             `json:"paid_amount"`
	FulfillMethod  int8              `json:"fulfill_method"`
	PickupAddress  string            `json:"pickup_address"`
	DeliverAddress string            `json:"deliver_address"`
	DeliverTime    *time.Time        `json:"deliver_time"`
	Status         int8              `json:"status"`
	Remark         string            `json:"remark"`
	CreatedAt      time.Time         `json:"created_at"`
	Items          []*PurchaseItemVO `json:"items"`
}

// PurchaseFilter 采购单列表筛选
type PurchaseFilter struct {
	ChefID        int64
	SupplierID    int64
	Status        *int8
	FulfillMethod int8
	Page          int
	Size          int
}

// PurchaseStatsVO 管理端采购统计
type PurchaseStatsVO struct {
	Total     int64          `json:"total"`
	Pending   int64          `json:"pending"`
	Completed int64          `json:"completed"`
	ByFulfill map[int8]int64 `json:"by_fulfill"` // 1 自取单量 2 送货单量
	GMV       int64          `json:"gmv"`        // 已支付采购总金额，分
}

// RejectReq 供应商拒单原因
type RejectReq struct {
	Reason string `json:"reason" binding:"required,max=255"`
}
