// Copyright@daidai53 2023
package repository

import (
	"context"
	"database/sql"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/repository/cache"
	"github.com/daidai53/webook/internal/repository/dao"
	"github.com/gin-gonic/gin"
	"log"
)

var (
	ErrDuplicateUser = dao.ErrDuplicateEmail
	ErrUserNotFound  = dao.ErrRecordNotFound
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx *gin.Context, id int64) (domain.User, error)
	FindByPhone(c context.Context, phone string) (domain.User, error)
	Update(ctx context.Context, idInt64 int64, user domain.User) error
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewCachedUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (u *CachedUserRepository) Create(ctx context.Context, user domain.User) error {
	return u.dao.Insert(ctx, u.ToEntity(user))
}

func (u *CachedUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	usr, err := u.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return u.toDomainUser(usr), nil
}

func (u *CachedUserRepository) FindById(ctx *gin.Context, id int64) (domain.User, error) {
	cu, err := u.cache.Get(ctx, id)
	if err == nil {
		return cu, nil
	}
	/*
		err不为nil就查询数据库，err有两种情况
		1、 key不存在，redis是正常的；
		2、访问redis有问题，可能是网络问题或redis崩溃
	*/
	usr, err := u.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	du := u.toDomainUser(usr)
	go func() {
		err := u.cache.Set(ctx, du)
		if err != nil {
			log.Println(err)
		}
	}()
	return du, nil
}

func (u *CachedUserRepository) toDomainUser(usr dao.User) domain.User {
	return domain.User{
		Id:       usr.Id,
		Email:    usr.Email.String,
		Password: usr.Password,
		Nickname: usr.Nickname,
		Phone:    usr.Phone.String,
		Birthday: usr.Birthday,
		AboutMe:  usr.AboutMe,
	}
}

func (u *CachedUserRepository) Update(ctx context.Context, idInt64 int64, user domain.User) error {
	return u.dao.Update(ctx, idInt64, dao.User{
		Nickname: user.Nickname,
		Birthday: user.Birthday,
		AboutMe:  user.AboutMe,
	})
}

func (u *CachedUserRepository) FindByPhone(c context.Context, phone string) (domain.User, error) {
	user, err := u.dao.FindByPhone(c, phone)
	if err != nil {
		return domain.User{}, err
	}
	return u.toDomainUser(user), nil
}

func (u *CachedUserRepository) ToEntity(usr domain.User) dao.User {
	return dao.User{
		Id: usr.Id,
		Email: sql.NullString{
			String: usr.Email,
			Valid:  usr.Email != "",
		},
		Password: usr.Password,
		Nickname: usr.Nickname,
		Phone: sql.NullString{
			String: usr.Phone,
			Valid:  usr.Phone != "",
		},
		Birthday: usr.Birthday,
		AboutMe:  usr.AboutMe,
	}
}
