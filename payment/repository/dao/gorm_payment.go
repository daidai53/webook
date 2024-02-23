package dao

import (
	"context"
	"github.com/daidai53/webook/payment/domain"
	"gorm.io/gorm"
	"time"
)

type GORMPaymentDAO struct {
	db *gorm.DB
}

func (g *GORMPaymentDAO) GetPayment(ctx context.Context, bizTradeNo string) (Payment, error) {
	var pmt Payment
	err := g.db.WithContext(ctx).Where("biz_trade_no=?", bizTradeNo).First(&pmt).Error
	return pmt, err
}

func (g *GORMPaymentDAO) FindExpiredPayment(ctx context.Context, offset, limit int, t time.Time) ([]Payment, error) {
	var res []Payment
	err := g.db.WithContext(ctx).Where("status=? AND u_time<?", domain.PaymentStatusInit, t.UnixMilli()).
		Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (g *GORMPaymentDAO) UpdateTxnIDAndStatus(
	ctx context.Context, tradeNo string, txnId string, status domain.PaymentStatus) error {
	return g.db.WithContext(ctx).Model(&Payment{}).
		Where("biz_trade_no=?", tradeNo).
		Updates(map[string]any{
			"biz_trade_no": tradeNo,
			"txn_id":       txnId,
			"status":       status,
		}).Error
}

func (g *GORMPaymentDAO) Insert(ctx context.Context, pmt Payment) error {
	now := time.Now().UnixMilli()
	pmt.CTime = now
	pmt.UTime = now
	return g.db.WithContext(ctx).Create(&pmt).Error
}
