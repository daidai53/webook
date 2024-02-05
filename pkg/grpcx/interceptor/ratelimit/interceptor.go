// Copyright@daidai53 2024
package ratelimit

import (
	"context"
	"github.com/daidai53/webook/pkg/limiter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InterceptorBuilder struct {
	limiter limiter.Limiter
	key     string
}

func NewInterceptorBuilder(limiter limiter.Limiter, key string) *InterceptorBuilder {
	return &InterceptorBuilder{limiter: limiter, key: key}
}

func (i *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := i.limiter.Limit(ctx, i.key)
		if err != nil {
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}

		if limited {
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}
		return handler(ctx, req)
	}
}
