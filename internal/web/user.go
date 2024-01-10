// Copyright@daidai53 2023
package web

import (
	"errors"
	"fmt"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/errs"
	"github.com/daidai53/webook/internal/service"
	ijwt "github.com/daidai53/webook/internal/web/jwt"
	"github.com/daidai53/webook/pkg/ginx"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

const (
	emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	passwordRegexPattern = "(?=.*([a-zA-Z].*))(?=.*[0-9].*)[a-zA-Z0-9-*/+.~!@#$%^&*()]{8,72}$"
	birthdayRegexPattern = "^(?:(?!0000)[0-9]{4}-(?:(?:0[1-9]|1[0-2])-(?:0[1-9]|1[0-9]|2[0-8])|(?:0[13-9]|1[0-2])-(?:29|30)" +
		"|(?:0[13578]|1[02])-31)|(?:[0-9]{2}(?:0[48]|[2468][048]|[13579][26])|(?:0[48]|[2468][048]|[13579][26])00)-02-29)$"
	bizLogin = "login"
)

type UserHandler struct {
	ijwt.Handler
	emailRegExp    *regexp.Regexp
	passwordRegExp *regexp.Regexp
	birthdayRegExp *regexp.Regexp
	svc            service.UserService
	codeSvc        service.CodeService
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, hdl ijwt.Handler) *UserHandler {
	return &UserHandler{
		emailRegExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		birthdayRegExp: regexp.MustCompile(birthdayRegexPattern, regexp.None),
		svc:            svc,
		codeSvc:        codeSvc,
		Handler:        hdl,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", ginx.WrapBody(u.SignUp))
	ug.POST("/login", ginx.WrapBody(u.LoginJWT))
	ug.POST("/edit", ginx.WrapBodyAndClaims(u.Edit))
	ug.GET("/profile", ginx.WrapClaims(u.Profile))
	ug.GET("/refresh_token", u.RefreshToken)
	ug.POST("/logout", u.LogOut)

	//手机验证码登录相关功能
	ug.POST("/login_sms/code/send", ginx.WrapBody(u.SendSMSLoginCode))
	ug.POST("/login_sms", ginx.WrapBody(u.LoginSMS))
}

func (u *UserHandler) SendSMSLoginCode(context *gin.Context, req SendSMSCodeReq) (ginx.Result, error) {
	if req.Phone == "" {
		return ginx.Result{
			Code: 4,
			Msg:  "请输入手机号",
		}, nil
	}
	err := u.codeSvc.Send(context, bizLogin, req.Phone)
	switch {
	case err == nil:
		return ginx.Result{
			Msg: "发送成功",
		}, nil
	case errors.Is(err, service.ErrCodeSendTooMany):
		return ginx.Result{
			Code: 4,
			Msg:  "短信发送太频繁，请稍后再试",
		}, nil
	default:
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
}

func (u *UserHandler) LoginSMS(context *gin.Context, req LoginSMSReq) (ginx.Result, error) {
	ok, err := u.codeSvc.Verify(context, bizLogin, req.Phone, req.Code)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	if !ok {
		return ginx.Result{
			Code: 4,
			Msg:  "验证码不对，请重新输入",
		}, nil
	}
	user, err := u.svc.FindOrCreate(context, req.Phone)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	err = u.SetLoginToken(context, user.Id)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Msg: "登录成功",
	}, nil
}

func (u *UserHandler) SignUp(context *gin.Context, req SignUpReq) (ginx.Result, error) {
	isMail, err := u.emailRegExp.MatchString(req.Email)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	if !isMail {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "邮箱输入非法",
		}, err
	}

	if req.Password != req.ConfirmPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "两次输入的密码不一致",
		}, err
	}

	isPassword, err := u.passwordRegExp.MatchString(req.Password)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	if !isPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "密码必须包含字母、数字、特殊字符，并且长度不能小于8位",
		}, err
	}

	err = u.svc.SignUp(context, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	switch {
	case err == nil:
		return ginx.Result{Msg: "OK"}, nil
	case errors.Is(err, service.ErrDuplicateEmail):
		return ginx.Result{Code: errs.UserDuplicateEmail, Msg: "邮箱冲突，请换一个"}, nil
	default:
		return ginx.Result{Code: errs.UserInternalServerError, Msg: "系统错误"}, err
	}

}

func (u *UserHandler) Login(context *gin.Context) {
	type Req struct {
		Email    string
		Password string
	}
	var req Req
	err := context.Bind(&req)
	if err != nil {
		return
	}
	usr, err := u.svc.Login(context, req.Email, req.Password)
	switch {
	case err == nil:
		sess := sessions.Default(context)
		sess.Set("userId", usr.Id)
		sess.Options(sessions.Options{
			MaxAge: 900,
		})
		err = sess.Save()
		if err != nil {
			context.String(http.StatusOK, "系统错误")
			return
		}
		context.String(http.StatusOK, "登录成功")
	case errors.Is(err, service.ErrInvalidUserOrPassword):
		context.String(http.StatusOK, "用户名或者密码不对")
	default:
		context.String(http.StatusOK, "系统错误")
	}

}

func (u *UserHandler) LoginJWT(context *gin.Context, req LoginJWTReq) (ginx.Result, error) {
	usr, err := u.svc.Login(context, req.Email, req.Password)
	switch {
	case err == nil:
		err = u.SetLoginToken(context, usr.Id)
		if err != nil {
			return ginx.Result{
				Code: 5,
				Msg:  "系统错误",
			}, err
		}
		return ginx.Result{
			Msg: "OK",
		}, nil
	case errors.Is(err, service.ErrInvalidUserOrPassword):
		return ginx.Result{Msg: "用户名或者密码不对"}, nil
	default:
		return ginx.Result{Msg: "系统错误"}, err
	}

}

func (u *UserHandler) Edit(context *gin.Context, req UserEditReq,
	uc ijwt.UserClaim) (ginx.Result, error) {
	if runes := []rune(req.Nickname); len(runes) > 30 {
		return ginx.Result{Code: 4, Msg: "昵称的长度不超过30个字符"}, errors.New("昵称的长度不超过30个字符")
	}
	if runes := []rune(req.AboutMe); len(runes) > 300 {
		return ginx.Result{Code: 4, Msg: "简介的长度不超过300个字符"}, errors.New("简介的长度不超过300个字符")
	}
	legalBirthday, err := u.birthdayRegExp.MatchString(req.Birthday)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	if !legalBirthday {
		return ginx.Result{Code: 4, Msg: "生日的输入格式不合法"}, errors.New("生日的输入格式不合法")
	}
	err = u.svc.Edit(context, uc.Uid, req.Nickname, req.Birthday, req.AboutMe)
	if err != nil {
		return ginx.Result{Code: 5, Msg: "系统错误"}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}

func (u *UserHandler) Profile(context *gin.Context, uc ijwt.UserClaim) (ginx.Result, error) {
	usr, err := u.svc.Profile(context, uc.Uid)
	switch {
	case err == nil:
		return ginx.Result{
			Data: UserProfile{
				Nickname: usr.Nickname,
				Email:    usr.Email,
				AboutMe:  usr.AboutMe,
				Birthday: usr.Birthday,
			},
		}, nil
	case errors.Is(err, service.ErrNoUserProfile):
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	default:
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
}

func (u *UserHandler) RefreshToken(context *gin.Context) {
	// 约定，前端在authorization里带上了refresh token
	tokenStr := u.ExtractToken(context)
	var rc ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RCJWTKey, nil
	})
	if err != nil {
		context.AbortWithStatus(http.StatusUnauthorized)
	}
	if token == nil || !token.Valid {
		context.AbortWithStatus(http.StatusUnauthorized)
	}

	err = u.CheckSession(context, fmt.Sprintf("users:ssid:%s", rc.Ssid))
	if err != nil {
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = u.SetJWTToken(context, rc.Uid, rc.Ssid)
	if err != nil {
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	context.JSON(http.StatusOK, ginx.Result{
		Msg: "OK",
	})
}

func (u *UserHandler) LogOut(context *gin.Context) {
	err := u.ClearToken(context)
	if err != nil {
		context.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	context.JSON(http.StatusOK, ginx.Result{
		Msg: "退出登录成功",
	})
}
