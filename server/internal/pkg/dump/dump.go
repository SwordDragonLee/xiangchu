// Package dump 提供类似 JS console.log 的万能调试打印，仅用于开发期排查。
package dump

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/davecgh/go-spew/spew"
)

// Dump 打印任意值（可传多个）：能 JSON 化的走 JSON（指针自动解开、格式可读）；
// 不能的（如 *gorm.DB 等框架对象）退回 spew 深打印（递归解出所有指针的值，输出较大），永不失败。
func Dump(vs ...any) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("==> %s:%d\n", filepath.Base(file), line)
	for _, v := range vs {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Print(spew.Sdump(v))
			continue
		}
		fmt.Println(string(b))
	}
}
