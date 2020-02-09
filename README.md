# Go Subcommand Implementation Performance Tests

With this project i like to test different ways of implementing a multi-language plugin system. This is a investigation for a Chat command Project i'm building. The general setup is as follows:

- The __Hub__ connects to __CommandProviders__
- The __Hub__ delegates command invocations to the __CommandProviders__
- A __CommandProvider__ handles a single command.

In this project the [main_test.go](main_test.go) is taking the role of the hub and has Benchmark tests for the different ways of implementation.

Before explaining the different implementations i'd like to make one thing clear. The hub is not needed to start the subcommands with `os/exec`. This is just for simplification. In my own project i may use this so the end-user doesn't have to start the subcommands himself. But for this performance test every subcommand will be started with `os/exec` to have an equal testbed.

Currently the following implementations are present:

- [web](web/) - Uses `net/http` and uses a client-server architecture to communicate between the hub and the subcommand. This is just for comparison with a naive http implementation.
- [cli](cli/) - Uses `os/exec` to start and communicate between hub and subcommands. This actually requires the hub to start the subcommand with `os/exec` to be able to communicate with it.
- [grpc](grpc/) - Uses the [grpc](https://grpc.io/) library for communication between subcommand and hub. This implementation contains a rpc with streaming and one without as well as support for unix sockets and tcp.

## Current results

These benchmarks are performed on an really old iMac (2010). These will be updated with more specific hardware information. Till then feel free to download the source and perform the tests by yourself.

Versions: Go 1.13.7, grpc v1.27.1

| Benchmark                    | Iterations | Speed        | Memory Usage | Allocation   |
| ---------------------------- | ---------- | ------------ | ------------ | ------------ |
| BenchmarkCli-2               | 201109     | 7278 ns/op   | 32 B/op      | 2 allocs/op  |
| BenchmarkWeb-2               | 4857       | 252173 ns/op | 3552 B/op    | 47 allocs/op |
| BenchmarkGrpcTcp-2           | 5744       | 230338 ns/op | 4844 B/op    | 98 allocs/op |
| BenchmarkGrpcSocket-2        | 6807       | 201812 ns/op | 4840 B/op    | 98 allocs/op |
| BenchmarkGrpcTcp_Stream-2    | 155290     | 7335 ns/op   | 532 B/op     | 15 allocs/op |
| BenchmarkGrpcSocket_Stream-2 | 188743     | 7541 ns/op   | 506 B/op     | 15 allocs/op |

## What not to test

While exploring the different ways to implement something like this also want to list things i don't want to test and why.

- Websocket - Websockets allow bidirectional communication, but reconnecting (and some other little things) have to be handled in order to use it properly. I would expect the Websocket implementation to have a performance between GRPC Streaming and the HTTP implementation.
- `go/exec` with starting a process every time a command has to be delegated - This was the basic implementation before using `os.Stdin` and `os.Stdout` and was by far slower than the basic http version.

## My Conclusion

Currently (09.02.2020) the CLI version seems to be the best fit for my project. As i want to control the installation and execution process of the subcommands i'll be using `os/exec` anyways. As i also want to provide the most simple environment for others to develop their subcommands with the `os/exec` implementation also seems most straight forward. No heavy handling of a framework like grpc. Just implement something that reads from stdin and write to stdout and it works.

But `grpc` also seems great for hosted environments. If i want to provide a server infrastructure, where people can host their own chat command listener or even Bots, this may nice as spawning another container providing functionality through `grpc` may be simpler than installing a cli script and wiring it into the hub.

I may think about implementing an abstraction over the actual implementation so i can provide local subcommands (with `os/exec`) and remote ones (with `grpc` and `tcp`). 