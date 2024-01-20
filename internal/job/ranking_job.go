// Copyright@daidai53 2024
package job

import (
	"context"
	"github.com/daidai53/webook/internal/service"
	"github.com/daidai53/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"golang.org/x/exp/rand"
	"sync"
	"sync/atomic"
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

	// 作业
	load int32
	low  int32
	high int32
}

func NewRankingJob(svc service.RankingService, timeout time.Duration, l logger.LoggerV1, client *rlock.Client) *RankingJob {
	res := &RankingJob{
		svc:       svc,
		timeout:   timeout,
		key:       "job:ranking",
		l:         l,
		client:    client,
		localLock: &sync.Mutex{},
		low:       30,
		high:      80,
	}

	go func() {
		ticker := time.NewTicker(time.Minute)
		for {
			select {
			case <-ticker.C:
				rand.Seed(uint64(time.Now().Unix()))
				newLoad := rand.Int31n(101)
				atomic.StoreInt32(&res.load, newLoad)
			}
		}
	}()
	return res
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
		if !r.needGetLock() {
			r.localLock.Unlock()
			return nil
		}

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

	if r.needReturnLock() {
		r.lock.Unlock(ctx)
		r.lock = nil
		r.localLock.Unlock()
		return nil
	}
	ctx, cancel = context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	err := r.svc.TopN(ctx)
	r.localLock.Unlock()
	return err
}

// 获取分布式锁之前，根据负载判断是否要获取
func (r *RankingJob) needGetLock() bool {
	// 获得当前负载
	load := atomic.LoadInt32(&r.load)
	// 低负载，直接尝试获取分布式锁
	if load < r.low {
		return true
	}
	// 高负载，直接不获取分布式锁
	if load > r.high {
		r.l.Error("负载过高，本节点不参与调度",
			logger.Int32("load", load))
		return false
	}

	// 中负载，概率尝试获取分布式锁
	rand.Seed(uint64(time.Now().Unix()))
	num := rand.Int31n(101)
	if num > load {
		return true
	}
	r.l.Error("负载过高，本节点不参与调度",
		logger.Int32("load", load))
	return false
}

// 在已有分布式锁调度任务时，或者自动刷新租期时，根据负载判断是否要立即结束本次调度
func (r *RankingJob) needReturnLock() bool {
	// 获得当前负载
	load := atomic.LoadInt32(&r.load)

	// 低负载，不释放分布式锁
	if load < r.low {
		return false
	}
	// 高负载，直接释放分布式锁
	if load > r.high {
		r.l.Error("负载过高，立即释放分布式锁",
			logger.Int32("load", load))
		return true
	}

	// 中负载，概率尝试释放分布式锁
	rand.Seed(uint64(time.Now().Unix()))
	num := rand.Int31n(101)
	if num > load {
		return false
	}
	r.l.Error("负载过高，立即释放分布式锁",
		logger.Int32("load", load))
	return true
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
