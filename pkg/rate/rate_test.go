package rate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFoo(t *testing.T) {
	assert.Equal(t, "foo", time.Now().Truncate(time.Hour).String())
}

func TestInMemorySlideWindowRateLimiter(t *testing.T) {
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
			ok, err := InMemorySlideWindowRateLimiter(inMemoryStore(), false)(c.limit, c.owner, c.resource)
			assert.NoError(t, err)
			assert.Equal(t, ok, c.ok)
		})
	}
}

func inMemoryStore() map[string][]time.Time {
	return map[string][]time.Time{
		"myservice_resource1": {
			time.Now().Add(-time.Hour * 2),
			time.Now().Add(-time.Minute * 45),
			time.Now().Add(-time.Minute * 30),
			time.Now().Add(-time.Second * 15),
		},
	}
}
