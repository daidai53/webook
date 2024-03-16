// Copyright@daidai53 2024
package app

import (
	"github.com/daidai53/webook/internal/events"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

type App struct {
	Server    *gin.Engine
	Consumers []events.Consumer
	Cron      *cron.Cron
}
