package rate

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

// NewSlideWindowStorageFromDSN creates a SlideWindowStorage based on a DSN.
// Example: redis://localhost:6379/0
func NewSlideWindowStorageFromDSN(raw string) (SlideWindowStorage, error) {
	dsn, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	var storage SlideWindowStorage
	switch dsn.Scheme {
	case "redis":
		ops := &redis.Options{
			Addr: dsn.Host,
			DB:   0,
		}

		db, err := strconv.Atoi(dsn.Path)
		if err == nil {
			ops.DB = db
		}

		storage = NewRedisSlideWindowStorage(
			redis.NewClient(ops),
		)
	case "inmemory":
		storage = NewInMemorySlideWindowStorage(make(map[string][]time.Time))
	default:
		return nil, errors.New("invalid slide window storage")
	}

	return storage, nil
}
