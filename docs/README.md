# ratio : A simple distributed rate limiter made in go

## Table of contents

- [Decisions and thoughts](decisions.md)

## Rate limit algorithm

The algorithm behind the `ratio` rate limit calculation is called "Slide window of timestamps". It could be considered 
as an improvement or adaptation of the [Token Bucket](decisions.md#token-bucket).
The idea remains on calculating the rate limit based on a time window that is always in movement (sliding), and with a 
fixed size according the time unit you use in the rate calculation (1 min, 1 hour, 2 days...).

### Redis implementation

Redis [Sorted Sets](https://redis.io/topics/data-types#sorted-sets) are lists of non repeating elements associated with 
a score.

- Each hit will add a new element to the Sorted Set, the score of it will be the timestamp (the key will be the same).
- On each hit, we run a `ZREMRANGEBYSCORE` Redis command in order to remove the elements of the Sorted Set with a 
  score (timestamp) lower than the current one.
- The remaining elements will contain the real hits that happened during the current time window. Running a `ZCOUNT min_score (now` 
  will give us the total hits count inside the time window. Then we add the current one by run `ZADD` command (always).
    - At this point, we can already know if the rate limit should apply.
- As last step, we set a TTL on the Sorted Set key with the rate limit interval. Redis will evict the Sorted Set if no 
  hits are received during that time.
  
All this operations can be done atomically but as consistency in `ratio` is not a priority, this should not need to happen. 
  