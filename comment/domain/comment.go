// Copyright@daidai53 2024
package domain

import "time"

type Comment struct {
	Id          int64 `json:"id"`
	Commentator User  `json:"commentator"`

	Biz   string `json:"biz"`
	BizId int64  `json:"biz_id"`

	Content string `json:"content"`

	RootComment   *Comment `json:"root_comment"`
	ParentComment *Comment `json:"parent_comment"`

	Children []Comment `json:"children"`

	CTime time.Time `json:"c_time"`
	UTime time.Time `json:"u_time"`
}

type User struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}
