// Copyright@daidai53 2023
package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/daidai53/webook/code/repository"
	"github.com/daidai53/webook/code/service/sms"
	"math/rand"
)

var ErrCodeSendTooMany = repository.ErrCodeSendTooMany

type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type CodeServiceImpl struct {
	repo repository.CodeRepository
	sms  sms.Service
}

func NewCodeService(c repository.CodeRepository, sms sms.Service) CodeService {
	return &CodeServiceImpl{
		repo: c,
		sms:  sms,
	}
}

func NewCodeServiceImpl(c repository.CodeRepository, sms sms.Service) *CodeServiceImpl {
	return &CodeServiceImpl{
		repo: c,
		sms:  sms,
	}
}

func (c *CodeServiceImpl) Send(ctx context.Context, biz, phone string) error {
	code := c.generateCode()
	err := c.repo.Set(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	const codeTplId = "1877556"
	return c.sms.Send(ctx, codeTplId, []string{code}, phone)
}

func (c *CodeServiceImpl) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	verify, err := c.repo.Verify(ctx, biz, phone, code)
	if errors.Is(err, repository.ErrCodeVerifyTooMany) {
		return false, nil
	}
	return verify, err
}

func (c *CodeServiceImpl) generateCode() string {
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}
