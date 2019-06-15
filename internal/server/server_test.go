package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	ratio "github.com/smoya/ratio/api/proto"
)

func TestGRPC_RateLimit(t *testing.T) {
	s := GRPC{}
	r := ratio.RateLimitRequest{}
	resp, err := s.RateLimit(context.TODO(), &r)

	assert.NoError(t, err)
	assert.Equal(t, &ratio.RateLimitResponse{}, resp) // Note that UNKNOWN code is 0, so the zero value.
}
