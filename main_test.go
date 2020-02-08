package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestCli(t *testing.T) {
	cmd := exec.Command("build/cliprov", "Kevin")
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	err := cmd.Run()
	if err != nil {
		t.Fatal(err, out.String(), errOut.String())
	}
	if strings.Trim(out.String(), " \n") != "Hello, Kevin!" {
		t.Fatalf("Invalid output '%s'", out.String())
	}
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

func BenchmarkCli(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cmd := exec.Command("build/cliprov", "Kevin")
		var out bytes.Buffer
		var errOut bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errOut
		err := cmd.Run()
		if err != nil {
			b.Fatal(err, out.String(), errOut.String())
		}
		if strings.Trim(out.String(), " \n") != "Hello, Kevin!" {
			b.Fatalf("Invalid output '%s'", out.String())
		}
	}
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
