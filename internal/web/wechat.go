// Copyright@daidai53 2023
package web

import (
	"fmt"
	"github.com/daidai53/webook/internal/service"
	"github.com/daidai53/webook/internal/service/oauth2/wechat"
	ijwt "github.com/daidai53/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
	"time"
)

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	ijwt.Handler
	stateCookieName string
}

func NewOAuth2WechatHandler(svc wechat.Service, usrSvc service.UserService, hdl ijwt.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:             svc,
		userSvc:         usrSvc,
		stateCookieName: "jwt-state",
		Handler:         hdl,
	}
}

func (o *OAuth2WechatHandler) ResiterRoutes(s *gin.Engine) {
	g := s.Group("/oauth2/wechat")
	g.GET("/authurl", o.Auth2URL)
	g.Any("/callback", o.Callback)
}

func (o *OAuth2WechatHandler) Auth2URL(context *gin.Context) {
	state := uuid.New()
	val, err := o.svc.AuthURL(context, state)
	if err != nil {
		context.JSON(http.StatusOK, Result{
			Msg:  "构造跳转URL失败",
			Code: 5,
		})
	}
	err = o.setStateCookie(context, state)
	if err != nil {
		context.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "服务器异常",
		})
	}
	context.JSON(http.StatusOK, Result{
		Data: val,
	})
}

func (o *OAuth2WechatHandler) Callback(context *gin.Context) {
	err := o.verifyState(context)
	if err != nil {
		context.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "非法请求",
		})
		return
	}
	code := context.Query("code")
	// state := context.Query("state")
	wechatInfo, err := o.svc.VerifyCode(context, code)
	if err != nil {
		context.JSON(http.StatusOK, Result{
			Msg:  "授权码有误",
			Code: 4,
		})
	}
	user, err := o.userSvc.FindOrCreateByWeChat(context, wechatInfo)
	if err != nil {
		context.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
	}
	err = o.SetLoginToken(context, user.Id)
	if err != nil {
		context.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "系统错误",
		})
	}
	context.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (o *OAuth2WechatHandler) setStateCookie(context *gin.Context, state string) error {
	sc := StateClaims{
		State: state,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, sc)
	tokenStr, err := token.SignedString(ijwt.JWTKey)
	if err != nil {
		return err
	}
	context.SetCookie(o.stateCookieName, tokenStr, 600, "/oauth2/wechat/callback",
		"", false, true)
	return nil
}

func (o *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	ck, err := ctx.Cookie(o.stateCookieName)
	if err != nil {
		return err
	}
	var sc StateClaims
	_, err = jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return ijwt.JWTKey, nil
	})
	if err != nil {
		return err
	}
	if state != sc.State {
		return fmt.Errorf("state不匹配")
	}
	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}
