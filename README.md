# git-mirror

A fast, parallel Git repository mirror tool written in Go. Mirror repositories from AWS CodeCommit and GitHub to your local filesystem, organized by provider and account/organization.

## Features

- **Multi-Provider Support**: Mirror from AWS CodeCommit and GitHub
- **Multi-Account/Organization Support**: Mirror from multiple AWS accounts or GitHub users/orgs in a single run
- **Multi-Region Support** (CodeCommit): Auto-discover all AWS regions or specify a subset
- **GitHub User + Org Repos**: Mirror personal repos via `GITHUB_USERNAME` and/or org repos via `GITHUB_ORGS`
- **Parallel Cloning**: Concurrent clone operations for performance
- **Live Progress Display**: Real-time animated progress bars per repo — shows spinner while cloning, ✓/✗ on completion
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

- Go 1.25 or later
- [just](https://github.com/casey/just) command runner (`brew install just`)
- Git (installed and in PATH)
- **AWS CodeCommit**: AWS CLI installed and configured (`aws --version`)
- **GitHub**: Personal access token (create at https://github.com/settings/tokens)

### Build and Install

```bash
# Build only
just build
# Output: ./build/git-mirror

# Build and install to ~/.local/bin
just install
# Copies binary to ~/.local/bin and creates config directory
# Next: Edit ~/.config/git-mirror/config.yaml with your settings
```

## Configuration

The tool searches for configuration in this order:
1. `~/.config/git-mirror/config.yaml` (recommended for user-wide config)
2. `config.yaml` in the current directory
3. `../config.yaml` in the parent directory
4. Falls back to `.env` search (legacy) if no `config.yaml` found

### Setup

```bash
# Option 1: Quick install (recommended)
just install
# Creates ~/.config/git-mirror/config.yaml from template
# Then edit: nano ~/.config/git-mirror/config.yaml

# Option 2: Manual setup
cp config.example.yaml ~/.config/git-mirror/config.yaml
nano ~/.config/git-mirror/config.yaml
```

### Example Configuration (`config.yaml`)

```yaml
# Always required
output_dir: ./mirror

# CodeCommit (optional — activate by listing accounts)
# codecommit:
#   accounts:
#     - id: "123456789012"
#   regions:
#     - us-east-1
#     - eu-west-1

# GitHub (optional — activate by setting username and/or orgs)
github:
  username: my-username
  token: ghp_xxxxxxxxxxxx
```

See `config.example.yaml` for all settings. Legacy `.env` is still supported
(see `.env.example.legacy` for variable names).

### Configuration Variables

### Configuration Variables

**Required:**
- `output_dir` — Local directory for cloned repositories

**Optional (at least one provider must be configured):**
- `codecommit.accounts` — List of AWS accounts (each with `id`, optional `profile`, `transport`, `ssh_key_id`, `ssh_key_file`)
- `codecommit.transport` — Clone transport for CodeCommit: `ssh` (default) or `grc` (git-remote-codecommit via AWS profile); overridable per account
- `github.username` — GitHub username to mirror personal repos (activates GitHub)
- `github.orgs` — List of GitHub organizations to mirror

**General optional:**
- `ssh.default_key_path` — Path to SSH private key
- `concurrent_clones` — Number of concurrent clone operations (default: 5)
- `log_level` — Logging level: debug, info, warn, error (default: "info")
- `git.path` — Path to git executable (default: "git")
- `git.path_format` — Path format: "linux" or "windows" (default: "linux")

## Usage

```bash
just run
```

Or run the binary directly:
```bash
./build/git-mirror
```

The tool will:
1. Load configuration from `config.yaml` (or `.env` legacy fallback)
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
just build
```

### Run Tests
```bash
just test
```

### Run Tests with Verbose Output
```bash
just test-verbose
```

### Run with Coverage
```bash
just test-cover
```

### Format Code
```bash
just fmt
```

### Check Code Issues
```bash
just vet
```

### Full Lint Check
```bash
just lint
```

### Run All (clean, format, vet, test, build)
```bash
just all
```

### Show All Available Commands
```bash
just
```

See [AGENTS.md](AGENTS.md) for detailed architecture and development guidelines.

## Known Issues

### AWS SDK Endpoint Resolution Error

The tool currently fails with a `ResolveEndpointV2` error when trying to discover CodeCommit repositories. This is a known issue with AWS SDK v2.

**Workaround:** The CodeCommit provider uses AWS CLI directly (via `os/exec`) instead of the SDK, which works correctly. See [KNOWN_ISSUES.md](KNOWN_ISSUES.md) for details.

**Note:** This issue only affects the CodeCommit provider. The GitHub provider uses direct HTTP calls and is unaffected.

## License

MIT
