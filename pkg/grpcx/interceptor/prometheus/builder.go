// Copyright@daidai53 2024
package prometheus

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

type InterceptorBuilder struct {
	Namespace  string
	Subsystem  string
	Name       string
	InstanceID string
}

func (i *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	labels := []string{
		"type",
		"service",
		"method",
		"code",
	}
	vector := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			// 这三个字段不能有 _ 以外的其他符号
			Namespace: i.Namespace,
			Subsystem: i.Subsystem,
			Name:      i.Name + "_resp_time",
			ConstLabels: map[string]string{
				"instance_id": i.InstanceID,
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
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		defer func() {
			sn, method := i.splitMethodName(info.FullMethod)
			code := "OK"
			if err != nil {
				st, _ := status.FromError(err)
				code = st.Code().String()
			}
			cost := float64(time.Since(start))
			vector.WithLabelValues("unary", sn, method, code).Observe(cost)
		}()
		resp, err = handler(ctx, req)
		return
	}
}

func (i *InterceptorBuilder) splitMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/")
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}
