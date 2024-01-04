// Copyright@daidai53 2023
package repository

import (
	"bytes"
	"context"
	"fmt"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/repository/dao"
	daomocks "github.com/daidai53/webook/internal/repository/dao/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/exp/rand"
	"os"
	"testing"
	"time"
)

func TestCachedArticleRepository_SyncV1(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO)
		art     domain.Article
		wantId  int64
		wantErr error
	}{
		{
			name: "同步成功",
			mock: func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO) {
				authorDao := daomocks.NewMockArticleAuthorDAO(ctrl)
				readerDao := daomocks.NewMockArticleReaderDAO(ctrl)
				authorDao.EXPECT().Create(gomock.Any(), dao.Article{
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
				}).Return(int64(1), nil)
				readerDao.EXPECT().Upsert(gomock.Any(), dao.Article{
					Id:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
				}).Return(nil)
				return authorDao, readerDao
			},
			art: domain.Article{
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 1,
		},
		{
			name: "修改同步成功",
			mock: func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO) {
				authorDao := daomocks.NewMockArticleAuthorDAO(ctrl)
				readerDao := daomocks.NewMockArticleReaderDAO(ctrl)
				authorDao.EXPECT().Update(gomock.Any(), dao.Article{
					Id:       11,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
				}).Return(nil)
				readerDao.EXPECT().Upsert(gomock.Any(), dao.Article{
					Id:       11,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
				}).Return(nil)
				return authorDao, readerDao
			},
			art: domain.Article{
				Id:      11,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 11,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			authorDAO, readerDAO := tc.mock(ctrl)
			repo := NewCachedArticleRepositoryV2(readerDAO, authorDAO)
			id, err := repo.SyncV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}

func Test_GenSql(t *testing.T) {
	file, _ := os.OpenFile("a.sql", os.O_CREATE, 0666)
	var pub = `insert into published_articles (id, title, content, author_id, status, ctime, utime) values ('%d','%s','%s','%d','1','100','100');
`
	var arti = `insert into articles (id, title, content, author_id, status, ctime, utime) values ('%d','%s','%s','%d','1','100','100');`
	var inter = `insert into interactives (id, biz_id, biz, read_cnt, like_cnt, collect_cnt, u_time, c_time) values ('%d','%d','article',0,'%d',0,'100','100')
`
	var buffer bytes.Buffer
	for i := 1; i <= 1000; i++ {
		rand.Seed(uint64(time.Now().UnixMilli()))
		likeCnt := rand.Intn(1000) + 1
		buffer.WriteString(fmt.Sprintf(pub, i, fmt.Sprintf("title-%d", i), fmt.Sprintf("content-%d", i),
			123))
		buffer.WriteString(fmt.Sprintf(arti, i, fmt.Sprintf("title-%d", i), fmt.Sprintf("content-%d", i),
			123))
		buffer.WriteString(fmt.Sprintf(inter, i, i, likeCnt))
	}
	file.WriteString(buffer.String())
	file.Sync()
	file.Close()
}
