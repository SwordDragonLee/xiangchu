package supplier

// ApplyReq 供应商入驻/资料更新申请
type ApplyReq struct {
	Name       string `json:"name" binding:"required,max=64"`
	License    string `json:"license" binding:"required,max=255"` // 营业执照/资质图片 URL
	Contact    string `json:"contact" binding:"required,max=32"`
	Phone      string `json:"phone" binding:"required,len=11"`
	Region     string `json:"region" binding:"max=64"`
	Address    string `json:"address" binding:"required,max=255"`
	Categories string `json:"categories" binding:"max=128"` // 经营品类，逗号分隔
}

// AuditReq 管理端审核入参
type AuditReq struct {
	Approve bool   `json:"approve"`
	Remark  string `json:"remark" binding:"max=255"`
}

// ListFilter 管理端列表筛选
type ListFilter struct {
	Keyword string // 名称/联系人/手机号模糊
	Status  *int8
	Page    int
	Size    int
}

// SupplierStatsVO 管理端统计
type SupplierStatsVO struct {
	Total   int64 `json:"total"`
	Pending int64 `json:"pending"`
}
