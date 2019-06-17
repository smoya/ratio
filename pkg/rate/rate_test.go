package rate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInMemorySlideWindowStorage_Add(t *testing.T) {
	s := make(map[string][]time.Time)
	store := NewInMemorySlideWindowStorage(s)

	now := time.Now()
	assert.NoError(t, store.Add("key1", now, 0))

	assert.Len(t, s["key1"], 1)
	assert.Equal(t, now, s["key1"][0])
}

func TestInMemorySlideWindowStorage_Count(t *testing.T) {
	s := make(map[string][]time.Time)
	store := NewInMemorySlideWindowStorage(s)

	now := time.Now()
	assert.NoError(t, store.Add("key1", now, 0))

	c, err := store.Count("key1", now)
	assert.NoError(t, err)
	assert.Equal(t, 1, c)
}

func TestInMemorySlideWindowStorage_Drop(t *testing.T) {
	s := make(map[string][]time.Time)
	store := NewInMemorySlideWindowStorage(s)

	now := time.Now()
	assert.NoError(t, store.Add("key1", now, 0))
	assert.NoError(t, store.Add("key1", now.Add(-time.Minute), 0))
	assert.NoError(t, store.Add("key1", now.Add(-time.Minute*2), 0))

	c, err := store.Drop("key1", now.Add(-time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, 1, c)

	assert.Len(t, s["key1"], 2)
	assert.Equal(t, []time.Time{now, now.Add(-time.Minute)}, s["key1"])
}

func TestInMemorySlideWindowStorage_Flush(t *testing.T) {
	store := NewInMemorySlideWindowStorage(inMemoryStore())
	assert.NoError(t, store.Flush())
	assert.Empty(t, store.(*inMemorySlideWindowStorage).store)
}

func TestSlideWindowLimiter_InMemoryStorage(t *testing.T) {
	store := NewInMemorySlideWindowStorage(inMemoryStore())
	limiter := SlideWindowRateLimiter(store, false)

	cases := []struct {
		desc     string
		limit    Limit
		owner    string
		resource string
		ok       bool
	}{
		{
			desc:     "Limit 4/h. 3 hits found for 1h window. 1 more is allowed.",
			limit:    NewLimit(time.Hour, 4),
			owner:    "myservice",
			resource: "resource1",
			ok:       true,
		},
		{
			desc:     "Limit 3/h. 3 hits found for 1h window. 1 more is NOT allowed.",
			limit:    NewLimit(time.Hour, 3),
			owner:    "myservice",
			resource: "resource1",
			ok:       false,
		},
		{
			desc:     "Limit 2/m. 1 hit found for 1m window. 1 more is allowed.",
			limit:    NewLimit(time.Minute, 2),
			owner:    "myservice",
			resource: "resource1",
			ok:       true,
		},
		{
			desc:     "Limit 1/m. 1 hit found for 1m window. 1 more is NOT allowed.",
			limit:    NewLimit(time.Minute, 1),
			owner:    "myservice",
			resource: "resource1",
			ok:       false,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			ok, err := limiter(c.limit, c.owner, c.resource)
			assert.NoError(t, err)
			assert.Equal(t, c.ok, ok)
			assert.NoError(t, store.Flush())
			store.(*inMemorySlideWindowStorage).store = inMemoryStore()
		})
	}
}

func inMemoryStore() map[string][]time.Time {
	return map[string][]time.Time{
		"myservice-resource1": {
			time.Now().Add(-time.Hour * 2),
			time.Now().Add(-time.Minute * 45),
			time.Now().Add(-time.Minute * 30),
			time.Now().Add(-time.Second * 15),
		},
	}
}
