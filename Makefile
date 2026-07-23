.PHONY: build test test-verbose test-cover fmt vet lint clean run deps help

help:
	@echo "AWS CodeCommit Mirror - Build Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build         - Build the codecommit-mirror binary"
	@echo "  test          - Run all unit tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-cover    - Run tests with coverage report"
	@echo "  fmt           - Format all Go code"
	@echo "  vet           - Run go vet analysis"
	@echo "  lint          - Run golangci-lint (requires installation)"
	@echo "  run           - Build and run the tool"
	@echo "  clean         - Remove build artifacts"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Build output: ./build/codecommit-mirror"
	@echo ""

build:
	@mkdir -p build
	cd src && go build -o ../build/codecommit-mirror ./cmd/main
	@echo "✓ Built: ./build/codecommit-mirror"

test:
	cd src && go test ./...

test-verbose:
	cd src && go test -v ./...

test-cover:
	cd src && go test -cover ./...

fmt:
	cd src && go fmt ./...
	@echo "✓ Code formatted"

vet:
	cd src && go vet ./...
	@echo "✓ Code vetted"

lint:
	cd src && golangci-lint run
	@echo "✓ Linting complete"

clean:
	rm -rf build/
	@echo "✓ Clean complete"

run: build
	./build/codecommit-mirror

deps:
	cd src && go mod download
	cd src && go mod tidy
	@echo "✓ Dependencies updated"

all: clean fmt vet test-cover build
	@echo "✓ Full build complete"
