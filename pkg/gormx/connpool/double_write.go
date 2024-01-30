// Copyright@daidai53 2024
package connpool

import (
	"context"
	"database/sql"
	"errors"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
)

type DoubleWritePool struct {
	src     gorm.ConnPool
	dst     gorm.ConnPool
	pattern *atomicx.Value[string]
	l       logger.LoggerV1
}

func NewDoubleWritePool(src *gorm.DB, dst *gorm.DB, l logger.LoggerV1) *DoubleWritePool {
	return &DoubleWritePool{
		src:     src.ConnPool,
		dst:     dst.ConnPool,
		l:       l,
		pattern: atomicx.NewValueOf(PatternSrcOnly),
	}
}

var ErrUnknownPattern = errors.New("未知的双写模式")

func (d *DoubleWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWritePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWriteTx{src: src, l: d.l, pattern: pattern}, err
	case PatternDstOnly:
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWriteTx{dst: dst, l: d.l, pattern: pattern}, err
	case PatternSrcFirst:
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			d.l.Error("双写目标表开启事务失败",
				logger.Error(err))
		}
		return &DoubleWriteTx{src: src, dst: dst, l: d.l, pattern: pattern}, err
	case PatternDstFirst:
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			d.l.Error("双写源表开启事务失败",
				logger.Error(err))
		}
		return &DoubleWriteTx{src: src, dst: dst, l: d.l, pattern: pattern}, err
	default:
		return nil, ErrUnknownPattern
	}
}

func (d *DoubleWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err == nil {
			_, err1 := d.dst.ExecContext(ctx, query, args...)
			if err1 != nil {
				d.l.Error("双写写入Dst失败",
					logger.Error(err1),
					logger.String("sql", query))
			}
		}
		return res, err
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err == nil {
			_, err1 := d.src.ExecContext(ctx, query, args...)
			if err1 != nil {
				d.l.Error("双写写入Src失败",
					logger.Error(err1),
					logger.String("sql", query))
			}
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	default:
		return nil, ErrUnknownPattern
	}
}

func (d *DoubleWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, ErrUnknownPattern
	}
}

func (d *DoubleWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		panic(ErrUnknownPattern)
	}
}

type DoubleWriteTx struct {
	src     *sql.Tx
	dst     *sql.Tx
	pattern string
	l       logger.LoggerV1
}

func (d *DoubleWriteTx) Commit() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Commit()
	case PatternDstOnly:
		return d.dst.Commit()
	case PatternSrcFirst:
		err := d.src.Commit()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err = d.dst.Commit()
			if err != nil {
				d.l.Error("目标表提交事务失败")
			}
		}
	case PatternDstFirst:
		err := d.dst.Commit()
		if err != nil {
			return err
		}
		if d.src != nil {
			err = d.src.Commit()
			if err != nil {
				d.l.Error("源表提交事务失败")
			}
		}
	default:
		return ErrUnknownPattern
	}
	return nil
}

func (d *DoubleWriteTx) Rollback() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Rollback()
	case PatternDstOnly:
		return d.dst.Rollback()
	case PatternSrcFirst:
		err := d.src.Rollback()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err = d.dst.Rollback()
			if err != nil {
				d.l.Error("目标表提交事务失败")
			}
		}
	case PatternDstFirst:
		err := d.dst.Rollback()
		if err != nil {
			return err
		}
		if d.src != nil {
			err = d.src.Rollback()
			if err != nil {
				d.l.Error("源表提交事务失败")
			}
		}
	default:
		return ErrUnknownPattern
	}
	return nil
}

func (d *DoubleWriteTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DoubleWriteTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err == nil && d.dst != nil {
			_, err1 := d.dst.ExecContext(ctx, query, args...)
			if err1 != nil {
				d.l.Error("双写写入Dst失败",
					logger.Error(err1),
					logger.String("sql", query))
			}
		}
		return res, err
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err == nil && d.src != nil {
			_, err1 := d.src.ExecContext(ctx, query, args...)
			if err1 != nil {
				d.l.Error("双写写入Src失败",
					logger.Error(err1),
					logger.String("sql", query))
			}
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	default:
		return nil, ErrUnknownPattern
	}
}

func (d *DoubleWriteTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, ErrUnknownPattern
	}
}

func (d *DoubleWriteTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		panic(ErrUnknownPattern)
	}
}

const (
	PatternSrcOnly  = "src_only"
	PatternSrcFirst = "src_first"
	PatternDstFirst = "dst_first"
	PatternDstOnly  = "dst_only"
)
