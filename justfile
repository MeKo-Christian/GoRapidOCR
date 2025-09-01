# Justfile for GoRapidOCR

# Run tests
test:
	go test -v ./...

# Build the project
build:
	go build -v ./...

# Clean build artifacts
clean:
	go clean ./...

# Run linter
lint:
	go vet ./...

# Format code using treefmt
fmt:
	treefmt --allow-missing-formatter

# Run all checks (fmt, lint, test)
check: fmt lint test

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests with coverage
test-cover:
	go test -v -cover ./...

# Build and test
all: deps build test
