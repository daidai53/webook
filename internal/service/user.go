// Copyright@daidai53 2023
package service

import (
	"context"
	"errors"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/repository"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户邮箱或者密码不存在")
	ErrNoUserProfile         = repository.ErrUserNotFound
)

type UserService interface {
	SignUp(ctx context.Context, user domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	Profile(c context.Context, userId int64) (domain.User, error)
	Edit(c *gin.Context, idInt64 int64, nickname, birthday, aboutMe string) error
	FindOrCreate(c *gin.Context, phone string) (domain.User, error)
	FindOrCreateByWeChat(c context.Context, info domain.WeChatInfo) (domain.User, error)
	IsActiveUser(ctx context.Context, uid int64) (bool, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(r repository.UserRepository) UserService {
	return &userService{
		repo: r,
	}
}

func (u *userService) IsActiveUser(ctx context.Context, uid int64) (bool, error) {
	return true, nil
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

func (u *userService) Profile(c context.Context, userId int64) (domain.User, error) {
	return u.repo.FindById(c, userId)
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
	// 这边意味着是一个新用户
	// json格式的wechatInfo
	zap.L().Info("新用户", zap.Any("unionid", wechatInfo))
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
