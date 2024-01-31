// Copyright@daidai53 2024
package validator

import (
	"context"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/daidai53/webook/pkg/migrator"
	"github.com/daidai53/webook/pkg/migrator/events"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
)

type Validator[T migrator.Entity] struct {
	// 数据迁移，肯定有源库和目标库
	base   *gorm.DB
	target *gorm.DB

	l             logger.LoggerV1
	producer      events.Producer
	direction     string
	batchSize     int
	utime         int64
	sleepInterval time.Duration

	fromBase func(ctx context.Context, offset int, limit int) ([]T, error)
}

func NewValidator[T migrator.Entity](base *gorm.DB,
	target *gorm.DB, l logger.LoggerV1, producer events.Producer, direction string) *Validator[T] {
	res := &Validator[T]{base: base, target: target, l: l, producer: producer, direction: direction, batchSize: 100}
	res.fromBase = res.fullFromBase
	return res
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group

	eg.Go(func() error {
		return v.ValidateBaseToTarget(ctx)
	})
	eg.Go(func() error {
		return v.ValidateTargetToBase(ctx)
	})
	return eg.Wait()
}

// 一般认为，Validate会把想校验的数据都校验一遍
func (v *Validator[T]) ValidateBaseToTarget(ctx context.Context) error {
	offset := 0
	for {
		srcs, err := v.fromBase(ctx, offset, v.batchSize)
		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound {
			if v.sleepInterval <= 0 {
				// 没有数据了
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			// 查询出错
			v.l.Error("base -> target 查询base失败",
				logger.Error(err))
			offset += len(srcs)
			continue
		}

		for _, src := range srcs {
			var dst T
			err = v.target.WithContext(ctx).Where("id=?", src.ID()).First(&dst).Error
			switch err {
			case gorm.ErrRecordNotFound:
				// target没有
				// 要丢一条消息到kafka上
				v.Notify(src.ID(), events.InconsistentEventTypeTargetMissing)
			case nil:
				equal := src.CompareTo(dst)
				if !equal {
					// 要丢一条消息到kafka上
					v.Notify(src.ID(), events.InconsistentEventTypeNEQ)
				}
			default:
				// 记录日志，然后继续
				// 做好监控
				v.l.Error("base -> target 查询target失败",
					logger.Error(err),
					logger.Int64("id", src.ID()))
			}
		}
		offset += len(srcs)
	}
}

func (v *Validator[T]) Utime(t int64) *Validator[T] {
	v.utime = t
	return v
}

func (v *Validator[T]) SleepInterval(in time.Duration) *Validator[T] {
	v.sleepInterval = in
	return v
}

func (v *Validator[T]) Full() {
	v.fromBase = v.fullFromBase
}

func (v *Validator[T]) Incr() *Validator[T] {
	v.fromBase = v.incrFromBase
	return v
}

func (v *Validator[T]) fullFromBase(ctx context.Context, offset int, limit int) ([]T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src []T
	err := v.base.WithContext(dbCtx).Order("id").Offset(offset).Limit(limit).Find(&src).Error
	return src, err
}

func (v *Validator[T]) incrFromBase(ctx context.Context, offset int, limit int) ([]T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src []T
	err := v.base.WithContext(dbCtx).Order("u_time").
		Where("u_time>?", v.utime).Limit(limit).
		Offset(offset).Find(&src).Error
	return src, err
}

func (v *Validator[T]) ValidateTargetToBase(ctx context.Context) error {
	offset := 0
	for {
		var ts []T

		err := v.target.WithContext(ctx).
			Select("id").
			Order("id").
			Offset(offset).
			Limit(v.batchSize).
			Find(&ts).Error

		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}
		if err == gorm.ErrRecordNotFound || len(ts) == 0 {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}
		if err != nil {
			v.l.Error("target->base 查询target失败", logger.Error(err))
			offset += len(ts)
			continue
		}

		var srcTs []T
		ids := slice.Map(ts, func(idx int, src T) int64 {
			return src.ID()
		})
		err = v.base.WithContext(ctx).Select("id").Where("id in ?", ids).Find(&srcTs).Error
		if err == gorm.ErrRecordNotFound || len(srcTs) == 0 {
			v.NotifyBaseMissing(ts)
			offset += len(ts)
			continue
		}
		if err != nil {
			v.l.Error("target -> base 查询base失败",
				logger.Error(err))
			offset += len(ts)
			continue
		}
		// 差集，diff里的，就是target有但是base没有的
		diff := slice.DiffSetFunc(ts, srcTs, func(src, dst T) bool {
			return src.ID() == dst.ID()
		})
		v.NotifyBaseMissing(diff)
		if len(ts) < v.batchSize {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
		}
		offset += len(ts)
	}
	return nil
}

func (v *Validator[T]) NotifyBaseMissing(diff []T) {
	for _, val := range diff {
		v.Notify(val.ID(), events.InconsistentEventTypeBaseMissing)
	}
}

func (v *Validator[T]) Notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := v.producer.ProduceIncosistentEvent(ctx, events.InconsistentEvent{
		ID:        id,
		Type:      typ,
		Direction: v.direction,
	})
	if err != nil {
		v.l.Error("发送不一致消息失败",
			logger.Error(err),
			logger.Int64("id", id),
			logger.String("type", typ))
	}
}
