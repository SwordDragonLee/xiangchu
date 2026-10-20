package wxpay

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// NotifyResult 回调解密后的支付结果。真实模式解出的明文即 V3 交易通知结构，
// out_trade_no / transaction_id / trade_state 三个字段与 mock 明文 JSON 一致，
// 下游 HandleWxNotify 无需区分模式。
type NotifyResult struct {
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	TradeState    string `json:"trade_state"`
}

// notifyEnvelope V3 回调请求体（密文信封）
type notifyEnvelope struct {
	ID        string `json:"id"`
	EventType string `json:"event_type"`
	Resource  struct {
		Algorithm      string `json:"algorithm"`
		Ciphertext     string `json:"ciphertext"`
		AssociatedData string `json:"associated_data"`
		Nonce          string `json:"nonce"`
	} `json:"resource"`
}

// ParseNotify 解析微信支付回调：
//   - mock/未启用：接受明文 JSON（NotifyResult 结构）
//   - 真实模式：Wechatpay-* 头验签（平台证书）+ resource 密文 AES-256-GCM 解密
func (c *Client) ParseNotify(headers http.Header, body []byte) (*NotifyResult, error) {
	if c.cfg.Mock || !c.cfg.Enabled {
		var plain NotifyResult
		if err := json.Unmarshal(body, &plain); err != nil {
			return nil, err
		}
		plain.TradeState = "SUCCESS"
		return &plain, nil
	}
	return c.parseReal(headers, body)
}

// parseReal 真实模式回调处理：头齐全性 → 时间戳窗口 → 平台证书验签 → 密文解密
func (c *Client) parseReal(headers http.Header, body []byte) (*NotifyResult, error) {
	serial := headers.Get("Wechatpay-Serial")
	tsStr := headers.Get("Wechatpay-Timestamp")
	nonceHdr := headers.Get("Wechatpay-Nonce")
	sigB64 := headers.Get("Wechatpay-Signature")
	if serial == "" || tsStr == "" || nonceHdr == "" || sigB64 == "" {
		return nil, errors.New("缺少 Wechatpay-* 签名头")
	}

	// 时间戳窗口 ±5 分钟，超窗拒绝（防重放）
	sec, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil || time.Since(time.Unix(sec, 0)) > 5*time.Minute {
		return nil, errors.New("回调时间戳缺失或超出 ±5 分钟窗口")
	}

	// 用 Wechatpay-Serial 选平台证书验签（未知 serial 自动刷新证书缓存）
	cert, err := c.platformCert(serial)
	if err != nil {
		return nil, err
	}

	message := tsStr + "\n" + nonceHdr + "\n" + string(body) + "\n"
	if err := verifySignature(cert, message, sigB64); err != nil {
		return nil, err
	}

	// 解密 resource 密文得到明文 JSON
	var env notifyEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("回调体非 JSON: %w", err)
	}
	plain, err := aesGCMDecrypt([]byte(c.cfg.APIv3Key), env.Resource.Nonce, env.Resource.AssociatedData, env.Resource.Ciphertext)
	if err != nil {
		return nil, err
	}
	var res NotifyResult
	if err := json.Unmarshal(plain, &res); err != nil { // 明文即 NotifyResult 同构 JSON
		return nil, fmt.Errorf("解密明文非 NotifyResult: %w", err)
	}
	return &res, nil
}

// verifySignature 平台证书公钥 RSA-SHA256 验签。
// 签名串 = Wechatpay-Timestamp + \n + Wechatpay-Nonce + \n + 请求原文 + \n
func verifySignature(cert *x509.Certificate, message, sigB64 string) error {
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("平台证书公钥非 RSA")
	}
	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		return fmt.Errorf("签名非 base64: %w", err)
	}
	sum := sha256.Sum256([]byte(message))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, sum[:], sig); err != nil {
		return fmt.Errorf("回调验签失败: %w", err)
	}
	return nil
}
