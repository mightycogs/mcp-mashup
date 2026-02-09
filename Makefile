.PHONY: help build test clean run install

help:
	@echo "Available targets:"
	@echo "  build    - Build mcp-mashup binary"
	@echo "  test     - Run tests"
	@echo "  clean    - Remove build artifacts"
	@echo "  run      - Build and run"
	@echo "  install  - Install to ~/.local/bin"

build:
	go build -o mcp-mashup ./cmd/mcp-mashup

test:
	go test ./...

clean:
	rm -f mcp-mashup

run: build
	./mcp-mashup

install: build
	mkdir -p ~/.local/bin
	cp mcp-mashup ~/.local/bin/
