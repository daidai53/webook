// Copyright@daidai53 2024
package repository

import (
	"context"
	"github.com/daidai53/webook/search/domain"
	"github.com/daidai53/webook/search/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type userRepository struct {
	dao dao.UserSearchDAO
}

func (u *userRepository) SyncUser(ctx context.Context, user domain.User) error {
	return u.dao.InputUser(ctx, dao.User{
		Id:       user.Id,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
	})
}

func (u *userRepository) SearchUser(ctx context.Context, keywords []string) ([]domain.User, error) {
	users, err := u.dao.Search(ctx, keywords)
	if err != nil {
		return nil, err
	}
	return slice.Map(users, func(idx int, src dao.User) domain.User {
		return domain.User{
			Id:       src.Id,
			Nickname: src.Nickname,
			Email:    src.Email,
			Phone:    src.Phone,
		}
	}), nil
}
