// Copyright@daidai53 2024
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
)

type Producer interface {
	ProduceInteractiveEvent(ctx context.Context, evt InteractiveEvent) error
}

type SaramaSyncProducer struct {
	client sarama.SyncProducer
}

func (s *SaramaSyncProducer) ProduceInteractiveEvent(ctx context.Context, evt InteractiveEvent) error {
	data, _ := json.Marshal(evt)
	event := SyncDataEvent{
		IndexName: "interactive_index",
		DocId:     fmt.Sprintf("%d-%s-%d"),
		Data:      string(data),
	}
	data, _ = json.Marshal(event)
	_, _, err := s.client.SendMessage(&sarama.ProducerMessage{
		Topic: "sync_any_events",
		Value: sarama.ByteEncoder(data),
	})
	return err
}

type InteractiveEvent struct {
	Uid   int64  `json:"uid"`
	Biz   string `json:"biz"`
	BizId int64  `json:"biz_id"`
	Type  uint8  `json:"type"`
}

type SyncDataEvent struct {
	IndexName string
	DocId     string
	Data      string
}

const (
	TypeLike    = 1
	TypeCollect = 2
)
