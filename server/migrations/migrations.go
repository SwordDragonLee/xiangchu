// Package migrations 数据库迁移 SQL（golang-migrate 命名，embed 进二进制）。
// 基线 000001 与 docs/schema.sql 保持一致；此后变更新增版本对，禁止改动历史文件
package migrations

import "embed"

// FS 迁移 SQL 资源
//
//go:embed *.sql
var FS embed.FS
