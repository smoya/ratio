FROM golang:1.12-alpine as builder
RUN apk update && apk add build-base git ca-certificates
WORKDIR /go/src/github.com/smoya/ratio
ENV GO111MODULE on

# Add dependencies first to make use of docker cache
COPY go.mod .
COPY go.sum .
RUN go get ./...

COPY . .
RUN go build -o build/ratio cmd/server/main.go

FROM alpine:3.8
RUN apk update && apk add ca-certificates
COPY --from=builder /go/src/github.com/smoya/ratio/build/ratio ratio
EXPOSE 50051
ENTRYPOINT ["./ratio"]
