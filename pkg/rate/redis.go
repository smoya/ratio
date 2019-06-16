package rate

func RedisSlideWindowRateLimiter() Limiter {
	return func(l Limit, owner, resource string) (bool, error) {
		//windowStartedAt := time.Now().Add(-l.Unit)

		// Lua
		// 1. ZREMRANGEBYSCORE windowStartedAt.Nanosecond() / 1000000 TODO CHECK CONVERSION
		// 2. ZRANGE(0, -1)

		// Asynchronously:
		// 1. ZADD
		// 2. TTL = time.Now().Add(l.Unit)

		fetchedTokens := 10

		return fetchedTokens < l.Quantity, nil
	}
}
