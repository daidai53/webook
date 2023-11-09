// Copyright@daidai53 2023
package login

import (
	"fmt"
	"github.com/daidai53/webook/internal/web"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
)

type MiddlewareJWTBuilder struct {
}

func (m *MiddlewareJWTBuilder) CheckLogin() gin.HandlerFunc {
	return func(context *gin.Context) {
		path := context.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" ||
			path == "/users/login_sms/code/send" || path == "/users/login_sms" {
			return
		}

		authCode := context.GetHeader("authorization")
		if authCode == "" {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		segs := strings.Split(authCode, " ")
		if len(segs) != 2 {
			context.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		var uc web.UserClaim
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return web.JWTKey, nil
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

		expireTime, _ := uc.GetExpirationTime()
		if expireTime.Sub(time.Now()) < time.Minute*29 {
			uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 30))
			tokenStr, err = token.SignedString(web.JWTKey)
			context.Header("x-jwt-token", tokenStr)
			if err != nil {
				fmt.Println(err)
			}
		}
		context.Set("user-id", uc.Uid)
	}
}
