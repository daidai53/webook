// Copyright@daidai53 2024
package job

import (
	"context"
	"fmt"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/service"
	"github.com/daidai53/webook/pkg/logger"
	"golang.org/x/sync/semaphore"
	"time"
)

// Executor 任务执行器
type Executor interface {
	Name() string
	// Exec ctx 这个是全局控制，Executor的实现者要正确处理ctx超时或取消
	Exec(ctx context.Context, j domain.Job) error
}

type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{
		funcs: map[string]func(context.Context, domain.Job) error{},
	}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未注册本地方法%s", j.Name)
	}
	return fn(ctx, j)
}

func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(context.Context, domain.Job) error) {
	l.funcs[name] = fn
}

type Scheduler struct {
	dbTimeout time.Duration

	svc       service.CronJobService
	executors map[string]Executor
	l         logger.LoggerV1

	limiter *semaphore.Weighted
}

func NewScheduler(svc service.CronJobService, l logger.LoggerV1) *Scheduler {
	return &Scheduler{
		svc:       svc,
		l:         l,
		dbTimeout: time.Second,
		limiter:   semaphore.NewWeighted(100),
		executors: map[string]Executor{},
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.executors[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		// 上下文出问题，终止调度器的循环
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// 每次循环，先拿令牌，防止一秒钟内抢到太多的任务，导致节点处理不过来
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			// limiter也出问题（不是没拿到令牌，而是limiter有问题），终止调度
			return err
		}

		// 本次循环中，给db的context
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 总之就是没抢到job
			// 最简单的就是直接下一轮
			continue
		}

		// 肯定要调度执行抢到的job
		exec, ok := s.executors[j.Executor]
		if !ok {
			// 可以直接中断，也可以下一轮
			s.l.Error("找不到执行器",
				logger.Int64("jid", j.Id),
				logger.String("executor", j.Executor))
			continue
		}

		go func() {
			defer func() {
				// 调度完释放掉
				s.limiter.Release(1)
				j.CancelFunc()
			}()
			err1 := exec.Exec(ctx, j)
			if err1 != nil {
				s.l.Error("执行任务失败",
					logger.Int64("jid", j.Id),
					logger.String("executor", j.Executor))
				return
			}

			err1 = s.svc.ResetNextTime(ctx, j)
			if err1 != nil {
				s.l.Error("重置下次执行任务使时间失败",
					logger.Int64("jid", j.Id),
					logger.String("executor", j.Executor))
			}
		}()

	}
}
