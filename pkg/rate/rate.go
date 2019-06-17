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

// SlideWindowStorage represents the storage behind the Slide Window limiter algorithm
type SlideWindowStorage interface {
	Add(key string, now time.Time, expireIn time.Duration) error
	Drop(key string, until time.Time) (int, error)
	Count(key string, until time.Time) (int, error)
	Flush() error
}

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

// NewInMemorySlideWindowStorage creates a new InMemory SlideWindowStorage.
// Not recommended for prod. Just testing purpose. In case you want to use an in memory storage, please implement it with
// the interface SlideWindowStorage. A good implementation could be https://github.com/patrickmn/go-cache.
func NewInMemorySlideWindowStorage(store map[string][]time.Time) SlideWindowStorage {
	return &inMemorySlideWindowStorage{store: store}
}

func (s inMemorySlideWindowStorage) Add(key string, now time.Time, _ time.Duration) error {
	if _, ok := s.store[key]; !ok {
		s.store[key] = make([]time.Time, 0)
	}

	s.store[key] = append(s.store[key], now)
	return nil
}

func (s *inMemorySlideWindowStorage) Drop(key string, until time.Time) (int, error) {
	if len(s.store[key]) == 0 {
		return 0, nil
	}

	var dropped int
	tsInWindow := s.store[key][:0]
	for _, t := range s.store[key] {
		if t.After(until) || t.Equal(until) {
			tsInWindow = append(tsInWindow, t)
		} else {
			dropped++
		}
	}

	s.store[key] = tsInWindow

	return dropped, nil
}

func (s inMemorySlideWindowStorage) Count(key string, until time.Time) (int, error) {
	var hits int
	for _, t := range s.store[key] {
		if t.Before(until) || t.Equal(until) {
			hits++
		}
	}

	return hits, nil
}

func (s *inMemorySlideWindowStorage) Flush() error {
	s.store = make(map[string][]time.Time)
	return nil
}
