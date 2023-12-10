// Copyright@daidai53 2023
package domain

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
}

type Author struct {
	Id   int64
	Name string
}
