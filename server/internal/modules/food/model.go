package food

import "time"

// 食材状态
const (
	IngStatusOff int8 = 0 // 下架
	IngStatusOn  int8 = 1 // 在售
)

// 采购单状态：0 待接单 1 备货中 2 待自取/待送达 3 已完成 4 已取消
const (
	POStatusPending   int8 = 0
	POStatusPrepping  int8 = 1
	POStatusAwaiting  int8 = 2
	POStatusCompleted int8 = 3
	POStatusCancelled int8 = 4
)

// 履约方式：1 乡厨自取 2 供应商送货
const (
	FulfillPickup  int8 = 1
	FulfillDeliver int8 = 2
)

// Ingredient 对应 ingredients 表（供应商食材商品）
type Ingredient struct {
	ID          int64     `gorm:"primaryKey" json:"id"`
	SupplierID  int64     `gorm:"index" json:"supplier_id"`
	Name        string    `gorm:"size:64" json:"name"`
	Category    string    `gorm:"size:32" json:"category"` // 蔬菜/肉类/水产/调料...
	Unit        string    `gorm:"size:16" json:"unit"`     // 斤/份/箱
	Price       int64     `json:"price"`                   // 单价，分
	Stock       int       `json:"stock"`                   // 可售库存
	LockedStock int       `json:"locked_stock"`            // 接单锁定库存
	Origin      string    `gorm:"size:64" json:"origin"`   // 产地
	Images      string    `gorm:"type:text" json:"images"` // JSON 数组字符串
	Status      int8      `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PurchaseOrder 对应 purchase_orders 表
type PurchaseOrder struct {
	ID             int64      `gorm:"primaryKey" json:"id"`
	PurchaseNo     string     `gorm:"size:32;uniqueIndex" json:"purchase_no"`
	ChefID         int64      `gorm:"index" json:"chef_id"`
	SupplierID     int64      `gorm:"index" json:"supplier_id"`
	OrderID        int64      `json:"order_id"` // 0=手动采购，>0 关联宴席订单
	TotalAmount    int64      `json:"total_amount"`
	PaidAmount     int64      `json:"paid_amount"` // 实付，0=未支付
	FulfillMethod  int8       `json:"fulfill_method"`
	PickupAddress  string     `gorm:"size:255" json:"pickup_address"` // 自取时=供应商地址快照
	DeliverAddress string     `gorm:"size:255" json:"deliver_address"`
	DeliverTime    *time.Time `json:"deliver_time"` // 送货必填
	Status         int8       `gorm:"default:0" json:"status"`
	Remark         string     `gorm:"size:255" json:"remark"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PurchaseItem 对应 purchase_items 表（下单时名称/规格/价格快照）
type PurchaseItem struct {
	ID              int64  `gorm:"primaryKey" json:"id"`
	PurchaseOrderID int64  `gorm:"index" json:"purchase_order_id"`
	IngredientID    int64  `gorm:"index" json:"ingredient_id"`
	Name            string `gorm:"size:64" json:"name"`
	Unit            string `gorm:"size:16" json:"unit"`
	Price           int64  `json:"price"`
	Quantity        int    `json:"quantity"`
	Subtotal        int64  `json:"subtotal"`
}
