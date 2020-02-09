package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/subcommands_test/grpc/pb"
	"github.com/subcommands_test/grpc/provider"
	"google.golang.org/grpc"
)

func TestCli(t *testing.T) {
	wg := sync.WaitGroup{}
	cmd := exec.Command("build/cliprov")
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	in, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	reader, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	buff := bufio.NewReader(reader)
	wg.Add(1)
	go func() {
		err = cmd.Run()
		if err != nil {
			t.Log(err, errOut.String())
		}
		wg.Done()
	}()
	_, err = fmt.Fprintln(in, "Kevin")
	if err != nil {
		t.Fatal(err, errOut.String())
	}
	response, err := buff.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	response = strings.TrimSpace(response)
	if response != "Hello, Kevin!" {
		t.Fatalf("Invalid output '%s' - %s", response, errOut.String())
	}
	err = cmd.Process.Kill()
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}

func TestWeb(t *testing.T) {
	srv := &http.Server{Addr: ":8080"}

	done := make(chan struct{})
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			t.Log(err)
		}
		close(done)
	}()
	resp, err := http.Get("http://localhost:8080?params=Kevin")
	if err != nil {
		t.Fatal(err)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	response := strings.Trim(string(respBody), " \n")
	if response != "Hello, Kevin!" {
		t.Fatalf("invalid response: '%s'", response)
	}
	err = srv.Shutdown(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	<-done
}

func BenchmarkCli(b *testing.B) {
	wg := sync.WaitGroup{}
	cmd := exec.Command("build/cliprov")
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	in, err := cmd.StdinPipe()
	if err != nil {
		b.Fatal(err)
	}
	reader, err := cmd.StdoutPipe()
	if err != nil {
		b.Fatal(err)
	}
	buff := bufio.NewReader(reader)
	wg.Add(1)
	go func() {
		_ = cmd.Run()
		wg.Done()
	}()
	waitc := make(chan struct{})
	go func() {
		defer close(waitc)
		for i := 0; i < b.N; i++ {
			response, err := buff.ReadString('\n')
			if err != nil {
				b.Log(err)
				b.Fail()
				return
			}
			response = strings.TrimSpace(response)
			if response != "Hello, Kevin!" {
				b.Logf("Invalid output '%s' - %s", response, errOut.String())
				b.Fail()
				return
			}
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = fmt.Fprintln(in, "Kevin")
		if err != nil {
			b.Fatal(err, errOut.String())
		}
	}
	select {
	case <-waitc:
	case <-time.After(time.Second):
		b.Log("didn't receive all responses")
		b.Fail()
	}
	b.StopTimer()
	if cmd.Process != nil {
		err = cmd.Process.Kill()
		if err != nil {
			b.Fatal(err)
		}
	}
	wg.Wait()
}

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s!", r.URL.Query().Get("params"))
	})
}

func BenchmarkWeb(b *testing.B) {
	srv := &http.Server{Addr: ":8080"}

	done := make(chan struct{})
	go func() {
		srv.ListenAndServe()
		close(done)
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Get("http://localhost:8080?params=Kevin")
		if err != nil {
			b.Fatal(err)
		}
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			b.Fatal(err)
		}
		response := strings.Trim(string(respBody), " \n")
		if response != "Hello, Kevin!" {
			b.Fatalf("invalid response: '%s'", response)
		}
	}
	b.StopTimer()
	err := srv.Shutdown(context.Background())
	if err != nil {
		b.Fatal(err)
	}
	select {
	case <-done:
	case <-time.NewTimer(100 * time.Millisecond).C:
		b.Fatal("didn't shutdown server properly")
	}
}

func BenchmarkGrpc(b *testing.B) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 8082))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterCommandServer(grpcServer, &provider.CommandProviderServer{})
	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			b.Log(err)
		}
	}()

	conn, err := grpc.Dial(":8082", grpc.WithInsecure())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewCommandClient(conn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Handle(context.Background(), &pb.CommandArguments{
			Args: []string{"Kevin"},
		})
		if err != nil {
			b.Fatal(err)
		}
		if resp.Result != "Hello, Kevin!" {
			b.Fatalf("invalid result: %s", resp.Result)
		}
	}
	b.StopTimer()
	grpcServer.Stop()
}

func BenchmarkGrpcStream(b *testing.B) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 8082))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterCommandServer(grpcServer, &provider.CommandProviderServer{})
	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			b.Log(err)
		}
	}()

	conn, err := grpc.Dial(":8082", grpc.WithInsecure())
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewCommandClient(conn)

	stream, err := client.HandleStream(context.Background())
	if err != nil {
		b.Fatal(err)
	}

	// Start listener for recv
	waitc := make(chan struct{})
	go func() {
		defer close(waitc)
		for i := 0; i < b.N; i++ {
			resp, err := stream.Recv()
			if err != nil {
				b.Log(err)
				b.Fail()
				break
			}
			if resp.Result != "Hello, Kevin!" {
				b.Logf("invalid result: %s", resp.Result)
				b.Fail()
				break
			}
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = stream.Send(&pb.CommandArguments{
			Args: []string{"Kevin"},
		})
		if err != nil {
			b.Fatal(err)
		}
	}
	select {
	case <-waitc:
	case <-time.After(time.Second):
		b.Log("didn't receive all responses")
		b.Fail()
	}
	b.StopTimer()
	stream.CloseSend()
	grpcServer.Stop()
}
