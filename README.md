# Go Subcommand Implementation Performance Tests

With this project i like to test different ways of implementing a multi-language plugin system. This is a investigation for a Chat command Project i'm building. The general setup is as follows:

- The __Hub__ connects to __CommandProviders__
- The __Hub__ delegates command invocations to the __CommandProviders__
- A __CommandProvider__ handles a single command.

In this project the [main_test.go](main_test.go) is taking the role of the hub and has Benchmark tests for the different ways of implementation.

Currently basic __HTTP Web__ and __CLI__ versions are implemented. The [web](web/) subdirectory contains the basic HTTP Web version which uses the `net/http` package. The [cli](cli/) subdirectory contains the CLI version which uses `os/exec` and `os.Stdin, os.Stdout` to communicate between hub and command provider. The [grpc](grpc/) subdirectory contains the grpc version supporting simple rpc and rpc streaming.

## Current results

These benchmarks are performed on an really old iMac (2010). These will be updated with more specific hardware information. Till then feel free to download the source and perform the tests by yourself.

| Benchmark             | Iterations | Speed        |
| --------------------- | ---------- | ------------ |
| BenchmarkCli-2        | 32542      | 33926 ns/op  |
| BenchmarkWeb-2        | 6307       | 183996 ns/op |
| BenchmarkGrpc-2       | 6385       | 183819 ns/op |
| BenchmarkGrpcStream-2 | 142116     | 10903 ns/op  |

## What not to test

While exploring the different ways to implement something like this also want to list things i don't want to test and why.

- Websocket - Websockets allow bidirectional communication, but reconnecting (and some other little things) have to be handled in order to use it properly. I would expect the Websocket implementation to have a performance between GRPC Streaming and the HTTP implementation.
- `go/exec` with starting a process every time a command has to be delegated - This was the basic implementation before using `os.Stdin` and `os.Stdout` and was by far slower than the basic http version.