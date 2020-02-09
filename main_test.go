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
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	done := make(chan struct{})
	go func() {
		cmd := exec.CommandContext(ctx, "build/webprov")
		var out bytes.Buffer
		var errOut bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errOut
		if err := cmd.Run(); err != nil {
			t.Log(err, out.String(), errOut.String())
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
	cancel()
	<-done
}

func TestGrpc(t *testing.T) {
	cmd := exec.Command("build/grpcprov", "-port", "8081")
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	done := make(chan struct{})
	go func() {
		err := cmd.Run()
		if err != nil {
			t.Log(err, out.String(), errOut.String())
		}
		close(done)
	}()
	conn, err := grpc.Dial(":8081", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewCommandClient(conn)
	resp, err := client.Handle(context.Background(), &pb.CommandArguments{
		Args: []string{"Kevin"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Result != "Hello, Kevin!" {
		t.Fatalf("invalid result: %s", resp.Result)
	}
	if cmd.Process != nil {
		err = cmd.Process.Kill()
		if err != nil {
			t.Fatal(err)
		}
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
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = fmt.Fprintln(in, "Kevin")
		if err != nil {
			b.Fatal(err, errOut.String())
		}
		response, err := buff.ReadString('\n')
		if err != nil {
			b.Fatal(err)
		}
		response = strings.TrimSpace(response)
		if response != "Hello, Kevin!" {
			b.Fatalf("Invalid output '%s' - %s", response, errOut.String())
		}
	}
	b.StopTimer()
	err = cmd.Process.Kill()
	if err != nil {
		b.Fatal(err)
	}
	wg.Wait()
}

func BenchmarkWeb(b *testing.B) {
	done := make(chan struct{})
	cmd := exec.Command("build/webprov")
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	go func() {
		if err := cmd.Run(); err != nil {
			b.Log(err, out.String(), errOut.String())
		}
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
	select {
	case <-done:
	case <-time.NewTimer(100 * time.Millisecond).C:
		err := cmd.Process.Kill()
		if err != nil {
			b.Fatalf("failed to kill process: %v", err)
		}
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
