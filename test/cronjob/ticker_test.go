// Copyright@daidai53 2024
package cronjob

import (
	"context"
	cron "github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	tick := time.NewTicker(time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	defer tick.Stop()
	var stop bool
	for !stop {
		select {
		case <-ctx.Done():
			stop = true
			t.Log("超时")
		case <-tick.C:
			t.Log("1 second pass")
		}
	}
	t.Log("over")
}

func TestCronExpr(t *testing.T) {
	expr := cron.New(cron.WithSeconds())

	id, err := expr.AddJob("@every 1s", JobFunc(func() {
		t.Log("执行了")
	}))
	assert.NoError(t, err)
	t.Log(id)
	expr.Start()
	time.Sleep(time.Second * 10)
	ctx := expr.Stop()
	t.Log("停止信号")
	<-ctx.Done()
	t.Log("彻底停下来")
}

type JobFunc func()

func (j JobFunc) Run() {
	j()
}
