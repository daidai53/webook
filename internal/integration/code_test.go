// Copyright@daidai53 2024
package integration

import (
	"context"
	"github.com/daidai53/webook/internal/integration/startup"
	"github.com/daidai53/webook/internal/repository/cache"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type CodeHandlerSuite struct {
	suite.Suite
	cmd redis.Cmdable
}

func (c *CodeHandlerSuite) SetupSuite() {
	c.cmd = startup.InitRedis()
}

func (c *CodeHandlerSuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := c.cmd.FlushDB(ctx).Err()
	assert.NoError(c.T(), err)
}

func (c *CodeHandlerSuite) TestSendAndVerify() {
	t := c.T()
	testCases := []struct {
		name string

		before  func(*testing.T)
		mid     func(*testing.T)
		after   func(*testing.T)
		getCode func(*testing.T) string

		biz   string
		phone string

		wantErr      error
		wantErr2     error
		wantVerified bool
	}{
		{
			name:   "发送成功，认证成功",
			biz:    "article",
			phone:  "123456789",
			before: func(t *testing.T) {},
			mid:    func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := c.cmd.Del(ctx, "phone_code:article:123456789").Err()
				assert.NoError(t, err)
			},
			getCode: func(t *testing.T) string {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				res, err := c.cmd.Get(ctx, "phone_code:article:123456789").Result()
				assert.NoError(t, err)
				return res
			},
			wantVerified: true,
		},
		{
			name:   "发送成功，认证失败",
			biz:    "article",
			phone:  "423456789",
			before: func(t *testing.T) {},
			mid:    func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := c.cmd.Del(ctx, "phone_code:article:423456789").Err()
				assert.NoError(t, err)
			},
			getCode: func(t *testing.T) string {
				return "invalid"
			},
			wantVerified: false,
		},
		{
			name:  "发送太频繁",
			biz:   "article",
			phone: "223456789",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				c.cmd.Set(ctx, "phone_code:article:223456789", "xxx", 10*time.Minute)
			},
			mid: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := c.cmd.Del(ctx, "phone_code:article:223456789").Err()
				assert.NoError(t, err)
			},
			getCode: func(t *testing.T) string {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				res, err := c.cmd.Get(ctx, "phone_code:article:223456789").Result()
				assert.NoError(t, err)
				return res
			},
			wantErr: cache.ErrCodeSendTooMany,
		},
		{
			name:   "验证太频繁",
			biz:    "article",
			phone:  "323456789",
			before: func(t *testing.T) {},
			mid: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := c.cmd.Set(ctx, "phone_code:article:323456789:cnt", "0", 10*time.Minute).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := c.cmd.Del(ctx, "phone_code:article:323456789").Err()
				assert.NoError(t, err)
			},
			getCode: func(t *testing.T) string {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				res, err := c.cmd.Get(ctx, "phone_code:article:323456789").Result()
				assert.NoError(t, err)
				return res
			},
			wantVerified: false,
		},
	}

	svc := startup.InitCodeService()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			err := svc.Send(context.Background(), tc.biz, tc.phone)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			tc.mid(t)
			verified, err := svc.Verify(context.Background(), tc.biz, tc.phone, tc.getCode(t))
			assert.Equal(t, tc.wantErr2, err)
			assert.Equal(t, tc.wantVerified, verified)
		})
	}
}

func TestCodeHandler(t *testing.T) {
	suite.Run(t, &CodeHandlerSuite{})
}
