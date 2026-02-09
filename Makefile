.PHONY: help build test clean run install

VERSION=$(shell date +v%y.%m%d.%H%M)
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

help:
	@echo "Available targets:"
	@echo "  build    - Build mcp-mashup binary"
	@echo "  test     - Run tests"
	@echo "  clean    - Remove build artifacts"
	@echo "  run      - Build and run"
	@echo "  install  - Install to ~/.local/bin"

build:
	mkdir -p ./bin
	go build $(LDFLAGS) -o ./bin/mcp-mashup ./cmd/mcp-mashup

test:
	go test ./...

clean:
	rm -rf ./bin

run: build
	./bin/mcp-mashup

install: build
	mkdir -p ~/.local/bin
	cp ./bin/mcp-mashup ~/.local/bin/
