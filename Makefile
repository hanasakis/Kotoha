.PHONY: run build test lint eval

run:
	go run ./cmd/server

build:
	go build -o bin/kotoha ./cmd/server

test:
	go test ./... -v

eval:
	go test -v -run "Eval" ./internal/search/ ./internal/agent/
