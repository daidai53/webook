// Copyright@daidai53 2024
package circuitbreaker

import (
	"context"
	"github.com/go-kratos/aegis/circuitbreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InterceptorBuilder struct {
	breaker circuitbreaker.CircuitBreaker
}

func (i *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		err = i.breaker.Allow()
		if err == nil {
			resp, err = handler(ctx, req)
			if err == nil {
				i.breaker.MarkSuccess()
			} else {
				i.breaker.MarkFailed()
			}
		} else {
			i.breaker.MarkFailed()
			return nil, status.Errorf(codes.Unavailable, "熔断")
		}
		return
	}
}
