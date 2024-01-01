// Copyright@daidai53 2024
package main

import (
	"github.com/daidai53/webook/internal/events"
	"github.com/gin-gonic/gin"
)

type App struct {
	server    *gin.Engine
	consumers []events.Consumer
}
