.PHONY: build test clean examples install

# Default target
all: test build

# Build the library
build:
	go build ./...

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	go clean ./...
	rm -f coverage.out coverage.html

# Build examples
examples: basic-example trading-bot

basic-example:
	go build -o bin/basic-example examples/basic/main.go

trading-bot:
	go build -o bin/trading-bot examples/trading-bot/main.go

# Run basic example
run-basic: basic-example
	./bin/basic-example

# Run trading bot example
run-bot: trading-bot
	./bin/trading-bot

# Install dependencies
install:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Run all checks
check: fmt test lint

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/ctrader-linux-amd64 ./...
	GOOS=windows GOARCH=amd64 go build -o bin/ctrader-windows-amd64 ./...
	GOOS=darwin GOARCH=amd64 go build -o bin/ctrader-darwin-amd64 ./...
	GOOS=darwin GOARCH=arm64 go build -o bin/ctrader-darwin-arm64 ./...

# Create release
release: clean test build-all
	tar -czf ctrader-go-release.tar.gz bin/ README.md LICENSE

# Development setup
dev-setup: install
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the library"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  examples       - Build all examples"
	@echo "  basic-example  - Build basic example"
	@echo "  trading-bot    - Build trading bot example"
	@echo "  run-basic      - Run basic example"
	@echo "  run-bot        - Run trading bot example"
	@echo "  install        - Install dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  check          - Run format, test, and lint"
	@echo "  build-all      - Build for multiple platforms"
	@echo "  release        - Create release tarball"
	@echo "  dev-setup      - Setup development environment"
	@echo "  help           - Show this help message"
