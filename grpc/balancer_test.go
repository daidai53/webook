// Copyright@daidai53 2024
package grpc

import (
	"context"
	_ "github.com/daidai53/webook/pkg/grpcx/balancer/wrr"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/balancer/weightedroundrobin"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type BalancerTestSuite struct {
	suite.Suite
	cli *etcdv3.Client
}

func (s *BalancerTestSuite) SetupSuite() {
	cli, err := etcdv3.NewFromURL("localhost:12379")
	require.NoError(s.T(), err)
	s.cli = cli

}

func (s *BalancerTestSuite) TestClient() {
	t := s.T()

	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(t, err)

	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`{
  "loadBalancingConfig": [ { "round_robin": {} } ],
  "methodConfig": [
    {
      "name": [{"service": "UserService"}],
      "retryPolicy": {
        "maxAttempts": 4,
        "initialBackoff": "0.01s",
        "maxBackoff": "0.1s",
        "backoffMultiplier": 2.0,
        "retryableStatusCodes": ["UNAVAILABLE"]
      }
    }
  ]
}`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)

	time.Sleep(time.Millisecond * 50)
	for i := 0; i < 10; i++ {
		resp, err := client.GetByID(context.Background(), &GetByIDRequest{Id: 123})
		require.NoError(t, err)
		t.Log(resp.User)
	}
}

func (s *BalancerTestSuite) TestServer() {
	go func() {
		s.startServer(":8090", 10, &Server{
			Name: ":8090",
		})
	}()
	go func() {
		s.startServer(":8091", 20, &Server{
			Name: ":8091",
		})
	}()
	s.startServer(":8092", 30, &FailServer{
		Name: ":8092",
	})
}

func (s *BalancerTestSuite) startServer(port string, weight int, svr UserServiceServer) {
	t := s.T()
	em, err := endpoints.NewManager(s.cli, "service/user")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	addr := "127.0.0.1" + port
	key := "service/user/" + addr
	l, err := net.Listen("tcp", port)
	ttl := int64(5)
	leaseRsp, _ := s.cli.Grant(ctx, ttl)
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
		Metadata: map[string]any{
			"weight": weight,
		},
	}, etcdv3.WithLease(leaseRsp.ID))
	require.NoError(t, err)

	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		_, _ = s.cli.KeepAlive(kaCtx, leaseRsp.ID)
	}()

	server := grpc.NewServer()
	RegisterUserServiceServer(server, svr)
	server.Serve(l)
	kaCancel()
	em.DeleteEndpoint(ctx, key)
	server.GracefulStop()
}

func TestBalancer(t *testing.T) {
	suite.Run(t, new(BalancerTestSuite))
}
