package main

import (
	"os"

	"github.com/subcommands_test/cli/lib"
)

func main() {
	provider := &lib.ReaderWriterProvider{
		Input:       os.Stdin,
		Output:      os.Stdout,
		HandlerFunc: lib.HelloProvider,
	}

	<-provider.Start()
}
