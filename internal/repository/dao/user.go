// Copyright@daidai53 2023
package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("邮箱冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	Update(ctx context.Context, idInt64 int64, user User) error
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByWeChat(ctx context.Context, openId string) (User, error)
}

type GormUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GormUserDAO{
		db: db,
	}
}

type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 代表这是一个可以为NULL的列，也可以写成*String的类型
	Email      sql.NullString `gorm:"unique"`
	Password   string
	CreateTime int64
	UpdateTime int64

	Nickname string
	Phone    sql.NullString `gorm:"unique"`

	// 如果查询要求同时使用openid和unionid，就要创建联合唯一索引
	// 如果查询只用openid，那么就在openid上创建唯一索引，或者<openid，unionid>联合索引
	// 如果查询只用unionid，那么就在unionid上创建唯一索引，或者<unionid,openid>联合索引
	WechatOpenId  sql.NullString `gorm:"unique"`
	WechatUnionId sql.NullString
	Birthday      string
	AboutMe       string
}

func (dao *GormUserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.CreateTime = now
	u.UpdateTime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			const duplicateErr uint16 = 1062
			if mysqlErr.Number == duplicateErr {
				return ErrDuplicateEmail
			}
		}
	}
	return err
}

func (dao *GormUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email=?", email).First(&u).Error
	return u, err
}

func (dao *GormUserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id=?", id).First(&u).Error
	return u, err
}

func (dao *GormUserDAO) Update(ctx context.Context, idInt64 int64, user User) error {
	return dao.db.WithContext(ctx).Where("id=?", idInt64).Updates(&user).Error
}

func (dao *GormUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone=?", phone).First(&u).Error
	return u, err
}

func (dao *GormUserDAO) FindByWeChat(ctx context.Context, openId string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("wechat_open_id=?", openId).First(&u).Error
	return u, err
}
