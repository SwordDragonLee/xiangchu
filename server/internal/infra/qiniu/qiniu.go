package qiniu

import (
	"fmt"
	"time"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/xiangchu/server/internal/config"
	"github.com/xiangchu/server/internal/pkg/errs"
)

// Uploader 七牛云 Kodo：后端只签发前端直传凭证（uptoken）
type Uploader struct {
	mac *qbox.Mac
	cfg config.QiniuConfig
}

func New(cfg config.QiniuConfig) (*Uploader, error) {
	if cfg.AK == "" || cfg.SK == "" || cfg.Bucket == "" {
		return &Uploader{cfg: cfg}, nil // 未配置时服务可启动，签发凭证时再报错
	}
	return &Uploader{mac: qbox.NewMac(cfg.AK, cfg.SK), cfg: cfg}, nil
}

// Token 返回前端直传所需凭证；Scope 固定 bucket:key，区域由配置显式指定并下发前端
func (u *Uploader) Token(ext string) (token, key, domain, region string, err error) {
	if u.mac == nil {
		return "", "", "", "", errs.New(50001, "七牛云尚未配置，请联系管理员")
	}
	key = fmt.Sprintf("%s/%d%s", time.Now().Format("200601"), time.Now().UnixNano(), ext)
	putPolicy := storage.PutPolicy{
		Scope:      fmt.Sprintf("%s:%s", u.cfg.Bucket, key),
		Expires:    uint64(time.Now().Add(30 * time.Minute).Unix()),
		InsertOnly: 0,
		FsizeLimit: 10 * 1024 * 1024,
		MimeLimit:  "image/jpeg;image/png;image/webp",
	}
	return putPolicy.UploadToken(u.mac), key, u.cfg.Domain, u.cfg.Region, nil
}
