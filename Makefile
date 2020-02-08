
build_cli:
	@go build -o build/cliprov cli/cli_main.go

build_web:
	@go build -o build/webprov web/web_main.go

test: build_cli build_web
	@go test

bench: build_cli build_web
	@go test -bench .