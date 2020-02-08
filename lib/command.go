package lib

// Command handles a command invocation with its Handle method.
type Command interface {
	Handle(args []string) string
}

// CommandFunc represents a Command as function directly.
type CommandFunc func(args []string) string
