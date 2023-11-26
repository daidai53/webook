// Copyright@daidai53 2023
package login

import (
	"fmt"
	ijwt "github.com/daidai53/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

type MiddlewareJWTBuilder struct {
	ijwt.Handler
}

func NewMiddlewareJWTBuilder(hdl ijwt.Handler) *MiddlewareJWTBuilder {
	return &MiddlewareJWTBuilder{
		Handler: hdl,
	}
}

func (m *MiddlewareJWTBuilder) CheckLogin() gin.HandlerFunc {
	return func(context *gin.Context) {
		path := context.Request.URL.Path
		if path == "/users/signup" ||
			path == "/users/login" ||
			path == "/users/login_sms/code/send" ||
			path == "/users/login_sms" ||
			path == "/oauth2/wechat/authurl" ||
			path == "/oauth2/wechat/callback" {
			return
		}

		tokenStr := m.ExtractToken(context)
		var uc ijwt.UserClaim
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return ijwt.JWTKey, nil
		})
		if err != nil {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if !strings.EqualFold(uc.UserAgent, context.GetHeader("User-Agent")) {
			context.AbortWithStatus(http.StatusUnauthorized)
			fmt.Println("疑似有仿冒token，已拒绝")
			return
		}

		err = m.CheckSession(context, uc.Ssid)
		if err != nil {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		context.Set("user", uc)
		context.Set("user-id", uc.Uid)
	}
}
