package repository

import (
	"context"
	"github.com/daidai53/webook/payment/domain"
	"time"
)

type PaymentRepository interface {
	AddPayment(ctx context.Context, pmt domain.Payment) error
	UpdatePayment(ctx context.Context, pmt domain.Payment) error
	FindExpiredPayment(ctx context.Context, offset, limit int, t time.Time) ([]domain.Payment, error)
	GetPayment(ctx context.Context, bizTradeId string) (domain.Payment, error)
	StorePaymentEvent(ctx context.Context, bizTradeNo string, status uint8) (int64, error)
	SetPaymentEventSent(ctx context.Context, pid int64) error
	FindUnsentPaymentEvents(ctx context.Context, offset, limit int) ([]domain.PaymentEvent, error)
}
