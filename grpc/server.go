// Copyright@daidai53 2024
package grpc

import "context"

type Server struct {
	UnimplementedUserServiceServer
}

func (s *Server) GetByID(ctx context.Context, request *GetByIDRequest) (*GetByIDResponse, error) {
	return &GetByIDResponse{
		User: &User{
			Id:   123,
			Name: "daidai53",
		},
	}, nil
}
