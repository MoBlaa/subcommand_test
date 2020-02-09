package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	waitc := make(chan struct{})
	go func() {
		defer close(waitc)
		err := grpcServer.Serve(lis)
		if err == grpc.ErrServerStopped {
			return
		}
		if err != nil {
			log.Fatal(err)
		}
	}()

	sigs := make(chan os.Signal)
	waitsig := make(chan struct{})

	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		sig := <-sigs
		log.Printf("Signal received: %v\n", sig)
		close(waitsig)
	}()

	select {
	case <-waitc:
		// Server has been closed for any reason
	case <-waitsig:
		// Signal received, server has to be closed now
		grpcServer.Stop()
		select {
		case <-waitc:
		case <-time.After(time.Second):
			log.Fatal("server wasn't closed after stop")
		}
	}
}
