package server

import (
	"context"
	"log"

	ratio "github.com/smoya/ratio/api/proto"
)

// GRPC is a GRPC server.
type GRPC struct{}

// RateLimit implements ratio.RateLimitService
func (s *GRPC) RateLimit(ctx context.Context, r *ratio.RateLimitRequest) (*ratio.RateLimitResponse, error) {
	log.Printf("RateLimit request: %s -> %s\n", r.Owner, r.Resource)

	return &ratio.RateLimitResponse{
		Code: ratio.RateLimitResponse_UNKNOWN,
	}, nil
}
