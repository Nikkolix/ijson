# Makefile for github.com/Nikkolix/ijson
# Usage examples:
#   make test         # run all tests
#   make cover-html   # generate coverage.out and coverage.html
#   make fmt vet      # format and vet all packages

GO ?= go
PKG ?= ./...

.PHONY: help test test-v cover cover-html fmt vet tidy build bench clean

help:
	@echo "Available targets:"
	@echo "  test        - Run all tests"
	@echo "  test-v      - Run all tests (verbose)"
	@echo "  cover       - Run tests with coverage and write coverage.out"
	@echo "  cover-html  - Generate coverage.html from coverage.out"
	@echo "  fmt         - Run go fmt on all packages"
	@echo "  vet         - Run go vet on all packages"
	@echo "  tidy        - Run go mod tidy"
	@echo "  build       - Build all packages"
	@echo "  bench       - Run benchmarks"
	@echo "  clean       - Clean test cache and coverage artifacts"

test:
	$(GO) test $(PKG)

test-v:
	$(GO) test -v $(PKG)

cover:
	$(GO) test -covermode=atomic -coverprofile=coverage.out $(PKG)

cover-html: cover
	$(GO) tool cover -html=coverage.out -o coverage.html

fmt:
	$(GO) fmt $(PKG)

vet:
	$(GO) vet $(PKG)

tidy:
	$(GO) mod tidy

build:
	$(GO) build $(PKG)

bench:
	$(GO) test -bench=. -benchmem $(PKG)

clean:
	$(GO) clean -testcache
	- powershell -NoProfile -Command "if (Test-Path coverage.out) { Remove-Item -Force coverage.out }; if (Test-Path coverage.html) { Remove-Item -Force coverage.html }"
	- rm -f coverage.out coverage.html >/dev/null 2>&1 || true

