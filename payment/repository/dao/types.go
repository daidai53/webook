package dao

import (
	"context"
	"database/sql"
	"github.com/daidai53/webook/payment/domain"
	"time"
)

type PaymentDAO interface {
	Insert(ctx context.Context, pmt Payment) error
	UpdateTxnIDAndStatus(ctx context.Context, tradeNo string, txnId string, status domain.PaymentStatus) error
	FindExpiredPayment(ctx context.Context, offset, limit int, t time.Time) ([]Payment, error)
	GetPayment(ctx context.Context, bizTradeNo string) (Payment, error)
	StorePaymentEvent(ctx context.Context, bizTradeNo string, status uint8) (int64, error)
	SetPaymentEventSent(ctx context.Context, pid int64) error
	FindUnsentPaymentEvents(ctx context.Context, offset, limit int) ([]PaymentEvent, error)
}

type Payment struct {
	ID          int64 `gorm:"primaryKey,autoIncrement"`
	Amt         int64
	Currency    string
	Description string
	BizTradeNO  string         `gorm:"column:biz_trade_no;type:varchar(256);unique"`
	TxnID       sql.NullString `gorm:"column:txn_id;type:varchar(128);unique"`
	Status      uint8
	UTime       int64 `gorm:"index"`
	CTime       int64
}

type PaymentEvent struct {
	Id         int64
	BizTradeNo string
	Status     uint8
	Sent       bool
	CTime      int64
	UTime      int64
}
