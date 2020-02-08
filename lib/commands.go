package lib

import (
	"fmt"
	"strings"
)

// EchoProvider joins the arguments.
func EchoProvider(args []string) string {
	return strings.Join(args, " ")
}

// HelloProvider returns a hello message and includes a name if given as first argument.
func HelloProvider(args []string) string {
	if len(args) == 0 {
		return "Hello!"
	}
	return fmt.Sprintf("Hello, %s!", args[0])
}
