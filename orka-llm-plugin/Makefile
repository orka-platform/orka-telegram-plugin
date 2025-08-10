.PHONY: build clean test run help

# Default target
help:
	@echo "Available targets:"
	@echo "  build    - Build the LLM plugin binary"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Build and run test client"
	@echo "  run      - Build and run the plugin on port 50052"
	@echo "  deps     - Download and tidy dependencies"
	@echo "  help     - Show this help message"

# Build the plugin
build: deps
	@echo "Building LLM plugin..."
	go build -o orka-llm-plugin main.go
	@echo "Build complete: orka-llm-plugin"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f orka-llm-plugin
	@echo "Clean complete"

# Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	go mod tidy
	go mod download

# Build and run the plugin
run: build
	@echo "Starting LLM plugin on port 50052..."
	./orka-llm-plugin --port 50052

# Build and run test client
test: build
	@echo "Building test client..."
	go build -o test_client test_client.go
	@echo "Test client built. Make sure the plugin is running on port 50052, then run:"
	@echo "  ./test_client"

# Install dependencies (for development)
install-deps:
	@echo "Installing development dependencies..."
	go get github.com/orka-platform/orka-plugin-sdk@latest
	go get github.com/vercel/ai-sdk-go@latest

# Format code
fmt:
	@echo "Formatting code..."
	go fmt .

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Run tests (if you add unit tests later)
test-unit:
	@echo "Running unit tests..."
	go test ./...

# Build for different platforms
build-linux: deps
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -o orka-llm-plugin-linux main.go

build-darwin: deps
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -o orka-llm-plugin-darwin main.go

build-windows: deps
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -o orka-llm-plugin-windows.exe main.go

# Build all platforms
build-all: build-linux build-darwin build-windows
	@echo "Built for all platforms"

# Show plugin info
info: build
	@echo "Plugin binary info:"
	@ls -la orka-llm-plugin
	@echo ""
	@echo "Plugin version and methods:"
	@cat config.json | jq '.name, .version, .methods | keys' 2>/dev/null || cat config.json