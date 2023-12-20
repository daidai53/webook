// Copyright@daidai53 2023
package dao

import (
	"bytes"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/daidai53/webook/internal/domain"
	"github.com/ecodeclub/ekit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"time"
)

type ArticleS3DAO struct {
	ArticleGormDAO
	oss *s3.S3
}

func NewArticleS3DAO(db *gorm.DB, client *s3.S3) *ArticleS3DAO {
	return &ArticleS3DAO{
		ArticleGormDAO: ArticleGormDAO{db: db},
		oss:            client,
	}
}

func (a *ArticleS3DAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
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
	res = tx.Model(&PublishedArticleV2{}).Where("id = ?", uid).Updates(map[string]interface{}{
		"utime":  now,
		"status": status,
	})
	if res.Error != nil {
		return res.Error
	}
	tx.Commit()
	var err error
	if status == uint8(domain.ArticleStatusPrivate) {
		_, err = a.oss.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: ekit.ToPtr[string]("webook-1"),
			Key:    ekit.ToPtr[string](strconv.FormatInt(id, 10)),
		})

	}
	return err
}

func (a *ArticleS3DAO) Sync(ctx context.Context, art Article) (int64, error) {
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
	now := time.Now().UnixMilli()
	pubArt := PublishedArticleV2{
		Id:       art.Id,
		Title:    art.Title,
		AuthorId: art.AuthorId,
		Ctime:    now,
		Utime:    now,
		Status:   art.Status,
	}
	pubArt.Ctime = now
	pubArt.Utime = now
	err = tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":  pubArt.Title,
			"utime":  now,
			"status": pubArt.Status,
		}),
	}).Create(&pubArt).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	_, err = a.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      ekit.ToPtr[string]("webook-1"),
		Key:         ekit.ToPtr[string](strconv.FormatInt(art.Id, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: ekit.ToPtr[string]("text/plain;charset=utf-8"),
	})
	return id, err
}

type PublishedArticleV2 struct {
	Id    int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	// 要根据创作者ID来查询
	AuthorId int64 `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	Ctime    int64 `bson:"ctime,omitempty"`
	Utime    int64 `bson:"utime,omitempty"`
}
