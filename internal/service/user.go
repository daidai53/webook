// Copyright@daidai53 2023
package service

import (
	"context"
	"errors"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户邮箱或者密码不存在")
	ErrNoUserProfile         = repository.ErrUserNotFound
)

type UserService interface {
	SignUp(ctx context.Context, user domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	Profile(c *gin.Context, userId int64) error
	Edit(c *gin.Context, idInt64 int64, nickname, birthday, aboutMe string) error
	FindOrCreate(c *gin.Context, phone string) (domain.User, error)
	FindOrCreateByWeChat(c context.Context, info domain.WeChatInfo) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(r repository.UserRepository) UserService {
	return &userService{
		repo: r,
	}
}

func (u *userService) SignUp(ctx context.Context, user domain.User) error {
	encryptedPwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(encryptedPwd)
	return u.repo.Create(ctx, user)
}

func (u *userService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	usr, err := u.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return usr, nil
}

func (u *userService) Profile(c *gin.Context, userId int64) error {
	usr, err := u.repo.FindById(c, userId)
	if errors.Is(err, repository.ErrUserNotFound) {
		c.String(http.StatusOK, "系统错误")
		return err
	}
	c.JSON(http.StatusOK, toJson(usr))
	return nil
}

func (u *userService) Edit(c *gin.Context, idInt64 int64, nickname, birthday, aboutMe string) error {
	return u.repo.Update(c, idInt64, domain.User{
		Nickname: nickname,
		Birthday: birthday,
		AboutMe:  aboutMe,
	})
}

func (u *userService) FindOrCreate(c *gin.Context, phone string) (domain.User, error) {
	usr, err := u.repo.FindByPhone(c, phone)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return usr, err
	}
	err = u.repo.Create(c, domain.User{
		Phone: phone,
	})
	// 有两种可能
	// err恰好是唯一索引冲突，或者是系统错误
	if err != nil && !errors.Is(err, repository.ErrDuplicateUser) {
		return domain.User{}, err
	}
	// err为nil或者重复，代表用户存在
	// 这里有坑：主从延迟，可能会从从库查询
	// 理论上应该强制走主库查询
	return u.repo.FindByPhone(c, phone)
}

func (u *userService) FindOrCreateByWeChat(c context.Context, wechatInfo domain.WeChatInfo) (domain.User, error) {
	usr, err := u.repo.FindByWeChat(c, wechatInfo.OpenId)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return usr, err
	}
	err = u.repo.Create(c, domain.User{
		WeChatInfo: wechatInfo,
	})
	// 有两种可能
	// err恰好是唯一索引冲突，或者是系统错误
	if err != nil && !errors.Is(err, repository.ErrDuplicateUser) {
		return domain.User{}, err
	}
	// err为nil或者重复，代表用户存在
	// 这里有坑：主从延迟，可能会从从库查询
	// 理论上应该强制走主库查询
	return u.repo.FindByWeChat(c, wechatInfo.OpenId)
}

func toJson(usr domain.User) gin.H {
	return gin.H{
		"Email":    usr.Email,
		"Nickname": usr.Nickname,
		"Phone":    usr.Phone,
		"Birthday": usr.Birthday,
		"AboutMe":  usr.AboutMe,
	}
}
