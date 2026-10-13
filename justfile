set shell := ["sh", "-cu"]

modules := "modules/core modules/exec modules/path modules/gitops modules/config modules/orchestrator modules/provider-github modules/provider-codecommit cmd/git-mirror"

# List available recipes
default:
    @just --list

# Build the git-mirror binary → ./build/git-mirror
build:
    @mkdir -p build
    cd cmd/git-mirror && go build -o ../../build/git-mirror .
    @echo "✓ Built: ./build/git-mirror"

# Run all unit tests (all modules)
test:
    @for m in {{modules}}; do echo "== $m"; (cd $m && go test ./...) || exit 1; done

# Run tests with verbose output
test-verbose:
    @for m in {{modules}}; do echo "== $m"; (cd $m && go test -v ./...) || exit 1; done

# Run tests with coverage report
test-cover:
    @for m in {{modules}}; do echo "== $m"; (cd $m && go test -cover ./...) || exit 1; done

# Format all Go code
fmt:
    @for m in {{modules}}; do (cd $m && go fmt ./...) || exit 1; done
    @echo "✓ Code formatted"

# Run go vet analysis
vet:
    @for m in {{modules}}; do (cd $m && go vet ./...) || exit 1; done
    @echo "✓ Code vetted"

# Full linting (requires golangci-lint)
lint:
    @for m in {{modules}}; do (cd $m && golangci-lint run) || exit 1; done
    @echo "✓ Linting complete"

# Remove build artifacts
clean:
    rm -rf build/
    @echo "✓ Clean complete"

# Build and run the tool
run: build
    ./build/git-mirror

# Build and install to ~/.local/bin, seed config template if missing
install: build
    @mkdir -p $HOME/.local/bin
    @mkdir -p $HOME/.config/git-mirror
    @cp build/git-mirror $HOME/.local/bin/git-mirror
    @chmod +x $HOME/.local/bin/git-mirror
    @echo "✓ Installed: $HOME/.local/bin/git-mirror"
    @echo "✓ Created config directory: $HOME/.config/git-mirror"
    @if [ ! -f $HOME/.config/git-mirror/config.yaml ] && [ ! -f $HOME/.config/git-mirror/.env ]; then \
        cp config.example.yaml $HOME/.config/git-mirror/config.yaml; \
        echo "✓ Created config template: $HOME/.config/git-mirror/config.yaml"; \
    else \
        echo "ℹ  Config file already exists in $HOME/.config/git-mirror/ (not overwritten)"; \
    fi
    @echo ""
    @echo "Next steps:"
    @echo "1. Edit the configuration file with your settings:"
    @echo "   nano $HOME/.config/git-mirror/config.yaml"
    @echo ""
    @echo "2. Set the REQUIRED output_dir and at least one provider:"
    @echo "   - output_dir (e.g. ./mirror)"
    @echo "   - github.username / github.orgs, or codecommit.accounts"
    @echo ""
    @echo "3. Run the tool:"
    @echo "   git-mirror"
    @echo ""
    @echo "The tool will automatically load config from:"
    @echo "  - $HOME/.config/git-mirror/config.yaml (recommended)"
    @echo "  - config.yaml in current or parent directory"
    @echo "  - .env search as legacy fallback"
    @echo ""
    @echo "To use this command globally, ensure $HOME/.local/bin is in your PATH:"
    @echo "  export PATH=\$HOME/.local/bin:\$PATH"

# Download and tidy dependencies in every module
tidy:
    @for m in {{modules}}; do echo "== $m"; (cd $m && go mod download && go mod tidy) || exit 1; done
    @echo "✓ Dependencies updated"

alias deps := tidy

# Sync the go workspace
sync:
    go work sync
    @echo "✓ Workspace synced"

# Full pipeline: clean, format, vet, test (coverage), build
all: clean fmt vet test-cover build
    @echo "✓ Full build complete"
