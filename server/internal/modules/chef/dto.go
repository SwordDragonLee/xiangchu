package chef

// ListFilter 公开列表筛选条件
type ListFilter struct {
	Region    string  // 区域，模糊匹配
	Cuisine   string  // 菜系，FIND_IN_SET 命中 CSV
	Keyword   string  // 关键字，姓名/手机号
	MinRating float64 // 最低评分
	OrderBy   string  // rating_desc | order_count_desc，默认 id DESC
	Page      int
	Size      int
}

// ApplyReq 入驻申请
type ApplyReq struct {
	RealName   string `json:"real_name" binding:"required,max=32"`
	Phone      string `json:"phone" binding:"required,max=20"`
	Region     string `json:"region" binding:"required,max=64"`
	Cuisines   string `json:"cuisines" binding:"required,max=128"`
	Experience string `json:"experience"`
}

// UpdateReq 乡厨维护自己的资料（空字段不更新）
type UpdateReq struct {
	RealName   string `json:"real_name" binding:"max=32"`
	Phone      string `json:"phone" binding:"max=20"`
	Region     string `json:"region" binding:"max=64"`
	Cuisines   string `json:"cuisines" binding:"max=128"`
	Experience string `json:"experience"`
}

// PackageReq 创建/更新套餐；金额单位分
type PackageReq struct {
	ID            int64  `json:"id"` // >0 视为更新（PUT /chef/packages/:id 时由路径参数覆盖）
	Name          string `json:"name" binding:"required,max=64"`
	BanquetType   string `json:"banquet_type" binding:"required,max=16"`
	PricePerTable int64  `json:"price_per_table" binding:"min=0"`
	Dishes        string `json:"dishes"` // JSON 数组字符串 [{name,category}]
	MinTables     int    `json:"min_tables" binding:"min=0"`
	Status        int8   `json:"status"` // 1上架 0下架；创建未传默认上架，更新按传值
	Sort          int    `json:"sort"`
}

// RestDaysReq 标记/取消休息日
type RestDaysReq struct {
	Dates []string `json:"dates" binding:"required,min=1"` // yyyy-MM-dd
	Rest  bool     `json:"rest"`                           // true 标休息日，false 恢复可约
}

// AuditReq 入驻审核
type AuditReq struct {
	Approve bool   `json:"approve"`
	Remark  string `json:"remark"` // 驳回原因
}

// ChefDetailVO 乡厨详情；评价由 order 模块挂载的 GET /chefs/:id/reviews 返回
type ChefDetailVO struct {
	Chef     *Chefs      `json:"chef"`
	Packages []*Package  `json:"packages"`
	Cases    []*ChefCase `json:"cases"`
}

// WorkbenchVO 乡厨工作台
type WorkbenchVO struct {
	TodayOrders   int64 `json:"today_orders"`
	PendingOrders int64 `json:"pending_orders"`
	MonthIncome   int64 `json:"month_income"` // 本月净结算收入（分），来自 pay 模块结算单汇总
}

// ChefStatsVO 管理端仪表盘统计
type ChefStatsVO struct {
	Total   int64 `json:"total"`
	Pending int64 `json:"pending"`
}

// ReviewEvent 评价创建事件体（order 模块发布，key review.created）
type ReviewEvent struct {
	ChefID int64 `json:"chef_id"`
}
