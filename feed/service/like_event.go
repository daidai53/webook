// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/feed/domain"
	"github.com/daidai53/webook/feed/repository"
	"time"
)

const likeEventName = "like_event"

type LikeEventHandler struct {
	repo repository.FeedEventRepo
}

func (l *LikeEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	// 字段校验，可以做或不做，看和业务方的协商
	// 需要被点赞的人
	uid, err := ext.Get("liked").Int64()
	if err != nil {
		return err
	}
	return l.repo.CreatePushEvents(
		ctx,
		[]domain.FeedEvent{
			{
				Uid:   uid,
				Ext:   ext,
				Type:  likeEventName,
				Ctime: time.Now(),
			},
		},
	)
}

func (l *LikeEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	return l.repo.FindPushEventsWithTyp(ctx, likeEventName, uid, timestamp, limit)
}
