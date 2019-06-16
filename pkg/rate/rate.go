package rate

import (
	"fmt"
	"sync"
	"time"
)

// Limit represents the rate limit of a repeating event (hits) per unit of time.
type Limit struct {
	Unit     time.Duration
	Quantity int
}

// NewLimit creates a Limit
func NewLimit(unit time.Duration, quantity int) Limit {
	return Limit{Unit: unit, Quantity: quantity}
}

// Limiter rate limits a resource for a given owner based on a Rate.
type Limiter func(l Limit, owner, resource string) (bool, error)

// InMemorySlideWindowRateLimiter implements Slide Window algorithm in local memory. Useful for testing purposes.
func InMemorySlideWindowRateLimiter(store map[string][]time.Time, safe bool) Limiter {
	var m sync.Mutex
	return func(l Limit, owner, resource string) (bool, error) {
		now := time.Now()
		key := fmt.Sprintf("%s_%s", owner, resource)

		if safe {
			m.Lock()
			defer m.Unlock()
		}

		if _, ok := store[key]; !ok {
			store[key] = []time.Time{
				now,
			}

			return 1 <= l.Quantity, nil
		}

		windowStartedAt := now.Add(-l.Unit)
		tsInWindow := store[key][:0]
		for _, t := range store[key] {
			if t.After(windowStartedAt) || t.Equal(windowStartedAt) {
				tsInWindow = append(tsInWindow, t)
			}
		}
		hits := len(tsInWindow)

		// Only store the new so we clean up ts prior to the time window.
		store[key] = append(tsInWindow, now)

		return hits < l.Quantity, nil
	}

}
