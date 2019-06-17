# Decisions

Find here a list of thoughts and decisions I made during the development of `ratio`.

## Architecture

The idea behind `ratio` is to provide a basic service for rate limiting web requests. 

## Good practices

- This projects follows the https://github.com/golang-standards/project-layout Go project layout.
  - I've been using it for a while. The layout gets easy to understand and it seems to be pretty accepted by the community.
- [Golang-ci](https://github.com/golangci/golangci-lint) is the preferred tool for running code linters.

## Deployment

This service is agnostic and it does not depends on any particular cloud provider or orchestration framework. However, 
as a facilitator, I chose Kubernetes for demo purposes. The config files for deploying it insight any cluster 
can be found in this repository.

## Communication

I decided to use gRPC. Some of the reasons behind are:
- It uses protocol buffers for both service declaration and data serialization. It gives a really good performance at
  serialize/deserialize and also a portable service schema definition, which can be used even for auto generate code.
- built on top of [HTTP/2](https://medium.com/@factoryhr/http-2-the-difference-between-http-1-1-benefits-and-how-to-use-it-38094fa0e95b).
  
## Storage

I choice a combination of local memory and Redis. The goal would be achieving a system that ensures 
the **AP** from the [CAP Theorem](https://en.wikipedia.org/wiki/CAP_theorem). We want a High Available service, and 
in second position a Partition Tolerant one. We do not really care about consistency if we can ensure the first two. 
 
- Redis will contain the distributed map of hits. 
  - In case we want to scale redis horizontally, Redis Cluster would 
    be preferred since it will give a better performance as the sync between master nodes is eventually consistent.
  - `allkeys-lru` as eviction policy so the availability is kept. 
  > Note: There are several key/value stores that could be used as well like bbolt (boltdb). At the end I choice Redis 
  >  because is more mature and I already have experience and some success stories behind using it.
- Local memory on each service instance will contain an eventually consistent list of the most recent hits (LRU).
- Background processes will be updating both storage asynchronously in the following way:
  - Every time a hit is received, it is stored in the Local memory.
  - The rate limit algorithm is going to read from local memory. In case no data is present, it queries Redis.
    > Note: It would be nice to assume that in case no data is present, no limit should be apply. However, this is not possible
    > because the cache will not contain a copy of the whole database but just a list of most recently used. Otherwise each service
    > instance will require a potentially big a mount of memory (at least the same as Redis). 
  - The hit wil not be directly persisted in Redis but eventually. The client should not wait such write.

## Rate limit algorithm

The chosen algorithm for calculating the rate limit is "Slide window of timestamps" (this is how I called).
Please see more details in the [algorithm](README.md#algorithm) section of the docs readme.  
  
# Discarded ideas

### Architecture

- In order to distribute the hits avoiding the needs for a single datastore, I though about adding a proxy in front of
  the services fleet in order to shard the calls, so each service instance will just keep a local in memory database as 
  all the hits for a given source will be always hitting the same instance. Two problems related:
    1. As the rate limiter service should be agnostic enough for providing a 
      global rate limiting experience, there is no field to consider for the shard algorithm. In a more precise scenario we 
      could shard by customer id so we would know that a given customer will always hit the same node.
    2. Resharding should be performed on each scale (up or down). 
    3. The proxy could become a single point of failure.
- In addition to the previous point, I also considered using sticky sessions as alternative to the sharding, but the 
  problems are very similar.

### Rate limit algorithm

#### Token bucket

At a first glance, the [Token bucket](https://en.wikipedia.org/wiki/Token_bucket), AKA [Leaky bucket as a meter](https://en.wikipedia.org/wiki/Leaky_bucket#As_a_meter), seems to be the most used algorithm 
in rate limit services. Here a brief explanation about the algorithm applied to our rate limit service:

- Each combination of owner + resource would have a bucket associated.
- Such bucket will contain tokens.
- Every time it receives a hit, the number of tokens should be checked:
  - If there are no tokens, the rate limit should apply.
  - Otherwise, a token get subtracted from the bucket.
- From time to time (the rate you want to set), the tokens should be refilled.

- Pros:
  - Space: It can be represented as a counter (integer) per owner + resource.
- Cons:
  - Need a process refilling the buckets periodically.
  - The algorithm is not slide window base: It could lead in to the problem of allowing eventually the double of hits 
    that it should. Explanation:
    - Let's say the bucket is full (100 tokens) at 09:00:00. 
    - At 09:00:59 100 hits are received and allowed, so the bucket gets empty.
    - At 09:01:00 the bucket gets refilled.
    - At 09:01:01 100 hits are received and allowed. 
    - Problem: 200 hits were performed in less than 3 seconds. 
    
I discarded this algorithm specially because of the second cons.

#### Token bucket + storing timestamp

The idea behind this alternative would be storing number of tokens and a timestamp of the last refill operation:

- When a hit is received, we fetch such timestamp in order to calculate the number of tokens that, in theory, has 
  been refilled since the last hit.
  
- Pros:
  - The refill process is not needed anymore.
  - Space needed is still small.
- Cons:
  - Still no slide window based.

I discarded this algorithm specially because of the cons.