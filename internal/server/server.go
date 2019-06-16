package server

import (
	"context"
	"log"

	"github.com/smoya/ratio/pkg/rate"

	ratio "github.com/smoya/ratio/api/proto"
)

type grpc struct {
	limit   rate.Limit
	limiter rate.Limiter
}

// NewGRPC creates a new GRPC RateLimitServiceServer
func NewGRPC(limit rate.Limit, limiter rate.Limiter) ratio.RateLimitServiceServer {
	return &grpc{limit: limit, limiter: limiter}
}

// RateLimit implements ratio.RateLimitService
func (s *grpc) RateLimit(ctx context.Context, r *ratio.RateLimitRequest) (*ratio.RateLimitResponse, error) {
	log.Printf("RateLimit request: %s -> %s\n", r.Owner, r.Resource)

	ok, err := s.limiter(s.limit, r.Owner, r.Resource)
	if err != nil {
		return &ratio.RateLimitResponse{
			Code: ratio.RateLimitResponse_UNKNOWN,
		}, err
	}

	code := ratio.RateLimitResponse_OK
	if !ok {
		code = ratio.RateLimitResponse_OVER_LIMIT
	}

	return &ratio.RateLimitResponse{
		Code: code,
	}, nil
}
