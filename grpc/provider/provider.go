package provider

import (
	"context"
	"fmt"

	"github.com/subcommands_test/grpc/pb"
)

type CommandProviderServer struct {
}

func (prov *CommandProviderServer) Handle(ctx context.Context, arg *pb.CommandArguments) (*pb.CommandResult, error) {
	var message string
	if len(arg.Args) == 0 {
		message = "Hello!"
	} else {
		message = fmt.Sprintf("Hello, %s!", arg.Args[0])
	}
	return &pb.CommandResult{
		Result: message,
	}, nil
}
