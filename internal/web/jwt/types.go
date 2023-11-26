// Copyright@daidai53 2023
package jwt

import "github.com/gin-gonic/gin"

type Handler interface {
	ExtractToken(context *gin.Context) string
	SetLoginToken(ctx *gin.Context, uid int64) error
	SetJWTToken(context *gin.Context, uid int64, ssid string) error
	CheckSession(ctx *gin.Context, ssid string) error
	ClearToken(ctx *gin.Context) error
}
