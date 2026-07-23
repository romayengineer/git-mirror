# AWS CodeCommit Mirror

A fast, parallel AWS CodeCommit mirror tool written in Go. Mirror all repositories from one or more AWS accounts across all regions to your local filesystem.

## Features

- **Multi-Account Support**: Mirror repositories from multiple AWS accounts in a single run
- **Multi-Region Support**: Automatically discover all AWS regions or specify a subset
- **Parallel Cloning**: Concurrent clone operations for performance
- **Smart Organization**: Organizes clones as `account_alias/region/repository_name`
- **Automatic Updates**: Detects existing repos and pulls latest changes
- **Error Resilience**: Continues if individual repos/regions fail
- **SSH-Based**: Uses SSH keys for CodeCommit access
- **Configuration via .env**: Simple environment-based configuration

## Installation

### Prerequisites

- Go 1.21 or later
- Git (installed and in PATH)
- AWS CLI credentials configured (via `~/.aws/credentials` or environment variables)
- SSH key configured for CodeCommit access

### Build and Install

```bash
# Build only
make build
# Output: ./build/codecommit-mirror

# Build and install to ~/.local/bin
make install
# Copies binary to ~/.local/bin and creates config directory
# Next: Edit ~/.config/codecommit-mirror/.env with your settings
```

## Configuration

The tool searches for configuration in this order:
1. `~/.config/codecommit-mirror/.env` (recommended for user-wide config)
2. `.env` in the current directory
3. `../.env` in the parent directory

### Setup

```bash
# Option 1: Quick install (recommended)
make install
# Creates ~/.config/codecommit-mirror/.env from template
# Then edit: nano ~/.config/codecommit-mirror/.env

# Option 2: Manual setup
cp .env.example ~/.config/codecommit-mirror/.env
nano ~/.config/codecommit-mirror/.env
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

## Usage

```bash
make run
```

Or run the binary directly:
```bash
./build/codecommit-mirror
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

## Known Issues

### AWS SDK Endpoint Resolution Error

The tool currently fails with a `ResolveEndpointV2` error when trying to discover repositories. This is a known issue with AWS SDK v2 (go version 1.21).

**Status:** The AWS credentials ARE valid and working (confirmed with AWS CLI), but the Go SDK cannot resolve service endpoints.

**Workaround:** Use AWS CLI directly to list repositories (see [KNOWN_ISSUES.md](KNOWN_ISSUES.md) for details and recommended solutions).

See [KNOWN_ISSUES.md](KNOWN_ISSUES.md) for more information and planned fixes.

## License

MIT
