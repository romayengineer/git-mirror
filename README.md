# git-mirror

A fast, parallel Git repository mirror tool written in Go. Mirror repositories from AWS CodeCommit and GitHub to your local filesystem, organized by provider and account/organization.

## Features

- **Multi-Provider Support**: Mirror from AWS CodeCommit and GitHub
- **Multi-Account/Organization Support**: Mirror from multiple AWS accounts or GitHub users/orgs in a single run
- **Multi-Region Support** (CodeCommit): Auto-discover all AWS regions or specify a subset
- **GitHub User + Org Repos**: Mirror personal repos via `GITHUB_USERNAME` and/or org repos via `GITHUB_ORGS`
- **Parallel Cloning**: Concurrent clone operations for performance
- **Smart Directory Organization**: Organizes clones as `provider/account-or-org/[region/]repository_name`
- **Automatic Updates**: Detect existing repos and pull latest changes
- **Error Resilience**: Continue if individual repos/regions/providers fail
- **Flexible Authentication**: SSH keys for CodeCommit, tokens for GitHub
- **Configuration via .env**: Simple environment-based configuration
- **Per-Provider Config**: Each provider activated by its own config variables
- **GitHub HTTP API**: Direct HTTP calls — no external CLI dependency for GitHub
- **WSL Path Conversion**: Optional Windows path format for using Windows Git from WSL

## Installation

### Prerequisites

- Go 1.21 or later
- Git (installed and in PATH)
- **AWS CodeCommit**: AWS CLI installed and configured (`aws --version`)
- **GitHub**: Personal access token (create at https://github.com/settings/tokens)

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
# Always required
MIRROR_OUTPUT_DIR=./mirror

# CodeCommit (optional — activate by setting AWS_ACCOUNT_IDS)
AWS_ACCOUNT_IDS=123456789012,210987654321
SSH_KEY_PATH=~/.ssh/id_rsa
AWS_REGIONS=us-east-1,eu-west-1

# GitHub (optional — activate by setting GITHUB_USERNAME or GITHUB_ORGS)
GITHUB_USERNAME=my-username
GITHUB_TOKEN=ghp_xxxxxxxxxxxx
```

### Configuration Variables

**Required:**
- `MIRROR_OUTPUT_DIR` — Local directory for cloned repositories

**Optional (at least one provider must be configured):**
- `AWS_ACCOUNT_IDS` — Comma-separated list of AWS account IDs (activates CodeCommit)
- `GITHUB_USERNAME` — GitHub username to mirror personal repos (activates GitHub)
- `GITHUB_ORGS` — Comma-separated GitHub organizations to mirror

**General optional:**
- `SSH_KEY_PATH` — Path to SSH private key
- `CONCURRENT_CLONES` — Number of concurrent clone operations (default: 5)
- `LOG_LEVEL` — Logging level: debug, info, warn, error (default: "info")
- `GIT_PATH` — Path to git executable (default: "git")
- `GIT_PATH_FORMAT` — Path format: "linux" or "windows" (default: "linux")

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
2. Discover repositories from all active providers in parallel
3. Clone repositories to `MIRROR_OUTPUT_DIR/provider/account_or_org/[region/]repo_name`
4. For existing repos, verify the URL and pull latest changes
5. Show a summary report

## Output Structure

```
mirror_output/
├── codecommit/
│   ├── account-alias-1/
│   │   ├── us-east-1/
│   │   │   ├── repo-name-1/
│   │   │   └── repo-name-2/
│   │   └── eu-west-1/
│   │       └── repo-name-3/
│   └── account-alias-2/
│       └── repo-name-4/
└── github/
    ├── org-name-1/
    │   ├── repo-name-1/
    │   └── repo-name-2/
    └── user-name/
        └── repo-name-3/
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

See [AGENTS.md](AGENTS.md) for detailed architecture and development guidelines.

## Known Issues

### AWS SDK Endpoint Resolution Error

The tool currently fails with a `ResolveEndpointV2` error when trying to discover CodeCommit repositories. This is a known issue with AWS SDK v2.

**Workaround:** The CodeCommit provider uses AWS CLI directly (via `os/exec`) instead of the SDK, which works correctly. See [KNOWN_ISSUES.md](KNOWN_ISSUES.md) for details.

**Note:** This issue only affects the CodeCommit provider. The GitHub provider uses direct HTTP calls and is unaffected.

## License

MIT
