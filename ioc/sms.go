// Copyright@daidai53 2023
package ioc

import (
	"github.com/daidai53/webook/code/service/sms"
	"github.com/daidai53/webook/code/service/sms/localsms"
	"github.com/daidai53/webook/code/service/sms/opentelemetry"
	"github.com/daidai53/webook/code/service/sms/tencent"
	"github.com/daidai53/webook/pkg/limiter"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"os"
	"time"
)

func InitSmsService() sms.Service {
	return opentelemetry.NewDecorator(localsms.NewService())
	//return initTencentSmsService()
}

func initTencentSmsService() sms.Service {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		return nil
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")

	c, err := tsms.NewClient(common.NewCredential(secretId, secretKey),
		"ap-nanjing",
		profile.NewClientProfile())
	if err != nil {
		return nil
	}

	return tencent.NewService(c, "1400864331", "妙影科技",
		limiter.NewRedisSlidingWindowLimiter(nil, time.Second, 100))
}
