// Copyright@daidai53 2024
package grpcx

import (
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	Server *grpc.Server
	Addr   string
}

func (s *Server) Serve() error {
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		panic(err)
	}
	return s.Server.Serve(l)
}