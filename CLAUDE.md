# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**git-mirror** is a multi-provider Git repository mirror tool written in Go. It discovers and clones repositories from multiple Git providers (currently AWS CodeCommit, with GitHub and GitLab support planned) to a local filesystem, maintaining organized mirrors by provider, account/organization, and region.

The tool uses a plugin-style provider architecture with an interface-based design that makes adding new providers straightforward without modifying core orchestration logic.

## Features

- **Multi-Provider Support**: Mirror from AWS CodeCommit, GitHub, GitLab, and other Git providers
- **Multi-Account/Organization Support**: Mirror from multiple AWS accounts or GitHub orgs in a single run
- **Multi-Region Support** (CodeCommit): Auto-discover all AWS regions or specify a subset
- **Parallel Cloning**: Concurrent clone operations for performance
- **Smart Directory Organization**: Organizes clones as `provider/account-or-org/region/repository_name`
- **.env Configuration**: Simple environment variable configuration without CLI flags
- **Error Resilience**: Gracefully handle failures - continue cloning if regions or repositories fail
- **Automatic Updates**: Detect existing repositories and pull latest changes
- **Provider-Specific Features**: Each provider implementation handles its own authentication and API patterns
- **Account/Org Aliases**: Uses readable aliases instead of just IDs (AWS account names, GitHub org names, etc.)
- **Flexible Authentication**: Support for SSH keys, tokens, and other auth methods per provider
- **Flexible Concurrency**: Configurable concurrent clone operations with per-provider rate limits
- **Git CLI Integration**: Uses git CLI directly with a thin wrapper (no git library dependencies)
- **Dependency Injection**: Interface-based design for easy testing, mocking, and extension

## High-Level Architecture

The tool follows a plugin-style provider pattern:

1. **Configuration Loading**: Parse `.env` file and load provider configs
2. **Provider Initialization**: Create provider instances (CodeCommit, GitHub, GitLab, etc.)
3. **Repository Discovery**: Each provider discovers its repositories in parallel
4. **Aggregation**: Merge results from all providers
5. **Clone Phase**: Parallel clone/update of all discovered repositories
6. **Mirror Sync**: Pull latest changes from cloned repositories

The codebase is structured around these concerns:
- **Provider Interface** (`src/internal/providers/provider.go`): Defines contract for all providers
- **Provider Implementations** (`src/internal/providers/*/`): Provider-specific logic (CodeCommit, GitHub, etc.)
- **Git Operations** (`src/internal/git/`): Unified clone/pull logic used by all providers
- **Orchestration** (`src/internal/mirror/`): Coordinates providers and cloning
- **Configuration** (`src/pkg/config/`): Environment-based configuration
- **Type Definitions** (`src/pkg/types/`): Shared data structures

## Repository Structure

```
git-mirror/
├── Makefile                                 # Build automation (run from root)
├── README.md
├── CLAUDE.md
├── KNOWN_ISSUES.md
├── .env.example                             # Example .env template
├── .env                                     # Configuration (git ignored)
├── .gitignore
├── src/                                     # All Go code
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── main/
│   │       └── main.go                      # Entry point - initializes providers and orchestrator
│   ├── internal/
│   │   ├── providers/                       # Git provider implementations
│   │   │   ├── provider.go                  # Provider interface
│   │   │   └── codecommit/
│   │   │       └── provider.go              # CodeCommit provider
│   │   ├── aws/
│   │   │   ├── client.go                    # AWS SDK wrapper
│   │   │   ├── cli_executor.go              # AWS CLI wrapper
│   │   │   ├── discoverer.go                # CodeCommit repository discovery
│   │   │   └── discoverer_test.go
│   │   ├── git/
│   │   │   ├── executor.go                  # Git CLI wrapper
│   │   │   ├── cloner.go                    # Clone/pull logic (provider-agnostic)
│   │   │   └── cloner_test.go
│   │   └── mirror/
│   │       └── orchestrator.go              # Coordinates providers and cloning
│   └── pkg/
│       ├── config/
│       │   ├── config.go                    # .env parsing & validation
│       │   └── config_test.go
│       └── types/
│           └── types.go                     # Shared data types (Repository, CloneResult, etc.)
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

### Installation and Setup

```bash
# Build and install to ~/.local/bin
make install
# Creates binary at ~/.local/bin/codecommit-mirror
# Creates config directory ~/.config/codecommit-mirror/
# Copies .env.example → ~/.config/codecommit-mirror/.env (if not exists)
```

## Configuration

The tool searches for configuration files in this order:
1. `~/.config/codecommit-mirror/.env` (recommended - user-wide config)
2. `.env` (repository root - project-specific config)
3. `../.env` (parent directory - fallback)

This search order allows both user-wide and project-specific configurations.

### Setup

**First-time setup (recommended):**
```bash
make install
# Then edit: nano ~/.config/codecommit-mirror/.env
```

**Manual setup:**
```bash
cp .env.example ~/.config/codecommit-mirror/.env
nano ~/.config/codecommit-mirror/.env
```

### Required Variables

```env
AWS_ACCOUNT_IDS=123456789012,210987654321    # Comma-separated list of AWS account IDs
SSH_KEY_PATH=~/.ssh/id_rsa                   # Path to SSH private key for CodeCommit
MIRROR_OUTPUT_DIR=./mirror                   # Local directory for cloned repositories
```

### Optional Variables

```env
AWS_DEFAULT_REGION=us-east-1                 # AWS region for initial API calls (default: us-east-1)
AWS_REGIONS=us-east-1,us-west-2,eu-west-1    # Comma-separated list of regions to mirror
                                             # If not set, discovers all regions
AWS_CREDENTIALS_PROFILE=default              # AWS credentials profile (default: "default")
AWS_ACCOUNT_PROFILES=ID:profile,ID:profile   # Map account IDs to AWS credential profiles
CONCURRENT_CLONES=5                          # Number of concurrent clone operations (default: 5)
LOG_LEVEL=info                               # Logging level: debug, info, warn, error (default: info)
```

The tool validates all required variables are present on startup.

**Regional Discovery**: If `AWS_REGIONS` is not specified, the tool will:
- Fetch all available AWS regions via the EC2 API
- Attempt to list CodeCommit repositories in each region
- Skip regions where CodeCommit is not available
- Clone from all regions that have repositories

**Per-Account Profiles**: Use `AWS_ACCOUNT_PROFILES` to specify different AWS credential profiles for different accounts (useful for cross-account access or STS AssumeRole):
```env
AWS_ACCOUNT_PROFILES=123456789012:prod-account,210987654321:dev-account
```

## Key Files & Packages

- `Makefile`: Build, test, and install automation
  - `make build`: Compiles binary to `./build/codecommit-mirror`
  - `make install`: Installs binary to `~/.local/bin/`, creates `~/.config/codecommit-mirror/`
  - `make test`: Runs all unit tests
- `.env.example`: Template configuration file (all variables commented and documented)
- `src/go.mod` / `src/go.sum`: Dependency management
- `src/cmd/main/main.go`: Entry point (searches for .env file and starts orchestration)
  - Config file search order: `~/.config/codecommit-mirror/.env` → `.env` → `../.env`
- `src/internal/aws/`: AWS SDK interactions and region enumeration (uses CLI wrapper for discovery)
- `src/internal/git/`: Thin wrapper around git CLI commands (clone, fetch, etc.)
- `src/internal/mirror/`: Orchestration of discovery + cloning workflow
- `src/pkg/config/`: .env parsing, validation, and configuration struct
- `src/pkg/types/`: Shared data structures (Repository, Region, CloneResult, etc.)

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

## Multi-Provider Architecture

**git-mirror** is architected from the ground up with a plugin-style provider system that makes it straightforward to add support for new Git providers without modifying core orchestration logic.

### Currently Implemented: AWS CodeCommit

**Package Structure:**
```
src/internal/aws/
├── client.go           # AccountClient interface + AWS SDK implementation
├── cli_executor.go     # AWS CLI wrapper for API calls
├── discoverer.go       # RepositoryDiscoverer for CodeCommit
└── discoverer_test.go  # Tests with mocked client
```

### Adding New Providers

**Required Components:**

1. **Provider Discovery Interface** (implements `RepositoryDiscoverer`)
   ```go
   type RepositoryDiscoverer interface {
     DiscoverRepositories(ctx context.Context, config ProviderConfig) ([]*types.Repository, error)
   }
   ```
   - Discovers all repositories from the provider
   - Returns normalized `types.Repository` objects
   - Handles provider-specific authentication and pagination

2. **Provider Configuration**
   - Add to `.env.example` and `config.go`
   - Example: `GITHUB_TOKEN`, `GITHUB_ORGS`, `GITLAB_URL`, etc.

3. **Update Orchestrator**
   - Modify `mirror/orchestrator.go` to handle multiple providers
   - Each provider can have different concurrency/rate limits
   - Results aggregated into single output directory structure

### Example: Adding GitHub Support

```go
// src/internal/github/discoverer.go
type RepositoryDiscoverer struct {
  client *github.Client
  owner  string
}

func (d *RepositoryDiscoverer) DiscoverRepositories(ctx context.Context, config *types.Config) ([]*types.Repository, error) {
  // Fetch repos from GitHub API
  // Return as types.Repository objects
}
```

**Configuration:**
```env
GIT_PROVIDERS=codecommit,github    # Which providers to use
GITHUB_TOKEN=ghp_xxxx              # GitHub personal access token
GITHUB_ORGS=org1,org2              # Organizations to mirror
```

**Orchestrator Changes:**
```go
// Mirror multiple providers in parallel
providers := []Discoverer{
  aws.NewDiscoverer(awsConfig),
  github.NewDiscoverer(githubConfig),
  gitlab.NewDiscoverer(gitlabConfig),
}

for _, provider := range providers {
  repos := provider.DiscoverRepositories(ctx, config)
  // Clone repos...
}
```

### Design Principles for New Providers

1. **Normalize Output**: All providers return `types.Repository` objects
2. **Consistent Organization**: Use provider name as top-level directory
   ```
   mirror_output/
   ├── codecommit/
   │   ├── account-alias/
   │   │   ├── region/
   │   │   │   └── repo-name/
   ├── github/
   │   ├── org-name/
   │   │   └── repo-name/
   ├── gitlab/
   │   ├── group-name/
   │   │   └── repo-name/
   ```
3. **Reuse Git Operations**: Use same `git.Cloner` for all providers
4. **Handle Rate Limits**: Each provider has different rate limit strategies
5. **Authentication Flexibility**: Support tokens, SSH keys, OAuth, etc.

## Important Considerations

### SSH Key Handling
- SSH keys are passed as file paths; never embed them in code or logs
- Ensure proper file permissions (600) are validated before use
- SSH key errors should fail fast with clear messaging
- The SSH key path is used by git CLI directly, so it must be in a standard location or explicitly added to ssh-agent
- Consider validating that the SSH key exists and is readable before starting the mirror process

### Concurrency
- Process multiple providers in parallel
- Within each provider, process accounts/organizations in parallel
- For CodeCommit: Process regions in parallel within each account
- Use goroutine pools/semaphores to avoid overwhelming APIs or system resources
- Handle context cancellation gracefully (e.g., Ctrl+C)
- Be mindful of rate limits for each provider (may differ significantly)

### Provider-Specific Considerations

**AWS CodeCommit:**
- Always paginate through ListRepositories results (repos > 25 per region)
- Handle region errors gracefully; failed regions shouldn't stop others
- Use STS AssumeRole for cross-account access or manage credentials via `~/.aws/credentials`
- Use AWS credential chain (env vars, credentials file, IAM role)
- Each account needs its own credential context

**GitHub (future):**
- Respect GitHub API rate limits (60 req/hr unauthenticated, 5000 req/hr authenticated)
- Use pagination for organization repos
- Handle GitHub organizations and user accounts differently
- Personal access tokens support OAuth scopes

**GitLab (future):**
- Support both self-hosted and cloud GitLab instances
- Handle group hierarchies (nested groups)
- Respect GitLab API rate limits
- Support both personal access tokens and OAuth

### Output Structure

Cloned repositories are organized by provider, then provider-specific hierarchy:

```
mirror_output/
├── codecommit/                          # AWS CodeCommit provider
│   ├── account-alias-1/                 # AWS account alias
│   │   ├── us-east-1/
│   │   │   ├── repo-name-1/
│   │   │   └── repo-name-2/
│   │   └── eu-west-1/
│   │       └── repo-name-3/
│   └── account-alias-2/
│       ├── us-east-1/
│       │   └── repo-name-4/
├── github/                              # GitHub provider (when implemented)
│   ├── org-name-1/
│   │   ├── repo-name-1/
│   │   └── repo-name-2/
│   └── org-name-2/
│       └── repo-name-3/
├── gitlab/                              # GitLab provider (when implemented)
│   ├── group-name-1/
│   │   └── repo-name-1/
```

**Directory Structure Rules:**
- Top level: Provider name (`codecommit`, `github`, `gitlab`, etc.)
- Second level: Account alias (AWS), organization name (GitHub), or group name (GitLab)
- Third level (CodeCommit only): AWS region
- Leaf level: Repository names
- Directories are created automatically if they don't exist
- Duplicate repository names across providers/accounts/regions are supported (stored in different directories)

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

## Known Issues

**AWS SDK Endpoint Resolution Error:** The tool currently fails when discovering repositories due to an AWS SDK v2 endpoint resolver bug. See [KNOWN_ISSUES.md](../KNOWN_ISSUES.md) for details, workarounds, and recommended solutions.

**Impact:** Blocks repository discovery but does not affect:
- Configuration loading
- Credential validation
- Architecture and design patterns
- Git operations (once repositories are provided)

**Recommended Fix:** Refactor AWS API calls to use AWS CLI via `os/exec` wrapper instead of SDK (proven working alternative).
