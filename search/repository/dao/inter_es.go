// Copyright@daidai53 2024
package dao

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
)

type InterEsDAO struct {
	client *elastic.Client
}

func (i *InterEsDAO) SearchCollectBizIds(ctx context.Context, uid int64, biz string) ([]int64, error) {
	query := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("uid", uid),
		elastic.NewTermQuery("biz", biz),
		elastic.NewMatchQuery("type", 1),
	)
	res, err := i.client.Search("interactive_index").Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	bizIds := make([]int64, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		var iEvt InteractiveEvent
		err := json.Unmarshal(hit.Source, &iEvt)
		if err != nil {
			return nil, err
		}
		bizIds = append(bizIds, iEvt.BizId)
	}
	return bizIds, nil
}

func (i *InterEsDAO) SearchLikeBizIds(ctx context.Context, uid int64, biz string) ([]int64, error) {
	query := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("uid", uid),
		elastic.NewTermQuery("biz", biz),
		elastic.NewMatchQuery("type", 2),
	)
	res, err := i.client.Search("interactive_index").Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	bizIds := make([]int64, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		var iEvt InteractiveEvent
		err := json.Unmarshal(hit.Source, &iEvt)
		if err != nil {
			return nil, err
		}
		bizIds = append(bizIds, iEvt.BizId)
	}
	return bizIds, nil
}

type InteractiveEvent struct {
	Uid   int64  `json:"uid"`
	Biz   string `json:"biz"`
	BizId int64  `json:"biz_id"`
	Type  uint8  `json:"type"`
}
