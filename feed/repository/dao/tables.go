// Copyright@daidai53 2024
package dao

// 对应的是收件箱
type FeedPushEvent struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 收件人
	Uid  int64 `gorm:"index"`
	Type string
	// 扩展字段，不同的事件类型有不同的解析方式，取决于Type
	Content string
	CTime   int64
	// 没有更新场景，不用定义UTime字段
}

// 对应的是发件箱
type FeedPullEvent struct {
	Id      int64 `gorm:"primaryKey,autoIncrement"`
	Uid     int64
	Type    string
	Content string
	CTime   int64
}
