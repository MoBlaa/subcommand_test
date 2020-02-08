package lib

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ReaderWriterProvider implements the ReaderWriterProvider interface
// by using io.Stdin and io.Stdout for communication.
type ReaderWriterProvider struct {
	Input  io.Reader
	Output io.Writer

	Handler     Command
	HandlerFunc CommandFunc
}

func (prov *ReaderWriterProvider) listen(input <-chan []string) <-chan string {
	output := make(chan string)
	go func() {
		defer close(output)
	loop:
		for {
			select {
			case args, ok := <-input:
				if !ok {
					break loop
				}
				if len(args) == 0 {
					continue
				}
				var value string
				if prov.Handler == nil {
					if prov.HandlerFunc == nil {
						value = EchoProvider(args)
					} else {
						value = prov.HandlerFunc(args)
					}
				} else {
					value = prov.Handler.Handle(args)
				}
				output <- value
			}
		}
	}()
	return output
}

// proxy between the raw output writer and output channel
func (prov *ReaderWriterProvider) outputProxy(output <-chan string) <-chan struct{} {
	out := make(chan struct{})
	go func() {
		defer close(out)
		for out := range output {
			_, err := fmt.Fprintln(prov.Output, out)
			if err != nil {
				break
			}
		}
	}()
	return out
}

func (prov *ReaderWriterProvider) inputProxy() <-chan []string {
	out := make(chan []string)
	go func() {
		defer close(out)
		reader := bufio.NewReader(prov.Input)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			out <- strings.Split(strings.TrimSpace(line), " ")
		}
	}()
	return out
}

// Start the Provider.
func (prov *ReaderWriterProvider) Start() <-chan struct{} {
	input := prov.inputProxy()
	output := prov.listen(input)
	return prov.outputProxy(output)
}
