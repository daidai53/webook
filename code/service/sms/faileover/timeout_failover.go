// Copyright@daidai53 2023
package faileover

import (
	"context"
	"github.com/daidai53/webook/code/service/sms"
	"sync/atomic"
)

type TimeoutFailOverSMSService struct {
	svcs []sms.Service

	// 当前正在使用的节点
	idx int32
	// 连续几个超时了
	cnt int32
	// 切换的阈值，只读的
	threshold int32
}

func NewTimeoutFailOverSMSService(svcs []sms.Service, threshold int32) *TimeoutFailOverSMSService {
	return &TimeoutFailOverSMSService{
		svcs:      svcs,
		threshold: threshold,
	}
}

func (t *TimeoutFailOverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)
	// 超过阈值，执行切换
	if cnt >= t.threshold {
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 重置cnt
			atomic.StoreInt32(&t.cnt, 0)
		}
		idx = newIdx
	}

	svc := t.svcs[idx]
	err := svc.Send(ctx, tplId, args, numbers...)
	switch err {
	case nil:
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1)
	default:
		// 灵活考虑
	}
	return err
}
