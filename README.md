# ratio : A simple distributed rate limiter made in go

## Overview [![Build Status](https://travis-ci.org/smoya/ratio.svg?branch=master)](https://travis-ci.org/smoya/ratio) [![Go Report Card](https://goreportcard.com/badge/github.com/smoya/ratio)](https://goreportcard.com/report/github.com/smoya/ratio)

Ratio is a distributed rate limiter designed for being used by any of your webservices.

It has been designed based on the following goals:

- Minimal response times.
- Requests can come from multiple sources and services.
- Limits don't have to be strictly enforced. Consistency can be eventual.

## Installation

### Docker

A part from the `Dockerfile`, this project also includes a `docker-compose.yml` in order to spin up a basic set of 
`ratio` + `redis` as storage.

```bash
docker-compose up
```

The GRPC server will be up and running, accessible via `localhost:50051`.
The default limit is `100/m`, but you can change it directly in the `docker-compose.yml` file.

### Kubernetes

```bash
kubectl apply -f deploy/kubernetes
```

It will create a service with only one instance. Please modify the files in [deploy/kubernetes](deploy/kubernetes) 
accordingly with your needs.

The default limit is `100/m`, but you can change it directly in the `deployment.yaml` file.

> Note: For simplification, the ratio is deployed with the `inmemory` slide window storage. We encourage to deploy it 
> with Redis as Storage, as the `inmemory` is not meant to be used in production.

#### Development

[Skaffold](https://github.com/GoogleContainerTools/skaffold) can build a Docker image locally and deploy it to 
your cluster without pushing it to any repository. Please install it following the [
official docs](https://skaffold.dev/docs/getting-started/#installing-skaffold)

Then run:

```bash
skaffold run
```

## Documentation

Please find ratio documentation in the [/docs](/docs) directory.

## Build

This project was tested and compiled with Go 1.12. Even though it will probably compile in lower versions, we strongly 
recommend using, at least, the mentioned one.

### Dependencies

This project manage its dependencies via [Go modules](https://github.com/golang/go/wiki/Modules).

Go modules automatically fetches the dependencies on compile time so no extra step should be performed.
However, make sure your `GO111MODULE` env var is set to `true`, otherwise go modules will not be activated.

### GRPC and Protobuf 

In order to compile protobuf files, please download the `protoc` compiler. Please find the instructions 
[here](https://github.com/protocolbuffers/protobuf/blob/master/README.md#protocol-compiler-installation).

The protobuf files can be found in [/api/proto](/api/proto) and the Go code can be regenerated by simply run:

```go
go generate ./...
``` 

#### Command line client
In case you want to do some calls to the server, you can install the `grpc_cli` tool from 
[here](https://github.com/grpc/grpc/blob/master/doc/command_line_tool.md). 

You can find an example at [/docs/](/docs/README.md#grpc-command-line-test-client)

## Load Testing

```
Summary:
  Count:        200
  Total:        398.93 ms
  Slowest:      152.28 ms
  Fastest:      16.55 ms
  Average:      72.22 ms
  Requests/sec: 501.35

Response time histogram:
  16.551 [1]    |∎
  30.124 [31]   |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  43.696 [21]   |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  57.269 [16]   |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  70.842 [25]   |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  84.415 [31]   |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  97.988 [23]   |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  111.561 [27]  |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  125.134 [13]  |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  138.707 [11]  |∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  152.280 [1]   |∎

Latency distribution:
  10% in 24.72 ms
  25% in 42.33 ms
  50% in 73.22 ms
  75% in 99.50 ms
  90% in 119.49 ms
  95% in 130.95 ms
  99% in 136.96 ms

Status code distribution:
  [OK]   200 responses

```

[ghz](https://ghz.sh) is a very useful load testing tool for GRPC.

Run:

```bash
ghz --config .ghz.json
```

## TODO

- Support for Redis Cluster. Read the reasons behind [here](/docs/decisions.md#storage)
- Support for distributed in memory storage, with autodiscovery of `ratio` instances and broadcast between them.
- Interface the logger and use a better implementation like [zap](https://github.com/uber-go/zap).
- Instrument the server with a Prometheus endpoint exposing basic metrics.
- Add benchmarks. 
- Improve the K8s deploy adding different use cases:
    - More than one instance behind a Load Balancer.
    - Provide helm chart.
- Allow the caller to configure its own Limits. 

## License

MIT.
