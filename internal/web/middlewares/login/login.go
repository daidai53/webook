// Copyright@daidai53 2023
package login

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type MiddlewareBuilder struct {
}

func (m *MiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(context *gin.Context) {
		path := context.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" {
			return
		}

		sess := sessions.Default(context)
		userId := sess.Get("userId")
		if userId == nil {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		const updateTime = "update_time"
		now := time.Now()
		key := sess.Get(updateTime)
		lastUpdTime, ok := key.(time.Time)
		if !ok || now.Sub(lastUpdTime) > time.Minute {
			sess.Set("userId", userId)
			sess.Set(updateTime, now)
			err := sess.Save()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
