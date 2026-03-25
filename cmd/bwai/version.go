package main

// version is set at build time via:
//
//	go build -ldflags "-X main.version=<tag>"
//
// The Makefile derives the value from `git describe --tags --always --dirty`.
// When built without ldflags (e.g. `go run`) it falls back to "dev".
var version = "dev"
