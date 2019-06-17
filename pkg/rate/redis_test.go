package rate

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

func TestRedisSlideWindowRateLimiter(t *testing.T) {
	r, m := createRedis()
	defer m.Close()

	limiter := RedisSlideWindowRateLimiter(r, false)

	cases := []struct {
		desc         string
		limit        Limit
		owner        string
		resource     string
		previousHits int
		fastForward  time.Duration
		ok           bool
	}{
		{
			desc:         "Limit 4/h. 3 hits found in 1h window. 1 more is allowed.",
			limit:        NewLimit(time.Hour, 4),
			owner:        "myservice",
			resource:     "resource1",
			previousHits: 3,
			ok:           true,
		},
		{
			desc:         "Limit 3/h. 3 hits found in 1h window. 1 more is NOT allowed.",
			limit:        NewLimit(time.Hour, 3),
			owner:        "myservice",
			resource:     "resource1",
			previousHits: 3,
			ok:           false,
		},
		{
			desc:         "Limit 2/m. 1 hit found in 1m window. 1 more is allowed.",
			limit:        NewLimit(time.Minute, 2),
			owner:        "myservice",
			resource:     "resource1",
			previousHits: 1,
			ok:           true,
		},
		{
			desc:         "Limit 1/m. 1 hit found in 1m window. 1 more is NOT allowed.",
			limit:        NewLimit(time.Minute, 1),
			owner:        "myservice",
			resource:     "resource1",
			previousHits: 1,
			ok:           false,
		},
		{
			desc:         "Limit 1/m. 1 hit found in 1m window. 1 more is allowed AFTER 1 min.",
			limit:        NewLimit(time.Minute, 1),
			owner:        "myservice",
			resource:     "resource1",
			previousHits: 1,
			fastForward:  time.Minute,
			ok:           true,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			for i := 0; i < c.previousHits; i++ {
				ok, err := limiter(c.limit, c.owner, c.resource)
				assert.True(t, ok, "error populating previous hits")
				assert.NoError(t, err, "error populating previous hits")

				// Needed because races can happen in miniredis.
				time.Sleep(time.Millisecond)
			}

			if c.fastForward > 0 {
				m.FastForward(c.fastForward)
			}

			ok, err := limiter(c.limit, c.owner, c.resource)
			assert.NoError(t, err)
			assert.Equal(t, ok, c.ok)

			r.FlushAll()
		})
	}
}

func createRedis() (*redis.Client, *miniredis.Miniredis) {
	mini, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	r := redis.NewClient(&redis.Options{
		Addr: mini.Addr(),
	})

	return r, mini
}
