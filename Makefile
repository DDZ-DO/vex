.PHONY: build install test clean lint coverage run

BINARY=vex
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/vex

install:
	go install $(LDFLAGS) ./cmd/vex

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.txt ./...
	go tool cover -html=coverage.txt -o coverage.html

lint:
	golangci-lint run

clean:
	rm -rf bin/
	rm -f coverage.txt coverage.html

run: build
	./bin/$(BINARY)

# Development helpers
deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

# Build for multiple platforms
build-all: build-linux build-darwin

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 ./cmd/vex
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-linux-arm64 ./cmd/vex

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-amd64 ./cmd/vex
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64 ./cmd/vex
