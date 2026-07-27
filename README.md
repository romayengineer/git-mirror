# git-mirror

A fast, parallel Git repository mirror tool written in Go. Mirror repositories from multiple Git providers (AWS CodeCommit, GitHub, GitLab, etc.) to your local filesystem, organized by provider and account/organization.

## Features

- **Multi-Provider Support**: Mirror from AWS CodeCommit, GitHub, GitLab, and more (GitHub/GitLab support coming soon)
- **Multi-Account/Org Support**: Mirror from multiple AWS accounts or GitHub organizations in one run
- **Multi-Region Support**: For AWS CodeCommit, automatically discover all regions or specify a subset
- **Parallel Cloning**: Concurrent clone operations for performance
- **Smart Organization**: Organizes clones by provider: `provider/account-or-org/region/repository_name`
- **Automatic Updates**: Detects existing repos and pulls latest changes
- **Error Resilience**: Continues if individual repos/regions/providers fail
- **Flexible Authentication**: SSH keys, tokens, and other authentication methods per provider
- **Configuration via .env**: Simple environment-based configuration
- **Extensible Architecture**: Plugin-style provider system makes adding new providers straightforward

## Installation

### Prerequisites

- Go 1.21 or later
- Git (installed and in PATH)
- Provider credentials configured:
  - **AWS CodeCommit**: AWS CLI credentials (via `~/.aws/credentials` or environment variables) + SSH key
  - **GitHub**: Personal access token or SSH key (future support)
  - **GitLab**: Personal access token or SSH key (future support)

### Build and Install

```bash
# Build only
make build
# Output: ./build/git-mirror

# Build and install to ~/.local/bin
make install
# Copies binary to ~/.local/bin and creates config directory
# Next: Edit ~/.config/git-mirror/.env with your settings
```

## Configuration

The tool searches for configuration in this order:
1. `~/.config/git-mirror/.env` (recommended for user-wide config)
2. `.env` in the current directory
3. `../.env` in the parent directory

### Setup

```bash
# Option 1: Quick install (recommended)
make install
# Creates ~/.config/git-mirror/.env from template
# Then edit: nano ~/.config/git-mirror/.env

# Option 2: Manual setup
cp .env.example ~/.config/git-mirror/.env
nano ~/.config/git-mirror/.env
```

### Example Configuration

```env
# Required settings
AWS_ACCOUNT_IDS=123456789012,210987654321
SSH_KEY_PATH=~/.ssh/id_rsa
MIRROR_OUTPUT_DIR=./mirror

# Optional settings
AWS_REGIONS=us-east-1,eu-west-1
AWS_CREDENTIALS_PROFILE=default
CONCURRENT_CLONES=5
```

### Configuration Variables

**Required:**
- `AWS_ACCOUNT_IDS` - Comma-separated list of AWS account IDs to mirror
- `SSH_KEY_PATH` - Path to SSH private key for CodeCommit
- `MIRROR_OUTPUT_DIR` - Local directory for cloned repositories

**Optional:**
- `AWS_REGIONS` - Comma-separated list of regions (auto-discovered if not set)
- `AWS_CREDENTIALS_PROFILE` - AWS credentials profile to use (default: "default")
- `CONCURRENT_CLONES` - Number of concurrent clone operations (default: 5)
- `LOG_LEVEL` - Logging level: debug, info, warn, error (default: "info")
- `GIT_PATH` - Path to git executable (default: "git" - assumes git is in PATH)

## Usage

```bash
make run
```

Or run the binary directly:
```bash
./build/git-mirror
```

The tool will:
1. Load configuration from `.env`
2. For each account, discover all repositories (in specified or all regions)
3. Clone repositories to `MIRROR_OUTPUT_DIR/account_alias/region/repo_name`
4. For existing repos, verify the URL and pull latest changes
5. Show a summary report

## Output Structure

```
mirror_output/
├── account-alias-1/
│   ├── us-east-1/
│   │   ├── repo-name-1/
│   │   └── repo-name-2/
│   ├── eu-west-1/
│   │   └── repo-name-3/
├── account-alias-2/
│   └── us-west-2/
│       └── repo-name-4/
```

## Development

All commands are run from the **project root directory**.

### Build
```bash
make build
```

### Run Tests
```bash
make test
```

### Run Tests with Verbose Output
```bash
make test-verbose
```

### Run with Coverage
```bash
make test-cover
```

### Format Code
```bash
make fmt
```

### Check Code Issues
```bash
make vet
```

### Full Lint Check
```bash
make lint
```

### Run All (clean, format, vet, test, build)
```bash
make all
```

### Show All Available Commands
```bash
make help
```

See [CLAUDE.md](CLAUDE.md) for detailed architecture and development guidelines.

## Multi-Provider Support (Roadmap)

The tool is architected to support multiple Git providers beyond AWS CodeCommit:

**Currently Supported:**
- AWS CodeCommit ✅

**Planned Support:**
- GitHub (organizations and repositories)
- GitLab (self-hosted and cloud)
- Gitea
- Other Git providers

**Architecture Notes:**
The codebase uses interface-based dependency injection to separate provider-specific logic from core orchestration. Adding support for new providers requires:
1. Implement a `RepositoryDiscoverer` for the new provider
2. Implement provider-specific authentication
3. Add configuration options for the provider
4. Update the orchestrator to handle multiple providers in a single run

This design ensures that adding new providers is straightforward and doesn't affect existing functionality.

## Known Issues

### AWS SDK Endpoint Resolution Error

The tool currently fails with a `ResolveEndpointV2` error when trying to discover repositories. This is a known issue with AWS SDK v2 (go version 1.21).

**Status:** The AWS credentials ARE valid and working (confirmed with AWS CLI), but the Go SDK cannot resolve service endpoints.

**Workaround:** Use AWS CLI directly to list repositories (see [KNOWN_ISSUES.md](KNOWN_ISSUES.md) for details and recommended solutions).

See [KNOWN_ISSUES.md](KNOWN_ISSUES.md) for more information and planned fixes.

## License

MIT
