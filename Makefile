.PHONY: help build test test-race test-coverage lint clean docker-build docker-run docker-stop install-tools

# Default target
help:
	@echo "Available targets:"
	@echo "  build          - Build the library and examples"
	@echo "  test           - Run tests"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run linter"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run with Docker Compose"
	@echo "  docker-stop    - Stop Docker Compose services"
	@echo "  install-tools  - Install development tools"

# Build the library and examples
build:
	@echo "Building CHORM library..."
	go build -v ./...
	@echo "Building examples..."
	cd examples && go build -v .

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	golint -set_exit_status ./...
	go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -f coverage.out coverage.html
	rm -f examples/chorm-example

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install golang.org/x/lint/golint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Docker targets
docker-build:
	@echo "Building Docker image..."
	docker build -t forester/chorm:latest .

docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-stop:
	@echo "Stopping Docker Compose services..."
	docker-compose down

# Development helpers
dev-setup: install-tools
	@echo "Development environment setup complete"

# Benchmark
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Generate documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060 &
	@echo "Documentation available at http://localhost:6060"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	go list -json -deps ./... | nancy sleuth

# Create release
release:
	@echo "Creating release..."
	git tag -a v$(shell git describe --tags --abbrev=0 | cut -d'v' -f2 | awk -F. '{print $$1"."$$2"."$$3+1}') -m "Release v$(shell git describe --tags --abbrev=0 | cut -d'v' -f2 | awk -F. '{print $$1"."$$2"."$$3+1}')"
	git push origin --tags

# All-in-one development target
dev: clean build test lint
	@echo "Development cycle complete" 