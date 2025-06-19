# SMLGOAPI Makefile
.PHONY: build clean test fmt vet deps check docker-build

# Build the application
build:
	go build -v -o smlgoapi .

# Clean build artifacts
clean:
	go clean
	rm -f smlgoapi smlgoapi.exe

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Download dependencies
deps:
	go mod download
	go mod verify

# Run all checks (CI pipeline)
check: fmt vet build test

# Docker build
docker-build:
	docker build -t smlgoapi:latest .

# Development server
dev:
	go run .

# Production build (for CI/CD)
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o smlgoapi .

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format code"
	@echo "  vet        - Vet code"
	@echo "  deps       - Download dependencies"
	@echo "  check      - Run all checks (CI pipeline)"
	@echo "  docker-build - Build Docker image"
	@echo "  dev        - Run development server"
	@echo "  build-prod - Production build"
	@echo "  help       - Show this help"