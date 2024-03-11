// Copyright@daidai53 2024
package domain

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

type SearchResult struct {
	Users    []User
	Articles []Article
}
