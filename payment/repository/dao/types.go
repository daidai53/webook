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
