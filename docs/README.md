# ratio : A simple distributed rate limiter made in go

## Table of contents

- [Usage](#usage)
- [Configuration](#configuration)
- [Decisions and thoughts](decisions.md)
- [Rate limit algorithm](#rate-limit-algorithm)

## Usage

`ratio` exposes it's API via [GRPC](https://grpc.io/). It means that you just need a GRPC client on your side.
> Install your GRPC deps for your language at https://packages.grpc.io/

### API

[GRPC](https://grpc.io/) API is declared via [Procol Buffers](https://developers.google.com/protocol-buffers/).
Please find the `.proto` file at [/api/proto/ratio.proto](/api/proto/ratio.proto).

There is only one RPC by now: `RateLimit`.

### Definition

- **Owner**: The owner of the target resource. Usually the service name from where  the request was made. 
Used for storing the hits in a dedicated bucket per service. 
  - Examples:
    1. `my-awesome-service`
* **Resource**: The target resource behind the rate limit.
  - Examples:
    1. `/v1/order/pay`
    2. `/v1/order/pay#customer123`
    3. `graphql_resolver_root`
    
As you may noticed, the combination of `owner` plus `resource`, makes an entry as unique.

### GRPC command line test client

In case you want to do some calls to the server, you can install the `grpc_cli` tool from 
[here](https://github.com/grpc/grpc/blob/master/doc/command_line_tool.md). 

Example:

```bash
grpc_cli call localhost:50051 RateLimit "owner: 'my-awesome-service', resource: '/v1/user/register'"
```

## Configuration

`ratio` can be configured via environment variables. Please find here the most important ones:

- `RATIO_PORT`: The GRPC port. Default `50051`.
- `RATIO_CONNECTION_TIMEOUT`: Timeout for all incoming connections. Default `1s`.
- `RATIO_STORAGE`: DSN Storage. Example: `inmemory://`. Default: `redis://redis:6379/0`.
- `RATIO_LIMIT`: The rate limit. Example: `2400/day`, `100/hour`, `2/minute`.

## Storage

`ratio` storage is configurable via the `RATIO_STORAGE` env var. Its value should be a `DSN` related to the storage you
want to use. e.g. `redis://localhost:6379/0`.

### In memory

The In memory implementation that you can find in `ratio` is not meant to be used in production. Is a non concurrency-safe 
storage for testing purpose.
A good in memory implementation could be https://github.com/patrickmn/go-cache, however a distributed in memory database 
will lead in to a better final storage.   

### Redis

`ratio` preferred storage is [Redis](https://redis.io/).
Please read why we chose Redis as preferred storage in the [Decisions and thoughts](decisions.md#storage) doc.
Read more details about the implementation [here](#redis-implementation).

### Your own storage

`ratio` library allows to quickly implement your own storage thanks to its design based on Interface Segregation.
Take a look to the `SlideWindowStorage` interface located in the [rate](/pkg/rate/storage.go) package:

```go
type SlideWindowStorage interface {
	Add(key string, now time.Time, expireIn time.Duration) error
	Drop(key string, until time.Time) (int, error)
	Count(key string, until time.Time) (int, error)
	Flush() error
}
```

Implementing this interface and adding your own DSN pattern (e.g. `mongodb://host:port/db`) in the factory 
[`NewSlideWindowStorageFromDSN`](/pkg/rate/factory.go) will let `ratio` to use your own storage through 
the env var `RATIO_STORAGE`.

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
  