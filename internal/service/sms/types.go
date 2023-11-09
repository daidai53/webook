// Copyright@daidai53 2023
package sms

import "context"

// Service 	发送短信的抽象
type Service interface {
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}
