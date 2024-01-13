// Copyright@daidai53 2024
package context

import (
	"context"
	"testing"
	"time"
)

type key struct {
}

func TestContextValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), key{}, "value1")
	val, ok := ctx.Value(key{}).(string)
	t.Log(val, ok)
}

func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	<-ctx.Done()
	t.Log("已经cancel了")
}

func TestContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	<-ctx.Done()
	t.Log("超时了")
}

func TestContextParentCancel(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	time.AfterFunc(time.Second, func() {
		cancel()
	})

	son, sonCancel := context.WithCancel(parent)
	<-son.Done()
	t.Log("son 已经过来了")
	sonCancel()
}
