// Copyright@daidai53 2024
package dao

import (
	"context"
	"encoding/json"
	"github.com/ecodeclub/ekit/slice"
	"github.com/olivere/elastic/v7"
	"strconv"
	"strings"
)

type ArticleElasticDAO struct {
	client *elastic.Client
}

func (a *ArticleElasticDAO) InputArticle(ctx context.Context, arti Article) error {
	_, err := a.client.Index().Index("article_idx").
		Id(strconv.FormatInt(arti.Id, 10)).
		BodyJson(arti).Do(ctx)
	return err
}

func (a *ArticleElasticDAO) Search(ctx context.Context, tagArgIds []int64, keywords []string) ([]Article, error) {
	queryString := strings.Join(keywords, " ")
	ids := slice.Map(tagArgIds, func(idx int, src int64) any {
		return src
	})
	status := elastic.NewTermQuery("status", 2)
	tags := elastic.NewTermsQuery("id", ids...).Boost(2)
	title := elastic.NewMatchQuery("title", queryString)
	content := elastic.NewMatchQuery("content", queryString)

	or := elastic.NewBoolQuery().Should(tags, title, content)
	query := elastic.NewBoolQuery().Must(status, or)
	resp, err := a.client.Search("article_idx").Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]Article, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var art Article
		err = json.Unmarshal(hit.Source, &art)
		if err != nil {
			return nil, err
		}
		res = append(res, art)
	}
	return res, nil
}
