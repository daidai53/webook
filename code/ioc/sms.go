// Copyright@daidai53 2024
package ioc

import (
	"github.com/daidai53/webook/code/service/sms"
	"github.com/daidai53/webook/code/service/sms/localsms"
)

func InitSmsService() sms.Service {
	return localsms.NewService()
	//return initTencentSmsService()
}
