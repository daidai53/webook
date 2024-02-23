// Copyright@daidai53 2024
package events

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type SaramaProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaProducer(client sarama.Client) (*SaramaProducer, error) {
	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}
	return &SaramaProducer{
		producer: p,
	}, nil
}

func (s *SaramaProducer) ProducePaymentEvent(ctx context.Context, event PaymentEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: event.Topic(),
		Key:   sarama.ByteEncoder(event.BizTradeNo),
		Value: sarama.ByteEncoder(data),
	})
	return err
}
