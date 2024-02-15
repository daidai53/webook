package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/daidai53/webook/payment/domain"
	"github.com/daidai53/webook/payment/repository"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"time"
)

var errUnknownTransactionState = errors.New("未知的Transaction状态")

type NativePaymentService struct {
	svc       *native.NativeApiService
	appID     string
	mchID     string
	notifyURL string
	repo      repository.PaymentRepository

	nativeCBTypeToStatus map[string]domain.PaymentStatus

	l logger.LoggerV1
}

func NewNativePaymentService(svc *native.NativeApiService, appID string, mchID string, l logger.LoggerV1,
	repo repository.PaymentRepository) *NativePaymentService {
	return &NativePaymentService{
		svc:       svc,
		appID:     appID,
		mchID:     mchID,
		notifyURL: "http://xxx.xx",
		repo:      repo,
		l:         l,
		nativeCBTypeToStatus: map[string]domain.PaymentStatus{
			"SUCCESS":  domain.PaymentStatusSuccess,
			"PAYERROR": domain.PaymentStatusFailed,
			"CLOSED":   domain.PaymentStatusFailed,
			"NOTPAY":   domain.PaymentStatusInit,
			"REVOKED":  domain.PaymentStatusFailed,
			"REFUND":   domain.PaymentStatusRefund,
		},
	}
}

func (n *NativePaymentService) PrePay(ctx context.Context, pmt domain.Payment) (string, error) {
	err := n.repo.AddPayment(ctx, pmt)
	if err != nil {
		return "", err
	}
	resp, _, err := n.svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(n.appID),
		Mchid:       core.String(n.mchID),
		Description: core.String(pmt.Description),
		Amount: &native.Amount{
			Currency: core.String(pmt.Amt.Currency),
			Total:    core.Int64(pmt.Amt.Total),
		},
		NotifyUrl:  core.String(n.notifyURL),
		OutTradeNo: core.String(pmt.BizTradeNo),
		TimeExpire: core.Time(time.Now().Add(30 * time.Minute)),
	})

	if err != nil {
		return "", err
	}
	return *resp.CodeUrl, nil
}

func (n *NativePaymentService) HandleCallback(ctx context.Context, trans *payments.Transaction) error {
	return n.updateByTxn(ctx, trans)
}

func (n *NativePaymentService) updateByTxn(ctx context.Context, trans *payments.Transaction) error {
	status, ok := n.nativeCBTypeToStatus[*trans.TradeState]
	if !ok {
		return fmt.Errorf("%w, 微信的状态是：%s", errUnknownTransactionState, trans.TradeState)
	}
	n.repo.UpdatePayment(ctx, domain.Payment{
		TxnID:      *trans.TransactionId,
		BizTradeNo: *trans.OutTradeNo,
		Status:     status,
	})
}

func (n *NativePaymentService) SyncWechatInfo(ctx context.Context, tradeNo string) error {
	trans, _, err := n.svc.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(tradeNo),
		Mchid:      core.String(n.mchID),
	})
	if err != nil {
		return err
	}
	return n.updateByTxn(ctx, trans)
}

func (n *NativePaymentService) FindExpiredPayment(ctx context.Context, offset int,
	limit int, t time.Time) ([]domain.Payment, error) {
	return n.repo.FindExpiredPayment(ctx, offset, limit, t)
}
