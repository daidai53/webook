// Copyright@daidai53 2023
package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type RedisJWTHandler struct {
	client        redis.Cmdable
	signingMethod jwt.SigningMethod
	rtExpiration  time.Duration
}

func NewRedisJWTHandler(cmd redis.Cmdable) Handler {
	return &RedisJWTHandler{
		client:        cmd,
		signingMethod: jwt.SigningMethodHS512,
		rtExpiration:  time.Hour * 7 * 24,
	}
}

func (r *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	cnt, err := r.client.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if err != nil {
		return err
	}
	if cnt > 0 {
		return errors.New("token 无效")
	}
	return nil
}

func (r *RedisJWTHandler) ExtractToken(context *gin.Context) string {
	authCode := context.GetHeader("authorization")
	if authCode == "" {
		return authCode
	}
	segs := strings.Split(authCode, " ")
	if len(segs) != 2 {
		return ""
	}
	return segs[1]
}

func (r *RedisJWTHandler) SetJWTToken(context *gin.Context, uid int64, ssid string) error {
	uc := UserClaim{
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: context.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(r.signingMethod, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	context.Header("x-jwt-token", tokenStr)
	return nil
}

func (r *RedisJWTHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	rc := RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(r.rtExpiration)),
		},
	}
	token := jwt.NewWithClaims(r.signingMethod, rc)
	tokenStr, err := token.SignedString(RCJWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (r *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := r.SetRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return r.SetJWTToken(ctx, uid, ssid)
}

func (r *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-refresh-token", "")
	ctx.Header("x-jwt-token", "")
	uc := ctx.MustGet("user").(UserClaim)
	return r.client.Set(ctx, fmt.Sprintf("users:ssid:%s", uc.Ssid),
		"", r.rtExpiration).Err()
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}

type UserClaim struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
	Ssid      string
}

var JWTKey = []byte("fjquxoimjdoiwqjoifjnoi")
var RCJWTKey = []byte("fjquxoimjdoifdasfqjoifjnoi")
