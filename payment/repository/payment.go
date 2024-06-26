package repository

import (
	"context"
	"database/sql"
	"github.com/daidai53/webook/payment/domain"
	"github.com/daidai53/webook/payment/repository/dao"
	"time"
)

type paymentRepository struct {
	dao dao.PaymentDAO
}

func (p *paymentRepository) FindUnsentPaymentEvents(ctx context.Context, offset, limit int) ([]domain.PaymentEvent, error) {
	//TODO implement me
	panic("implement me")
}

func (p *paymentRepository) StorePaymentEvent(ctx context.Context, bizTradeNo string, status uint8) (int64, error) {
	return p.dao.StorePaymentEvent(ctx, bizTradeNo, status)
}

func (p *paymentRepository) SetPaymentEventSent(ctx context.Context, pid int64) error {
	return p.dao.SetPaymentEventSent(ctx, pid)
}

func (p *paymentRepository) GetPayment(ctx context.Context, bizTradeId string) (domain.Payment, error) {
	pmt, err := p.dao.GetPayment(ctx, bizTradeId)
	return p.toDomain(&pmt), err
}

func (p *paymentRepository) FindExpiredPayment(ctx context.Context, offset, limit int, t time.Time) ([]domain.Payment, error) {
	pmts, err := p.dao.FindExpiredPayment(ctx, offset, limit, t)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Payment, 0, len(pmts))
	for _, pmt := range pmts {
		res = append(res, p.toDomain(&pmt))
	}
	return res, nil
}

func (p *paymentRepository) AddPayment(ctx context.Context, pmt domain.Payment) error {
	return p.dao.Insert(ctx, p.toEntity(&pmt))
}

func (p *paymentRepository) UpdatePayment(ctx context.Context, pmt domain.Payment) error {
	return p.dao.UpdateTxnIDAndStatus(ctx, pmt.BizTradeNo, pmt.TxnID, pmt.Status)
}

func (p *paymentRepository) toEntity(pmt *domain.Payment) dao.Payment {
	return dao.Payment{
		Amt:         pmt.Amt.Total,
		Currency:    pmt.Amt.Currency,
		Description: pmt.Description,
		BizTradeNO:  pmt.Description,
		TxnID: sql.NullString{
			String: pmt.TxnID,
			Valid:  true,
		},
	}
}

func (p *paymentRepository) toDomain(pmt *dao.Payment) domain.Payment {
	return domain.Payment{
		Amt: domain.Amount{
			Total:    pmt.Amt,
			Currency: pmt.Currency,
		},
		BizTradeNo:  pmt.BizTradeNO,
		Description: pmt.Description,
		Status:      domain.PaymentStatus(pmt.Status),
		TxnID:       pmt.TxnID.String,
	}
}
