.PHONY: build test clean run

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	go build -ldflags "-X main.Version=$(VERSION)" -o launch-pilot ./cmd/launch-pilot

test:
	go test ./... -count=1

clean:
	rm -f launch-pilot

run: build
	./launch-pilot
