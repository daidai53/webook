// Copyright@daidai53 2024
package grpcx

import (
	"context"
	"github.com/daidai53/webook/pkg/netx"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"time"
)

type Server struct {
	Server   *grpc.Server
	cli      *etcdv3.Client
	EtcdUrl  string
	kaCancel func()
	Port     int
	Name     string
}

func (s *Server) Serve() error {
	addr := ":" + strconv.Itoa(s.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	s.register()
	return s.Server.Serve(l)
}

func (s *Server) register() error {
	client, err := etcdv3.NewFromURL(s.EtcdUrl)
	if err != nil {
		return err
	}

	s.cli = client
	em, err := endpoints.NewManager(s.cli, "service/"+s.Name)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	addr := netx.GetOutBoundIP() + ":" + strconv.Itoa(s.Port)
	key := "service/" + s.Name + "/" + addr

	ttl := int64(5)
	leaseRsp, err := s.cli.Grant(ctx, ttl)
	if err != nil {
		return err
	}
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr:     addr,
		Metadata: time.Now().String(),
	}, etcdv3.WithLease(leaseRsp.ID))
	if err != nil {
		return err
	}

	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.kaCancel = kaCancel
	_, err = s.cli.KeepAlive(kaCtx, leaseRsp.ID)
	return err
}

func (s *Server) Close() error {
	if s.kaCancel != nil {
		s.kaCancel()
	}
	if s.cli != nil {
		// 依赖注入的方式就不要关
		s.cli.Close()
	}
	s.Server.GracefulStop()
	return nil
}
