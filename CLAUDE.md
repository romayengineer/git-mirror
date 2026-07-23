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
- **AWS Credential Chain Support**: Uses `~/.aws/credentials` profiles with optional STS AssumeRole for cross-account access
- **Flexible Concurrency**: Configurable number of concurrent clone operations
- **Git CLI Integration**: Uses git CLI directly with a thin wrapper (no external git library dependencies)
- **Dependency Injection**: Uses dependency injection for AWS clients and git operations (enables easy testing and mocking)

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
├── Makefile                                 # Build automation (run from root)
├── README.md
├── CLAUDE.md
├── .env.example                             # Example .env template
├── .env                                     # Configuration (git ignored)
├── .gitignore
├── src/                                     # All Go code
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── main/
│   │       └── main.go                      # Entry point
│   ├── internal/
│   │   ├── aws/
│   │   │   ├── client.go                    # AWS SDK wrapper
│   │   │   ├── discoverer.go                # Repository discovery
│   │   │   └── discoverer_test.go
│   │   ├── git/
│   │   │   ├── executor.go                  # Git CLI wrapper
│   │   │   ├── cloner.go                    # Clone logic with URL verification
│   │   │   └── cloner_test.go
│   │   └── mirror/
│   │       └── orchestrator.go              # Workflow orchestration
│   └── pkg/
│       ├── config/
│       │   ├── config.go                    # .env parsing & validation
│       │   └── config_test.go
│       └── types/
│           └── types.go                     # Shared data types
└── build/                                   # Build output (git ignored)
    └── codecommit-mirror                    # Compiled binary
```

## Development Commands

All commands are run from the **project root directory** using the Makefile.

### Building
```bash
make build
# Output: ./build/codecommit-mirror
```

### Testing
```bash
make test              # Run all tests
make test-verbose      # Run with verbose output
make test-cover        # Show coverage report
```

To run a specific test:
```bash
cd src/
go test ./... -run TestName
```

### Linting & Code Quality
```bash
make fmt               # Format code
make vet               # Check for common errors
make lint              # Full linting (requires golangci-lint)
```

### Running
```bash
make run               # Build and run the tool
# Or run the binary directly:
./build/codecommit-mirror
```

### Cleanup
```bash
make clean             # Remove build artifacts
```

### Show Available Commands
```bash
make help
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
- `src/internal/git/`: Thin wrapper around git CLI commands (clone, pull, etc.)
- `src/internal/mirror/`: Orchestration of discovery + cloning workflow
- `src/pkg/config/`: .env parsing, validation, and configuration struct
- `src/pkg/types/`: Shared data structures (Repository, Region, etc.)

## Git Command Implementation

The tool uses **git CLI commands directly** via Go's `os/exec` package rather than a pure-Go git library. A thin wrapper in `src/internal/git/` encapsulates common git operations:

- `Clone(url, path)` — executes `git clone`
- `Pull(repoPath)` — executes `git pull` in an existing repository
- Other git operations as needed

**Advantages:**
- No external Go dependencies for git functionality
- Uses system-installed git (familiar, reliable)
- Easy to debug (can see exact git commands being run)
- Full feature parity with git CLI

**Requirements:**
- Git must be installed and available in `$PATH`
- SSH must be configured for CodeCommit access

## Dependency Injection Architecture

The codebase uses **interface-based dependency injection throughout** to maximize testability and flexibility:

**Core Principles:**
- Define interfaces for ALL external dependencies (AWS, git, file I/O, etc.)
- Accept interfaces in constructors and function parameters
- Keep interfaces small and focused (single responsibility)
- Implement interfaces in separate implementation packages when needed

**Key Interfaces (Examples):**
- `aws.AccountClient` — Interface for AWS account-level operations (list regions, get account alias)
- `aws.RepositoryDiscoverer` — Interface for discovering repositories
- `git.Executor` — Interface for git CLI operations (clone, pull, verify origin)
- `config.Loader` — Interface for loading configuration
- `log.Logger` — Interface for logging (if needed)

**Implementation Pattern:**
```go
// Define interface in the package
type AccountClient interface {
  GetAccountAlias(ctx context.Context, accountID string) (string, error)
  ListRegions(ctx context.Context) ([]string, error)
}

// Accept interface in constructor
type Discoverer struct {
  client AccountClient
}

func NewDiscoverer(client AccountClient) *Discoverer {
  return &Discoverer{client: client}
}

// In tests, pass mock
type mockAccountClient struct{}
func (m *mockAccountClient) GetAccountAlias(...) (string, error) { return "test-account", nil }
```

**Benefits:**
- 100% unit test coverage possible without AWS/git infrastructure
- Easy to test error cases, edge cases, and timeouts
- Tests run fast (no real API calls)
- Easy to add new implementations (e.g., future git library changes)

## Important Considerations

### SSH Key Handling
- SSH keys are passed as file paths; never embed them in code or logs
- Ensure proper file permissions (600) are validated before use
- SSH key errors should fail fast with clear messaging
- The SSH key path is used by git CLI directly, so it must be in a standard location or explicitly added to ssh-agent
- Consider validating that the SSH key exists and is readable before starting the mirror process

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

### Existing Repository Handling
When a repository already exists locally:
1. **Verify the URL**: Check that the git origin URL matches the expected source URL from AWS CodeCommit
2. **If URL matches**: Pull latest changes (`git pull`) to update the repository
3. **If URL doesn't match**: Fail with a clear error message (prevents accidental overwrite of mismatched repos)
4. **Error Recovery**: On URL mismatch, log which repository has the mismatch and continue with other repositories

## Testing Strategy

- **Unit Tests**: Mock all external dependencies (AWS, git) using interfaces
- **AWS Discovery Tests**: Mock `AccountClient` and `RepositoryDiscoverer` interfaces
- **Git Operation Tests**: Mock `git.Executor` interface for clone/pull/verify operations
- **Configuration Tests**: Unit test .env parsing and validation
- **Error Scenario Tests**: Use mocks to simulate AWS/git failures, timeouts, etc.
- **Integration Tests** (optional): Use a test AWS account if available for end-to-end testing
- **Test Coverage Goal**: Aim for high coverage through interface mocking (no external dependencies needed for unit tests)
