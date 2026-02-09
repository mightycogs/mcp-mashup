.PHONY: help build test clean run example install

help:
	@echo "Available targets:"
	@echo "  build    - Build mcp-mashup binary"
	@echo "  test     - Run tests"
	@echo "  clean    - Remove build artifacts"
	@echo "  run      - Build and run"
	@echo "  example  - Build and run example test client"
	@echo "  install  - Install to ~/.local/bin"

build:
	go build -o mcp-mashup ./cmd/mcp-mashup

test:
	go test ./...

clean:
	rm -f mcp-mashup
	rm -f examples/test-server/test-server
	rm -f examples/test_client

run: build
	./mcp-mashup

example: build
	go build -o examples/test-server/test-server ./examples/test-server
	go build -o examples/test_client examples/test_client.go
	export MCP_CONFIG=./examples/config.json && ./examples/test_client

install: build
	mkdir -p ~/.local/bin
	cp mcp-mashup ~/.local/bin/