// Copyright@daidai53 2024
package service

import (
	"context"
	"errors"
	"fmt"
	accountv1 "github.com/daidai53/webook/api/proto/gen/account/v1"
	pmtv1 "github.com/daidai53/webook/api/proto/gen/payment/v1"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/daidai53/webook/reward/domain"
	"github.com/daidai53/webook/reward/repository"
	"strconv"
	"strings"
)

type WechatNativeRewardService struct {
	client pmtv1.WechatPaymentServiceClient
	repo   repository.RewardRepository
	l      logger.LoggerV1
	accCli accountv1.AccountServiceClient
}

func (w *WechatNativeRewardService) PreReward(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	// 本地repo获取url，防止重复申请
	res, err := w.repo.GetCachedCodeURL(ctx, r)
	if err == nil {
		return res, nil
	}
	rid, err := w.repo.CreateReward(ctx, r)
	if err != nil {
		return domain.CodeURL{}, err
	}
	resp, err := w.client.NativePrePay(ctx, &pmtv1.PrePayRequest{
		Amt: &pmtv1.Amount{
			Total:    r.Amt,
			Currency: "CNY",
		},
		BizTradeNo:  w.bizTradeNO(rid),
		Description: r.Target.BizName,
	})
	if err != nil {
		return domain.CodeURL{}, err
	}
	cu := domain.CodeURL{
		Rid: rid,
		URL: resp.CodeUrl,
	}
	err1 := w.repo.CacheCodeURL(ctx, cu, r)
	if err1 != nil {
		w.l.Error("缓存二维码失败",
			logger.Error(err),
			logger.Int64("rid", rid))
	}
	return cu, nil
}

func (w *WechatNativeRewardService) GetReward(ctx context.Context, rid, uid int64) (domain.Reward, error) {
	// 快路径，查本地
	res, err := w.repo.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}
	if res.Uid != uid {
		return domain.Reward{}, errors.New("非法访问别人的打赏记录")
	}
	// 降级或者限流时，不走慢路径
	if ctx.Value("limited") == "true" {
		return res, nil
	}
	if !res.Completed() {
		// 去问一下支付
		pmtRes, err := w.client.GetPayment(ctx, &pmtv1.GetPaymentRequest{
			BizTradeNo: w.bizTradeNO(rid),
		})
		if err != nil {
			w.l.Error("慢路径查询支付状态失败",
				logger.Error(err),
				logger.Int64("rid", rid))
			return res, nil
		}
		switch pmtRes.GetStatus() {
		case pmtv1.PaymentStatus_PaymentStatusRefund:
			res.Status = domain.RewardStatusFailed
		case pmtv1.PaymentStatus_PaymentStatusSuccess:
			res.Status = domain.RewardStatusSuccess
		case pmtv1.PaymentStatus_PaymentStatusInit:
			res.Status = domain.RewardStatusInit
		case pmtv1.PaymentStatus_PaymentStatusFailed:
			res.Status = domain.RewardStatusFailed
		case pmtv1.PaymentStatus_PaymentStatusUnknown:
		}
		err = w.UpdateReward(ctx, w.bizTradeNO(rid), res.Status)
		if err != nil {
			w.l.Error("慢路径更新本地状态失败",
				logger.Error(err),
				logger.Int64("rid", rid))
		}
	}
	return res, nil
}

func (w *WechatNativeRewardService) UpdateReward(ctx context.Context, bizTradeNo string, status domain.RewardStatus) error {
	rid := w.toRid(bizTradeNo)
	err := w.repo.UpdateStatus(ctx, rid, status)
	if err != nil {
		return err
	}
	// 入账部分，调用记账服务，进行分账和记账
	if status == domain.RewardStatusPayed {
		rew, err := w.repo.GetReward(ctx, rid)
		if err != nil {
			return err
		}
		// 抽成
		pltAmt := int64(float64(rew.Amt) * 0.1)
		_, err = w.accCli.Credit(ctx, &accountv1.CreditRequest{
			Biz:   "reward",
			BizId: rid,
			Items: []*accountv1.CreditItem{
				{
					AccountType: accountv1.AccountType_AccountTypePlatform,
					Amt:         pltAmt,
					Currency:    "CNY",
				},
				{
					AccountType: accountv1.AccountType_AccountTypeReward,
					Amt:         rew.Amt - pltAmt,
					Currency:    "CNY",
				},
			},
		})
		if err != nil {
			w.l.Error("入账失败",
				logger.Error(err),
				logger.String("biz_trade_no", bizTradeNo))
			return err
		}
	}
	return nil
}

func (w *WechatNativeRewardService) bizTradeNO(rid int64) string {
	return fmt.Sprintf("reward-%d", rid)
}

func (w *WechatNativeRewardService) toRid(bizTradeNo string) int64 {
	ridStr := strings.Split(bizTradeNo, "-")
	val, _ := strconv.ParseInt(ridStr[1], 10, 64)
	return val
}
