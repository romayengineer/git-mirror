# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an AWS CodeCommit mirror tool written in Go. Given SSH credentials and an AWS account ID, it discovers and clones all CodeCommit repositories across all AWS regions, maintaining a local mirror of the entire account's repositories.

## High-Level Architecture

The tool follows a simple pipeline:

1. **Discovery Phase**: Uses AWS SDK to list CodeCommit repositories in each region
2. **Authentication**: Uses provided SSH credentials to clone repositories
3. **Clone Phase**: Clones each discovered repository to the local filesystem
4. **Mirror Sync**: Pulls latest changes from all repositories (for updates)

The codebase is structured around these concerns:
- AWS API interaction (account alias lookup, region enumeration, repository listing)
- SSH/Git operations (cloning, pulling)
- Concurrency and error handling (parallel cloning across regions and repositories)
- Path management (building consistent `account_alias/region/repository_name` structure)
- Configuration management (SSH keys, AWS credentials, output directory)

## Repository Structure

```
codecommit-mirror/
├── src/                                     # All Go code lives here
│   ├── cmd/
│   │   └── main/
│   │       └── main.go                      # Entry point (.env loading & orchestration)
│   ├── internal/
│   │   ├── aws/
│   │   │   ├── client.go                    # AWS SDK wrapper/initialization
│   │   │   ├── discoverer.go                # List repos across regions
│   │   │   └── discoverer_test.go
│   │   ├── git/
│   │   │   ├── cloner.go                    # SSH-based clone logic
│   │   │   ├── syncer.go                    # Pull/update logic
│   │   │   └── cloner_test.go
│   │   └── mirror/
│   │       ├── orchestrator.go              # Workflow coordination
│   │       └── orchestrator_test.go
│   ├── pkg/
│   │   ├── config/
│   │   │   └── config.go                    # .env parsing & validation
│   │   └── types/
│   │       └── types.go                     # Repository, Region types
│   ├── go.mod
│   ├── go.sum
│   └── Makefile                             # Optional: build automation
├── .env.example                             # Example .env template
├── .env                                     # Configuration (git ignored)
├── README.md
├── CLAUDE.md
└── .gitignore
```

## Development Commands

All commands are run from the `src/` directory.

### Building
```bash
cd src/
go build -o codecommit-mirror ./cmd/main
```

### Testing
```bash
cd src/
go test ./...                  # Run all tests
go test -v ./...               # Run with verbose output
go test ./... -run TestName    # Run specific test
go test -cover ./...           # Show coverage
```

### Linting & Code Quality
```bash
cd src/
go fmt ./...                   # Format code
go vet ./...                   # Vet for common errors
golangci-lint run              # Full linting (if installed)
```

### Running
```bash
cd src/
./codecommit-mirror
```

Configuration is loaded from a `.env` file in the repository root.

## Configuration

All configuration is loaded from a `.env` file in the repository root. Required variables:

```env
AWS_ACCOUNT_ID=123456789012         # AWS account ID to mirror
AWS_REGION=us-east-1                # Primary AWS region for API calls
SSH_KEY_PATH=/path/to/ssh/key       # Path to SSH private key for CodeCommit
MIRROR_OUTPUT_DIR=./mirror          # Local directory for cloned repositories
```

Optional variables:

```env
AWS_CREDENTIALS_PROFILE=default     # AWS credentials profile (if using profiles)
CONCURRENT_CLONES=5                 # Number of concurrent clone operations
LOG_LEVEL=info                      # Logging level (debug, info, warn, error)
```

The tool reads the `.env` file on startup and validates that all required variables are present and valid.

## Key Files & Packages

- `src/go.mod` / `src/go.sum`: Dependency management
- `src/cmd/main/main.go`: Entry point (loads .env and starts orchestration)
- `src/internal/aws/`: AWS SDK interactions and region enumeration
- `src/internal/git/`: SSH-based clone and sync operations
- `src/internal/mirror/`: Orchestration of discovery + cloning workflow
- `src/pkg/config/`: .env parsing, validation, and configuration struct
- `src/pkg/types/`: Shared data structures (Repository, Region, etc.)

## Important Considerations

### SSH Key Handling
- SSH keys are passed as file paths; never embed them in code or logs
- Ensure proper file permissions (600) are validated before use
- SSH key errors should fail fast with clear messaging

### Concurrency
- Repository cloning should happen in parallel across regions to improve performance
- Use goroutine pools or semaphores to avoid overwhelming the system or AWS API limits
- Handle context cancellation gracefully (e.g., Ctrl+C during mirror operations)

### AWS API Interactions
- Always paginate through ListRepositories results (repositories > 25 per region)
- Handle region errors gracefully; a failed region shouldn't stop other regions
- Use proper AWS credential chain (env vars, ~/.aws/credentials, IAM role, etc.)

### Output Structure
Cloned repositories follow this directory structure:
```
mirror_output/
├── account_alias/
│   ├── us-east-1/
│   │   ├── repo-name-1/
│   │   └── repo-name-2/
│   ├── eu-west-1/
│   │   └── repo-name-3/
```

- Use the AWS account alias for the top-level directory
- Organize by region underneath
- Repository names at the leaf level
- Create directories if they don't exist
- Handle duplicate repository names across regions (same name, different regions)
- Skip cloning if a repository already exists locally (add an update flag to re-sync)

## Testing Strategy

- Unit tests for AWS discovery logic (mock AWS SDK responses)
- Unit tests for git operations (mock git commands or use temp directories)
- Integration tests should use a test AWS account/credentials (if available)
- CLI flag parsing should be tested independently
