// Copyright@daidai53 2023
package faileover

import (
	"context"
	"errors"
	"github.com/daidai53/webook/code/service/sms"
	"sync/atomic"
)

type FailOverSMSService struct {
	svcs []sms.Service

	// v1的字段
	// 当前服务商下标
	idx uint64
}

func NewFailOverSMSService(svcs []sms.Service) *FailOverSMSService {
	return &FailOverSMSService{
		svcs: svcs,
	}
}

func (f *FailOverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tplId, args, numbers...)
		if err == nil {
			return nil
		}
		// log
	}
	return errors.New("轮询了所有服务商，但是发送都失败了")
}

// 按起始下标轮询
// 并且出错也轮询
func (f *FailOverSMSService) SendV1(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	for i := idx; i < idx+length; i++ {
		svc := f.svcs[i%length]
		err := svc.Send(ctx, tplId, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.Canceled, context.DeadlineExceeded:

		}
		// log
	}
	return errors.New("轮询了所有服务商，但是发送都失败了")
}
