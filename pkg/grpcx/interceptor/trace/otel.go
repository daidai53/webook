// Copyright@daidai53 2024
package trace

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type OTELInterceptorBuilder struct {
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
}

func (o *OTELInterceptorBuilder) BuildUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	tracer := o.tracer
	if tracer == nil {
		tracer = otel.Tracer("github.com/daidai53/webook/pkg/grpcx")
	}
	propagator := o.propagator
	if propagator == nil {
		propagator = otel.GetTextMapPropagator()
	}
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		attribute.Key("rpc.grpc.kind").String("unary"),
		attribute.Key("rpc.component").String("server"),
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithAttributes(attrs...),
			trace.WithSpanKind(trace.SpanKindServer))
		defer func() {
			span.End()
		}()
		span.SetAttributes(
			semconv.RPCMethodKey.String(info.FullMethod),
		)
		defer func() {
			if err != nil {
				span.RecordError(err)
				if e, _ := status.FromError(err); e != nil {
					span.SetAttributes(semconv.RPCGRPCStatusCodeKey.String(e.Code().String()))
				}
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "OK")
			}
		}()
		return handler(ctx, req)
	}
}

//func extract(ctx context.Context, propagators propagation.TextMapPropagator) context.Context {
//	md, ok := metadata.FromIncomingContext(ctx)
//	if !ok {
//		md = metadata.MD{}
//	}
//	return propagators.Extract(ctx, GrpcHeaderCarrier(md))
//}
//
//type GrpcHeaderCarrier metadata.MD
