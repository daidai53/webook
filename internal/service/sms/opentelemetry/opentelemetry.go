// Copyright@daidai53 2024
package opentelemetry

import (
	"context"
	"github.com/daidai53/webook/internal/service/sms"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Decorator struct {
	svc   sms.Service
	trace trace.Tracer
}

func NewDecorator(svc sms.Service) sms.Service {
	res := &Decorator{
		svc: svc,
	}
	res.trace = otel.Tracer("github.com/daidai53/webook/internal/service/sms")
	return res
}

func (d *Decorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	ctx, span := d.trace.Start(ctx, "sms")
	defer span.End()
	span.SetAttributes(attribute.String("tpl_id", tplId))
	span.AddEvent("发短信")
	err := d.svc.Send(ctx, tplId, args, numbers...)
	if err != nil {
		span.RecordError(err)
	}
	return err
}
