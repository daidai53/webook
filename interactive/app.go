// Copyright@daidai53 2024
package main

import (
	"github.com/daidai53/webook/internal/events"
	"github.com/daidai53/webook/pkg/grpcx"
)

type App struct {
	consumers []events.Consumer
	server    *grpcx.Server
}
