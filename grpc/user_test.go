// Copyright@daidai53 2024
package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	grpcServer := grpc.NewServer()
	userServer := &Server{}
	RegisterUserServiceServer(grpcServer, userServer)

	l, err := net.Listen("tcp", ":8090")
	require.NoError(t, err)
	err = grpcServer.Serve(l)
	t.Log(err)
}

func TestClient(t *testing.T) {
	cc, err := grpc.Dial("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	resp, err := client.GetByID(context.Background(), &GetByIDRequest{
		Id: 123,
	})
	require.NoError(t, err)
	t.Log(resp.User)
}
