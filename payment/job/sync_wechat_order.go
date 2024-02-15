package job

import (
	"context"
	"github.com/daidai53/webook/payment/service"
	"github.com/daidai53/webook/pkg/logger"
	"time"
)

type SyncWechatOrderJob struct {
	svc *service.NativePaymentService
	l   logger.LoggerV1
}

func (s *SyncWechatOrderJob) Name() string {
	return "sync_wechat_order_job"
}

// 不必频繁运行，一分钟运行一次即可
func (s *SyncWechatOrderJob) Run() error {
	t := time.Now().Add(-time.Minute * 30)
	offset := 0
	const limit = 100
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		pmts, err := s.svc.FindExpiredPayment(ctx, offset, limit, t)
		cancel()
		if err != nil {
			return err
		}

		for _, pmt := range pmts {
			sCtx, sCancel := context.WithTimeout(context.Background(), time.Second*3)
			err = s.svc.SyncWechatInfo(sCtx, pmt.BizTradeNo)
			sCancel()
			if err != nil {
				s.l.Error("同步微信订单状态失败", logger.Error(err),
					logger.String("biz_trade_no", pmt.BizTradeNo))
			}
		}

		if len(pmts) < limit {
			return nil
		}
		offset += limit
	}
}
