// Copyright@daidai53 2024
package job

import (
	"github.com/daidai53/webook/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"strconv"
	"time"
)

type CronJobBuilder struct {
	l      logger.LoggerV1
	vector *prometheus.SummaryVec
}

func NewCronJobBuilder(l logger.LoggerV1, opts prometheus.SummaryOpts) *CronJobBuilder {
	return &CronJobBuilder{
		l: l,
		vector: prometheus.NewSummaryVec(opts,
			[]string{
				"job",
				"success",
			}),
	}
}

func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobAdaptorFunc(func() {
		start := time.Now()
		b.l.Debug("开始运行",
			logger.String("name", name))
		err := job.Run()
		if err != nil {
			b.l.Error("执行失败", logger.Error(err),
				logger.String("name", name))
		}
		b.l.Debug("结束运行",
			logger.String("name", name))
		duration := time.Since(start)
		b.vector.WithLabelValues(name, strconv.FormatBool(err == nil)).
			Observe(float64(duration.Milliseconds()))
	})
}

type cronJobAdaptorFunc func()

func (c cronJobAdaptorFunc) Run() {
	c()
}
