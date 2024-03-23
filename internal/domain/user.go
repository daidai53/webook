// Copyright@daidai53 2023
package domain

type User struct {
	Id       int64
	Email    string
	Password string

	Nickname string
	Birthday string
	AboutMe  string
	Phone    string

	Active     bool
	WeChatInfo WeChatInfo
}
