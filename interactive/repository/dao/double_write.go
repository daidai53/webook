// Copyright@daidai53 2024
package dao

import (
	"context"
	"errors"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/ecodeclub/ekit/syncx/atomicx"
)

type DoubleWriteDAO struct {
	src     InteractiveDAO
	dst     InteractiveDAO
	pattern *atomicx.Value[string]
	l       logger.LoggerV1
}

var ErrUnknownPattern = errors.New("未知的双写模式")

func (d *DoubleWriteDAO) UpdatePattern(pattern string) {
	d.pattern.Store(pattern)
}

func (d *DoubleWriteDAO) IncrReadCnt(ctx context.Context, biz string, id int64) error {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.src.IncrReadCnt(ctx, biz, id)
	case PatternSrcFirst:
		err := d.src.IncrReadCnt(ctx, biz, id)
		if err != nil {
			return err
		}
		err = d.dst.IncrReadCnt(ctx, biz, id)
		if err != nil {
			// 双写阶段，认为src成功了就算业务成功，写dst失败不返回错误
			d.l.Error("双写写入Dst失败", logger.Error(err),
				logger.String("biz", biz), logger.Int64("biz_id", id))
		}
		return nil
	case PatternDstFirst:
		err := d.dst.IncrReadCnt(ctx, biz, id)
		if err != nil {
			return err
		}
		err = d.src.IncrReadCnt(ctx, biz, id)
		if err != nil {
			// 双写阶段，认为src成功了就算业务成功，写dst失败不返回错误
			d.l.Error("双写写入src失败", logger.Error(err),
				logger.String("biz", biz), logger.Int64("biz_id", id))
		}
		return nil
	case PatternDstOnly:
		return d.dst.IncrReadCnt(ctx, biz, id)
	default:
		return ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) BatchIncrReadCnt(ctx context.Context, biz []string, id []int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) GetLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserLikeBiz, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) GetCollectInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserCollectionBiz, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.Get(ctx, biz, bizId)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.Get(ctx, biz, bizId)
	default:
		return Interactive{}, ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetAllArticleLikes() ([]Likes, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteDAO) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	//TODO implement me
	panic("implement me")
}

const (
	PatternSrcOnly  = "src_only"
	PatternSrcFirst = "src_first"
	PatternDstFirst = "dst_first"
	PatternDstOnly  = "dst_only"
)
