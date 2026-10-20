package notify

import (
	"context"
	"encoding/json"

	"github.com/xiangchu/server/internal/infra/logger"
	"github.com/xiangchu/server/internal/infra/mq"
	"go.uber.org/zap"
)

// Event 通知事件体（走 MQ xc.event 交换机，key 前缀 notify.）
type Event struct {
	Type    string  `json:"type"`
	Title   string  `json:"title"`
	Content string  `json:"content"`
	UserIDs []int64 `json:"user_ids"`
}

// Notifier 通知接口：当前 Log 实现，后续可替换短信/微信模板消息
type Notifier interface {
	Notify(ctx context.Context, e *Event)
}

type LogNotifier struct{}

func NewLogNotifier() *LogNotifier { return &LogNotifier{} }

// Notify 日志通知实现（短信暂缓）
func (LogNotifier) Notify(ctx context.Context, e *Event) {
	logger.L().Info("通知",
		zap.String("type", e.Type),
		zap.String("title", e.Title),
		zap.String("content", e.Content),
		zap.Int64s("user_ids", e.UserIDs),
	)
}

// Send 发布通知事件到 MQ，由消费者落日志/后续对接真实通道
// 注意：直接传 Event，由 mq.Publish 统一序列化；先 Marshal 成 []byte 会二次编码为 base64 字符串
func Send(b *mq.MQ, e *Event) {
	if b == nil {
		logger.L().Warn("MQ 未连接，通知降级为日志",
			zap.String("type", e.Type), zap.String("title", e.Title))
		return
	}
	_ = b.Publish(mq.EventExchange, mq.KeyNotify, e)
}

// RunConsumer 启动通知消费者（notify.* 事件 → Notifier）
func RunConsumer(b *mq.MQ, n Notifier) {
	b.Consume(mq.QueueNotify, func(ctx context.Context, body []byte) error {
		var e Event
		if err := json.Unmarshal(body, &e); err != nil {
			return err
		}
		n.Notify(ctx, &e)
		return nil
	})
}
