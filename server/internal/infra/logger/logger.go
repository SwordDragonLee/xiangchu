package logger

import (
	"os"

	"github.com/xiangchu/server/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var l *zap.Logger

// Init 初始化全局 zap 日志（Zap + Lumberjack 滚动）
func Init(cfg config.LogConfig) error {
	if cfg.MaxSizeMB <= 0 {
		cfg.MaxSizeMB = 100
	}
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.File,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   true,
	})
	stdout := zapcore.AddSync(os.Stdout)

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}
	enc := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
	core := zapcore.NewTee(
		zapcore.NewCore(enc, writer, level),
		zapcore.NewCore(enc, stdout, level),
	)
	l = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return nil
}

// L 取全局 logger；未初始化时兜底 Nop，避免空指针
func L() *zap.Logger {
	if l == nil {
		return zap.NewNop()
	}
	return l
}

func Sync() {
	if l != nil {
		_ = l.Sync()
	}
}
