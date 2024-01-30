// Copyright@daidai53 2024
package ioc

import (
	"github.com/IBM/sarama"
	"github.com/daidai53/webook/interactive/repository/dao"
	"github.com/daidai53/webook/pkg/ginx"
	"github.com/daidai53/webook/pkg/gormx/connpool"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/daidai53/webook/pkg/migrator/events"
	"github.com/daidai53/webook/pkg/migrator/events/fixer"
	"github.com/daidai53/webook/pkg/migrator/scheduler"
	"github.com/gin-gonic/gin"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

// InitGinxServer 管理后台的 server
func InitGinxServer(l logger.LoggerV1,
	src SrcDB,
	dst DstDB,
	pool *connpool.DoubleWritePool,
	producer events.Producer) *ginx.Server {
	engine := gin.Default()
	group := engine.Group("/migrator")
	ginx.InitCounter(prometheus2.CounterOpts{
		Namespace: "daidai53",
		Subsystem: "webook_intr_admin",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})
	sch := scheduler.NewScheduler[dao.Interactive](src, dst, pool, l, producer)
	sch.RegisterRoutes(group)
	return &ginx.Server{
		Engine: engine,
		Addr:   viper.GetString("migrator.http.addr"),
	}
}

func InitInteractiveProducer(p sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer(p, "inconsistent_interactive")
}

func InitFixerConsumer(client sarama.Client,
	l logger.LoggerV1,
	src SrcDB,
	dst DstDB) *fixer.Consumer[dao.Interactive] {
	res, err := fixer.NewConsumer[dao.Interactive](client, l, "inconsistent_interactive", src, dst)
	if err != nil {
		panic(err)
	}
	return res
}
