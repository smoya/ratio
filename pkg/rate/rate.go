package rate

import (
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
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

// SlideWindowRateLimiter limits based on a time window that is always in movement (sliding).
func SlideWindowRateLimiter(s SlideWindowStorage, async bool) Limiter {
	return func(l Limit, owner, resource string) (bool, error) {
		now := time.Now()
		windowStartedAt := now.Add(-l.Unit)

		key := fmt.Sprintf("%s-%s", owner, resource) // TODO COMPRESS?

		_, err := s.Drop(key, windowStartedAt)
		if err != nil && err != redis.Nil {
			log.Printf("error dropping out of window hits: %s\n", err.Error())
		}

		hits, err := s.Count(key, now)
		if err != nil && err != redis.Nil {
			return false, fmt.Errorf("getting hits count: %s", err.Error())
		}

		if async {
			// Asynchronously, we do not want the caller to wait as ratio is eventually consistent.
			go func() {
				_ = s.Add(key, now, l.Unit)
			}()
		} else {
			err = s.Add(key, now, l.Unit)
			if err != nil {
				log.Printf("error adding hit: %s\n", err.Error())
			}
		}

		return int(hits) < l.Quantity, nil
	}
}

type inMemorySlideWindowStorage struct {
	store map[string][]time.Time
}
