package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")

	flag.Parse()

	srv := http.Server{Addr: fmt.Sprintf(":%d", *port)}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s!", r.URL.Query().Get("params"))
	})

	waitc := make(chan struct{})
	go func() {
		defer close(waitc)
		err := srv.ListenAndServe()
		if err == http.ErrServerClosed {
			return
		}
		if err != nil {
			log.Println(err)
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
		err := srv.Shutdown(context.Background())
		if err != nil {
			log.Fatalf("failed to shutdown server: %v", err)
		}
		select {
		case <-waitc:
		case <-time.After(time.Second):
			log.Fatal("server wasn't closed after shutdown")
		}
	}
}
