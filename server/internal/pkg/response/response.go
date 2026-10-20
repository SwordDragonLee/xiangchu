package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiangchu/server/internal/pkg/errs"
)

// R 统一响应结构
type R struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, R{Code: 0, Msg: "ok", Data: data})
}

// Fail 业务错误统一出口；未识别错误按 50000 处理并打日志
func Fail(c *gin.Context, err error) {
	var e *errs.Err
	if errors.As(err, &e) {
		c.JSON(http.StatusOK, R{Code: e.Code, Msg: e.Msg})
		return
	}
	c.JSON(http.StatusOK, R{Code: errs.ErrServer.Code, Msg: errs.ErrServer.Msg})
}

// PageList 分页响应结构
type PageList struct {
	List  any   `json:"list"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

// PageData 分页响应
func PageData(list any, total int64, page, size int) *PageList {
	return &PageList{List: list, Total: total, Page: page, Size: size}
}
