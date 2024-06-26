// Copyright@daidai53 2024
package prometheus

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type Builder struct {
	Namespace  string
	Subsystem  string
	Name       string
	InstanceID string
}

func NewBuilder(namespace string, subsystem string, name string, instanceID string) *Builder {
	return &Builder{Namespace: namespace, Subsystem: subsystem, Name: name, InstanceID: instanceID}
}

func (b *Builder) BuildResponseTime() gin.HandlerFunc {
	labels := []string{
		"method",
		"pattern",
		"status",
	}
	vector := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			// 这三个字段不能有 _ 以外的其他符号
			Namespace: b.Namespace,
			Subsystem: b.Subsystem,
			Name:      b.Name + "_resp_time",
			ConstLabels: map[string]string{
				"instance_id": b.InstanceID,
			},
			Objectives: map[float64]float64{
				0.5:   0.01,
				0.75:  0.01,
				0.9:   0.01,
				0.99:  0.001,
				0.999: 0.0001,
			},
		},
		labels,
	)
	prometheus.MustRegister(vector)
	return func(ctx *gin.Context) {
		start := time.Now()
		defer func() {
			duration := time.Since(start).Milliseconds()
			method := ctx.Request.Method
			pattern := ctx.FullPath()
			status := ctx.Writer.Status()
			vector.WithLabelValues(method, pattern, strconv.Itoa(status)).
				Observe(float64(duration))
		}()
		ctx.Next()
	}
}

func (b *Builder) BuildActiveRequest() gin.HandlerFunc {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      b.Name + "_active_req",
		ConstLabels: map[string]string{
			"instance_id": b.InstanceID,
		},
	})
	prometheus.MustRegister(gauge)
	return func(ctx *gin.Context) {
		gauge.Inc()
		defer gauge.Dec()
		ctx.Next()
	}
}
