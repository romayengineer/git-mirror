.PHONY: build test test-verbose test-cover fmt vet lint clean run deps install help

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
	@echo "  install       - Build and install to ~/.local/bin"
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

install: build
	@mkdir -p $(HOME)/.local/bin
	@mkdir -p $(HOME)/.config/codecommit-mirror
	@cp build/codecommit-mirror $(HOME)/.local/bin/codecommit-mirror
	@chmod +x $(HOME)/.local/bin/codecommit-mirror
	@echo "✓ Installed: $(HOME)/.local/bin/codecommit-mirror"
	@echo "✓ Created config directory: $(HOME)/.config/codecommit-mirror"
	@if [ ! -f $(HOME)/.config/codecommit-mirror/.env ]; then \
		cp .env.example $(HOME)/.config/codecommit-mirror/.env; \
		echo "✓ Created config template: $(HOME)/.config/codecommit-mirror/.env"; \
	else \
		echo "ℹ  Config file already exists: $(HOME)/.config/codecommit-mirror/.env (not overwritten)"; \
	fi
	@echo ""
	@echo "Next steps:"
	@echo "1. Edit the configuration file with your settings:"
	@echo "   nano $(HOME)/.config/codecommit-mirror/.env"
	@echo ""
	@echo "2. Uncomment the REQUIRED variables and fill in your values:"
	@echo "   - AWS_ACCOUNT_IDS"
	@echo "   - SSH_KEY_PATH"
	@echo "   - MIRROR_OUTPUT_DIR"
	@echo ""
	@echo "3. Run the tool:"
	@echo "   codecommit-mirror"
	@echo ""
	@echo "The tool will automatically load config from:"
	@echo "  - $(HOME)/.config/codecommit-mirror/.env (recommended)"
	@echo "  - .env in current or parent directory"
	@echo ""
	@echo "To use this command globally, ensure $(HOME)/.local/bin is in your PATH:"
	@echo "  export PATH=\$$HOME/.local/bin:\$$PATH"

deps:
	cd src && go mod download
	cd src && go mod tidy
	@echo "✓ Dependencies updated"

all: clean fmt vet test-cover build
	@echo "✓ Full build complete"
