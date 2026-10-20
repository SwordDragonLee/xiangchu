package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xiangchu/server/internal/config"
	"github.com/xiangchu/server/internal/infra/logger"
)

// New 建立 Redis 连接（go-redis v9）
func New(cfg config.RedisConfig) (*redis.Client, error) {
	cli := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	logger.L().Info("Redis 连接成功")
	return cli, nil
}
