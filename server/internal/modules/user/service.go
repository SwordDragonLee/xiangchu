package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xiangchu/server/internal/config"
	"github.com/xiangchu/server/internal/infra/logger"
	"github.com/xiangchu/server/internal/infra/mq"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xiangchu/server/internal/pkg/errs"
	"github.com/xiangchu/server/internal/pkg/jwtx"
	"github.com/xiangchu/server/internal/pkg/password"
)

var phoneRe = regexp.MustCompile(`^1\d{10}$`)

// Service 用户模块服务接口（跨模块只允许通过此接口调用）
type Service interface {
	SendCode(ctx context.Context, phone string) error
	LoginByCode(ctx context.Context, phone, code string) (*LoginResult, error)
	WxLogin(ctx context.Context, code string) (*LoginResult, error)
	WebLogin(ctx context.Context, phone, pwd string) (*LoginResult, error)
	RefreshSession(ctx context.Context, uid int64, wantRole string) (*LoginResult, error)
	Profile(ctx context.Context, uid int64) (*ProfileVO, error)
	UpdateProfile(ctx context.Context, uid int64, req *UpdateProfileReq) error
	SetPassword(ctx context.Context, uid int64, pwd string) error
	GetUser(ctx context.Context, id int64) (*User, error)
	ListForAdmin(ctx context.Context, keyword string, page, size int) ([]*AdminUserVO, int64, error)
	SetStatus(ctx context.Context, id int64, status int8) error
	ListAddress(ctx context.Context, uid int64) ([]*UserAddress, error)
	GetAddress(ctx context.Context, uid, id int64) (*UserAddress, error)
	SaveAddress(ctx context.Context, uid, id int64, req *AddressReq) (int64, error)
	DeleteAddress(ctx context.Context, uid, id int64) error
	ListCoupon(ctx context.Context, onlyActive bool, page, size int) ([]*CouponVO, int64, error)
	ReceiveCoupon(ctx context.Context, uid, couponID int64) error
	MyCoupons(ctx context.Context, uid int64) ([]*UserCouponVO, error)
	UseCoupon(ctx context.Context, uid, ucID, orderID int64) error
	ReleaseCoupon(ucID int64) error
}

// ProfileChecker 由 bootstrap 注入，实现方为 chef / supplier 模块
type ProfileChecker interface {
	HasApprovedProfile(ctx context.Context, uid int64) bool
	// FilterApproved 批量版：整页用户只查一次，避免逐个判角色的 N+1
	FilterApproved(ctx context.Context, uids []int64) (map[int64]bool, error)
}

// ---------- impl ----------

type serviceImpl struct {
	repo        *Repo
	rdb         *redis.Client
	mk          *jwtx.Maker
	cfg         *config.Config
	bus         *mq.MQ
	chefChecker ProfileChecker
	supChecker  ProfileChecker
}

func NewService(repo *Repo, rdb *redis.Client, mk *jwtx.Maker, cfg *config.Config, bus *mq.MQ, chefCk ProfileChecker, supCk ProfileChecker) Service {
	return &serviceImpl{repo: repo, rdb: rdb, mk: mk, cfg: cfg, bus: bus, chefChecker: chefCk, supChecker: supCk}
}

const codeTTL = 5 * time.Minute

func (s *serviceImpl) SendCode(ctx context.Context, phone string) error {
	if !phoneRe.MatchString(phone) {
		return errs.ErrPhoneInvalid
	}
	code := s.cfg.SMS.DevCode
	if !s.cfg.SMS.DevMode {
		// TODO: 接入真实短信通道时改为调用 Notifier
		code = genCode()
	}
	if err := s.rdb.Set(ctx, "sms:code:"+phone, code, codeTTL).Err(); err != nil {
		return err
	}
	logger.L().Info("send code", zap.String("phone", phone))
	return nil
}

func genCode() string {
	return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
}

// resolveRole 登录角色判定：chef 优先于 supplier，默认 user
func (s *serviceImpl) resolveRole(ctx context.Context, uid int64) string {
	role := "user"
	if s.chefChecker != nil && s.chefChecker.HasApprovedProfile(ctx, uid) {
		role = "chef"
	} else if s.supChecker != nil && s.supChecker.HasApprovedProfile(ctx, uid) {
		role = "supplier"
	}
	return role
}

// login 颁发 JWT
func (s *serviceImpl) login(ctx context.Context, u *User) (*LoginResult, error) {
	if u.Status == StatusDisabled {
		return nil, errs.ErrUserDisabled
	}
	role := s.resolveRole(ctx, u.ID)
	token, err := s.mk.Sign(u.ID, role)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: token, Role: role, UserID: u.ID}, nil
}

// RefreshSession 角色变更（如入驻审核通过）后凭有效 token 换发含最新角色的新 JWT，无需重新登录。
// wantRole 非空时（chef/supplier）：双身份账号（既是乡厨又是供应商）按期望角色签发——
// resolveRole 固定 chef 优先，供应商小程序不指定角色会永远换发到 chef。
func (s *serviceImpl) RefreshSession(ctx context.Context, uid int64, wantRole string) (*LoginResult, error) {
	u, err := s.repo.Get(ctx, uid)
	if err != nil {
		return nil, err
	}
	role := s.resolveRole(ctx, u.ID)
	if (wantRole == "chef" || wantRole == "supplier") && wantRole != role && s.hasApproved(ctx, u.ID, wantRole) {
		role = wantRole
	}
	if u.Status == StatusDisabled {
		return nil, errs.ErrUserDisabled
	}
	token, err := s.mk.Sign(u.ID, role)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: token, Role: role, UserID: u.ID}, nil
}

// hasApproved 该 uid 是否已有对应角色的已通过档案
func (s *serviceImpl) hasApproved(ctx context.Context, uid int64, role string) bool {
	switch role {
	case "chef":
		return s.chefChecker != nil && s.chefChecker.HasApprovedProfile(ctx, uid)
	case "supplier":
		return s.supChecker != nil && s.supChecker.HasApprovedProfile(ctx, uid)
	}
	return false
}

// LoginByCode 手机号验证码登录（未注册自动注册）
func (s *serviceImpl) LoginByCode(ctx context.Context, phone, code string) (*LoginResult, error) {
	if !phoneRe.MatchString(phone) {
		return nil, errs.ErrPhoneInvalid
	}
	key := "sms:code:" + phone
	stored, err := s.rdb.Get(ctx, key).Result()
	if err != nil || stored != code {
		return nil, errs.ErrCodeInvalid
	}
	_ = s.rdb.Del(ctx, key).Err()
	u, err := s.repo.GetByPhone(ctx, phone)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		u = &User{Phone: phone, Nickname: "用户" + phone[7:], Status: StatusEnabled}
		if err := s.repo.Create(ctx, u); err != nil {
			return nil, err
		}
	}
	return s.login(ctx, u)
}

// WxLogin 微信授权登录（MP 未配置时降级 mock openid 便于联调）
func (s *serviceImpl) WxLogin(ctx context.Context, code string) (*LoginResult, error) {
	openID := "mock_" + code
	if s.cfg.Wx.MP.Enabled {
		var err error
		openID, err = s.wxOpenID(ctx, code)
		if err != nil {
			return nil, err
		}
	}
	u, err := s.repo.GetByOpenID(ctx, openID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		u = &User{OpenID: openID, Nickname: "微信用户", Status: StatusEnabled}
		if err := s.repo.Create(ctx, u); err != nil {
			return nil, err
		}
		return s.login(ctx, u)
	}
	if err != nil {
		return nil, err
	}
	return s.login(ctx, u)
}

// wxOpenID 调微信 code2session 换取 openid
func (s *serviceImpl) wxOpenID(ctx context.Context, code string) (string, error) {
	var out struct {
		OpenID  string `json:"openid"`
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	api := "https://api.weixin.qq.com/sns/jscode2session"
	params := url.Values{}
	params.Set("appid", s.cfg.Wx.MP.AppID)
	params.Set("secret", s.cfg.Wx.MP.Secret)
	params.Set("js_code", code)
	params.Set("grant_type", "authorization_code")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Errcode != 0 {
		return "", fmt.Errorf("code2session 失败: %d %s", out.Errcode, out.Errmsg)
	}
	return out.OpenID, nil
}

// Profile 用户资料
func (s *serviceImpl) Profile(ctx context.Context, uid int64) (*ProfileVO, error) {
	u, err := s.repo.Get(ctx, uid)
	if err != nil {
		return nil, err
	}
	return &ProfileVO{ID: u.ID, Phone: u.Phone, Nickname: u.Nickname, Avatar: u.Avatar}, nil
}

// UpdateProfile 更新昵称/头像
func (s *serviceImpl) UpdateProfile(ctx context.Context, uid int64, req *UpdateProfileReq) error {
	u, err := s.repo.Get(ctx, uid)
	if err != nil {
		return err
	}
	fields := make([]string, 0, 2)
	if req.Nickname != "" {
		u.Nickname = req.Nickname
		fields = append(fields, "nickname")
	}
	if req.Avatar != "" {
		u.Avatar = req.Avatar
		fields = append(fields, "avatar")
	}
	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(ctx, u, fields...)
}

// SetPassword 设置/修改 Web 登录密码（供应商/乡厨后台登录用），bcrypt 存储
func (s *serviceImpl) SetPassword(ctx context.Context, uid int64, pwd string) error {
	if len(pwd) < 6 || len(pwd) > 32 {
		return errs.BadRequest("密码长度需为 6-32 位")
	}
	hash, err := password.Hash(pwd)
	if err != nil {
		return err
	}
	u, err := s.repo.Get(ctx, uid)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	return s.repo.Update(ctx, u, "password_hash")
}

// GetUser 内部/管理端查询
func (s *serviceImpl) GetUser(ctx context.Context, id int64) (*User, error) {
	return s.repo.Get(ctx, id)
}

// ListForAdmin 管理端用户列表。roles 由档案派生（双身份账号返回多个角色），
// 两个批量查询拿整页角色，避免逐用户判角色的 N+1
func (s *serviceImpl) ListForAdmin(ctx context.Context, keyword string, page, size int) ([]*AdminUserVO, int64, error) {
	users, total, err := s.repo.List(ctx, keyword, page, size)
	if err != nil {
		return nil, 0, err
	}
	uids := make([]int64, 0, len(users))
	for _, u := range users {
		uids = append(uids, u.ID)
	}
	var chefSet, supSet map[int64]bool
	if s.chefChecker != nil {
		if chefSet, err = s.chefChecker.FilterApproved(ctx, uids); err != nil {
			return nil, 0, err
		}
	}
	if s.supChecker != nil {
		if supSet, err = s.supChecker.FilterApproved(ctx, uids); err != nil {
			return nil, 0, err
		}
	}
	list := make([]*AdminUserVO, 0, len(users))
	for _, u := range users {
		// 有档案的角色都返回（双身份 ["chef","supplier"]），均无档案归为 ["user"]
		roles := make([]string, 0, 2)
		if chefSet[u.ID] {
			roles = append(roles, "chef")
		}
		if supSet[u.ID] {
			roles = append(roles, "supplier")
		}
		if len(roles) == 0 {
			roles = append(roles, "user")
		}
		list = append(list, &AdminUserVO{
			ID:        u.ID,
			Phone:     u.Phone,
			Nickname:  u.Nickname,
			Avatar:    u.Avatar,
			Status:    u.Status,
			Roles:     roles,
			CreatedAt: u.CreatedAt,
		})
	}
	return list, total, nil
}

// SetStatus 管理端禁用/启用
func (s *serviceImpl) SetStatus(ctx context.Context, id int64, status int8) error {
	return s.repo.SetStatus(ctx, id, status)
}

// ---------- 地址簿 ----------

func (s *serviceImpl) ListAddress(ctx context.Context, uid int64) ([]*UserAddress, error) {
	return s.repo.ListAddress(ctx, uid)
}

func (s *serviceImpl) GetAddress(ctx context.Context, uid, id int64) (*UserAddress, error) {
	return s.repo.GetAddress(ctx, uid, id)
}

// SaveAddress 新建/更新地址；设为默认时清除其他默认
func (s *serviceImpl) SaveAddress(ctx context.Context, uid, id int64, req *AddressReq) (int64, error) {
	addr := &UserAddress{
		UserID:       uid,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		Region:       req.Region,
		Address:      req.Address,
		IsDefault:    req.IsDefault,
	}
	if req.IsDefault {
		_ = s.repo.ClearDefault(ctx, uid)
	}
	if id > 0 {
		addr.ID = id
		if err := s.repo.UpdateAddress(ctx, addr, "contact_name", "contact_phone", "region", "address", "is_default"); err != nil {
			return 0, err
		}
		return id, nil
	}
	if err := s.repo.CreateAddress(ctx, addr); err != nil {
		return 0, err
	}
	return addr.ID, nil
}

func (s *serviceImpl) DeleteAddress(ctx context.Context, uid, id int64) error {
	return s.repo.DeleteAddress(ctx, uid, id)
}

// ---------- 优惠券 ----------

func (s *serviceImpl) ListCoupon(ctx context.Context, onlyActive bool, page, size int) ([]*CouponVO, int64, error) {
	list, total, err := s.repo.ListCoupon(ctx, onlyActive, page, size)
	if err != nil {
		return nil, 0, err
	}
	vos := make([]*CouponVO, 0, len(list))
	for _, c := range list {
		vos = append(vos, &CouponVO{
			ID:        c.ID,
			Name:      c.Name,
			Type:      c.Type,
			Threshold: c.Threshold,
			Amount:    c.Amount,
			ValidTo:   c.ValidTo,
		})
	}
	return vos, total, nil
}

// ReceiveCoupon 领取优惠券（校验有效期与剩余数量）
func (s *serviceImpl) ReceiveCoupon(ctx context.Context, uid, couponID int64) error {
	c, err := s.repo.GetCoupon(ctx, int(couponID))
	if err != nil {
		return err
	}
	if c.Status != 1 || time.Now().After(c.ValidTo) {
		return errs.BadRequest("优惠券已过期或下架")
	}
	if c.ReceivedCount >= c.TotalCount {
		return errs.BadRequest("优惠券已抢完")
	}
	if err := s.repo.IncrCouponReceived(ctx, c.ID); err != nil {
		return err
	}
	return s.repo.CreateUserCoupon(ctx, &UserCoupon{
		UserID:   uid,
		CouponID: couponID,
		ExpireAt: c.ValidTo,
	})
}

// MyCoupons 我持有的优惠券列表（联表带出券模板信息）
func (s *serviceImpl) MyCoupons(ctx context.Context, uid int64) ([]*UserCouponVO, error) {
	list, err := s.repo.MyCoupons(ctx, uid)
	if err != nil {
		return nil, err
	}
	vos := make([]*UserCouponVO, 0, len(list))
	for _, uc := range list {
		vo := &UserCouponVO{
			ID:       uc.ID,
			CouponID: uc.CouponID,
			Status:   uc.Status,
			ExpireAt: uc.ExpireAt,
		}
		if c, err := s.repo.GetCoupon(ctx, int(uc.CouponID)); err == nil {
			vo.Name = c.Name
			vo.Type = c.Type
			vo.Threshold = c.Threshold
			vo.Amount = c.Amount
		}
		vos = append(vos, vo)
	}
	return vos, nil
}

// UseCoupon 支付成功后标记优惠券已使用
func (s *serviceImpl) UseCoupon(ctx context.Context, uid, ucID, orderID int64) error {
	return s.repo.UseCoupon(ctx, ucID, orderID)
}

// ReleaseCoupon 取消订单时释放优惠券（MQ 消费场景无调用方 ctx，用 Background）
func (s *serviceImpl) ReleaseCoupon(ucID int64) error {
	return s.repo.ReleaseCoupon(context.Background(), ucID)
}

// WebLogin 乡厨/供应商 Web 端登录（手机号+密码）
func (s *serviceImpl) WebLogin(ctx context.Context, phone, pwd string) (*LoginResult, error) {
	if !phoneRe.MatchString(phone) {
		return nil, errs.ErrPhoneInvalid
	}
	u, err := s.repo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, errs.BadRequest("账号或密码错误")
	}
	if u.PasswordHash == "" {
		return nil, errs.BadRequest("账号尚未设置密码，请先在小程序完成入驻")
	}
	if !password.Check(u.PasswordHash, pwd) {
		return nil, errs.BadRequest("账号或密码错误")
	}
	return s.login(ctx, u)
}
