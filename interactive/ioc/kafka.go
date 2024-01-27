// Copyright@daidai53 2024
package ioc

import (
	"github.com/IBM/sarama"
	events2 "github.com/daidai53/webook/interactive/events"
	"github.com/daidai53/webook/internal/events"
	"github.com/spf13/viper"
)

func InitSaramaClient() sarama.Client {
	type Config struct {
		Addr []string
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	scfg := sarama.NewConfig()
	scfg.Producer.Return.Successes = true
	client, err := sarama.NewClient(cfg.Addr, scfg)
	if err != nil {
		panic(err)
	}
	return client
}

func InitSyncProducer(c sarama.Client) sarama.SyncProducer {
	p, err := sarama.NewSyncProducerFromClient(c)
	if err != nil {
		panic(err)
	}
	return p
}

func InitConsumers(c *events2.InteractiveReadEventConsumer) []events.Consumer {
	return []events.Consumer{
		c,
	}
}
