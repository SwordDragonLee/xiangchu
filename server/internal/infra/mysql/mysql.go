package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/xiangchu/server/internal/config"
	"github.com/xiangchu/server/internal/infra/logger"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Open 建立 MySQL 连接（GORM v2），日志桥接到 zap
func Open(cfg config.MySQLConfig, mode string) (*gorm.DB, error) {
	level := gormlogger.Warn
	if mode == "dev" {
		level = gormlogger.Info
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: gormlogger.New(zapWriter{}, gormlogger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      level,
			Colorful:      false,
		}),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}
	logger.L().Info("MySQL 连接成功")
	return db, nil
}

type zapWriter struct{}

func (zapWriter) Printf(format string, args ...any) {
	logger.L().Info("gorm", zap.String("detail", sprintf(format, args...)))
}

func sprintf(format string, args ...any) string { return fmt.Sprintf(format, args...) }
