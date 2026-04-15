BINARY  := bwai
CMD     := ./cmd/bwai
BIN_DIR := bin

VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -ldflags "-X main.version=$(VERSION)"

.PHONY: all build clean test lint fmt install install-hooks

all: build

build:
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY) $(CMD)

install:
	go install $(LDFLAGS) $(CMD)

test:
	go test ./...

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || { echo "The following files are not formatted:"; gofmt -l .;  exit 1; }

lint:
	golangci-lint run ./...

install-hooks:
	cp scripts/hooks/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit scripts/check.sh

clean:
	rm -rf $(BIN_DIR)/$(BINARY)
