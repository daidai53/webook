// Copyright@daidai53 2024
package dao

import "context"

type UserSearchDAO interface {
	InputUser(ctx context.Context, user User) error
	Search(ctx context.Context, keywords []string) ([]User, error)
}

type ArticleSearchDAO interface {
	InputArticle(ctx context.Context, arti Article) error
	Search(ctx context.Context, tagArgIds []int64, colIds []int64, likeIds []int64, keywords []string) ([]Article, error)
}

type TagSearchDAO interface {
	SearchBizIds(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error)
}

type InterSearchDAO interface {
	SearchCollectBizIds(ctx context.Context, uid int64, biz string) ([]int64, error)
	SearchLikeBizIds(ctx context.Context, uid int64, biz string) ([]int64, error)
}

type AnyDAO interface {
	Input(ctx context.Context, index, docId, data string) error
}

type User struct {
	Id       int64
	Nickname string
	Email    string
	Phone    string
}

type Article struct {
	Id      int64
	Title   string
	Content string
	Status  int32
}
