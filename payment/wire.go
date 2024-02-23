//go:build wireinject

// Copyright@daidai53 2024
package payment

import (
	"github.com/daidai53/webook/payment/ioc"
	"github.com/daidai53/webook/pkg/app"
	"github.com/google/wire"
)

func InitApp() *app.App {
	wire.Build(
		ioc.InitGRPCServer,
		wire.Struct(new(app.App), "*"),
	)
	return new(app.App)
}
