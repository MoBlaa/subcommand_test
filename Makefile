
build_cli:
	@go build -o build/cliprov cli/cli_main.go

build_web:
	@go build -o build/webprov web/web_main.go

gen_grpc:
	@protoc -I grpc/provider grpc/provider/provider.proto --go_out=plugins=grpc:grpc/provider

build_grpc: gen_grpc
	@go build -o build/grpcprov grpc/grpc_main.go

test: build_cli build_web build_grpc
	@go test

bench: build_cli build_web build_grpc
	@go test -bench .