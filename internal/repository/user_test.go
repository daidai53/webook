// Copyright@daidai53 2023
package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/repository/cache"
	cachemocks "github.com/daidai53/webook/internal/repository/cache/mocks"
	"github.com/daidai53/webook/internal/repository/dao"
	daomocks "github.com/daidai53/webook/internal/repository/dao/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO)

		// 输入
		ctx context.Context
		id  int64

		// 预期输出
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "查找成功，缓存未命中",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(
					domain.User{}, ErrUserNotFound)
				d.EXPECT().FindById(gomock.Any(), int64(123)).Return(
					dao.User{
						Id: int64(123),
						Email: sql.NullString{
							String: "123@qq.com",
							Valid:  true,
						},
						Password: "123456",
						Birthday: "2023-01-01",
						AboutMe:  "自我介绍",
						Phone: sql.NullString{
							String: "123456789",
							Valid:  true,
						},
						CreateTime: 101,
						UpdateTime: 102,
					}, nil)

				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       int64(123),
					Email:    "123@qq.com",
					Password: "123456",
					Birthday: "2023-01-01",
					AboutMe:  "自我介绍",
					Phone:    "123456789",
				}).Return(nil)

				return c, d
			},
			id: 123,

			wantUser: domain.User{
				Id:       int64(123),
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: "2023-01-01",
				AboutMe:  "自我介绍",
				Phone:    "123456789",
			},
			wantErr: nil,
		},
		{
			name: "缓存命中",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(
					domain.User{
						Id:       int64(123),
						Email:    "123@qq.com",
						Password: "123456",
						Birthday: "2023-01-01",
						AboutMe:  "自我介绍",
						Phone:    "123456789",
					}, nil)
				return c, d
			},
			id: 123,

			wantUser: domain.User{
				Id:       int64(123),
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: "2023-01-01",
				AboutMe:  "自我介绍",
				Phone:    "123456789",
			},
			wantErr: nil,
		},
		{
			name: "缓存未命中，数据库也未找到",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(
					domain.User{}, ErrUserNotFound)
				d.EXPECT().FindById(gomock.Any(), int64(123)).Return(
					dao.User{}, dao.ErrRecordNotFound)

				return c, d
			},
			id: 123,

			wantUser: domain.User{},
			wantErr:  dao.ErrRecordNotFound,
		},
		{
			name: "回写缓存失败",
			mock: func(ctrl *gomock.Controller) (cache.UserCache, dao.UserDAO) {
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				c.EXPECT().Get(gomock.Any(), int64(123)).Return(
					domain.User{}, ErrUserNotFound)
				d.EXPECT().FindById(gomock.Any(), int64(123)).Return(
					dao.User{
						Id: int64(123),
						Email: sql.NullString{
							String: "123@qq.com",
							Valid:  true,
						},
						Password: "123456",
						Birthday: "2023-01-01",
						AboutMe:  "自我介绍",
						Phone: sql.NullString{
							String: "123456789",
							Valid:  true,
						},
						CreateTime: 101,
						UpdateTime: 102,
					}, nil)

				c.EXPECT().Set(gomock.Any(), domain.User{
					Id:       int64(123),
					Email:    "123@qq.com",
					Password: "123456",
					Birthday: "2023-01-01",
					AboutMe:  "自我介绍",
					Phone:    "123456789",
				}).Return(errors.New("redis error"))

				return c, d
			},
			id: 123,

			wantUser: domain.User{
				Id:       int64(123),
				Email:    "123@qq.com",
				Password: "123456",
				Birthday: "2023-01-01",
				AboutMe:  "自我介绍",
				Phone:    "123456789",
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userCache, userDao := tc.mock(ctrl)
			userRepo := NewCachedUserRepository(userDao, userCache)
			user, err := userRepo.FindById(tc.ctx, tc.id)
			time.Sleep(10 * time.Millisecond)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
