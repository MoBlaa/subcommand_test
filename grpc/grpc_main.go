package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/subcommands_test/grpc/provider"
	"google.golang.org/grpc"
)

type commandProviderServer struct {
}

func (prov *commandProviderServer) Handle(ctx context.Context, arg *provider.CommandArguments) (*provider.CommandResult, error) {
	var message string
	if len(arg.Args) == 0 {
		message = "Hello!"
	} else {
		message = fmt.Sprintf("Hello, %s!", arg.Args[0])
	}
	return &provider.CommandResult{
		Result: message,
	}, nil
}

func main() {
	port := flag.Int("port", 8080, "Port to serve on")

	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	provider.RegisterCommandServer(grpcServer, &commandProviderServer{})
	log.Fatal(grpcServer.Serve(lis))
}
