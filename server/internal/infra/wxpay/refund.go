package wxpay

import (
	"net/http"
)

// Refund 发起退款（V3 原路退回）；Mock 模式直接成功
func (c *Client) Refund(outTradeNo, outRefundNo string, refundFen, totalFen int64) error {
	if c.cfg.Mock || !c.cfg.Enabled {
		return nil
	}
	body := map[string]any{
		"out_trade_no":  outTradeNo,
		"out_refund_no": outRefundNo,
		"amount": map[string]any{
			"refund":   refundFen,
			"total":    totalFen,
			"currency": "CNY",
		},
	}
	return c.do(http.MethodPost, "/v3/refund/domestic/refunds", body, nil)
}
