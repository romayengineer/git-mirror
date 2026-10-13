# git-mirror-provider-codecommit

CodeCommit discovery across accounts/regions.

- `aws/client.go` — SDK credential check
- `aws/cli.go` — AWS CLI JSON execution/parsing
- `aws/discovery.go` — alias fallback + region fan-out
- `provider.go` — per-account SSH mapping
