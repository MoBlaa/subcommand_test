package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/subcommands_test/grpc/pb"
	"google.golang.org/grpc"
)

type testHandler func(in io.WriteCloser, reader *bufio.Reader, errOut *bytes.Buffer)

func testStart(t *testing.T, command []string, iteration testHandler) {
	cmd := exec.Command(command[0], command[1:]...)
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

	waitc := make(chan struct{})
	go func() {
		defer close(waitc)
		err = cmd.Run()
		if err != nil {
			t.Log(err, errOut.String())
		}
	}()

	<-time.After(100 * time.Millisecond)

	iteration(in, buff, &errOut)

	select {
	case <-waitc:
	case <-time.After(time.Second):
		err = cmd.Process.Kill()
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-waitc:
		case <-time.After(time.Second):
			t.Fatal("failed to kill process")
		}
	}
}

func TestCli(t *testing.T) {
	testStart(t, []string{"build/cliprov"}, func(in io.WriteCloser, out *bufio.Reader, errOut *bytes.Buffer) {
		_, err := fmt.Fprintln(in, "Kevin")
		if err != nil {
			t.Error(err, errOut.String())
			return
		}
		response, err := out.ReadString('\n')
		if err != nil {
			t.Error(err)
			return
		}
		response = strings.TrimSpace(response)
		if response != "Hello, Kevin!" {
			t.Errorf("Invalid output '%s' - %s", response, errOut.String())
		}
	})
}

func TestWeb(t *testing.T) {
	testStart(t, []string{"build/webprov"}, func(in io.WriteCloser, out *bufio.Reader, errOut *bytes.Buffer) {
		var err error
		var resp *http.Response
		for i := 0; i < 5; i++ {
			<-time.After(250 * time.Millisecond)
			resp, err = http.Get("http://localhost:8080?params=Kevin")
			if err == nil {
				break
			}
		}
		if err != nil {
			t.Error(err)
			return
		}
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
			return
		}
		response := strings.Trim(string(respBody), " \n")
		if response != "Hello, Kevin!" {
			t.Errorf("invalid response: '%s'", response)
		}
	})
}

func TestGrpc(t *testing.T) {
	testStart(t, []string{"build/grpcprov"}, func(in io.WriteCloser, out *bufio.Reader, errOut *bytes.Buffer) {
		var conn *grpc.ClientConn
		var err error
		for i := 0; i < 5; i++ {
			<-time.After(250 * time.Millisecond)
			conn, err = grpc.Dial(":8080", grpc.WithInsecure())
			if err == nil {
				break
			}
		}
		if err != nil {
			t.Error(err, errOut.String())
			return
		}
		defer conn.Close()
		client := pb.NewCommandClient(conn)

		resp, err := client.Handle(context.Background(), &pb.CommandArguments{
			Args: []string{"Kevin"},
		})
		if err != nil {
			t.Error(err, errOut.String())
			return
		}
		if resp.Result != "Hello, Kevin!" {
			t.Errorf("invalid result: %s", resp.Result)
		}

		stream, err := client.HandleStream(context.Background())
		if err != nil {
			t.Error(err, errOut.String())
			return
		}
		err = stream.Send(&pb.CommandArguments{
			Args: []string{"Kevin"},
		})
		if err != nil {
			t.Error(err, errOut.String())
			return
		}
		if resp.Result != "Hello, Kevin!" {
			t.Errorf("invalid result: %s", resp.Result)
		}
	})
}

func benchStart(b *testing.B, command []string, iteration testHandler) {
	cmd := exec.Command(command[0], command[1:]...)
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

	waitc := make(chan struct{})
	go func() {
		defer close(waitc)
		err = cmd.Run()
		if err != nil {
			b.Log(err, errOut.String())
		}
	}()

	<-time.After(100 * time.Millisecond)

	b.ResetTimer()
	iteration(in, buff, &errOut)
	b.StopTimer()

	select {
	case <-waitc:
	case <-time.After(time.Second):
		err = cmd.Process.Kill()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCli(b *testing.B) {
	benchStart(b, []string{"build/cliprov"}, func(in io.WriteCloser, out *bufio.Reader, errOut *bytes.Buffer) {
		waitc := make(chan struct{})
		go func() {
			defer close(waitc)
			for i := 0; i < b.N; i++ {
				response, err := out.ReadString('\n')
				if err != nil {
					b.Error(err, errOut.String())
					return
				}
				response = strings.TrimSpace(response)
				if response != "Hello, Kevin!" {
					b.Errorf("Invalid output '%s' - %s", response, errOut.String())
					return
				}
			}
		}()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := fmt.Fprintln(in, "Kevin")
			if err != nil {
				b.Error(err, errOut.String())
				return
			}
		}
		select {
		case <-waitc:
		case <-time.After(time.Second):
			b.Error("didn't receive all responses")
		}
		b.StopTimer()
	})
}

func BenchmarkWeb(b *testing.B) {
	benchStart(b, []string{"build/webprov"}, func(in io.WriteCloser, out *bufio.Reader, errOut *bytes.Buffer) {
		// Wait till subcommand is ready
		var err error
		for i := 0; i < 5; i++ {
			<-time.After(250 * time.Millisecond)
			_, err = http.Get("http://localhost:8080?params=Kevin")
			if err == nil {
				break
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := http.Get("http://localhost:8080?params=Kevin")
			if err != nil {
				b.Error(err)
				return
			}
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				b.Error(err)
				return
			}
			response := strings.Trim(string(respBody), " \n")
			if response != "Hello, Kevin!" {
				b.Errorf("invalid response: '%s'", response)
				return
			}
		}
	})
}

func BenchmarkGrpc(b *testing.B) {
	benchStart(b, []string{"build/grpcprov"}, func(in io.WriteCloser, out *bufio.Reader, errOut *bytes.Buffer) {
		var conn *grpc.ClientConn
		var err error
		for i := 0; i < 5; i++ {
			<-time.After(250 * time.Millisecond)
			conn, err = grpc.Dial(":8080", grpc.WithInsecure())
			if err == nil {
				break
			}
		}
		if err != nil {
			b.Error(err)
			return
		}
		defer conn.Close()
		client := pb.NewCommandClient(conn)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := client.Handle(context.Background(), &pb.CommandArguments{
				Args: []string{"Kevin"},
			})
			if err != nil {
				b.Error(err)
				return
			}
			if resp.Result != "Hello, Kevin!" {
				b.Errorf("invalid result: %s", resp.Result)
				return
			}
		}
	})
}

func BenchmarkGrpcStream(b *testing.B) {
	benchStart(b, []string{"build/grpcprov"}, func(in io.WriteCloser, out *bufio.Reader, errOut *bytes.Buffer) {
		var err error
		var conn *grpc.ClientConn
		for i := 0; i < 5; i++ {
			<-time.After(250 * time.Millisecond)
			conn, err = grpc.Dial(":8080", grpc.WithInsecure())
			if err == nil {
				break
			}
		}
		if err != nil {
			b.Error(err, errOut.String())
			return
		}
		defer conn.Close()
		client := pb.NewCommandClient(conn)

		stream, err := client.HandleStream(context.Background())
		if err != nil {
			b.Error(err, errOut.String())
			return
		}

		// Start listener for recv
		waitc := make(chan struct{})
		go func() {
			defer close(waitc)
			for i := 0; i < b.N; i++ {
				resp, err := stream.Recv()
				if err != nil {
					b.Error(err)
					break
				}
				if resp.Result != "Hello, Kevin!" {
					b.Errorf("invalid result: %s", resp.Result)
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
				b.Error(err)
				break
			}
		}
		select {
		case <-waitc:
		case <-time.After(time.Second):
			b.Error("didn't receive all responses")
		}
		stream.CloseSend()
	})
}
