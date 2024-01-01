// Copyright@daidai53 2023
package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, id int64) error
	BatchIncrReadCnt(ctx context.Context, biz []string, id []int64) error
	InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error
	InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error
	GetLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserLikeBiz, error)
	GetCollectInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserCollectionBiz, error)
	Get(ctx context.Context, biz string, bizId int64) (Interactive, error)
}

type GORMInteractiveDAO struct {
	db *gorm.DB
}

func NewGORMInteractiveDAO(db *gorm.DB) InteractiveDAO {
	return &GORMInteractiveDAO{
		db: db,
	}
}

func (g *GORMInteractiveDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	var res Interactive
	err := g.db.WithContext(ctx).Where("biz=? AND biz_id=?", biz, bizId).First(&res).Error
	return res, err
}

func (g *GORMInteractiveDAO) GetCollectInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := g.db.WithContext(ctx).Where("biz=? AND biz_id=? AND uid=?", biz, bizId, uid).First(&res).Error
	return res, err
}

func (g *GORMInteractiveDAO) GetLikeInfo(ctx context.Context, biz string, bizId int64, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := g.db.WithContext(ctx).Where("biz=? AND biz_id=? AND uid=? AND status=?", biz, bizId, uid, 1).First(&res).Error
	return res, err
}

func (g *GORMInteractiveDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cb.CTime = now
	cb.UTime = now
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&cb).Error
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"collect_cnt": gorm.Expr("`collect_cnt`+1"),
				"u_time":      now,
			}),
		}).Create(&Interactive{
			Biz:        cb.Biz,
			BizId:      cb.BizId,
			CollectCnt: 1,
			CTime:      now,
			UTime:      now,
		}).Error
	})
}

func (g *GORMInteractiveDAO) InsertLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"u_time": now,
				"status": 1,
			}),
		}).Create(&UserLikeBiz{
			Uid:    uid,
			Biz:    biz,
			BizId:  id,
			Status: 1,
			UTime:  now,
			CTime:  now,
		}).Error
		if err != nil {
			return err
		}
		return tx.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt": gorm.Expr("`like_cnt`+1"),
				"u_time":   now,
			}),
		}).Create(&Interactive{
			Biz:     biz,
			BizId:   id,
			LikeCnt: 1,
			CTime:   now,
			UTime:   now,
		}).Error
	})
}

func (g *GORMInteractiveDAO) DeleteLikeInfo(ctx context.Context, biz string, id int64, uid int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&UserLikeBiz{}).
			Where("uid=? AND biz_id=? AND biz=?", uid, id, biz).
			Updates(map[string]interface{}{
				"u_time": now,
				"status": 0,
			}).Error
		if err != nil {
			return err
		}
		return tx.Model(&Interactive{}).Where("biz=? AND biz_id=?", biz, id).Updates(map[string]interface{}{
			"like_cnt": gorm.Expr("`like_cnt`-1"),
			"u_time":   now,
		}).Error
	})
}

func (g *GORMInteractiveDAO) BatchIncrReadCnt(ctx context.Context, biz []string, id []int64) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDAO := NewGORMInteractiveDAO(tx)
		for i := 0; i < len(biz); i++ {
			err := txDAO.IncrReadCnt(ctx, biz[i], id[i])
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *GORMInteractiveDAO) IncrReadCnt(ctx context.Context, biz string, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"read_cnt": gorm.Expr("`read_cnt`+1"),
			"u_time":   now,
		}),
	}).Create(&Interactive{
		Biz:     biz,
		BizId:   id,
		ReadCnt: 1,
		CTime:   now,
		UTime:   now,
	}).Error
}

type UserLikeBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// <bizId, biz>
	Uid    int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId  int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz    string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	Status int
	UTime  int64
	CTime  int64
}

type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// <bizId, biz>
	Cid   int64  `gorm:"index"`
	Uid   int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	BizId int64  `gorm:"uniqueIndex:uid_biz_type_id"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:uid_biz_type_id"`
	UTime int64
	CTime int64
}

type Interactive struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// <bizId, biz>
	BizId      int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz        string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`
	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	UTime      int64
	CTime      int64
}
