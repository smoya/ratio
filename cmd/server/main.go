//go:generate protoc -I ../../api/proto --go_out=plugins=grpc:../../api/proto ../../api/proto/ratio.proto

package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/smoya/ratio/pkg/rate"

	"github.com/kelseyhightower/envconfig"

	"github.com/smoya/ratio/internal/server"

	"google.golang.org/grpc/reflection"

	ratio "github.com/smoya/ratio/api/proto"
	"google.golang.org/grpc"
)

type config struct {
	Port              int           `default:"50051"`
	ConnectionTimeout time.Duration `default:"1s" help:"Timeout for all incoming connections" split_words:"true"`
	Storage           string        `default:"redis://redis:6379/0"`
}

func main() {
	var c config
	err := envconfig.Process("ratio", &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", strconv.Itoa(c.Port)))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.ConnectionTimeout(c.ConnectionTimeout))
	reflection.Register(s)

	storage, err := rate.NewSlideWindowStorageFromDSN(c.Storage)
	if err != nil {
		log.Fatal(err.Error())
	}

	grpcServer := server.NewGRPC(
		rate.NewLimit(time.Minute, 5),
		rate.SlideWindowRateLimiter(storage, true),
	)

	ratio.RegisterRateLimitServiceServer(s, grpcServer)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
