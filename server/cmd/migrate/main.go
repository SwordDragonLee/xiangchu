// 数据库迁移命令（golang-migrate）：迁移文件内嵌于 migrations 包，随代码版本化
//
//	go run ./cmd/migrate -c configs/config.yaml up          # 迁移到最新版本
//	go run ./cmd/migrate -c configs/config.yaml down [N]    # 回滚 N 步（默认 1）
//	go run ./cmd/migrate -c configs/config.yaml status      # 查看当前版本
//	go run ./cmd/migrate -c configs/config.yaml baseline    # 存量库（AutoMigrate 时代建的表）标记基线 v1
//	go run ./cmd/migrate -c configs/config.yaml force <N>   # 恢复：清除 dirty 标记并强制置版本 N
//	go run ./cmd/migrate -c configs/config.yaml create <名>  # 生成下一对 up/down 迁移文件
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/xiangchu/server/internal/config"
	"github.com/xiangchu/server/migrations"
)

func main() {
	cfgPath := "configs/config.yaml"
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "-c" {
		if len(args) < 2 {
			fatal("-c 需要配置文件路径")
		}
		cfgPath = args[1]
		args = args[2:]
	}
	if len(args) == 0 {
		fatal("缺少子命令：up / down / status / baseline / create（-c 指定配置，默认 configs/config.yaml）")
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fatal("加载配置失败: %v", err)
	}
	m, err := newMigrator(cfg.MySQL.DSN)
	if err != nil {
		fatal("%v", err)
	}
	defer m.Close()

	switch args[0] {
	case "up":
		doUp(m)
	case "down":
		n := 1
		if len(args) > 1 {
			if v, err := strconv.Atoi(args[1]); err != nil || v < 1 {
				fatal("down 步数非法: %s", args[1])
			} else {
				n = v
			}
		}
		doDown(m, n)
	case "status":
		doStatus(m)
	case "baseline":
		doBaseline(cfg.MySQL.DSN, m)
	case "force":
		// 恢复手段：迁移中断留下 dirty 标记后，人工核对表结构再强制置版本
		if len(args) < 2 {
			fatal("force 需要版本号，如: force 1")
		}
		v, err := strconv.Atoi(args[1])
		if err != nil || v < 0 {
			fatal("版本号非法: %s", args[1])
		}
		if err := m.Force(v); err != nil {
			fatal("强制置版本失败: %v", err)
		}
		fmt.Printf("已强制置版本 %06d（dirty 已清除，请核对表结构后继续）\n", v)
	case "create":
		if len(args) < 2 {
			fatal("create 需要迁移名称，如: create add_orders_channel")
		}
		doCreate(args[1])
	default:
		fatal("未知子命令 %q", args[0])
	}
}

// newMigrator 构造迁移器：DSN 强制 multiStatements（基线文件含多语句），迁移源为内嵌 FS
func newMigrator(dsn string) (*migrate.Migrate, error) {
	c, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("DSN 解析失败: %w", err)
	}
	c.MultiStatements = true
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, err
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, "mysql://"+c.FormatDSN())
	if err != nil {
		return nil, err
	}
	return m, nil
}

// doUp 迁移到最新版本
func doUp(m *migrate.Migrate) {
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("已是最新版本，无变更")
			return
		}
		fatal("迁移失败: %v", err)
	}
	v, dirty, _ := m.Version()
	fmt.Printf("迁移完成，当前版本 %06d（dirty=%v）\n", v, dirty)
}

// doDown 回滚 N 步（到底后自动停止）
func doDown(m *migrate.Migrate, n int) {
	for i := 0; i < n; i++ {
		if err := m.Steps(-1); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Printf("已回滚 %d 步，到底停止\n", i)
				return
			}
			fatal("回滚失败: %v", err)
		}
	}
	fmt.Printf("已回滚 %d 步\n", n)
}

// doStatus 打印当前版本
func doStatus(m *migrate.Migrate) {
	v, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("未迁移（无版本记录）")
		return
	}
	if err != nil {
		fatal("读取版本失败: %v", err)
	}
	fmt.Printf("当前版本 %06d（dirty=%v）\n", v, dirty)
}

// doBaseline 存量库标记基线：仅当业务表已存在（AutoMigrate 时代建的库）时，
// 将版本号直接置为 1 而不执行建表 SQL；空库误用会报错提示直接 up
func doBaseline(dsn string, m *migrate.Migrate) {
	c, err := mysql.ParseDSN(dsn)
	if err != nil {
		fatal("DSN 解析失败: %v", err)
	}
	db, err := sql.Open("mysql", c.FormatDSN())
	if err != nil {
		fatal("%v", err)
	}
	defer db.Close()
	var cnt int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'users'",
	).Scan(&cnt); err != nil {
		fatal("%v", err)
	}
	if cnt == 0 {
		fatal("未发现存量表（users 不存在），空库请直接执行 up")
	}
	if err := m.Force(1); err != nil {
		fatal("标记基线失败: %v", err)
	}
	fmt.Println("已标记基线版本 000001（未执行建表 SQL）")
}

// doCreate 生成下一对迁移文件（server/migrations/ 下，序号自增）
func doCreate(name string) {
	entries, err := os.ReadDir("migrations")
	if err != nil {
		fatal("读取 migrations 目录失败（请在 server/ 目录下执行）: %v", err)
	}
	next := 1
	for _, e := range entries {
		var v int
		if _, err := fmt.Sscanf(e.Name(), "%06d", &v); err == nil && v >= next {
			next = v + 1
		}
	}
	tpl := "-- 需求：TODO\n-- 说明：TODO\n\n"
	upPath := fmt.Sprintf("migrations/%06d_%s.up.sql", next, name)
	downPath := fmt.Sprintf("migrations/%06d_%s.down.sql", next, name)
	if err := os.WriteFile(upPath, []byte(tpl), 0o644); err != nil {
		fatal("%v", err)
	}
	if err := os.WriteFile(downPath, []byte(tpl), 0o644); err != nil {
		fatal("%v", err)
	}
	fmt.Printf("已生成 %s 与 %s\n", upPath, downPath)
}

func fatal(format string, a ...any) {
	fmt.Printf("错误: "+format+"\n", a...)
	os.Exit(1)
}
