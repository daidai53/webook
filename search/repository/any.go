// Copyright@daidai53 2024
package repository

import (
	"context"
	"github.com/daidai53/webook/search/repository/dao"
)

type anyRepository struct {
	dao dao.AnyDAO
}

func (a *anyRepository) Input(ctx context.Context, index, docId, data string) error {
	return a.dao.Input(ctx, index, docId, data)
}
