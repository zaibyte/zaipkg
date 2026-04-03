# zaipkg

[![Unit Test](https://github.com/templexxx/zaipkg/actions/workflows/unit-test.yml/badge.svg)](https://github.com/templexxx/zaipkg/actions/workflows/unit-test.yml)

A collection of Go toolkit packages for distributed system development.

## Install

```bash
go get github.com/zaibyte/zaipkg@latest
```

## Usage

Import the package(s) you need:

```go
import "github.com/zaibyte/zaipkg/<package>"
```

## Testing

Run all tests:

```bash
make test
```

Or run directly with Go:

```bash
go test -race -cover ./...
```
