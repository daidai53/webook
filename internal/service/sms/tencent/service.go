// Copyright@daidai53 2023
package tencent

import (
	"context"
	"errors"
	"fmt"
	"github.com/daidai53/webook/pkg/limiter"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	client   *sms.Client
	appId    *string
	signName *string
	limiter  limiter.Limiter
}

func NewService(client *sms.Client, appId string, signName string, l limiter.Limiter) *Service {
	return &Service{
		client:   client,
		appId:    ekit.ToPtr(appId),
		signName: ekit.ToPtr(signName),
		limiter:  l,
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, "tencent-sms-service")
	if err != nil {
		return err
	}
	if limited {
		return errors.New("触发了限流")
	}
	request := sms.NewSendSmsRequest()
	request.SetContext(ctx)
	request.SmsSdkAppId = s.appId
	request.SignName = s.signName
	request.TemplateId = ekit.ToPtr(tplId)
	request.TemplateParamSet = s.toPtrSlice(args)
	request.PhoneNumberSet = s.toPtrSlice(numbers)
	response, err := s.client.SendSms(request)
	// 处理异常
	if err != nil {
		fmt.Printf("An API error has returned: %s", err)
		return err
	}

	for _, statusPtr := range response.Response.SendStatusSet {
		if statusPtr == nil {
			continue
		}
		status := *statusPtr
		if status.Code == nil || *(status.Code) != "Ok" {
			return errors.New("短信发送失败")
		}
	}
	return nil
}

func (s *Service) toPtrSlice(data []string) []*string {
	return slice.Map[string, *string](data, func(idx int, src string) *string {
		return ekit.ToPtr(src)
	})
}
