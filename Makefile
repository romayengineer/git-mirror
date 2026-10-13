.PHONY: build test test-verbose test-cover fmt vet lint clean run deps install help tidy sync

MODULES := modules/core modules/exec modules/path modules/gitops modules/config modules/orchestrator modules/provider-github modules/provider-codecommit cmd/git-mirror

help:
	@echo "git-mirror (multi-module monorepo) - Build Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build         - Build the git-mirror binary"
	@echo "  test          - Run all unit tests (all modules)"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-cover    - Run tests with coverage report"
	@echo "  fmt           - Format all Go code"
	@echo "  vet           - Run go vet analysis"
	@echo "  lint          - Run golangci-lint (requires installation)"
	@echo "  tidy          - go mod tidy in every module"
	@echo "  sync          - go work sync"
	@echo "  run           - Build and run the tool"
	@echo "  install       - Build and install to ~/.local/bin"
	@echo "  clean         - Remove build artifacts"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Build output: ./build/git-mirror"
	@echo "Modules (each with own go.mod): $(MODULES)"
	@echo ""

build:
	@mkdir -p build
	cd cmd/git-mirror && go build -o ../../build/git-mirror .
	@echo "✓ Built: ./build/git-mirror"

test:
	@for m in $(MODULES); do echo "== $$m"; (cd $$m && go test ./...) || exit 1; done

test-verbose:
	@for m in $(MODULES); do echo "== $$m"; (cd $$m && go test -v ./...) || exit 1; done

test-cover:
	@for m in $(MODULES); do echo "== $$m"; (cd $$m && go test -cover ./...) || exit 1; done

fmt:
	@for m in $(MODULES); do (cd $$m && go fmt ./...) || exit 1; done
	@echo "✓ Code formatted"

vet:
	@for m in $(MODULES); do (cd $$m && go vet ./...) || exit 1; done
	@echo "✓ Code vetted"

lint:
	@for m in $(MODULES); do (cd $$m && golangci-lint run) || exit 1; done
	@echo "✓ Linting complete"

clean:
	rm -rf build/
	@echo "✓ Clean complete"

run: build
	./build/git-mirror

install: build
	@mkdir -p $(HOME)/.local/bin
	@mkdir -p $(HOME)/.config/git-mirror
	@cp build/git-mirror $(HOME)/.local/bin/git-mirror
	@chmod +x $(HOME)/.local/bin/git-mirror
	@echo "✓ Installed: $(HOME)/.local/bin/git-mirror"
	@echo "✓ Created config directory: $(HOME)/.config/git-mirror"
	@if [ ! -f $(HOME)/.config/git-mirror/.env ]; then \
		cp .env.example $(HOME)/.config/git-mirror/.env; \
		echo "✓ Created config template: $(HOME)/.config/git-mirror/.env"; \
	else \
		echo "ℹ  Config file already exists: $(HOME)/.config/git-mirror/.env (not overwritten)"; \
	fi
	@echo ""
	@echo "Next steps:"
	@echo "1. Edit the configuration file with your settings:"
	@echo "   nano $(HOME)/.config/git-mirror/.env"
	@echo ""
	@echo "2. Uncomment the REQUIRED variables and fill in your values:"
	@echo "   - AWS_ACCOUNT_IDS"
	@echo "   - SSH_KEY_PATH"
	@echo "   - MIRROR_OUTPUT_DIR"
	@echo ""
	@echo "3. Run the tool:"
	@echo "   git-mirror"
	@echo ""
	@echo "The tool will automatically load config from:"
	@echo "  - $(HOME)/.config/git-mirror/.env (recommended)"
	@echo "  - .env in current or parent directory"
	@echo ""
	@echo "To use this command globally, ensure $(HOME)/.local/bin is in your PATH:"
	@echo "  export PATH=\$$HOME/.local/bin:\$$PATH"

deps tidy:
	@for m in $(MODULES); do echo "== $$m"; (cd $$m && go mod download && go mod tidy) || exit 1; done
	@echo "✓ Dependencies updated"

sync:
	go work sync
	@echo "✓ Workspace synced"

all: clean fmt vet test-cover build
	@echo "✓ Full build complete"
