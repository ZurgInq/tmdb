.PHONY: all run demo lint check build

BINARY := ./bin/app

all: check lint build

run:
	go run main.go

demo:
	RENDER_DEMO=1 go run main.go
	mkdir -p ../../docs
	cp ../static/demo/index.html ../../docs/index.html
	cp ../static/demo/series.html ../../docs/series.html

check:
	go fmt ./...
	go vet ./...
	go test ./...

lint:
	golangci-lint run

build:
	mkdir -p ./bin
	go build -o $(BINARY) main.go