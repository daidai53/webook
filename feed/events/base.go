package events

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/daidai53/webook/feed/domain"
	"github.com/daidai53/webook/feed/service"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/daidai53/webook/pkg/saramax"
	"time"
)

// FeedEvent 业务方就按照这个格式，将放到feed里面的数据，丢到feed_event这个topic下
type FeedEvent struct {
	Type string
	// 为了序列化和反序列化不出问题
	Metadata map[string]string
}

type FeedEventConsumer struct {
	client sarama.Client
	l      logger.LoggerV1
	svc    service.FeedService
}

func (f *FeedEventConsumer) Start() error {
	consumerGroup, err := sarama.NewConsumerGroupFromClient("feed_event", f.client)
	if err != nil {
		return err
	}
	go func() {
		err := consumerGroup.Consume(context.Background(),
			[]string{"feed_event"},
			saramax.NewHandler[FeedEvent](f.l, f.Consume))
		if err != nil {
			f.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (f *FeedEventConsumer) Consume(msg *sarama.ConsumerMessage, evt FeedEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return f.svc.CreateFeedEvent(ctx, domain.FeedEvent{
		Type: evt.Type,
		Ext:  evt.Metadata,
	})
}
