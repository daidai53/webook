package domain

type Amount struct {
	Currency string
	Total    int64
}

type Payment struct {
	Amt         Amount
	BizTradeNo  string
	Description string
	Status      PaymentStatus
	TxnID       string
}

type PaymentStatus uint8

func (p PaymentStatus) Uint8() uint8 {
	return uint8(p)
}

const (
	PaymentStatusUnknown = iota
	PaymentStatusInit
	PaymentStatusSuccess
	PaymentStatusFailed
	PaymentStatusRefund
)
