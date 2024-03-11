// Copyright@daidai53 2024
package dao

import (
	"context"
	"github.com/olivere/elastic/v7"
)

type AnyESDAO struct {
	client *elastic.Client
}

func (a *AnyESDAO) Input(ctx context.Context, index, docId, data string) error {
	_, err := a.client.Index().Index(index).Id(docId).BodyString(data).Do(ctx)
	return err
}
