// Copyright@daidai53 2024
package dao

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
)

type TagESDAO struct {
	client *elastic.Client
}

// Search 通过用户、资源领域、关键词搜索到相关的资源id
func (t *TagESDAO) SearchBizIds(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error) {
	query := elastic.NewBoolQuery().Must(
		elastic.NewTermsQuery("uid", uid),
		elastic.NewTermQuery("biz", biz),
		elastic.NewTermsQueryFromStrings("tags", keywords...),
	)
	resp, err := t.client.Search("tags_index").Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]int64, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var tb TagBizs
		err = json.Unmarshal(hit.Source, &tb)
		if err != nil {
			return nil, err
		}
		res = append(res, tb.BizId)
	}
	return res, nil
}

type TagBizs struct {
	Uid   int64    `json:"uid"`
	Biz   string   `json:"biz"`
	BizId int64    `json:"biz_id"`
	Tags  []string `json:"tags"`
}
