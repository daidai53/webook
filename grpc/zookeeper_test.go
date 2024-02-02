// Copyright@daidai53 2024
package grpc

import (
	"context"
	"github.com/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

func TestZooKeeperServer(t *testing.T) {
	var hosts = []string{"localhost:2181"}
	conn, _, err := zk.Connect(hosts, time.Second)
	defer conn.Close()
	assert.NoError(t, err)

	_, err = conn.Create("/service", []byte("127.0.0.1:8090"), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	assert.NoError(t, err)

	listener, err := net.Listen("tcp", ":8090")
	assert.NoError(t, err)
	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Server{})
	server.Serve(listener)
	err = conn.Delete("/service/server", 0)
	assert.NoError(t, err)
}

func TestZooKeeperClient(t *testing.T) {
	var hosts = []string{"localhost:2181"}
	conn, _, err := zk.Connect(hosts, time.Second)
	defer conn.Close()
	assert.NoError(t, err)

	target, _, err := conn.Get("/service")
	assert.NoError(t, err)

	cc, err := grpc.Dial(string(target),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client := NewUserServiceClient(cc)
	resp, err := client.GetByID(context.Background(), &GetByIDRequest{Id: 123})
	require.NoError(t, err)
	t.Log(resp.User)
}
