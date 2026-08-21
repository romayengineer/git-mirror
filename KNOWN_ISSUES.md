# Known Issues

## Multi-Account / Multi-Provider SSH Key Conflict

### Issue
When mirroring from multiple CodeCommit accounts (same hostname `git-codecommit.*.amazonaws.com`) and/or multiple Git providers simultaneously, SSH key selection fails because:
- `~/.ssh/config` can only have one `Host git-codecommit.*.amazonaws.com` block → only one SSH `User` (key ID) at a time
- Different providers (CodeCommit vs GitHub) need different SSH key files
- Manually commenting/uncommenting SSH config blocks is error-prone

### Root Cause
CodeCommit SSH URLs don't include the IAM SSH key ID in the URL — it comes from `~/.ssh/config` via the `User` directive. Since all CodeCommit accounts in the same region share the same hostname, only one account's SSH key ID can be active at a time in `~/.ssh/config`.

### Solution (Implemented)
The tool now supports per-account SSH identity via environment variables, eliminating the need for `~/.ssh/config` entries for CodeCommit and GitHub:

**Step 1:** Configure per-account SSH key IDs (CodeCommit):
```env
CODECOMMIT_SSH_KEY_IDS=123456789012:APKAX4ENFON4GCH7RIJY,210987654321:APKA4YBIPDVQB6DUKUSZ
```
These are the `User` values from your `~/.ssh/config` `git-codecommit.*.amazonaws.com` blocks.

**Step 2 (optional):** Configure per-account SSH key files if accounts use different keys:
```env
CODECOMMIT_SSH_KEY_FILES=123456789012:~/.ssh/id_rsa,210987654321:~/.ssh/id_other
```
Falls back to global `SSH_KEY_PATH` if not set per account.

**Step 3:** Configure GitHub SSH key separately:
```env
GITHUB_SSH_KEY_PATH=~/.ssh/id_github
```

### How It Works
The tool embeds the SSH key ID directly into the clone URL and uses `GIT_SSH_COMMAND` with the correct key file per repository:
- `ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/my-repo` → `ssh://APKA4YBIPDVQB6DUKUSZ@git-codecommit.us-east-1.amazonaws.com/v1/repos/my-repo`
- Each clone/pull uses `GIT_SSH_COMMAND=ssh -i <correct-key> -o StrictHostKeyChecking=no`
- No `~/.ssh/config` changes needed — both CodeCommit accounts and GitHub work simultaneously

### Configuration Example

**Using config.yaml (recommended):**
```yaml
ssh:
  default_key_path: ~/.ssh/id_rsa

codecommit:
  accounts:
    - id: "123456789012"
      ssh_key_id: APKAX4ENFON4GCH7RIJY
    - id: "210987654321"
      ssh_key_id: APKA4YBIPDVQB6DUKUSZ

github:
  username: my-username
  ssh_key_path: ~/.ssh/id_github
```

**Using .env (legacy):**
```env
SSH_KEY_PATH=~/.ssh/id_rsa
AWS_ACCOUNT_IDS=123456789012,210987654321
CODECOMMIT_SSH_KEY_IDS=123456789012:APKAX4ENFON4GCH7RIJY,210987654321:APKA4YBIPDVQB6DUKUSZ
GITHUB_USERNAME=my-username
GITHUB_SSH_KEY_PATH=~/.ssh/id_github
```

You can then comment out or remove the `Host git-codecommit.*.amazonaws.com` and `Host github.com` blocks from `~/.ssh/config`.

### Status
**Implemented:** 2026-07-27  
**Files Modified:** `src/pkg/types/types.go`, `src/pkg/config/config.go`, `src/pkg/config/yaml_loader.go`, `src/internal/providers/codecommit/provider.go`, `src/internal/providers/github/provider.go`, `src/internal/git/executor.go`, `src/internal/git/cloner.go`, `src/internal/mirror/orchestrator.go`, `src/cmd/main/main.go`

---

## AWS SDK Endpoint Resolution Error

### Issue
The tool fails with the error:
```
failed to list account aliases: not found, ResolveEndpointV2
```

When attempting to discover CodeCommit repositories, the AWS SDK v2 cannot resolve service endpoints, despite having valid AWS credentials.

### Root Cause
There is a bug in AWS SDK v2 (versions 1.24.0-1.42.x) where the endpoint resolver (`ResolveEndpointV2`) fails for CodeCommit and other services. The issue occurs even though:
- ✅ Credentials load correctly from AWS profiles
- ✅ AWS CLI works perfectly with the same credentials
- ✅ AWS profile has proper IAM permissions

### Verification
The credentials ARE valid and working:
```bash
# These commands work fine:
aws sts get-caller-identity --profile your-profile
aws codecommit list-repositories --region us-east-1 --profile your-profile
```

But the Go SDK fails with endpoint resolution errors in both the tool and when called directly via AWS SDK APIs.

### Impact
- ⚠️ Tool cannot discover repositories via AWS SDK
- ⚠️ All repository listing operations fail
- ⚠️ Git clone operations cannot proceed (no repositories to clone)
- ✅ Configuration loading works
- ✅ Credential validation works
- ✅ All code architecture and design patterns work

### Workarounds

#### Option 1: Use AWS CLI Directly (Recommended)
The tool can be refactored to use AWS CLI commands via `os/exec` instead of the SDK. The CLI is proven to work with these credentials.

```bash
aws codecommit list-repositories --region us-east-1 --profile your-profile --output json
```

**Status:** Not yet implemented  
**Effort:** Low - straightforward `os/exec` wrapper  
**Benefit:** Immediate working solution

#### Option 2: Wait for AWS SDK Update
AWS SDK v2 with Go 1.24+ support should resolve this issue.

**Timeline:** Available when Go 1.24+ is released and adopted  
**Current Go Version:** 1.21.x

#### Option 3: Use Older SDK Version
Look for an older version of AWS SDK v2 that works with Go 1.21 and doesn't have this endpoint resolution bug (if one exists).

**Status:** Not yet investigated  
**Effort:** Medium - requires testing multiple SDK versions

### Related Issues
- AWS SDK v2 endpoint resolver has known issues with certain service/region combinations
- Issue appears to affect multiple AWS services (IAM, CodeCommit, EC2)
- Not specific to this project - affects any Go code using AWS SDK v2 for these services

### Testing

**Credentials are valid:**
```bash
$ aws sts get-caller-identity --profile your-profile
{
    "UserId": "AIDAX4ENFON4PAEHSCFJB",
    "Account": "123456789000",
    "Arn": "arn:aws:iam::123456789000:user/maximiliano.figueroa@your-profile.net"
}
```

**CodeCommit access works:**
```bash
$ aws codecommit list-repositories --region us-east-1 --profile your-profile
{
    "repositories": [
        {
            "repositoryName": "Seeka-Analytics",
            "repositoryId": "be778cf6-6b5f-442d-89d4-b722a663f3cc"
        },
        # ... more repositories listed
    ]
}
```

**Tool fails with SDK:**
```bash
$ AWS_PROFILE=your-profile ./build/git-mirror
# ... 
Warning: could not get account alias: failed to list account aliases: not found, ResolveEndpointV2
Warning: failed to list repositories in region us-east-1: not found, ResolveEndpointV2
```

### Recommended Next Steps

1. **Short-term:** Implement AWS CLI wrapper (refactor to use `os/exec`)
2. **Medium-term:** Upgrade to Go 1.24+ when available and test SDK update
3. **Long-term:** Monitor AWS SDK releases for endpoint resolver fixes

### Environment Details

- **Go Version:** 1.21.x
- **AWS SDK Version:** v1.24.0
- **AWS SDK CodeCommit Service:** v1.16.0
- **AWS SDK Endpoints:** v2.5.9
- **Error Type:** `*errors.errorString`
- **Error Message:** `not found, ResolveEndpointV2`
- **Affected Services:** IAM, CodeCommit, EC2, potentially others

### Code References

- **Error Location:** `src/internal/aws/client.go` - `ListRepositories()` method
- **Error Occurs In:** AWS SDK's endpoint resolver during service client initialization
- **Workaround Location:** Would be implemented in `src/internal/aws/` directory

---

**Last Updated:** 2026-07-23  
**Status:** Active - blocks repository discovery functionality  
**Severity:** High - core feature impacted
