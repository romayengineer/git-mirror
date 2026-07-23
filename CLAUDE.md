# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an AWS CodeCommit mirror tool written in Go. Given SSH credentials and a list of AWS account IDs, it discovers and clones all CodeCommit repositories across all AWS regions for each account, maintaining a local mirror of all repositories organized by account and region.

## Features

- **Multi-Account Support**: Mirror repositories from multiple AWS accounts in a single run
- **Multi-Region Support**: Automatically discover and clone from all AWS regions or specify a subset
- **SSH Authentication**: Uses SSH keys for secure CodeCommit access
- **Parallel Cloning**: Concurrent clone operations to improve performance across regions and accounts
- **Smart Directory Organization**: Organizes cloned repositories as `account_alias/region/repository_name`
- **.env Configuration**: Simple environment variable configuration without CLI flags
- **Regional Discovery**: Auto-discover available regions or manually specify which regions to clone from
- **Error Resilience**: Continue cloning even if individual regions or repositories fail
- **Update/Sync Support**: Re-clone existing repositories to pull latest changes
- **Account Alias Resolution**: Uses AWS account aliases for more readable directory structures
- **AWS Credential Chain Support**: Works with environment variables, credential files, IAM roles, and STS AssumeRole
- **Flexible Concurrency**: Configurable number of concurrent clone operations

## High-Level Architecture

The tool follows a pipeline for each configured account:

1. **Account Setup**: Assume IAM role or use credentials for each account ID
2. **Discovery Phase**: Uses AWS SDK to list CodeCommit repositories in each region
3. **Authentication**: Uses provided SSH credentials to clone repositories
4. **Clone Phase**: Clones each discovered repository to the local filesystem
5. **Mirror Sync**: Pulls latest changes from all repositories (for updates)

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
AWS_ACCOUNT_IDS=123456789012,210987654321    # Comma-separated list of AWS account IDs
SSH_KEY_PATH=/path/to/ssh/key                # Path to SSH private key for CodeCommit
MIRROR_OUTPUT_DIR=./mirror                   # Local directory for cloned repositories
```

Optional variables:

```env
AWS_REGIONS=us-east-1,us-west-2,eu-west-1     # Comma-separated list of regions
                                              # If not set, discovers all regions
AWS_CREDENTIALS_PROFILE=default     # AWS credentials profile (if using profiles)
CONCURRENT_CLONES=5                 # Number of concurrent clone operations
LOG_LEVEL=info                      # Logging level (debug, info, warn, error)
```

The tool reads the `.env` file on startup and validates that all required variables are present and valid.

**Regional Discovery**: If `AWS_REGIONS` is not specified, the tool will:
- Fetch all available AWS regions via the EC2 API
- Attempt to list CodeCommit repositories in each region
- Skip regions where CodeCommit is not available
- Clone from all regions that have repositories

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
- Process multiple accounts in parallel (each account gets its own credential context)
- Clone repositories in parallel across regions within each account
- Use goroutine pools or semaphores to avoid overwhelming the system or AWS API rate limits
- Handle context cancellation gracefully (e.g., Ctrl+C during mirror operations)
- Ensure account-level operations (like STS AssumeRole) are thread-safe

### AWS API Interactions
- Always paginate through ListRepositories results (repositories > 25 per region)
- Handle region errors gracefully; a failed region shouldn't stop other regions
- For multiple accounts: Use STS AssumeRole to switch between accounts, or manage credentials via AWS credential files
- Use proper AWS credential chain (env vars, ~/.aws/credentials, IAM role, etc.)
- Each account should have a separate AWS API session with appropriate credentials

### Output Structure
Cloned repositories follow this directory structure:
```
mirror_output/
├── account-alias-1/
│   ├── us-east-1/
│   │   ├── repo-name-1/
│   │   └── repo-name-2/
│   ├── eu-west-1/
│   │   └── repo-name-3/
├── account-alias-2/
│   ├── us-east-1/
│   │   └── repo-name-4/
│   └── ap-southeast-1/
│       └── repo-name-5/
```

- Top level: Account alias (for each account in `AWS_ACCOUNT_IDS`)
- Second level: AWS region
- Leaf level: Repository names
- Create directories if they don't exist
- Handle duplicate repository names across regions or accounts (same name, different regions/accounts)
- Skip cloning if a repository already exists locally (add an update flag to re-sync)

## Testing Strategy

- Unit tests for AWS discovery logic (mock AWS SDK responses)
- Unit tests for git operations (mock git commands or use temp directories)
- Integration tests should use a test AWS account/credentials (if available)
- CLI flag parsing should be tested independently
