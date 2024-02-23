package service

import (
	"context"
	"github.com/daidai53/webook/payment/domain"
)

type PaymentService interface {
	PrePay(ctx context.Context, pmt domain.Payment) (string, error)
	GetPayment(ctx context.Context, bizTradeId string) (domain.Payment, error)
}
