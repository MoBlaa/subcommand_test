package main

import (
	"fmt"
	"os"
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
	select {
	case args, ok := <-prov.input:
		if !ok {
			break
		}
		if len(args) == 0 {
			return fmt.Errorf("zero arguments given")
		}
		prov.output <- fmt.Sprintf("Hello, %s!", args[0])
	}
	return nil
}

func main() {
	in := make(chan []string, 1)
	done := make(chan struct{})
	provider := NewCliProvider(in)
	go provider.Start()
	go func() {
		outChan := provider.Output()
		for out := range outChan {
			fmt.Println(out)
		}
		close(done)
	}()
	in <- os.Args[1:]
	close(in)
	<-done
}
