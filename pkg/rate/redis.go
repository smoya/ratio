package rate

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

// Rediser is a Redis Client interface. As redis lib does not have any interface, this gets useful for testing.
type Rediser interface {
	ZRemRangeByScore(key, min, max string) *redis.IntCmd
	ZCount(key, min, max string) *redis.IntCmd
	ZAdd(key string, members ...redis.Z) *redis.IntCmd
	Expire(key string, expiration time.Duration) *redis.BoolCmd
	FlushAll() *redis.StatusCmd
}

type redisSlideWindowStorage struct {
	r Rediser
}

// NewRedisSlideWindowStorage creates a new Redis SlideWindowStorage.
func NewRedisSlideWindowStorage(r Rediser) SlideWindowStorage {
	return &redisSlideWindowStorage{r: r}
}

func (s redisSlideWindowStorage) Add(key string, now time.Time, expireIn time.Duration) error {
	nowMs := s.toMilliseconds(now)
	err := s.r.ZAdd(key, redis.Z{Score: float64(nowMs), Member: nowMs}).Err()
	if err != nil {
		return err
	}

	if expireIn > 0 {
		err = s.r.Expire(key, expireIn).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s redisSlideWindowStorage) Drop(key string, until time.Time) (int, error) {
	hits, err := s.r.ZRemRangeByScore(key, "-inf", fmt.Sprintf("(%s", strconv.Itoa(s.toMilliseconds(until)))).Result()
	if err != nil && err != redis.Nil {
		return 0, err
	}

	return int(hits), nil
}

func (s redisSlideWindowStorage) Count(key string, until time.Time) (int, error) {
	hits, err := s.r.ZCount(key, "-inf", fmt.Sprintf("%d", s.toMilliseconds(until))).Result()
	if err != nil && err != redis.Nil {
		return 0, err
	}

	return int(hits), nil
}

func (s redisSlideWindowStorage) Flush() error {
	return s.r.FlushAll().Err()
}

func (s redisSlideWindowStorage) toMilliseconds(t time.Time) int {
	return int(t.UnixNano() / 1000000)
}
