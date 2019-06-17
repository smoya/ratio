package server

import (
	"context"
	"testing"
	"time"

	"github.com/smoya/ratio/pkg/rate"

	"github.com/stretchr/testify/assert"

	ratio "github.com/smoya/ratio/api/proto"
)

func noopLimiter(_ rate.Limit, _, _ string) (bool, error) {
	return true, nil
}

func TestGRPC_RateLimit(t *testing.T) {

	s := NewGRPC(rate.NewLimit(time.Minute, 5), noopLimiter)
	r := ratio.RateLimitRequest{}
	resp, err := s.RateLimit(context.TODO(), &r)

	assert.NoError(t, err)
	assert.Equal(t, &ratio.RateLimitResponse{
		Code: ratio.RateLimitResponse_OK,
	}, resp)
}
