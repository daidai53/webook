// Copyright@daidai53 2024
package events

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
)

type Producer interface {
	ProduceIncosistentEvent(ctx context.Context, evt InconsistentEvent) error
}

type SaramaProducer struct {
	p     sarama.SyncProducer
	topic string
}

func NewSaramaProducer(p sarama.SyncProducer, topic string) *SaramaProducer {
	return &SaramaProducer{p: p, topic: topic}
}

func (s *SaramaProducer) ProduceIncosistentEvent(ctx context.Context, evt InconsistentEvent) error {
	val, _ := json.Marshal(evt)
	_, _, err := s.p.SendMessage(&sarama.ProducerMessage{
		Topic: s.topic,
		Value: sarama.ByteEncoder(val),
	})
	return err
}
