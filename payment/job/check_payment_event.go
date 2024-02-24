// Copyright@daidai53 2024
package job

import (
	"context"
	"github.com/daidai53/webook/payment/events"
	"github.com/daidai53/webook/payment/repository"
	"github.com/daidai53/webook/pkg/logger"
	"time"
)

type CheckPaymentEvent struct {
	repo     repository.PaymentRepository
	producer events.Producer
	l        logger.LoggerV1
}

func (c *CheckPaymentEvent) Name() string {
	return "check_unsent_payment_event"
}

func (c *CheckPaymentEvent) Run() error {
	offset := 0
	const limit = 100
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		evts, err := c.repo.FindUnsentPaymentEvents(ctx, offset, limit)
		cancel()
		if err != nil {
			return err
		}

		ctx, cancel = context.WithTimeout(context.Background(), time.Second*3)
		for _, event := range evts {
			err = c.producer.ProducePaymentEvent(ctx, events.PaymentEvent{
				BizTradeNo: event.BizTradeNo,
				Status:     event.Status,
			})
			if err == nil {
				err1 := c.repo.SetPaymentEventSent(ctx, event.Pid)
				if err1 != nil {
					c.l.Error(
						"设置PaymentEvent状态为已发送失败",
						logger.Error(err1),
						logger.Int64("pid", event.Pid),
					)
				}
			}
		}
		cancel()

		if len(evts) < limit {
			return nil
		}
		offset += limit
	}
}
