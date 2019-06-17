package rate

import (
	"errors"
	"fmt"
	"log"
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
}

// RedisSlideWindowRateLimiter implements Slide Window algorithm using redis as store.
func RedisSlideWindowRateLimiter(r Rediser, async bool) Limiter {
	return func(l Limit, owner, resource string) (bool, error) {
		now := int(time.Now().UnixNano() / 1000000)
		windowStartedAtStr := strconv.Itoa(int(time.Now().Add(-l.Unit).UnixNano() / 1000000))

		key := fmt.Sprintf("%s-%s", owner, resource) // TODO COMPRESS?

		// Note: We do not run the following using MULTI nor via LUA scripting since we do not want to block other operations.

		// 1. ZREMRANGEBYSCORE removing hits that happen before the current time window.
		err := r.ZRemRangeByScore(key, "-inf", fmt.Sprintf("(%s", windowStartedAtStr)).Err()
		if err != nil && err != redis.Nil {
			log.Println("error during redis ZREMRANGEBYSCORE command")
		}

		// 2. ZCOUNT since the beginning of the windows until now
		hits, err := r.ZCount(key, windowStartedAtStr, fmt.Sprintf("(%d", now)).Result()
		if err != nil && err != redis.Nil {
			return false, errors.New("getting hits count")
		}

		if async {
			// Asynchronously, we do not want the caller to wait as ratio is eventually consistent.
			go appendHit(r, l, key, now)
		} else {
			appendHit(r, l, key, now)
		}

		return int(hits) < l.Quantity, nil
	}
}

func appendHit(r Rediser, l Limit, key string, now int) {
	// 1. ZADD
	err := r.ZAdd(key, redis.Z{Score: float64(now), Member: now}).Err()
	if err != nil {
		log.Println("error during redis ZADD command")
	}

	// 2. Set TTL
	err = r.Expire(key, l.Unit).Err()
	if err != nil {
		log.Println("error during redis EXPIRE command")
	}
}
