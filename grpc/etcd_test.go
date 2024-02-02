// Copyright@daidai53 2024
package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type EtcdTestSuite struct {
	suite.Suite
	cli *etcdv3.Client
}

func (s *EtcdTestSuite) SetupSuite() {
	cli, err := etcdv3.NewFromURL("localhost:12379")
	require.NoError(s.T(), err)
	s.cli = cli

}

func (s *EtcdTestSuite) TestClient() {
	t := s.T()

	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(t, err)

	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client := NewUserServiceClient(cc)
	resp, err := client.GetByID(context.Background(), &GetByIDRequest{Id: 123})
	require.NoError(t, err)
	t.Log(resp.User)
}

func (s *EtcdTestSuite) TestServer() {
	t := s.T()
	em, err := endpoints.NewManager(s.cli, "service/user")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	addr := "127.0.0.1:8090"
	key := "service/user/" + addr
	l, err := net.Listen("tcp", ":8090")
	ttl := int64(5)
	leaseRsp, _ := s.cli.Grant(ctx, ttl)
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr:     addr,
		Metadata: time.Now().String(),
	}, etcdv3.WithLease(leaseRsp.ID))
	require.NoError(t, err)

	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		ch, _ := s.cli.KeepAlive(kaCtx, leaseRsp.ID)
		for kaResp := range ch {
			t.Log(kaResp.String())
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		for now := range ticker.C {
			ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
			err1 := em.AddEndpoint(ctx1, key, endpoints.Endpoint{
				Addr:     addr,
				Metadata: now.String(),
			})
			cancel1()
			if err1 != nil {
				t.Log(err1)
			}
		}
	}()

	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Server{})
	server.Serve(l)
	kaCancel()
	em.DeleteEndpoint(ctx, key)
	server.GracefulStop()
	s.cli.Close()
}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}
