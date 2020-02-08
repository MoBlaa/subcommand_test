package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

type cliProvider struct {
	input  <-chan []string
	output chan string
}

func NewCliProvider(input <-chan []string) *cliProvider {
	return &cliProvider{
		input:  input,
		output: make(chan string),
	}
}

func (prov *cliProvider) Output() <-chan string {
	return prov.output
}

func (prov *cliProvider) Start() error {
	defer close(prov.output)
loop:
	for {
		select {
		case args, ok := <-prov.input:
			log.Printf("Got: %v", args)
			if !ok {
				break loop
			}
			if len(args) == 0 {
				return fmt.Errorf("zero arguments given")
			}
			prov.output <- fmt.Sprintf("Hello, %s!", args[0])
		}
	}
	return nil
}

func main() {
	in := make(chan []string)
	wg := sync.WaitGroup{}
	provider := NewCliProvider(in)
	go provider.Start()
	wg.Add(1)
	go func() {
		outChan := provider.Output()
		for out := range outChan {
			fmt.Println(out)
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			in <- strings.Split(strings.TrimSpace(line), " ")
		}
		close(in)
		wg.Done()
	}()

	// Listen to signals
	sigs := make(chan os.Signal, 1)
	done := make(chan struct{})

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Printf("Kill: %d\n", sig)
		close(in)
		close(done)
	}()
	go func() {
		wg.Wait()
		close(done)
	}()
	<-done
}
