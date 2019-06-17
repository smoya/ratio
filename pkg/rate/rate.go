package rate

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

// Frequency is the unit of a rate based on time.
type Frequency time.Duration

// Duration returns the duration equivalence.
func (f Frequency) Duration() time.Duration {
	return time.Duration(f)
}

// Default Frequencies
const (
	PerDay    = Frequency(time.Hour * 60)
	PerHour   = Frequency(time.Hour)
	PerMinute = Frequency(time.Minute)
)

// ParseFrequency returns a Frequency from a string representation.
func ParseFrequency(s string) (Frequency, error) {
	switch s {
	case "day", "DAY", "d":
		return PerDay, nil
	case "hour", "HOUR", "h":
		return PerHour, nil
	case "minute", "MINUTE", "m":
		return PerMinute, nil
	}

	return 0, fmt.Errorf("%s is not a valid frequency for a rate", s)
}

// Limit represents the rate limit of a repeating event (hits) per unit of time.
type Limit struct {
	Unit     Frequency
	Quantity int
}

// NewLimit creates a Limit
func NewLimit(unit Frequency, quantity int) Limit {
	return Limit{Unit: unit, Quantity: quantity}
}

// ParseLimit parses a Limit from a string representation.
// Example: "100/day", "50/minute", "1/m"
func ParseLimit(s string) (Limit, error) {
	var l Limit
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return l, fmt.Errorf("%s is not a valid limit", s)
	}

	f, err := ParseFrequency(parts[1])
	if err != nil {
		return l, err
	}

	q, err := strconv.Atoi(parts[0])
	if err != nil {
		return l, fmt.Errorf("%s is not a valid quantity", parts[0])
	}

	l.Unit = f
	l.Quantity = q

	return l, nil
}

// Limiter rate limits a resource for a given owner based on a Rate.
type Limiter func(l Limit, owner, resource string) (bool, error)

// SlideWindowRateLimiter limits based on a time window that is always in movement (sliding).
func SlideWindowRateLimiter(s SlideWindowStorage, async bool) Limiter {
	return func(l Limit, owner, resource string) (bool, error) {
		now := time.Now()
		windowStartedAt := now.Add(-l.Unit.Duration())

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
				_ = s.Add(key, now, l.Unit.Duration())
			}()
		} else {
			err = s.Add(key, now, l.Unit.Duration())
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
