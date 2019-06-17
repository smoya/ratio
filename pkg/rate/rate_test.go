package rate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseFrequency(t *testing.T) {
	cases := []struct {
		string      string
		frequency   Frequency
		shouldError bool
	}{
		{string: "day", frequency: PerDay},
		{string: "d", frequency: PerDay},
		{string: "DAY", frequency: PerDay},
		{string: "minute", frequency: PerMinute},
		{string: "m", frequency: PerMinute},
		{string: "MINUTE", frequency: PerMinute},
		{string: "hour", frequency: PerHour},
		{string: "h", frequency: PerHour},
		{string: "HOUR", frequency: PerHour},

		{string: "WEEK", shouldError: true},
		{string: "YEAR", shouldError: true},
		{string: "", shouldError: true},
	}

	for _, c := range cases {
		f, err := ParseFrequency(c.string)
		if c.shouldError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		assert.Equal(t, c.frequency, f)
	}
}

func TestParseLimit(t *testing.T) {
	cases := []struct {
		string      string
		limit       Limit
		shouldError bool
	}{
		{string: "100/day", limit: NewLimit(PerDay, 100)},
		{string: "50/h", limit: NewLimit(PerHour, 50)},
		{string: "2/MINUTE", limit: NewLimit(PerMinute, 2)},
		{string: "1/WEEK", shouldError: true},
	}

	for _, c := range cases {
		l, err := ParseLimit(c.string)
		if c.shouldError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		assert.Equal(t, c.limit, l)
	}
}

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
			limit:    NewLimit(PerHour, 4),
			owner:    "myservice",
			resource: "resource1",
			ok:       true,
		},
		{
			desc:     "Limit 3/h. 3 hits found for 1h window. 1 more is NOT allowed.",
			limit:    NewLimit(PerHour, 3),
			owner:    "myservice",
			resource: "resource1",
			ok:       false,
		},
		{
			desc:     "Limit 2/m. 1 hit found for 1m window. 1 more is allowed.",
			limit:    NewLimit(PerMinute, 2),
			owner:    "myservice",
			resource: "resource1",
			ok:       true,
		},
		{
			desc:     "Limit 1/m. 1 hit found for 1m window. 1 more is NOT allowed.",
			limit:    NewLimit(PerMinute, 1),
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
