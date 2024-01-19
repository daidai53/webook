// Copyright@daidai53 2024
package domain

import (
	"github.com/robfig/cron/v3"
	"time"
)

type Job struct {
	Id   int64
	Name string
	// cron
	Expression string
	Executor   string
	CancelFunc func()
}

func (j Job) NextTime() time.Time {
	c := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	s, _ := c.Parse(j.Expression)
	return s.Next(time.Now())
}
