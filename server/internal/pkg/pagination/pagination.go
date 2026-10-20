package pagination

import (
	"github.com/xiangchu/server/internal/pkg/errs"
)

const (
	DefaultPage = 1
	DefaultSize = 10
	MaxSize     = 100
)

// Normalize 校验并规整分页参数
func Normalize(page, size int) (int, int, error) {
	if page <= 0 {
		page = DefaultPage
	}
	if size <= 0 {
		size = DefaultSize
	}
	if size > MaxSize {
		return 0, 0, errs.BadRequest("每页数量最多 100")
	}
	return page, size, nil
}
