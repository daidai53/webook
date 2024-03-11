// Copyright@daidai53 2024
package es

import (
	"context"
	"encoding/json"
	elastic "github.com/elastic/go-elasticsearch/v8"
	olivere "github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
	"time"
)

type ElasticSearchTestSuite struct {
	suite.Suite
	es      *elastic.Client
	olivere *olivere.Client
}

func (s *ElasticSearchTestSuite) SetupSuite() {
	client, err := elastic.NewClient(elastic.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	})
	assert.NoError(s.T(), err)
	s.es = client
	ol, err := olivere.NewClient(
		olivere.SetURL("http://localhost:9200"),
		olivere.SetSniff(false),
	)
	assert.NoError(s.T(), err)
	s.olivere = ol
}

func (s *ElasticSearchTestSuite) TestCreateIndex() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	def := `{
  "settings":{
    "number_of_shards":3,
    "number_of_replicas":2
  },
  "mappings": {
    "properties": {
      "email": {
        "type": "text"
      },
      "phone": {
        "type": "keyword"
      },
      "birthday": {
        "type": "date"
      }
    }
  }
}`

	// style1
	resp, err := s.es.Indices.Create("user_idx_go", s.es.Indices.Create.WithContext(ctx),
		s.es.Indices.Create.WithBody(strings.NewReader(def)))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 200, resp.StatusCode)

	//style2
	_, err = s.olivere.CreateIndex("user_idx_go_v1").Body(def).Do(ctx)
	require.NoError(s.T(), err)
}

func (s *ElasticSearchTestSuite) TestPutDoc() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	//	data := `{
	//  "email": "daidai@qq.com",
	//  "phone": "123456789",
	//  "birthday": "1995-10-20"
	//}`
	//
	//	_, err := s.es.Index("user_idx_go", strings.NewReader(data), s.es.Index.WithContext(ctx))
	//	require.NoError(s.T(), err)

	_, err := s.olivere.Index().Index("user_idx_go_v1").Id("1").BodyJson(User{
		Email:    "daidai2@qq.com",
		Phone:    "12345678",
		Birthday: "2000-01-02",
	}).Do(ctx)
	assert.NoError(s.T(), err)
}

func (s *ElasticSearchTestSuite) TestGetDoc() {
	olQuery := olivere.NewMatchQuery("email", "daidai2")
	resp, err := s.olivere.Search("user_idx_go_v1").Query(olQuery).Do(context.Background())
	assert.NoError(s.T(), err)
	for _, hit := range resp.Hits.Hits {
		var u User
		json.Unmarshal(hit.Source, &u)
		s.T().Log(u)
	}
}

func TestElasticSearchTestSuite(t *testing.T) {
	suite.Run(t, &ElasticSearchTestSuite{})
}

type User struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Birthday string `json:"birthday"`
}
