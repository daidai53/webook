// Copyright@daidai53 2024
package job

import (
	"context"
	"github.com/daidai53/webook/internal/service"
	"github.com/daidai53/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
)

type RankingJob struct {
	svc     service.RankingService
	timeout time.Duration
	client  *rlock.Client
	key     string
	l       logger.LoggerV1

	localLock *sync.Mutex
	lock      *rlock.Lock
}

func NewRankingJob(svc service.RankingService, timeout time.Duration, l logger.LoggerV1, client *rlock.Client) *RankingJob {
	return &RankingJob{
		svc:       svc,
		timeout:   timeout,
		key:       "job:ranking",
		l:         l,
		client:    client,
		localLock: &sync.Mutex{},
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	r.localLock.Lock()
	// 抢分布式锁
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()
	if r.lock == nil {
		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      3,
		}, time.Second)
		if err != nil {
			r.localLock.Unlock()
			r.l.Warn("获取分布式锁失败", logger.Error(err))
			return nil
		}
		r.lock = lock
		go func() {
			// 并不是非得一半就续约
			er := lock.AutoRefresh(r.timeout/2, r.timeout)
			if er != nil {
				// 续约失败了
				r.localLock.Lock()
				r.lock = nil
				r.localLock.Unlock()
			}
		}()
	}

	ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	err := r.svc.TopN(ctx)
	r.localLock.Unlock()
	return err
}

//func (r *RankingJob) Run() error {
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
//	defer cancel()
//	lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
//		Interval: time.Millisecond * 100,
//		Max:      3,
//	}, time.Second)
//
//	if err != nil {
//		return err
//	}
//	defer func() {
//		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//		defer cancel()
//		er := lock.Unlock(ctx)
//		if er != nil {
//			r.l.Error("ranking job释放分布式锁失败", logger.Error(er))
//		}
//	}()
//	ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
//	defer cancel()
//	return r.svc.TopN(ctx)
//}
