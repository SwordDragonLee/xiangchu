package orderid

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// GenNo 生成业务单号：前缀 + 时间(17位) + 4位随机数，如 XB26010112345678901
func GenNo(prefix string) string {
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("%s%s%04d", prefix, time.Now().Format("060102150405000"), n.Int64())
}
