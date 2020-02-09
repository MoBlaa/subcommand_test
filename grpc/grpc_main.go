package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/subcommands_test/grpc/pb"
	"github.com/subcommands_test/grpc/provider"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 8080, "Port to serve on")

	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterCommandServer(grpcServer, &provider.CommandProviderServer{})
	log.Fatal(grpcServer.Serve(lis))
}
