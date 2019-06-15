//go:generate protoc -I ../../api/proto --go_out=plugins=grpc:../../api/proto ../../api/proto/ratio.proto

package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc/reflection"

	ratio "github.com/smoya/ratio/api/proto"
	"google.golang.org/grpc"
)

type server struct{}

// RateLimit implements ratio.RateLimitService
func (s *server) RateLimit(ctx context.Context, r *ratio.RateLimitRequest) (*ratio.RateLimitResponse, error) {
	log.Printf("RateLimit request: %s -> %s\n", r.Owner, r.Resource)
	return &ratio.RateLimitResponse{
		Code: ratio.RateLimitResponse_UNKNOWN,
	}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	reflection.Register(s)
	ratio.RegisterRateLimitServiceServer(s, &server{})
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
