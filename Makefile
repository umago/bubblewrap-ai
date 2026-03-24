BINARY  := bwai
CMD     := ./cmd/bwai
BIN_DIR := bin

VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -ldflags "-X main.version=$(VERSION)"

.PHONY: all build clean test lint install

all: build

build:
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY) $(CMD)

install:
	go install $(LDFLAGS) $(CMD)

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BIN_DIR)/$(BINARY)
