// Copyright@daidai53 2024
package events

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/daidai53/webook/follow/repository"
	"github.com/daidai53/webook/follow/repository/dao"
	"github.com/daidai53/webook/pkg/canalx"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/daidai53/webook/pkg/saramax"
	"time"
)

type FollowBinlogConsumer struct {
	client sarama.Client
	repo   repository.CachedRelationRepository
	l      logger.LoggerV1
}

func (f *FollowBinlogConsumer) Start() error {
	consumerGroup, err := sarama.NewConsumerGroupFromClient("follow_relation_cache",
		f.client)
	if err != nil {
		return err
	}
	go func() {
		err := consumerGroup.Consume(
			context.Background(),
			[]string{"webook_binlog"},
			saramax.NewHandler[canalx.Message[dao.FollowRelation]](f.l, f.Consume),
		)
		if err != nil {
			f.l.Error("退出消费循环异常",
				logger.Error(err))
		}
	}()
	return err
}

func (f *FollowBinlogConsumer) Consume(msg *sarama.ConsumerMessage,
	val canalx.Message[dao.FollowRelation]) error {
	if val.Table != "follow_relations" {
		return nil
	}
	if val.Type != "INSERT" && val.Type != "UPDATE" {
		f.l.Error("操作类型不能被处理",
			logger.String("type", val.Type))
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, row := range val.Data {
		var err error
		switch row.Status {
		case dao.FollowRelationStatusActive:
			err = f.repo.Cache().Follow(ctx, row.Follower, row.Followee)
			if err != nil {
				f.l.Error("更新缓存失败",
					logger.Error(err),
					logger.Uint8("new_status", row.Status),
					logger.Int64("follower", row.Follower),
					logger.Int64("followee", row.Followee),
				)
			}
		case dao.FollowRelationStatusInactive:
			err = f.repo.Cache().CancelFollow(ctx, row.Follower, row.Followee)
			if err != nil {
				f.l.Error("更新缓存失败",
					logger.Error(err),
					logger.Uint8("new_status", row.Status),
					logger.Int64("follower", row.Follower),
					logger.Int64("followee", row.Followee),
				)
			}
		default:
			f.l.Error("未知状态")
		}
	}
	return nil
}
