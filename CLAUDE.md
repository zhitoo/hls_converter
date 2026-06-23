# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`github.com/zhitoo/hls_converter` — a Go CLI tool for HLS (HTTP Live Streaming) conversion. Currently in early/stub state.

## Commands

```bash
# Run
go run main.go

# Build
go build -o hls_converter .

# Test
go test ./...

# Run a single test
go test -run TestFunctionName ./path/to/package

# Add a dependency
go get <module>
```

## Architecture

Single-module Go project (`go 1.26`). Entry point is `main.go`. No packages or dependencies yet beyond the standard library.
