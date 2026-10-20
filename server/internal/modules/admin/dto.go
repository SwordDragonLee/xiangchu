package admin

// LoginReq 管理员账号密码登录
type LoginReq struct {
	Username string `json:"username" binding:"required,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
}

// LoginResult 登录返回
type LoginResult struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

// MeVO 当前登录管理员信息
type MeVO struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
	Role     string `json:"role"`
}

// LogFilter 操作日志查询条件
type LogFilter struct {
	AdminID *int64
	Action  string
	Page    int
	Size    int
}

// DictsVO 平台字典（宴席类型 / 菜系 / 订单与采购单状态）
type DictsVO struct {
	BanquetTypes   []string        `json:"banquet_types"`
	Cuisines       []string        `json:"cuisines"`
	OrderStatus    map[int8]string `json:"order_status"`
	PurchaseStatus map[int8]string `json:"purchase_status"`
}

// BanquetTypes 宴席类型字典
var BanquetTypes = []string{"婚宴", "寿宴", "白事", "满月", "节庆"}

// Cuisines 菜系字典
var Cuisines = []string{"川菜", "湘菜", "粤菜", "鲁菜", "苏菜", "浙菜", "闽菜", "徽菜", "家常菜", "其他"}

// OrderStatusName 订单状态：0待支付 1待接单 2已接单 3备菜中 4服务中 5已完成
var OrderStatusName = map[int8]string{
	0: "待支付",
	1: "待接单",
	2: "已接单",
	3: "备菜中",
	4: "服务中",
	5: "已完成",
}

// PurchaseStatusName 采购单状态：0待接单 1备货中 2待自取/待送达 3已完成 4已取消
var PurchaseStatusName = map[int8]string{
	0: "待接单",
	1: "备货中",
	2: "待自取/待送达",
	3: "已完成",
	4: "已取消",
}
