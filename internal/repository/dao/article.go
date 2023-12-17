// Copyright@daidai53 2023
package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error
}

type ArticleGormDAO struct {
	db *gorm.DB
}

func NewArticleGormDAO(db *gorm.DB) ArticleDAO {
	return &ArticleGormDAO{
		db: db,
	}
}

func (a *ArticleGormDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	tx := a.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	now := time.Now().UnixMilli()
	res := tx.Model(&Article{}).Where("id = ? and author_id = ?", uid, id).Updates(map[string]interface{}{
		"utime":  now,
		"status": status,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return errors.New("更新失败，ID不对或作者不对")
	}
	res = tx.Model(&PublishedArticle{}).Where("id = ?", uid).Updates(map[string]interface{}{
		"utime":  now,
		"status": status,
	})
	if res.Error != nil {
		return res.Error
	}
	tx.Commit()
	return nil
}

func (a *ArticleGormDAO) Sync(ctx context.Context, art Article) (int64, error) {
	tx := a.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer tx.Rollback()

	var id = art.Id
	var err error
	dao := NewArticleGormDAO(tx)
	if id > 0 {
		err = dao.UpdateById(ctx, art)
	} else {
		id, err = dao.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	pubArt := PublishedArticle(art)
	now := time.Now().UnixMilli()
	pubArt.Ctime = now
	pubArt.Utime = now
	err = tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   pubArt.Title,
			"content": pubArt.Content,
			"utime":   now,
			"status":  pubArt.Status,
		}),
	}).Create(&pubArt).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, err
}

func (a *ArticleGormDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := a.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (a *ArticleGormDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	res := a.db.WithContext(ctx).Model(&art).Where("id = ? AND author_id=?", art.Id, art.AuthorId).Updates(map[string]any{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"utime":   now,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("更新失败，ID不对或作者不对")
	}
	return nil
}

type Article struct {
	Id      int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title   string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	Content string `gorm:"type=BLOB" bson:"content,omitempty"`
	// 要根据创作者ID来查询
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	Ctime    int64 `bson:"ctime,omitempty"`
	Utime    int64 `bson:"utime,omitempty"`
}

type PublishedArticle Article
