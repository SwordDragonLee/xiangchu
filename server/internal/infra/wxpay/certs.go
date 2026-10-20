package wxpay

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"time"
)

// 证书列表最小刷新间隔：防止伪造 Wechatpay-Serial 触发频繁请求微信接口
const certRefreshInterval = 10 * time.Second

// certificatesResp GET /v3/certificates 响应
type certificatesResp struct {
	Data []struct {
		SerialNo           string `json:"serial_no"`
		EncryptCertificate struct {
			Algorithm      string `json:"algorithm"`
			Nonce          string `json:"nonce"`
			AssociatedData string `json:"associated_data"`
			Ciphertext     string `json:"ciphertext"`
		} `json:"encrypt_certificate"`
	} `json:"data"`
}

// platformCert 取 serial 对应的微信支付平台证书（公钥）。
// 缓存未命中时拉取 /v3/certificates 刷新（微信轮换证书后新回调携带新 serial，自动适配）。
func (c *Client) platformCert(serial string) (*x509.Certificate, error) {
	c.certMu.Lock()
	defer c.certMu.Unlock()

	if cert, ok := c.certs[serial]; ok {
		return cert, nil
	}
	if time.Since(c.certsAt) < certRefreshInterval {
		return nil, fmt.Errorf("平台证书 %s 不在缓存且刚刷新过", serial)
	}
	if err := c.refreshCerts(); err != nil {
		return nil, err
	}
	cert, ok := c.certs[serial]
	if !ok {
		return nil, fmt.Errorf("微信平台证书列表中无 %s", serial)
	}
	return cert, nil
}

// refreshCerts 拉取并解密平台证书列表，重建缓存。调用方需持有 certMu。
func (c *Client) refreshCerts() error {
	var out certificatesResp
	if err := c.do("GET", "/v3/certificates", nil, &out); err != nil {
		return fmt.Errorf("拉取平台证书失败: %w", err)
	}
	if len(out.Data) == 0 {
		return errors.New("平台证书列表为空")
	}
	certs := make(map[string]*x509.Certificate, len(out.Data))
	for _, item := range out.Data {
		pemBytes, err := aesGCMDecrypt(
			[]byte(c.cfg.APIv3Key),
			item.EncryptCertificate.Nonce,
			item.EncryptCertificate.AssociatedData,
			item.EncryptCertificate.Ciphertext,
		)
		if err != nil {
			return fmt.Errorf("解密平台证书 %s 失败: %w", item.SerialNo, err)
		}
		block, _ := pem.Decode(pemBytes)
		if block == nil {
			return fmt.Errorf("平台证书 %s 不是有效 PEM", item.SerialNo)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return fmt.Errorf("解析平台证书 %s 失败: %w", item.SerialNo, err)
		}
		certs[item.SerialNo] = cert
	}
	c.certs = certs
	c.certsAt = time.Now()
	return nil
}

// aesGCMDecrypt V3 通用 AEAD_AES_256_GCM 解密：
// nonce 为明文 ASCII（12 字节），密文 base64 解码后末 16 字节为 GCM tag，附加数据可空
func aesGCMDecrypt(key []byte, nonce, aad, ciphertextB64 string) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("apiv3_key 必须为 32 字节")
	}
	if len(nonce) != 12 {
		return nil, fmt.Errorf("nonce 长度须为 12 字节，实际 %d", len(nonce))
	}
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, fmt.Errorf("密文非 base64: %w", err)
	}
	if len(ciphertext) < 16 {
		return nil, errors.New("密文过短")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plain, err := gcm.Open(nil, []byte(nonce), ciphertext, []byte(aad))
	if err != nil {
		return nil, fmt.Errorf("AES-GCM 解密失败: %w", err)
	}
	return plain, nil
}
