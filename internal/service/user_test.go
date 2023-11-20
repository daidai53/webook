// Copyright@daidai53 2023
package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/repository"
	repomocks "github.com/daidai53/webook/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func Test_PasswordEncrypt(t *testing.T) {
	password := []byte("abc_123456")
	encrypted, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	assert.NoError(t, err)
	fmt.Println(string(encrypted))
}

func TestUserService_Login(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository

		// 预期输入
		ctx      context.Context
		email    string
		password string

		// 预期输出
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$i6yXSQiXbu7dYYDsSa4YTeqY5JInvIvm59bft8PWU9d4fxdd3Uc9.",
						Phone:    "123456789",
					}, nil)
				return repo
			},
			email:    "123@qq.com",
			password: "abc_123456",

			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "$2a$10$i6yXSQiXbu7dYYDsSa4YTeqY5JInvIvm59bft8PWU9d4fxdd3Uc9.",
				Phone:    "123456789",
			},
			wantErr: nil,
		},
		{
			name: "用户未找到",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			email:    "123@qq.com",
			password: "abc_123456",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{}, errors.New("dao error"))
				return repo
			},
			email:    "123@qq.com",
			password: "abc_123456",

			wantUser: domain.User{},
			wantErr:  errors.New("dao error"),
		},
		{
			name: "密码不对",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").
					Return(domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$i6yXSQiXbu7dYYDsSa4YTeqY5JInvIvm59bft8PWU9d4fxdd3Uc9.",
						Phone:    "123456789",
					}, nil)
				return repo
			},
			email:    "123@qq.com",
			password: "abc_12345",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewUserService(tc.mock(ctrl))
			user, err := svc.Login(tc.ctx, tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
