// Copyright@daidai53 2023
package cache

import (
	"context"
	"errors"
	"fmt"
	localcache "github.com/daidai53/localcache"
	"time"
)

var (
	errCntExhausted = errors.New("errCntExhausted")
	errCodeNotEqual = errors.New("errCodeNotEqual")
)

// LocalCodeCache 基于go-cache的CodeCache本地缓存实现
type LocalCodeCache struct {
	goCache localcache.LocalCache
}

func NewLocalCodeCache(goCache localcache.LocalCache) CodeCache {
	return &LocalCodeCache{
		goCache: goCache,
	}
}

func (l *LocalCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	key := l.key(biz, phone)
	cntKey := fmt.Sprintf("%s:%s", key, "cnt")
	return l.goCache.SafeOperate(key, func(c localcache.LocalCache) error {
		ttl, err := l.goCache.NLTTL(key)
		if errors.Is(err, localcache.ErrCodeNoExpireTime) {
			l.goCache.NLSet(key, []byte(code), 10*time.Minute)
			l.goCache.NLSet(cntKey, []byte{3}, 10*time.Minute)
		} else if err == nil && ttl < 540 {
			l.goCache.NLSet(key, []byte(code), 10*time.Minute)
			l.goCache.NLSet(cntKey, []byte{3}, 10*time.Minute)
		} else if err == nil {
			return ErrCodeSendTooMany
		} else {
			return err
		}
		return nil
	})
}

func (l *LocalCodeCache) Verify(ctx context.Context, biz, phone, expectedCode string) (bool, error) {
	key := l.key(biz, phone)
	cntKey := fmt.Sprintf("%s:%s", key, "cnt")
	err := l.goCache.SafeOperate(key, func(c localcache.LocalCache) error {
		cntData, err := l.goCache.NLGet(cntKey)
		if err != nil {
			return err
		}
		cnt := int(cntData[0])
		if cnt <= 0 {
			return errCntExhausted
		}
		codeData, err := l.goCache.NLGet(key)
		if err != nil {
			return err
		}
		code := string(codeData)
		if code == expectedCode {
			return nil
		} else {
			return errCodeNotEqual
		}
	})
	switch {
	case err == nil, errors.Is(err, localcache.ErrCodeRecordNotFound):
		return true, nil
	case errors.Is(err, errCntExhausted):
		return false, ErrVerifySendTooMany
	case errors.Is(err, errCodeNotEqual):
		return false, nil
	}
	return false, nil
}

func (l *LocalCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("%s:%s", biz, phone)
}
