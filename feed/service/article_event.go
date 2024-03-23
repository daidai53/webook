// Copyright@daidai53 2024
package service

import (
	"context"
	followv1 "github.com/daidai53/webook/api/proto/gen/follow/v1"
	"github.com/daidai53/webook/feed/domain"
	"github.com/daidai53/webook/feed/repository"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"sort"
	"sync"
	"time"
)

type ArticleEventHandler struct {
	repo         repository.FeedEventRepo
	followClient followv1.FollowServiceClient
}

// 压测之后才能判定
const threshold = 100
const articleEvent = "article_event"

func (a *ArticleEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	followee, err := ext.Get("followee").AsInt64()
	if err != nil {
		return err
	}
	// 找到该人的粉丝数量，判断拉还是推模型
	resp, err := a.followClient.GetFollowStatic(ctx, &followv1.GetFollowStaticRequest{
		Followee: followee,
	})
	if err != nil {
		return err
	}

	if resp.GetFollowStatic().GetFollowers() > threshold {
		return a.repo.CreatePullEvent(ctx, domain.FeedEvent{
			Uid:   followee,
			Type:  articleEvent,
			Ctime: time.Now(),
			Ext:   ext,
		})
	}

	// 推模型
	// 先查粉丝
	fResp, err := a.followClient.GetFollower(ctx, &followv1.GetFollowerRequest{
		Followee: followee,
	})
	if err != nil {
		return err
	}
	events := slice.Map(fResp.GetFollowRelations(), func(idx int, src *followv1.FollowRelation) domain.FeedEvent {
		if src == nil {
			return domain.FeedEvent{}
		}
		return domain.FeedEvent{
			Uid:   src.Follower,
			Type:  articleEvent,
			Ctime: time.Now(),
			Ext:   ext,
		}
	})
	return a.repo.CreatePushEvents(ctx, events)
}

func (a *ArticleEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	var eg errgroup.Group
	var lock sync.Mutex
	events := make([]domain.FeedEvent, 0, limit*2)
	eg.Go(func() error {
		// 查询发件箱
		resp, err := a.followClient.GetFollowee(ctx, &followv1.GetFolloweeRequest{
			Follower: uid,
			Limit:    10000,
		})
		if err != nil {
			return err
		}
		followeeIds := slice.Map(resp.GetFollowRelations(), func(idx int, src *followv1.FollowRelation) int64 {
			return src.Followee
		})
		evts, err := a.repo.FindPullEventsWithTyp(ctx, articleEvent, followeeIds, timestamp, limit)
		if err != nil {
			return err
		}
		lock.Lock()
		events = append(events, evts...)
		lock.Unlock()
		return nil
	})

	eg.Go(func() error {
		evts, err := a.repo.FindPushEventsWithTyp(ctx, articleEvent, uid, timestamp, limit)
		if err != nil {
			return err
		}

		lock.Lock()
		events = append(events, evts...)
		lock.Unlock()
		return nil
	})

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	// 排序
	sort.Slice(events, func(i, j int) bool {
		return events[i].Ctime.UnixMilli() > events[j].Ctime.UnixMilli()
	})
	return events[:min(int(limit), len(events))], nil
}
