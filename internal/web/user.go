// Copyright@daidai53 2023
package web

import (
	"errors"
	"fmt"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/service"
	ijwt "github.com/daidai53/webook/internal/web/jwt"
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
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.LoginJWT)
	ug.POST("/edit", u.Edit)
	ug.GET("/profile", u.Profile)
	ug.GET("/refresh_token", u.RefreshToken)
	ug.POST("/logout", u.LogOut)

	//手机验证码登录相关功能
	ug.POST("/login_sms/code/send", u.SendSMSLoginCode)
	ug.POST("/login_sms", u.LoginSMS)
}

func (u *UserHandler) SendSMSLoginCode(context *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := context.Bind(&req); err != nil {
		return
	}
	if req.Phone == "" {
		context.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "请输入手机号",
		})
	}
	err := u.codeSvc.Send(context, bizLogin, req.Phone)
	switch {
	case err == nil:
		context.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case errors.Is(err, service.ErrCodeSendTooMany):
		context.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "短信发送太频繁，请稍后再试",
		})
	default:
		context.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
}

func (u *UserHandler) LoginSMS(context *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := context.Bind(&req); err != nil {
		return
	}

	ok, err := u.codeSvc.Verify(context, bizLogin, req.Phone, req.Code)
	if err != nil {
		context.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		context.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码不对，请重新输入",
		})
		return
	}
	user, err := u.svc.FindOrCreate(context, req.Phone)
	if err != nil {
		context.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	err = u.SetLoginToken(context, user.Id)
	if err != nil {
		context.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "系统错误",
		})
	}
	context.JSON(http.StatusOK, Result{
		Msg: "登录成功",
	})
}

func (u *UserHandler) SignUp(context *gin.Context) {
	type SignUpReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var signUpReq SignUpReq
	if err := context.Bind(&signUpReq); err != nil {
		fmt.Println(err)
		return
	}

	isMail, err := u.emailRegExp.MatchString(signUpReq.Email)
	if err != nil {
		context.String(http.StatusOK, "系统错误")
		return
	}
	if !isMail {
		context.String(http.StatusOK, "邮箱输入非法")
		return
	}

	if signUpReq.Password != signUpReq.ConfirmPassword {
		context.String(http.StatusOK, "两次输入的密码不一致")
		return
	}

	isPassword, err := u.passwordRegExp.MatchString(signUpReq.Password)
	if err != nil {
		context.String(http.StatusOK, "系统错误")
		return
	}
	if !isPassword {
		context.String(http.StatusOK, "密码必须包含字母、数字、特殊字符，并且长度不能小于8位")
		return
	}

	err = u.svc.SignUp(context, domain.User{
		Email:    signUpReq.Email,
		Password: signUpReq.Password,
	})
	switch {
	case err == nil:
		context.String(http.StatusOK, "注册成功")
	case errors.Is(err, service.ErrDuplicateEmail):
		context.String(http.StatusOK, "邮箱冲突，请换一个")
	default:
		context.String(http.StatusOK, "系统错误:%v", err)
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

func (u *UserHandler) LoginJWT(context *gin.Context) {
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
		err = u.SetLoginToken(context, usr.Id)
		if err != nil {
			context.JSON(http.StatusOK, Result{
				Code: 4,
				Msg:  "系统错误",
			})
		}
		context.String(http.StatusOK, "登录成功")
	case errors.Is(err, service.ErrInvalidUserOrPassword):
		context.String(http.StatusOK, "用户名或者密码不对")
	default:
		context.String(http.StatusOK, "系统错误")
	}

}

func (u *UserHandler) Edit(context *gin.Context) {
	type Req struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	data, ok := context.Get("user-id")
	if !ok {
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userId, ok := data.(int64)
	if !ok {
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var req Req
	err := context.Bind(&req)
	if err != nil {
		context.String(http.StatusOK, "系统错误")
		return
	}
	if runes := []rune(req.Nickname); len(runes) > 30 {
		context.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "昵称的长度不超过30个字符",
		})
		return
	}
	if runes := []rune(req.AboutMe); len(runes) > 300 {
		context.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "简介的长度不超过300个字符",
		})
		return
	}
	legalBirthday, err := u.birthdayRegExp.MatchString(req.Birthday)
	if err != nil {
		context.String(http.StatusOK, "系统错误")
		return
	}
	if !legalBirthday {
		context.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "生日的输入格式不合法",
		})
		return
	}
	err = u.svc.Edit(context, userId, req.Nickname, req.Birthday, req.AboutMe)
	if err != nil {
		context.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  err,
		})
	}
	context.JSON(http.StatusOK, gin.H{
		"code": 0,
	})
}

func (u *UserHandler) Profile(context *gin.Context) {
	data, ok := context.Get("user-id")
	if !ok {
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	userId, ok := data.(int64)
	if !ok {
		context.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err := u.svc.Profile(context, userId)
	switch {
	case err == nil:
	case errors.Is(err, service.ErrNoUserProfile):
		context.String(http.StatusOK, "系统错误")
	default:
		context.String(http.StatusOK, "系统错误")
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
	context.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (u *UserHandler) LogOut(context *gin.Context) {
	err := u.ClearToken(context)
	if err != nil {
		context.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	context.JSON(http.StatusOK, Result{
		Msg: "退出登录成功",
	})
}
