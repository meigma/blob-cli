# blob-cli build tasks
set shell := ["bash", "-euo", "pipefail", "-c"]

# Binary name
binary := "blob"

# Default recipe: validate code
default: fmt vet lint test

# CI recipe: full validation pipeline
ci: fmt vet lint test build

# Format check (fails if code needs formatting)
fmt:
    @echo "Checking formatting..."
    @test -z "$(gofmt -l .)" || (echo "Files need formatting:"; gofmt -l .; exit 1)

# Run go vet
vet:
    @echo "Running go vet..."
    go vet ./...

# Run golangci-lint
lint:
    @echo "Running golangci-lint..."
    golangci-lint run

# Run tests
test:
    @echo "Running tests..."
    go test -race -cover ./...

# Build the binary
build:
    @echo "Building..."
    go build -o {{binary}} .

# Install development tools
tools:
    @echo "Installing development tools..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Format code (modifies files)
fmt-write:
    @echo "Formatting code..."
    gofmt -w .

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    rm -f {{binary}}

# Show available recipes
help:
    @just --list
