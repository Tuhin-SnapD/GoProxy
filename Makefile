.PHONY: build clean test run-proxy run-backend help

# Default target
all: build

# Build the project
build:
	@echo "Building GoProxy..."
	@go mod tidy
	@go build -o goproxy main.go
	@go build -tags testserver -o test_server test_server.go
	@echo "Build complete!"

# Build for Windows
build-windows:
	@echo "Building GoProxy for Windows..."
	@go mod tidy
	@go build -o goproxy.exe main.go
	@go build -tags testserver -o test_server.exe test_server.go
	@echo "Build complete!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f goproxy goproxy.exe
	@rm -f test_server test_server.exe
	@echo "Clean complete!"

# Run the proxy server
run-proxy:
	@echo "Starting proxy server..."
	@./goproxy

# Run the test backend server
run-backend:
	@echo "Starting test backend server..."
	@./test_server

# Test the proxy (requires both servers running)
test:
	@echo "Testing proxy server..."
	@curl -s http://localhost:8080/health
	@echo "\nTesting metrics endpoint..."
	@curl -s http://localhost:8080/metrics

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the project"
	@echo "  build-windows - Build for Windows"
	@echo "  clean         - Clean build artifacts"
	@echo "  run-proxy     - Run the proxy server"
	@echo "  run-backend   - Run the test backend server"
	@echo "  test          - Test the proxy server"
	@echo "  help          - Show this help" 