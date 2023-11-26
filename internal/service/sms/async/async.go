// Copyright@daidai53 2023
package async

import (
	"context"
	"github.com/daidai53/webook/internal/service/sms"
	"github.com/daidai53/webook/pkg/limiter"
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	defaultRetryTimes    = 8
	defaultRetryInterval = 10 * time.Second
)

type SmsService struct {
	svc         sms.Service
	limiter     limiter.Limiter
	redisClient redis.Cmdable

	retryTimes    uint32
	retryInterval time.Duration

	blackList []error
}

func NewAsyncSMSService(s sms.Service, l limiter.Limiter, cmd redis.Cmdable, b []error) *SmsService {
	return &SmsService{
		svc:           s,
		limiter:       l,
		redisClient:   cmd,
		retryTimes:    defaultRetryTimes,
		retryInterval: defaultRetryInterval,
		blackList:     b,
	}
}

func (a *SmsService) WithRetryTimes(t uint32) *SmsService {
	a.retryTimes = t
	return a
}

func (a *SmsService) WithRetryInterval(i time.Duration) *SmsService {
	a.retryInterval = i
	return a
}

func (a *SmsService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	return nil
}

func (a *SmsService) ReSend(key string, remainTimes uint32) {

}
