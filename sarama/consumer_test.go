// Copyright@daidai53 2023
package sarama

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"log"
	"testing"
	"time"
)

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	consumer, err := sarama.NewConsumerGroup(addr, "demo", cfg)
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err = consumer.Consume(ctx, []string{"test_topic"}, &ConsumerHandler{})
	assert.NoError(t, err)
}

type ConsumerHandler struct {
}

func (c *ConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	var offset int64 = 0
	for _, part := range session.Claims()["test_topic"] {
		session.ResetOffset("test_topic", part, offset, "")
	}
	return nil
}

func (c *ConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	const batchSize = 10
	for {
		batch := make([]*sarama.ConsumerMessage, 0, batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var eg errgroup.Group
		var done = false
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					return nil
				}
				batch = append(batch, msg)
				eg.Go(func() error {
					log.Println(string(msg.Value))
					return nil
				})
			}
			if done {
				break
			}
		}
		cancel()
		err := eg.Wait()
		if err != nil {
			log.Println(err)
			continue
		}
		for _, msg := range batch {
			session.MarkMessage(msg, "")
		}
	}
	return nil
}

func (c *ConsumerHandler) ConsumeClaimV1(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		log.Println(string(msg.Value))
		session.MarkMessage(msg, "")
	}
	return nil
}
