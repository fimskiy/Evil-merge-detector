BINARY_NAME=evilmerge
VERSION?=dev
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

.PHONY: build test lint clean install

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/evilmerge

install:
	go install $(LDFLAGS) ./cmd/evilmerge

test:
	go test -v ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

run:
	go run $(LDFLAGS) ./cmd/evilmerge $(ARGS)
