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
- AWS API interaction (region enumeration, repository listing)
- SSH/Git operations (cloning, pulling)
- Concurrency and error handling (parallel cloning across regions)
- Configuration management (SSH keys, AWS credentials, output directory)

## Development Commands

### Building
```bash
go build -o codecommit-mirror ./cmd/main.go
```

### Testing
```bash
go test ./...              # Run all tests
go test -v ./...           # Run with verbose output
go test ./... -run TestName  # Run specific test
go test -cover ./...       # Show coverage
```

### Linting & Code Quality
```bash
go fmt ./...               # Format code
go vet ./...               # Vet for common errors
golangci-lint run          # Full linting (if installed)
```

### Running
```bash
./codecommit-mirror -account <ACCOUNT_ID> -ssh-key <PATH_TO_KEY> -output <MIRROR_DIR>
./codecommit-mirror --help  # Show available flags
```

## Key Configuration Files

- `go.mod` / `go.sum`: Dependency management
- `cmd/main.go`: Entry point (CLI flags and initialization)
- `internal/`: Core logic (AWS discovery, git operations, concurrency)
- `internal/aws/`: AWS SDK interactions and region enumeration
- `internal/git/`: SSH-based clone and sync operations
- `internal/mirror/`: Orchestration of discovery + cloning workflow

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
- Preserve repository naming to avoid collisions (e.g., `region/repo-name/`)
- Create directories if they don't exist
- Skip cloning if a repository already exists locally (add an update flag)

## Testing Strategy

- Unit tests for AWS discovery logic (mock AWS SDK responses)
- Unit tests for git operations (mock git commands or use temp directories)
- Integration tests should use a test AWS account/credentials (if available)
- CLI flag parsing should be tested independently
