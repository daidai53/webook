// Copyright@daidai53 2023
package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/pkg/logger"
	"net/http"
	"net/url"
)

type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WeChatInfo, error)
}

var redirectURL = url.PathEscape(`https://meoying.com/oauth2/wechat/callback`)

type service struct {
	appID     string
	appSecret string
	client    *http.Client
	l         logger.LoggerV1
}

func NewService(appId string, secret string, l logger.LoggerV1) Service {
	return &service{
		appID:     appId,
		appSecret: secret,
		client:    http.DefaultClient,
		l:         l,
	}
}

func (s *service) AuthURL(ctx context.Context, state string) (string, error) {
	const authURLPattern = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect"
	return fmt.Sprintf(authURLPattern, s.appID, redirectURL, state), nil
}

func (s *service) VerifyCode(ctx context.Context, code string) (domain.WeChatInfo, error) {
	var accessTokenURL = fmt.Sprintf(`https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code`,
		s.appID, s.appSecret, code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, accessTokenURL, nil)
	if err != nil {
		return domain.WeChatInfo{}, err
	}
	httpResp, err := s.client.Do(req)
	if err != nil {
		return domain.WeChatInfo{}, err
	}

	var res Result
	err = json.NewDecoder(httpResp.Body).Decode(&res)
	if err != nil {
		return domain.WeChatInfo{}, err
	}
	if res.ErrorCode != 0 {
		return domain.WeChatInfo{}, fmt.Errorf("调用微信接口失败 errorcode %d errmsg %s",
			res.ErrorCode, res.ErrorMsg)
	}
	return domain.WeChatInfo{
		UnionId: res.UnionId,
		OpenId:  res.OpenId,
	}, nil
}

type Result struct {
	AccessToken  string `json:"accessToken"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionId      string `json:"unionid"`

	ErrorCode int    `json:"errcode"`
	ErrorMsg  string `json:"errormsg"`
}
