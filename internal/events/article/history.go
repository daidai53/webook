// Copyright@daidai53 2024
package article

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/repository"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/daidai53/webook/pkg/saramax"
	"time"
)

type HistoryRecordConsumer struct {
	repo   repository.HistoryRecordRepository
	client sarama.Client
	l      logger.LoggerV1
}

func NewHistoryRecordConsumer(repo repository.HistoryRecordRepository, client sarama.Client, l logger.LoggerV1) *HistoryRecordConsumer {
	return &HistoryRecordConsumer{repo: repo, client: client, l: l}
}

func (h *HistoryRecordConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", h.client)
	if err != nil {
		return err
	}

	go func() {
		er := cg.Consume(context.Background(),
			[]string{TopicReadEvent},
			saramax.NewHandler[ReadEvent](h.l, h.Consume),
		)
		if er != nil {
			h.l.Error("退出消费",
				logger.Error(er))
		}
	}()
	return nil
}

func (h *HistoryRecordConsumer) Consume(msg *sarama.ConsumerMessage, event ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return h.repo.AddRecord(ctx, domain.HistoryRecord{
		Biz:   "article",
		BizId: event.Aid,
		Uid:   event.Uid,
	})
}
