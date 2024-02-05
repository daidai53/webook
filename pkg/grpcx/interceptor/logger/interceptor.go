// Copyright@daidai53 2024
package logger

import (
	"context"
	"fmt"
	"github.com/daidai53/webook/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime"
	"time"
)

type InterceptorBuilder struct {
	l logger.LoggerV1
}

func (i *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		event := "normal"
		defer func() {
			if rec := recover(); rec != nil {
				switch re := rec.(type) {
				case error:
					err = re
				default:
					err = fmt.Errorf("%v", rec)
				}
				event = "recover"
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				err = status.New(codes.Internal, "panic, err"+err.Error()).Err()
			}
			cost := time.Since(start)
			fields := []logger.Field{
				logger.String("type", "unary"),
				logger.Int64("cost", cost.Milliseconds()),
				logger.String("event", event),
				logger.String("method", info.FullMethod),
			}
			st, _ := status.FromError(err)
			if st != nil {
				fields = append(fields, logger.String("code", st.Code().String()),
					logger.String("code_msg", st.Message()))
			}
			i.l.Info("RPC调用", fields...)
		}()
		resp, err = handler(ctx, req)
		return
	}
}
