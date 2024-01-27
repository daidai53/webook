// Copyright@daidai53 2024
package events

import (
	"context"
	"github.com/IBM/sarama"
	interrepov1 "github.com/daidai53/webook/api/proto/gen/inter/interrepo/v1"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/daidai53/webook/pkg/saramax"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const TopicReadEvent = "article_read"

type InteractiveReadEventConsumer struct {
	repo   interrepov1.InteractiveRepositoryClient
	client sarama.Client
	l      logger.LoggerV1
}

func NewInteractiveReadEventConsumer(repo interrepov1.InteractiveRepositoryClient, client sarama.Client, l logger.LoggerV1) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{
		repo:   repo,
		client: client,
		l:      l,
	}
}

func (i *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}

	go func() {
		er := cg.Consume(context.Background(),
			[]string{TopicReadEvent},
			saramax.NewBatchHandler[ReadEvent](i.BatchConsume, i.l,
				prometheus.CounterOpts{
					Namespace: "daidai53",
					Subsystem: "webook",
					Name:      "biz_kafka",
				}),
		)
		if er != nil {
			i.l.Error("退出消费",
				logger.Error(er))
		}
	}()
	return nil
}

func (i *InteractiveReadEventConsumer) StartSingle() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}

	go func() {
		er := cg.Consume(context.Background(),
			[]string{TopicReadEvent},
			saramax.NewHandler[ReadEvent](i.l, i.Consume),
		)
		if er != nil {
			i.l.Error("退出消费",
				logger.Error(er))
		}
	}()
	return nil
}

func (i *InteractiveReadEventConsumer) BatchConsume(msgs []*sarama.ConsumerMessage, events []ReadEvent) error {
	bizs := make([]string, 0, len(events))
	bizIds := make([]int64, 0, len(events))
	for _, evt := range events {
		bizs = append(bizs, "article")
		bizIds = append(bizIds, evt.Aid)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := i.repo.BatchIncrReadCnt(ctx, &interrepov1.BatchIncrReadCntRequest{
		Biz:   bizs,
		BizId: bizIds,
	})
	return err
}

func (i *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, event ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := i.repo.IncrReadCnt(ctx, &interrepov1.IncrReadCntRequest{
		Biz:   "article",
		BizId: event.Aid,
	})
	return err
}

type ReadEvent struct {
	Aid int64
	Uid int64
}
