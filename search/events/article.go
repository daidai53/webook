// Copyright@daidai53 2024
package events

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/daidai53/webook/pkg/saramax"
	"github.com/daidai53/webook/search/domain"
	"github.com/daidai53/webook/search/service"
	"time"
)

type ArticleConsumer struct {
	syncSvc service.SyncService
	client  sarama.Client
	l       logger.LoggerV1
}

type ArticleEvent struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Status  int32  `json:"status"`
}

func (a *ArticleConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("search_sync_data", a.client)
	if err != nil {
		return err
	}
	go func() {
		err = cg.Consume(context.Background(),
			[]string{"sync_article_events"},
			saramax.NewHandler[ArticleEvent](a.l, a.Consume))
		if err != nil {
			a.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (a *ArticleConsumer) Consume(sg *sarama.ConsumerMessage,
	evt ArticleEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return a.syncSvc.SyncArticle(ctx, domain.Article{
		Id:      evt.Id,
		Title:   evt.Title,
		Content: evt.Content,
		Status:  evt.Status,
	})
}
