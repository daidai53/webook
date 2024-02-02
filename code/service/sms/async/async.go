// Copyright@daidai53 2023
package async

import (
	"context"
	"encoding/json"
	"github.com/daidai53/webook/code/service/sms"
	"github.com/daidai53/webook/pkg/limiter"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"sync/atomic"
	"time"
)

const (
	defaultRetryTimes    = 8
	defaultRetryInterval = 10 * time.Second
)

var gs successRateCrash

// 用一段时间的成功率判断服务商是否崩溃
type successRateCrash struct {
	reqCnt int64
	rspCnt int64
}

func (s *successRateCrash) crash() bool {
	if s == nil {
		return false
	}

	req := atomic.LoadInt64(&s.reqCnt)
	rsp := atomic.LoadInt64(&s.rspCnt)
	// 本周期内请求小于10时，容易有过大误差，认为没有崩溃，即使真的崩溃尝试10次也没什么
	if req < 10 {
		return false
	}
	succRate := float64(rsp) / float64(req)
	if succRate < 0.5 {
		return true
	}
	return false
}

type SmsService struct {
	svc         sms.Service
	redisClient redis.Cmdable
	limiter     limiter.Limiter
	key         string

	retryTimes    uint32
	retryInterval time.Duration

	blackList        map[string]struct{}
	successRateCrash *successRateCrash
}

func NewAsyncSMSService(s sms.Service, cmd redis.Cmdable, limiter limiter.Limiter, key string, clearTimeArg time.Duration, errs ...error) *SmsService {
	res := &SmsService{
		svc:              s,
		key:              key,
		redisClient:      cmd,
		limiter:          limiter,
		retryTimes:       defaultRetryTimes,
		retryInterval:    defaultRetryInterval,
		blackList:        make(map[string]struct{}),
		successRateCrash: &gs,
	}
	for _, err := range errs {
		res.blackList[err.Error()] = struct{}{}
	}
	clearTime := 15 * time.Second
	if clearTimeArg > 0 {
		clearTime = clearTimeArg
	}
	go func() {
		// 每15s周期清空successRateCrash，只保存一周期内的发送数据
		for {
			time.Sleep(clearTime)
			atomic.StoreInt64(&res.successRateCrash.reqCnt, 0)
			atomic.StoreInt64(&res.successRateCrash.rspCnt, 0)
		}
	}()
	return res
}

func (a *SmsService) WithRetryTimes(t uint32) *SmsService {
	a.retryTimes = t
	return a
}

func (a *SmsService) WithRetryInterval(i time.Duration) *SmsService {
	a.retryInterval = i
	return a
}

func (a *SmsService) limit(ctx context.Context, key string) bool {
	limited, err := a.limiter.Limit(ctx, key)
	if err != nil {
		// log
		return false
	}
	return limited
}

func (a *SmsService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	if a.successRateCrash.crash() || a.limit(ctx, a.key) {
		// 本周期成功率较低，判断已崩溃，或触发限流，启动异步Resend并返回成功，此处不统计，待异步Resend中统计
		id := uuid.New().String()
		msg, err := json.Marshal(&Message{
			TplId:   tplId,
			Args:    args,
			Numbers: numbers,
		})
		if err != nil {
			return errors.New("系统错误")
		}
		err = a.redisClient.Set(ctx, id, string(msg), a.retryInterval*time.Duration(a.retryTimes)+time.Second*10).Err()
		if err != nil {
			// log
			return err
		}
		go a.ReSend(a.retryInterval, a.retryTimes, ctx, id)
		return nil
	}
	err := a.svc.Send(ctx, tplId, args, numbers...)
	switch {
	case a.BlackList(err):
		// 直接返回错误，不计入成功率统计
		//return errors.Wrap(err, "SmsService system error")
		return err
	case err == nil:
		// 发送成功，直接将请求和响应数都加一
		atomic.AddInt64(&a.successRateCrash.reqCnt, 1)
		atomic.AddInt64(&a.successRateCrash.rspCnt, 1)
		return nil
	default:
		// log
		// 既非限流，也非判断服务商崩溃，返回error并将请求数记录
		atomic.AddInt64(&a.successRateCrash.reqCnt, 1)
		return err
	}
}

func (a *SmsService) BlackList(e error) bool {
	if e == nil {
		return false
	}
	_, ok := a.blackList[e.Error()]
	if !ok {
		return false
	}
	return true
}

func (a *SmsService) ReSend(sleepTime time.Duration, remainTimes uint32, ctx context.Context, id string) {
	// 先判断是否有剩余次数，没有的话将请求数加一，返回
	if remainTimes == 0 {
		atomic.AddInt64(&a.successRateCrash.reqCnt, 1)
		return
	}
	time.Sleep(sleepTime)
	stored, err := a.redisClient.Get(ctx, id).Result()
	if err != nil {
		// log
		return
	}
	var msg Message
	err = json.Unmarshal([]byte(stored), &msg)
	if err != nil {
		// log
		return
	}
	err = a.svc.Send(ctx, msg.TplId, msg.Args, msg.Numbers...)
	switch {
	case a.BlackList(err):
	case err == nil:
		// 发送成功，直接将请求和响应数都加一
		atomic.AddInt64(&a.successRateCrash.reqCnt, 1)
		atomic.AddInt64(&a.successRateCrash.rspCnt, 1)
	default:
		// 其他类型的错误，启动异步Resend并返回成功，此处不统计，待异步Resend中统计
		go a.ReSend(sleepTime, remainTimes-1, ctx, id)
	}
}

type Message struct {
	TplId   string   `json:"tplId"`
	Args    []string `json:"args"`
	Numbers []string `json:"numbers"`
}
