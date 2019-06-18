package server

import (
	"context"
	"errors"
	"testing"

	"github.com/smoya/ratio/pkg/rate"

	"github.com/stretchr/testify/assert"

	ratio "github.com/smoya/ratio/api/proto"
)

func noopLimiter(ok bool, err error) rate.Limiter {
	return func(_ rate.Limit, _, _ string) (bool, error) {
		return ok, err
	}
}

func TestGRPC_RateLimit(t *testing.T) {
	cases := []struct {
		code ratio.RateLimitResponse_Code
		ok   bool
		err  error
	}{
		{code: ratio.RateLimitResponse_OK, ok: true},
		{code: ratio.RateLimitResponse_OVER_LIMIT, ok: false},
		{code: ratio.RateLimitResponse_UNKNOWN, err: errors.New("whatever error")},
	}

	for _, c := range cases {
		s := NewGRPC(rate.NewLimit(rate.PerMinute, 5), noopLimiter(c.ok, c.err))
		resp, err := s.RateLimit(context.Background(), &ratio.RateLimitRequest{})

		if c.err != nil {
			assert.EqualError(t, err, c.err.Error())
		} else {
			assert.NoError(t, err)
		}

		assert.Equal(t, c.code, resp.Code)
	}

}
