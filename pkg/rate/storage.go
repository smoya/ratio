package rate

import "time"

// SlideWindowStorage represents the storage behind the Slide Window limiter algorithm
type SlideWindowStorage interface {
	Add(key string, now time.Time, expireIn time.Duration) error
	Drop(key string, until time.Time) (int, error)
	Count(key string, until time.Time) (int, error)
	Flush() error
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
