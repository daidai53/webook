// Copyright@daidai53 2023
package ioc

import (
	"github.com/daidai53/webook/config"
	"github.com/daidai53/webook/internal/repository/dao"
	"github.com/daidai53/webook/pkg/gormx"
	"github.com/daidai53/webook/pkg/logger"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/prometheus"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN), &gorm.Config{
		//Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
		//	SlowThreshold: 0,
		//	LogLevel:      glogger.Info,
		//}),
	})
	if err != nil {
		panic(err)
	}
	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	cb := gormx.NewCallbacks(prometheus2.SummaryOpts{
		Namespace: "daidai53",
		Subsystem: "webook",
		Name:      "gorm_db",
		ConstLabels: map[string]string{
			"instance_id": "ins_1",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})

	err = db.Use(cb)
	if err != nil {
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(s string, i ...any) {
	g(s, logger.Field{Key: "args", Val: i})
}
