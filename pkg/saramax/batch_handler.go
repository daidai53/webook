// Copyright@daidai53 2024
package saramax

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/daidai53/webook/pkg/logger"
	"time"
)

type BatchHandler[T any] struct {
	fn func(msgs []*sarama.ConsumerMessage, ts []T) error
	l  logger.LoggerV1
}

func NewBatchHandler[T any](fn func(msgs []*sarama.ConsumerMessage, ts []T) error, l logger.LoggerV1) *BatchHandler[T] {
	return &BatchHandler[T]{
		fn: fn,
		l:  l,
	}
}

func (b *BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	const batchSize = 10
	for {
		batch := make([]*sarama.ConsumerMessage, 0, batchSize)
		ts := make([]T, 0, batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		var done = false
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					return nil
				}
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					b.l.Error("反序列化消息体失败",
						logger.Error(err),
						logger.String("topic", msg.Topic),
						logger.Int32("partition", msg.Partition),
						logger.Int64("offset", msg.Offset))
					continue
				}
				batch = append(batch, msg)
				ts = append(ts, t)
			}
			if done {
				break
			}
		}
		cancel()
		err := b.fn(batch, ts)
		if err != nil {
			b.l.Error("处理消息失败",
				logger.Error(err))
		}
		for _, msg := range batch {
			session.MarkMessage(msg, "")
		}
	}
	return nil
}
