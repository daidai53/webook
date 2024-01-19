// Copyright@daidai53 2024
package ioc

import (
	"github.com/daidai53/webook/internal/job"
	"github.com/daidai53/webook/internal/service"
	"github.com/daidai53/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"time"
)

func InitRankingJob(svc service.RankingService, l logger.LoggerV1, client *rlock.Client) *job.RankingJob {
	return job.NewRankingJob(svc, time.Second*30, l, client)
}

func InitJobs(l logger.LoggerV1, rJob *job.RankingJob) *cron.Cron {
	builder := job.NewCronJobBuilder(l, prometheus.SummaryOpts{
		Namespace: "daidai53",
		Subsystem: "webook",
		Name:      "cron_job",
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	expr := cron.New(cron.WithSeconds())
	_, err := expr.AddJob("@every 1m", builder.Build(rJob))
	if err != nil {
		panic(err)
	}
	return expr
}
