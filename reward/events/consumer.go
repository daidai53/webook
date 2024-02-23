// Copyright@daidai53 2024
package events

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/daidai53/webook/pkg/saramax"
	"github.com/daidai53/webook/reward/domain"
	"github.com/daidai53/webook/reward/service"
	"strings"
	"time"
)

type PaymentEvent struct {
	BizTradeNo string
	Status     uint8
}

func (p PaymentEvent) ToDomainStatus() domain.RewardStatus {
	switch p.Status {
	case 1:
		return domain.RewardStatusInit
	case 2:
		return domain.RewardStatusPayed
	case 3, 4:
		return domain.RewardStatusFailed
	default:
		return domain.RewardStatusUnknown
	}
}

type PaymentEventConsumer struct {
	client sarama.Client
	l      logger.LoggerV1
	svc    service.RewardService
}

func (p *PaymentEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("reward",
		p.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"payment_events"},
			saramax.NewHandler[PaymentEvent](p.l, p.Consume))
		if err != nil {
			p.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (p *PaymentEventConsumer) Consume(msg *sarama.ConsumerMessage, event PaymentEvent) error {
	if !strings.HasPrefix(event.BizTradeNo, "reward") {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return p.svc.UpdateReward(ctx, event.BizTradeNo, event.ToDomainStatus())
}
