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

type UserConsumer struct {
	syncSvc service.SyncService
	client  sarama.Client
	l       logger.LoggerV1
}

type UserEvent struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
}

func (u *UserConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("sync_user", u.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"sync_user_events"},
			saramax.NewHandler[UserEvent](u.l, u.Consume))
		if err != nil {
			u.l.Error("退出消费循环异常",
				logger.Error(err))
		}
	}()
	return err
}

func (u *UserConsumer) Consume(sg *sarama.ConsumerMessage,
	evt UserEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return u.syncSvc.SyncUser(ctx, domain.User{
		Id:       evt.Id,
		Nickname: evt.Nickname,
		Email:    evt.Email,
		Phone:    evt.Phone,
	})
}
