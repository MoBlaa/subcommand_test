package provider

import (
	"context"
	"fmt"
	"io"

	"github.com/subcommands_test/grpc/pb"
)

type CommandProviderServer struct {
}

func (prov *CommandProviderServer) handle(args []string) string {
	var message string
	if len(args) == 0 {
		message = "Hello!"
	} else {
		message = fmt.Sprintf("Hello, %s!", args[0])
	}
	return message
}

func (prov *CommandProviderServer) Handle(ctx context.Context, arg *pb.CommandArguments) (*pb.CommandResult, error) {
	return &pb.CommandResult{
		Result: prov.handle(arg.Args),
	}, nil
}

func (prov *CommandProviderServer) HandleStream(stream pb.Command_HandleStreamServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		err = stream.Send(&pb.CommandResult{
			Result: prov.handle(in.Args),
		})
		if err != nil {
			return err
		}
	}
}
