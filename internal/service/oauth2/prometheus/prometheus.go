// Copyright@daidai53 2024
package prometheus

import (
	"context"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/service/oauth2/wechat"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type Decorator struct {
	wechat.Service
	sum prometheus.Summary
}

func NewDecorator(svc wechat.Service, sum prometheus.Summary) *Decorator {
	return &Decorator{
		Service: svc,
		sum:     sum,
	}
}

func (d *Decorator) VerifyCode(ctx context.Context, code string) (domain.WeChatInfo, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		d.sum.Observe(float64(duration))
	}()
	return d.Service.VerifyCode(ctx, code)
}
