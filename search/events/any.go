// Copyright@daidai53 2024
package events

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/daidai53/webook/pkg/saramax"
	"github.com/daidai53/webook/search/service"
	"time"
)

type AnyConsumer struct {
	syncSvc service.SyncService
	client  sarama.Client
	l       logger.LoggerV1
}

type AnyEvent struct {
	IndexName string `json:"index_name"`
	DocId     string `json:"doc_id"`
	Data      string `json:"data"`
}

func (a *AnyConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("search_sync_data", a.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"sync_any_events"},
			saramax.NewHandler[AnyEvent](a.l, a.Consume))
		if err != nil {
			a.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (a *AnyConsumer) Consume(sg *sarama.ConsumerMessage,
	evt AnyEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return a.syncSvc.SyncAny(ctx, evt.IndexName, evt.DocId, evt.Data)
}
