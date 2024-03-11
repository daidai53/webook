// Copyright@daidai53 2024
package dao

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"
	"strconv"
	"strings"
)

type UserElasticDAO struct {
	client *elastic.Client
}

func (u *UserElasticDAO) InputUser(ctx context.Context, user User) error {
	_, err := u.client.Index().Index("user_idx").
		Id(strconv.FormatInt(user.Id, 10)).
		BodyJson(user).Do(ctx)
	return err
}

func (u *UserElasticDAO) Search(ctx context.Context, keywords []string) ([]User, error) {
	queryString := strings.Join(keywords, " ")
	query := elastic.NewMatchQuery("nickname", queryString)
	resp, err := u.client.Search("user_idx").Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]User, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var user User
		err := json.Unmarshal(hit.Source, &user)
		if err != nil {
			return nil, err
		}
		res = append(res, user)
	}
	return res, nil
}
