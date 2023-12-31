// Copyright@daidai53 2023
package ioc

import (
	"github.com/daidai53/webook/internal/service/oauth2/wechat"
	"github.com/daidai53/webook/pkg/logger"
	"os"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	appID, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("找不到环境变量WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("找不到环境变量WECHAT_APP_SECRET")
	}
	return wechat.NewService(appID, appSecret, l)
}
