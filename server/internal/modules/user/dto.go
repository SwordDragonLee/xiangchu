package user

import "time"

// AdminUserVO 管理端用户列表项（roles 为按档案派生的全部角色，均无档案则 ["user"]）
type AdminUserVO struct {
	ID        int64     `json:"id"`
	Phone     string    `json:"phone"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Status    int8      `json:"status"`
	Roles     []string  `json:"roles"` // chef/supplier，均无则 ["user"]
	CreatedAt time.Time `json:"created_at"`
}

// SmsCodeReq sends a login sms code.
type SmsCodeReq struct {
	Phone string `json:"phone" binding:"required,len=11"`
}

// LoginByCodeReq logs in with phone + code.
type LoginByCodeReq struct {
	Phone string `json:"phone" binding:"required,len=11"`
	Code  string `json:"code" binding:"required,len=6"`
}

// WxLoginReq logs in via wechat code.
type WxLoginReq struct {
	Code string `json:"code" binding:"required"`
}

// WebLoginReq logs into web admin via phone + password.
type WebLoginReq struct {
	Phone    string `json:"phone" binding:"required,len=11"`
	Password string `json:"password" binding:"required,min=6,max=32"`
}

// LoginResult returned after login.
type LoginResult struct {
	Token  string `json:"token"`
	Role   string `json:"role"`
	UserID int64  `json:"user_id"`
	IsNew  bool   `json:"is_new"`
}

// RefreshReq 换发 token 的可选 body：双身份账号按期望角色签发
type RefreshReq struct {
	Role string `json:"role"`
}

// StatusReq 禁用/启用状态
type StatusReq struct {
	Status *int8 `json:"status" binding:"required"`
}

// UpdateProfileReq updates nickname/avatar.
type UpdateProfileReq struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// SetPasswordReq 设置/修改 Web 登录密码
type SetPasswordReq struct {
	Password string `json:"password" binding:"required,min=6,max=32"`
}

// ProfileVO is the user profile view.
type ProfileVO struct {
	ID       int64    `json:"id"`
	Phone    string   `json:"phone"`
	Nickname string   `json:"nickname"`
	Avatar   string   `json:"avatar"`
	Roles    []string `json:"roles"`
}

// AddressReq creates or updates an address.
type AddressReq struct {
	ContactName  string `json:"contact_name" binding:"required"`
	ContactPhone string `json:"contact_phone" binding:"required,len=11"`
	Region       string `json:"region"`
	Address      string `json:"address" binding:"required"`
	IsDefault    bool   `json:"is_default"`
}

// CouponVO is the platform coupon view.
type CouponVO struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      int8      `json:"type"`
	Threshold int64     `json:"threshold"`
	Amount    int64     `json:"amount"`
	ValidTo   time.Time `json:"valid_to"`
}

// UserCouponVO is a coupon held by a user.
type UserCouponVO struct {
	ID        int64     `json:"id"`
	CouponID  int64     `json:"coupon_id"`
	Name      string    `json:"name"`
	Type      int8      `json:"type"`
	Threshold int64     `json:"threshold"`
	Amount    int64     `json:"amount"`
	Status    int8      `json:"status"`
	ExpireAt  time.Time `json:"expire_at"`
}
