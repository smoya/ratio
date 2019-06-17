FROM golang:1.12-alpine as builder
RUN apk update && apk add build-base git ca-certificates
WORKDIR /go/src/github.com/smoya/ratio
ENV GO111MODULE on

# Add dependencies first to make use of docker cache
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN make build

FROM alpine:3.9
RUN apk update && apk add ca-certificates
COPY --from=builder /go/src/github.com/smoya/ratio/bin/ratio ratio
EXPOSE 50051
ENTRYPOINT ["./ratio"]
