.PHONY: run build test lint

run:
	go run ./cmd/server

build:
	go build -o bin/kotoha ./cmd/server

test:
	go test ./... -v
