// Copyright@daidai53 2024
package grpc

import "context"

type Server struct {
	UnimplementedUserServiceServer
	Name string
}

func (s *Server) GetByID(ctx context.Context, request *GetByIDRequest) (*GetByIDResponse, error) {
	return &GetByIDResponse{
		User: &User{
			Id:   123,
			Name: "from " + s.Name,
		},
	}, nil
}
