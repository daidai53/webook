// Copyright@daidai53 2024
package events

type PaymentEvent struct {
	BizTradeNo string
	Status     uint8
}

func (p PaymentEvent) Topic() string {
	return "payment_events"
}
