// Copyright@daidai53 2024
package fixer

import (
	"context"
	"github.com/daidai53/webook/pkg/migrator"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OverrideFixer[T migrator.Entity] struct {
	base, target *gorm.DB

	columns []string
}

func NewOverrideFixer[T migrator.Entity](base *gorm.DB, target *gorm.DB) (*OverrideFixer[T], error) {
	rows, err := base.Order("id").Rows()
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	return &OverrideFixer[T]{
		base:    base,
		target:  target,
		columns: columns,
	}, err
}

func (f *OverrideFixer[T]) Fix(ctx context.Context, id int64) error {
	var t T
	err := f.base.Where("id = ?", id).First(&t).Error
	switch err {
	case gorm.ErrRecordNotFound:
		return f.target.WithContext(ctx).Model(&t).Delete("id =?", id).Error
	case nil:
		return f.target.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(f.columns),
		}).
			Create(&t).Error
	default:
		return err
	}
}
