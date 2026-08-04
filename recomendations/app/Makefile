.PHONY: all run demo lint check build

BINARY := ./bin/app

all: check lint build

run:
	go run main.go

demo:
	RENDER_STATIC=1 go run main.go
	mkdir -p ../../docs
	cp ../static/index_demo.html ../../docs/

check:
	go fmt ./...
	go vet ./...
	go test ./...

lint:
	golangci-lint run

build:
	mkdir -p ./bin
	go build -o $(BINARY) main.go