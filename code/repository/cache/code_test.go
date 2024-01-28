// Copyright@daidai53 2023
package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/daidai53/webook/internal/repository/cache/redismocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRedisCodeCache_Set(t *testing.T) {
	kenGenFunc := func(biz, phone string) string {
		return fmt.Sprintf("phone_code:%s:%s", biz, phone)
	}
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) redis.Cmdable

		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			name: "设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(gomock.NewController(t))
				res := redis.NewCmd(context.Background())
				res.SetErr(nil)
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(gomock.Any(),
					luaSetCode,
					[]string{kenGenFunc("test", "123456789")},
					any("123456"),
				).Return(res)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "123456789",
			code:    "123456",
			wantErr: nil,
		},
		{
			name: "redis返回error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(gomock.NewController(t))
				res := redis.NewCmd(context.Background())
				res.SetErr(errors.New("redis error"))
				cmd.EXPECT().Eval(gomock.Any(),
					luaSetCode,
					[]string{kenGenFunc("test", "123456789")},
					any("123456"),
				).Return(res)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "123456789",
			code:    "123456",
			wantErr: errors.New("redis error"),
		},
		{
			name: "验证码没有过期时间",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(gomock.NewController(t))
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(-2))
				cmd.EXPECT().Eval(gomock.Any(),
					luaSetCode,
					[]string{kenGenFunc("test", "123456789")},
					any("123456"),
				).Return(res)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "123456789",
			code:    "123456",
			wantErr: errors.New("验证码存在，但是没有过期时间"),
		},
		{
			name: "发送太频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := redismocks.NewMockCmdable(gomock.NewController(t))
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(-1))
				cmd.EXPECT().Eval(gomock.Any(),
					luaSetCode,
					[]string{kenGenFunc("test", "123456789")},
					any("123456"),
				).Return(res)
				return cmd
			},
			ctx:     context.Background(),
			biz:     "test",
			phone:   "123456789",
			code:    "123456",
			wantErr: ErrCodeSendTooMany,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			redisClient := tc.mock(ctrl)
			codeCache := NewRedisCodeCache(redisClient)
			err := codeCache.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
