// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/feed/domain"
	"github.com/daidai53/webook/feed/repository"
	"github.com/daidai53/webook/internal/service"
	"time"
)

const likeEventName = "like_event"

type LikeEventHandler struct {
	repo        repository.FeedEventRepo
	userService service.UserService
}

func (l *LikeEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	// 字段校验，可以做或不做，看和业务方的协商
	// 需要被点赞的人
	uid, err := ext.Get("liked").Int64()
	if err != nil {
		return err
	}

	if act, err := l.userService.IsActiveUser(ctx, uid); err == nil && act {
		return l.repo.CreatePullEvent(ctx,
			domain.FeedEvent{
				Uid:   uid,
				Ext:   ext,
				Type:  likeEventName,
				Ctime: time.Now(),
			})
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
	if act, err := l.userService.IsActiveUser(ctx, uid); err == nil && act {
		return l.repo.FindPullEventsWithTyp(ctx, likeEventName, []int64{uid}, timestamp, limit)
	}
	return l.repo.FindPushEventsWithTyp(ctx, likeEventName, uid, timestamp, limit)
}
