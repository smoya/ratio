package rate

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

func TestRedisSlideWindowStorage_Add(t *testing.T) {
	r, m := createRedis()
	defer m.Close()
	defer m.FlushAll()

	store := NewRedisSlideWindowStorage(r)
	now := time.Now()

	assert.NoError(t, store.Add("key1", now, 0))
	hits, err := r.ZRange("key1", 0, -1).Result()

	assert.NoError(t, err)
	assert.Len(t, hits, 1)
	assert.Equal(t, fmt.Sprintf("%d", now.UnixNano()/1000000), hits[0])
}

func TestRedisSlideWindowStorage_Count(t *testing.T) {
	r, m := createRedis()
	defer m.Close()
	defer m.FlushAll()

	store := NewRedisSlideWindowStorage(r)
	now := time.Now()

	assert.NoError(t, store.Add("key1", now, 0))
	assert.NoError(t, store.Add("key1", now.Add(-time.Minute), 0))
	assert.NoError(t, store.Add("key1", now.Add(-time.Minute*2), 0))

	c, err := store.Drop("key1", now.Add(-time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, 1, c)
}

func TestRedisSlideWindowStorage_Drop(t *testing.T) {
	r, m := createRedis()
	defer m.Close()
	defer m.FlushAll()

	store := NewRedisSlideWindowStorage(r)
	now := time.Now()

	assert.NoError(t, store.Add("key1", now, 0))
	assert.NoError(t, store.Add("key1", now.Add(-time.Minute), 0))
	assert.NoError(t, store.Add("key1", now.Add(-time.Minute*2), 0))

	c, err := store.Drop("key1", now.Add(-time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, 1, c)

	hits, err := r.ZRevRange("key1", 0, -1).Result()
	assert.NoError(t, err)

	assert.Equal(t, []string{
		strconv.Itoa(int(now.UnixNano() / 1000000)),
		strconv.Itoa(int(now.Add(-time.Minute).UnixNano() / 1000000)),
	}, hits)
}

func TestRedisSlideWindowStorage_Flush(t *testing.T) {
	r, m := createRedis()
	defer m.Close()
	defer m.FlushAll()

	store := NewRedisSlideWindowStorage(r)
	now := time.Now()

	assert.NoError(t, store.Add("key1", now, 0))
	assert.NoError(t, store.Flush())

	hits, err := r.ZCount("key1", "-inf", "inf").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), hits)
}

func TestSlideWindowLimiter_RedisStorage(t *testing.T) {
	r, m := createRedis()
	defer m.Close()

	s := redisSlideWindowStorage{r}
	limiter := SlideWindowRateLimiter(s, false)

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
			limit:        NewLimit(PerHour, 4),
			owner:        "myservice",
			resource:     "resource1",
			previousHits: 3,
			ok:           true,
		},
		{
			desc:         "Limit 3/h. 3 hits found in 1h window. 1 more is NOT allowed.",
			limit:        NewLimit(PerHour, 3),
			owner:        "myservice",
			resource:     "resource1",
			previousHits: 3,
			ok:           false,
		},
		{
			desc:         "Limit 2/m. 1 hit found in 1m window. 1 more is allowed.",
			limit:        NewLimit(PerMinute, 2),
			owner:        "myservice",
			resource:     "resource1",
			previousHits: 1,
			ok:           true,
		},
		{
			desc:         "Limit 1/m. 1 hit found in 1m window. 1 more is NOT allowed.",
			limit:        NewLimit(PerMinute, 1),
			owner:        "myservice",
			resource:     "resource1",
			previousHits: 1,
			ok:           false,
		},
		{
			desc:         "Limit 1/m. 1 hit found in 1m window. 1 more is allowed AFTER 1 min.",
			limit:        NewLimit(PerMinute, 1),
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
			assert.Equal(t, c.ok, ok)

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
