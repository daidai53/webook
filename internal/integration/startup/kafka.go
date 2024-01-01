// Copyright@daidai53 2024
package startup

import (
	"github.com/IBM/sarama"
)

func InitSaramaClient() sarama.Client {
	scfg := sarama.NewConfig()
	scfg.Producer.Return.Successes = true
	client, err := sarama.NewClient([]string{"localhost:9094"}, scfg)
	if err != nil {
		panic(err)
	}
	return client
}
