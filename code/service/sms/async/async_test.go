// Copyright@daidai53 2023
package async

import (
	"context"
	"errors"
	"github.com/daidai53/webook/code/service/sms"
	smsmocks "github.com/daidai53/webook/code/service/sms/mocks"
	"github.com/daidai53/webook/internal/integration/startup"
	"github.com/daidai53/webook/pkg/limiter"
	limitermocks "github.com/daidai53/webook/pkg/limiter/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"sync"
	"testing"
	"time"
)

func TestSmsService_Send(t *testing.T) {
	sdb := startup.InitRedis()
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller, wg *sync.WaitGroup) (sms.Service, limiter.Limiter)

		changer func(svc *SmsService)
		wg      sync.WaitGroup

		ctx            context.Context
		tplId          string
		args           []string
		numbers        []string
		errors         []error
		key            string
		sleepTimeAfter time.Duration

		reqCnt  int64
		rspCnt  int64
		wantErr error
	}{
		{
			name: "运营商返回错误码在黑名单中，直接接入失败",
			mock: func(ctrl *gomock.Controller, wg *sync.WaitGroup) (sms.Service, limiter.Limiter) {
				smsSvc := smsmocks.NewMockService(ctrl)
				limiterSvc := limitermocks.NewMockLimiter(ctrl)
				limiterSvc.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				smsSvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("error1"))
				return smsSvc, limiterSvc
			},

			ctx:     context.Background(),
			tplId:   "template1",
			args:    []string{"", ""},
			numbers: []string{"123456789"},
			errors:  []error{errors.New("error1")},
			key:     "async-sms-service",

			reqCnt:  0,
			rspCnt:  0,
			wantErr: errors.New("error1"),
		},
		{
			name: "发送成功",
			mock: func(ctrl *gomock.Controller, wg *sync.WaitGroup) (sms.Service, limiter.Limiter) {
				smsSvc := smsmocks.NewMockService(ctrl)
				limiterSvc := limitermocks.NewMockLimiter(ctrl)
				limiterSvc.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				smsSvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				return smsSvc, limiterSvc
			},

			ctx:     context.Background(),
			tplId:   "template1",
			args:    []string{"", ""},
			numbers: []string{"123456789"},
			errors:  []error{errors.New("error1")},
			key:     "async-sms-service",

			reqCnt:  1,
			rspCnt:  1,
			wantErr: nil,
		},
		{
			name: "发送失败，判断运营商崩溃,每秒重发一次，第一次重发成功",
			mock: func(ctrl *gomock.Controller, wg *sync.WaitGroup) (sms.Service, limiter.Limiter) {
				smsSvc := smsmocks.NewMockService(ctrl)
				limiterSvc := limitermocks.NewMockLimiter(ctrl)
				wg.Add(1)
				smsSvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_, _, _, _ any) {
						wg.Done()
					}).
					Return(nil)
				return smsSvc, limiterSvc
			},
			changer: func(svc *SmsService) {
				svc.successRateCrash.reqCnt = 20
				svc.successRateCrash.rspCnt = 5
				svc.retryInterval = time.Second
			},

			ctx:     context.Background(),
			tplId:   "template1",
			args:    []string{"", ""},
			numbers: []string{"123456789"},
			errors:  []error{errors.New("error1")},
			key:     "async-sms-service",

			reqCnt:  21,
			rspCnt:  6,
			wantErr: nil,
		},
		{
			name: "发送失败，判断客户端限流,每秒重发一次，第一次重发成功",
			mock: func(ctrl *gomock.Controller, wg *sync.WaitGroup) (sms.Service, limiter.Limiter) {
				smsSvc := smsmocks.NewMockService(ctrl)
				limiterSvc := limitermocks.NewMockLimiter(ctrl)
				wg.Add(1)
				limiterSvc.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil)
				smsSvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_, _, _, _ any) {
						wg.Done()
					}).
					Return(nil)
				return smsSvc, limiterSvc
			},

			changer: func(svc *SmsService) {
				svc.retryInterval = time.Second
			},

			ctx:     context.Background(),
			tplId:   "template1",
			args:    []string{"", ""},
			numbers: []string{"123456789"},
			errors:  []error{errors.New("error1")},
			key:     "async-sms-service",

			reqCnt:  1,
			rspCnt:  1,
			wantErr: nil,
		},
		{
			name: "发送失败，服务返回错误，直接返回失败",
			mock: func(ctrl *gomock.Controller, wg *sync.WaitGroup) (sms.Service, limiter.Limiter) {
				smsSvc := smsmocks.NewMockService(ctrl)
				limiterSvc := limitermocks.NewMockLimiter(ctrl)
				limiterSvc.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				smsSvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("error2"))
				return smsSvc, limiterSvc
			},

			changer: func(svc *SmsService) {
				svc.retryInterval = time.Second
			},

			ctx:     context.Background(),
			tplId:   "template1",
			args:    []string{"", ""},
			numbers: []string{"123456789"},
			errors:  []error{errors.New("error1")},
			key:     "async-sms-service",

			reqCnt:  1,
			rspCnt:  0,
			wantErr: errors.New("error2"),
		},
		{
			name: "发送成功,15s后统计被清除",
			mock: func(ctrl *gomock.Controller, wg *sync.WaitGroup) (sms.Service, limiter.Limiter) {
				smsSvc := smsmocks.NewMockService(ctrl)
				limiterSvc := limitermocks.NewMockLimiter(ctrl)
				limiterSvc.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				smsSvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
				return smsSvc, limiterSvc
			},

			ctx:            context.Background(),
			tplId:          "template1",
			args:           []string{"", ""},
			numbers:        []string{"123456789"},
			errors:         []error{errors.New("error1")},
			key:            "async-sms-service",
			sleepTimeAfter: time.Second * 15,

			reqCnt:  0,
			rspCnt:  0,
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			smsSvc, limiter := tc.mock(ctrl, &tc.wg)
			asyncSmsSvc := NewAsyncSMSService(smsSvc, sdb, limiter, tc.key, 0, tc.errors...)
			defer func() {
				asyncSmsSvc.successRateCrash.reqCnt = 0
				asyncSmsSvc.successRateCrash.rspCnt = 0
			}()
			if tc.changer != nil {
				tc.changer(asyncSmsSvc)
			}
			err := asyncSmsSvc.Send(tc.ctx, tc.tplId, tc.args, tc.numbers...)
			assert.Equal(t, tc.wantErr, err)
			tc.wg.Wait()
			time.Sleep(tc.sleepTimeAfter)
			assert.Equal(t, tc.reqCnt, asyncSmsSvc.successRateCrash.reqCnt)
			assert.Equal(t, tc.rspCnt, asyncSmsSvc.successRateCrash.rspCnt)
		})
	}
}

// 测试100个Goroutine并发发送，每个Goroutine发送10000条消息，共1000000条，全部成功，测试消息统计是否准确及处理时间
func Test_AsyncCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sdb := startup.InitRedis()

	smsSvc := smsmocks.NewMockService(ctrl)
	limiterSvc := limitermocks.NewMockLimiter(ctrl)
	limiterSvc.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	smsSvc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	asyncSmsSvc := NewAsyncSMSService(smsSvc, sdb, limiterSvc, "async-sms-service", time.Hour)
	var wg sync.WaitGroup
	wg.Add(1000000)
	startTime := time.Now()
	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 10000; j++ {
				go func() {
					err1 := asyncSmsSvc.Send(context.Background(), "template1", []string{})
					wg.Done()
					assert.NoError(t, err1)
				}()
			}
		}()
	}
	wg.Wait()
	endTime := time.Now()
	secs := endTime.Sub(startTime).Seconds()
	proc := 1000000 / secs
	t.Logf("process send callings: %0.2f/s, total cost %f seconds.", proc, secs)
	assert.Equal(t, int64(1000000), asyncSmsSvc.successRateCrash.reqCnt)
	assert.Equal(t, int64(1000000), asyncSmsSvc.successRateCrash.rspCnt)
}
