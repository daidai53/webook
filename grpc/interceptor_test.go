// Copyright@daidai53 2024
package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
)

type InterceptorTestSuite struct {
	suite.Suite
}

func (i *InterceptorTestSuite) TestServer() {
	t := i.T()
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(NewLogInterceptor(t)),
	)
	RegisterUserServiceServer(server, &Server{
		Name: "interceptor_test",
	})
	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	server.Serve(l)
}

func (i *InterceptorTestSuite) TestClient() {
	t := i.T()

	cc, err := grpc.Dial("localhost:8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client := NewUserServiceClient(cc)
	resp, err := client.GetByID(context.Background(), &GetByIDRequest{Id: 123})
	require.NoError(t, err)
	t.Log(resp.User)
}

func NewLogInterceptor(t *testing.T) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		t.Log("请求处理前", req, info)
		resp, err = handler(ctx, req)
		t.Log("请求处理后", resp, err)
		return
	}
}

func TestInterceptor(t *testing.T) {
	suite.Run(t, new(InterceptorTestSuite))
}
