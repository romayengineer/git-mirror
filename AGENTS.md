# AGENTS.md

This file provides guidance to AI agents when working with code in this repository.

## Project Overview

**git-mirror** is a multi-provider Git repository mirror tool written in Go. It discovers and clones repositories from multiple Git providers (currently AWS CodeCommit and GitHub) to a local filesystem, maintaining organized mirrors by provider, account/organization, and region.

The tool uses a plugin-style provider architecture with an interface-based design that makes adding new providers straightforward without modifying core orchestration logic.

## Features

- **Multi-Provider Support**: Mirror from AWS CodeCommit and GitHub
- **Multi-Account/Organization Support**: Mirror from multiple AWS accounts or GitHub users/orgs in a single run
- **Multi-Region Support** (CodeCommit): Auto-discover all AWS regions or specify a subset
- **GitHub User + Org Repos**: Mirror personal repos via `GITHUB_USERNAME` and/or org repos via `GITHUB_ORGS`
- **Parallel Cloning**: Concurrent clone operations for performance
- **Smart Directory Organization**: Organizes clones as `provider/account-or-org/[region/]repository_name`
- **Per-Provider Config Blocks**: Each provider has its own config struct — providers are activated by setting their config
- **.env Configuration**: Simple environment variable configuration without CLI flags
- **Error Resilience**: Gracefully handle failures - continue cloning if regions or repositories fail
- **Automatic Updates**: Detect existing repositories and pull latest changes
- **Provider-Specific Authentication**: Each provider handles its own auth (AWS CLI profiles, GitHub tokens, etc.)
- **Account/Org Aliases**: Uses readable aliases instead of just IDs (AWS account names, GitHub usernames/org names)
- **Flexible Concurrency**: Configurable concurrent clone operations
- **Live Progress Display**: Real-time animated progress bars per repo via `mpb` — shows spinner while cloning, ✓/✗ on completion
- **Git CLI Integration**: Uses git CLI directly with a thin wrapper (no git library dependencies)
- **GitHub HTTP API**: GitHub provider uses direct HTTP calls via `net/http` — no external CLI dependency
- **WSL Path Conversion**: Optional Windows path format for using Windows Git from WSL
- **Composable Architecture**: Single-constructor providers with per-provider config types

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
- **Type Definitions** (`src/pkg/types/`): Shared data structures including per-provider config types

## Repository Structure

```
git-mirror/
├── Makefile                                 # Build automation (run from root)
├── README.md
├── AGENTS.md
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
│   │   │   ├── codecommit/
│   │   │   │   └── provider.go              # CodeCommit provider
│   │   │   └── github/
│   │   │       ├── provider.go              # GitHub provider (direct HTTP API)
│   │   │       └── provider_test.go         # Tests with httptest server
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
│       ├── exec/
│       │   ├── commander.go                 # Command execution abstraction
│       │   └── commander_test.go            # Tests for Commander interface
│       ├── path/
│       │   ├── converter.go                 # WSL to Windows path conversion
│       │   └── converter_test.go            # Tests for path conversion
│       └── types/
│           └── types.go                     # Shared data types (Repository, CloneResult, etc.)
└── build/                                   # Build output (git ignored)
    └── git-mirror                            # Compiled binary
```

## Development Commands

All commands are run from the **project root directory** using the Makefile.

### Building
```bash
make build
# Output: ./build/git-mirror
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
./build/git-mirror
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
# Creates binary at ~/.local/bin/git-mirror
# Creates config directory ~/.config/git-mirror/
# Copies config.example.yaml → ~/.config/git-mirror/config.yaml (if not exists)
```

## Configuration

The tool searches for configuration files in this order:
1. `~/.config/git-mirror/config.yaml` (recommended - user-wide config)
2. `config.yaml` (repository root - project-specific config)
3. `../config.yaml` (parent directory - fallback)
4. Falls back to `.env` search (legacy) if no `config.yaml` found

The YAML format is the primary configuration method. Legacy `.env` support is maintained for backward compatibility.

### Setup

**First-time setup (recommended):**
```bash
make install
# Then edit: nano ~/.config/git-mirror/config.yaml
```

**Manual setup:**
```bash
cp config.example.yaml ~/.config/git-mirror/config.yaml
nano ~/.config/git-mirror/config.yaml
```

### YAML Configuration Reference

```yaml
# REQUIRED: Local directory for cloned repositories
output_dir: ./mirror

# OPTIONAL: Number of concurrent clone operations (default: 5)
# concurrent_clones: 5

# OPTIONAL: Logging level: debug, info, warn, error (default: info)
# log_level: info

# OPTIONAL: Log all executed commands (default: false)
# log_commands: false

# OPTIONAL: Git executable settings
# git:
#   path: git                               # Path to git (default: "git")
#   path_format: linux                      # "linux" or "windows" (default: "linux")

# OPTIONAL: Global SSH key (fallback for all providers)
# ssh:
#   default_key_path: ~/.ssh/id_rsa

# --- AWS CodeCommit ---
# Activated when accounts are listed
# codecommit:
#   default_region: us-east-1
#   regions:
#     - us-east-1
#     - us-west-2
#   accounts:
#     - id: "123456789012"
#       profile: production                  # AWS credentials profile
#       ssh_key_id: APKAX4ENFON4GCH7RIJY     # CodeCommit SSH key ID
#       ssh_key_file: ~/.ssh/id_rsa          # Per-account SSH key path

# --- GitHub ---
# Activated when username and/or orgs are set
# github:
#   username: my-username
#   orgs:
#     - my-org
#   token: ghp_xxxxxxxxxxxx
#   ssh_key_path: ~/.ssh/id_github
```

**Provider Activation**:
- **CodeCommit**: Activated when `accounts` has at least one entry
- **GitHub**: Activated when `username` or `orgs` is set
- At least one provider must be configured

**Command Logging**: Set `log_commands: true` to see every command being executed:
```
Executing: git clone git@codecommit:test-repo.git /path/to/mirror
Executing: aws codecommit list-repositories --region us-east-1 --output json
```

**Git Path Format**: Controls how paths are passed to git commands:
- `linux` (default): Passes paths as `/mnt/c/path` (for WSL Git, standard Linux git)
- `windows`: Converts paths to `C:\path` format (for Windows Git from WSL)

**Example for Windows Git:**
```yaml
git:
  path: wgit
  path_format: windows
```

**Example for WSL Git (default):**
```yaml
git:
  path: git
  path_format: linux
```

**Regional Discovery**: If `regions` is not specified, the tool will:
- Fetch all available AWS regions via the EC2 API
- Attempt to list CodeCommit repositories in each region
- Skip regions where CodeCommit is not available
- Clone from all regions that have repositories

**Per-Account Profiles**: Use per-account `profile` field instead of global profile:
```yaml
codecommit:
  accounts:
    - id: "123456789012"
      profile: prod-account
    - id: "210987654321"
      profile: dev-account
```

## Per-Provider Config Architecture

The `Config` struct in `src/pkg/types/types.go` uses per-provider config blocks instead of a flat list of fields:

```go
type Config struct {
    MirrorOutputDir  string
    ConcurrentClones int
    SSHKeyPath       string
    LogLevel         string
    GitPath          string
    LogCommands      bool
    GitPathFormat    string
    GitHub           *GitHubConfig       // nil when not configured
    CodeCommit       *CodeCommitConfig   // nil when not configured
}

type GitHubConfig struct {
    Username string
    Orgs     []string
    Token    string
}

type CodeCommitConfig struct {
    Accounts           []CodeCommitAccount
    Regions            []string
    DefaultRegion      string
    CredentialsProfile string
    LogCommands        bool
}

type CodeCommitAccount struct {
    ID         string
    Profile    string
    SSHKeyID   string
    SSHKeyFile string
}
```

Providers are activated by checking if their config pointer is non-nil:
```go
if cfg.GitHub != nil {
    providers = append(providers, github.NewGitHubProvider(cfg.GitHub))
}
```

Each provider has a **single constructor** that takes its config struct:
```go
func NewGitHubProvider(cfg *types.GitHubConfig) *GitHubProvider
func NewCodeCommitProvider(cfg *types.CodeCommitConfig) *CodeCommitProvider
```

This replaces the previous pattern of multiple constructor variants (e.g., `NewXWithLogging`, `NewXWithClient`).

## Key Files & Packages

- `Makefile`: Build, test, and install automation
  - `make build`: Compiles binary to `./build/git-mirror`
  - `make install`: Installs binary to `~/.local/bin/`, creates `~/.config/git-mirror/`
  - `make test`: Runs all unit tests
- `config.example.yaml`: Template YAML configuration file (all settings commented and documented)
- `src/go.mod` / `src/go.sum`: Dependency management
- `src/cmd/main/main.go`: Entry point (searches for config.yaml file and starts orchestration)
  - Config file search order: `~/.config/git-mirror/config.yaml` → `config.yaml` → `../config.yaml` → fallback `.env`
- `src/internal/aws/`: AWS SDK interactions and region enumeration (uses CLI wrapper for discovery)
- `src/internal/git/`: Thin wrapper around git CLI commands (clone, fetch, etc.)
- `src/internal/mirror/`: Orchestration of discovery + cloning workflow with `mpb` progress bars
- `src/pkg/config/`: .env parsing, validation, and configuration struct
- `src/pkg/exec/`: Command execution abstraction (Commander interface, logging support)
- `src/pkg/path/`: Path conversion utilities (WSL ↔ Windows path conversion)
- `src/pkg/types/`: Shared data structures (Repository, CloneResult, etc.) including per-provider config types (GitHubConfig, CodeCommitConfig)

## Git Command Implementation

The tool uses **git CLI commands directly** via Go's `os/exec` package rather than a pure-Go git library. A thin wrapper in `src/internal/git/` encapsulates common git operations:

- `Clone(url, path)` — executes `git clone`
- `Pull(repoPath)` — executes `git fetch --all` in an existing repository
- Other git operations as needed

**Path Format Support:** All paths passed to git are optionally converted from WSL to Windows format (if `GIT_PATH_FORMAT=windows`), enabling use of Windows Git binaries from WSL.

**Advantages:**
- No external Go dependencies for git functionality
- Uses system-installed git (familiar, reliable)
- Easy to debug (can see exact git commands being run with `LOG_COMMANDS=true`)
- Full feature parity with git CLI
- Support for both WSL Git and Windows Git from WSL

**Requirements:**
- Git must be installed and available in `$PATH`
- SSH must be configured for CodeCommit access
- When using Windows Git from WSL, set `GIT_PATH_FORMAT=windows` for proper path handling

## Command Execution Abstraction

The tool abstracts OS command execution through the **`Commander` interface** (`src/pkg/exec/commander.go`) to enable:
- **Testability**: Mock commands without executing real processes
- **Logging**: Optional logging of all executed commands (controlled by `LOG_COMMANDS` config)
- **Flexibility**: Future implementations can swap command execution strategies

**Core Interfaces:**
```go
type Command interface {
  CombinedOutput() ([]byte, error)
  Run() error
  Output() ([]byte, error)
  SetEnv(env []string)
}

type Commander interface {
  Command(ctx context.Context, name string, args ...string) Command
}
```

**Implementation:**
- `osCommand` — Wraps `exec.Cmd` with standard output/run methods
- `osCommander` — Creates commands with context, optionally logs them

**Usage Pattern:**
```go
// With logging disabled (default)
executor := git.NewExecutor(gitPath)

// With logging enabled
executor := git.NewExecutorWithLogging(gitPath, true)
```

**Testing:**
The `Commander` interface is injected into both `git.Executor` and `aws.CLIExecutor`, allowing tests to pass mock implementations that simulate command execution without actually running processes.

When `LOG_COMMANDS=true` in configuration, all git and AWS CLI commands are logged to stdout in the format:
```
Executing: <command> <arg1> <arg2> ...
```

## WSL Path Handling

When running on Windows Subsystem for Linux (WSL), paths can be represented in two formats:
- **WSL format**: `/mnt/c/Users/Admin/Documents/Projects` (used by WSL/Linux tools)
- **Windows format**: `C:\Users\Admin\Documents\Projects` (used by Windows tools)

### The Problem

If you're using Windows Git (like `wgit`) from WSL, the tool needs to convert paths to Windows format because Windows Git doesn't understand WSL paths. When the tool tries to clone a repository to `/mnt/c/mirror/repo`, Windows Git will fail because it expects `C:\mirror\repo`.

### Solution: GIT_PATH_FORMAT Configuration

Set `GIT_PATH_FORMAT=windows` in your `.env` to automatically convert all git paths from WSL format to Windows format.

**Configuration Example for Windows Git:**
```env
GIT_PATH=wgit
GIT_PATH_FORMAT=windows
```

### Path Converter Implementation

The path conversion is abstracted through the `PathConverter` interface (`src/pkg/path/converter.go`):

```go
type PathConverter interface {
  // Convert transforms a path according to the converter's format
  Convert(path string) string
}
```

**Two implementations:**
- **`linuxConverter`** — No conversion, paths stay as-is (default)
  - `/mnt/c/mirror/repo` stays `/mnt/c/mirror/repo`
  - Used when `GIT_PATH_FORMAT=linux` or not set
  
- **`windowsConverter`** — Converts WSL paths to Windows format
  - `/mnt/c/mirror/repo` becomes `C:\mirror\repo`
  - `/mnt/d/backups` becomes `D:\backups`
  - Non-WSL paths are passed through unchanged
  - Used when `GIT_PATH_FORMAT=windows`

### How It Works

When `GIT_PATH_FORMAT=windows`, the git executor converts every path before passing it to git:

1. User sets `MIRROR_OUTPUT_DIR=/mnt/c/mirror` in `.env`
2. When cloning a repo, the executor receives path `/mnt/c/mirror/repo-name`
3. Path converter transforms it to `C:\mirror\repo-name`
4. Git command receives Windows path it can understand: `wgit clone ... C:\mirror\repo-name`

### Integration Points

- **Configuration** (`src/pkg/config/config.go`): Parses `GIT_PATH_FORMAT` from environment
- **Config Struct** (`src/pkg/types/types.go`): Stores `GitPathFormat` field
- **Git Executor** (`src/internal/git/executor.go`): Injects `PathConverter` and applies conversion to all paths
- **Main** (`src/cmd/main/main.go`): Initializes executor with path format from config

### Testing

The path converter includes comprehensive tests (`src/pkg/path/converter_test.go`) covering:
- Windows path conversion: `/mnt/c/path` → `C:\path`
- Case handling: Uppercase drive letter conversion
- Non-WSL paths: Unchanged when not `/mnt/X/` format
- Case-insensitive format selection: "windows", "WINDOWS", "linux" all work correctly

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
- `exec.Commander` — Interface for OS command execution (create commands, run them, capture output)
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

### Currently Implemented Providers

**AWS CodeCommit:**

**Package Structure:**
```
src/internal/aws/
├── client.go           # AccountClient interface + AWS SDK implementation
├── cli_executor.go     # AWS CLI wrapper for API calls
├── discoverer.go       # RepositoryDiscoverer for CodeCommit
└── discoverer_test.go  # Tests with mocked client
```

**GitHub:**

Package Structure:
```
src/internal/providers/github/
├── provider.go         # GitHub provider (GitHub REST API via net/http)
└── provider_test.go    # Tests with httptest server
```

### Adding New Providers

**Required Components:**

1. **Per-Provider Config Type** — Add a config struct to `src/pkg/types/types.go`
   ```go
   type GitLabConfig struct {
     URL    string
     Token  string
     Groups []string
   }
   ```

2. **Provider Implementation** — Create `src/internal/providers/<name>/provider.go` implementing `providers.Provider`
   ```go
   func NewXProvider(cfg *types.XConfig) *XProvider
   func (p *XProvider) Name() string
   func (p *XProvider) DiscoverRepositories(ctx context.Context) ([]*types.Repository, error)
   ```

3. **Configuration Parsing** — Add a `parseXConfig()` function in `src/pkg/config/config.go`

4. **Wire in main.go** — Add provider to the providers slice when its config is non-nil:
   ```go
   if cfg.GitLab != nil {
     providers = append(providers, gitlab.NewGitLabProvider(cfg.GitLab))
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

**GitHub:**
- Uses direct HTTP calls to the GitHub REST API (`net/http`) — no external CLI dependency
- Lists repos via `GET /users/{username}/repos` (personal) and `GET /orgs/{org}/repos` (org repos)
- Respect GitHub API rate limits (60 req/hr unauthenticated, 5000 req/hr authenticated with token)
- Token sent as `Authorization: Bearer` header when configured via `GITHUB_TOKEN`
- Handles pagination automatically (100 repos per page)
- Supports both personal user repos (`GITHUB_USERNAME`) and organization repos (`GITHUB_ORGS`)

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
├── github/                              # GitHub provider
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
1. **Verify the URL**: Check that the git origin URL matches the expected source URL
2. **If URL matches**: Pull latest changes (`git fetch --all`) to update the repository
3. **If URL doesn't match**: Fail with a clear error message (prevents accidental overwrite of mismatched repos)
4. **Error Recovery**: On URL mismatch, log which repository has the mismatch and continue with other repositories

## Testing Strategy

- **Unit Tests**: Mock all external dependencies (AWS, git) using interfaces
- **AWS Discovery Tests**: Mock `AccountClient` and `RepositoryDiscoverer` interfaces
- **GitHub Provider Tests**: Use `httptest.NewServer` to mock the GitHub REST API
- **Git Operation Tests**: Mock `git.Executor` interface for clone/pull/verify operations
- **Configuration Tests**: Unit test .env parsing and validation
- **Error Scenario Tests**: Use mocks to simulate AWS/git failures, timeouts, etc.
- **Integration Tests** (optional): Use a test AWS account if available for end-to-end testing
- **Test Coverage Goal**: Aim for high coverage through interface mocking (no external dependencies needed for unit tests)

## Known Issues

**AWS SDK Endpoint Resolution Error:** The tool currently fails when discovering repositories due to an AWS SDK v2 endpoint resolver bug. See [KNOWN_ISSUES.md](KNOWN_ISSUES.md) for details, workarounds, and recommended solutions.

**Impact:** Blocks repository discovery but does not affect:
- Configuration loading
- Credential validation
- Architecture and design patterns
- Git operations (once repositories are provided)

**Recommended Fix:** Refactor AWS API calls to use AWS CLI via `os/exec` wrapper instead of SDK (proven working alternative).
