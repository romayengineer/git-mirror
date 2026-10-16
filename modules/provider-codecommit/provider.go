// Package codecommit discovers AWS CodeCommit repositories.
// Single responsibility: per-account discovery + clone-URL/SSH mapping. No cloning.
package codecommit

import (
	"context"
	"fmt"
	"strings"

	"github.com/romayengineer/git-mirror-config"
	core "github.com/romayengineer/git-mirror-core"
	"github.com/romayengineer/git-mirror-provider-codecommit/aws"
)

// Transport identifies how CodeCommit repositories are cloned.
const (
	// TransportSSH clones over ssh://git-codecommit.<region>.amazonaws.com using
	// an SSH key and the account's SSH key ID as the SSH user. Default.
	TransportSSH = "ssh"
	// TransportGRC clones via git-remote-codecommit using the account's AWS
	// profile (codecommit::<region>://<profile>@<name>). No SSH key required.
	TransportGRC = "grc"
)

// CodeCommitProvider discovers repos across accounts/regions.
type CodeCommitProvider struct {
	cfg *config.CodeCommitConfig
}

// NewCodeCommitProvider creates a CodeCommitProvider.
func NewCodeCommitProvider(cfg *config.CodeCommitConfig) *CodeCommitProvider {
	return &CodeCommitProvider{cfg: cfg}
}

// Name returns the provider name used in output paths.
func (p *CodeCommitProvider) Name() string {
	return "codecommit"
}

// DiscoverRepositories iterates accounts, continuing on per-account errors.
func (p *CodeCommitProvider) DiscoverRepositories(ctx context.Context) ([]*core.Repository, error) {
	var allRepositories []*core.Repository

	for _, account := range p.cfg.Accounts {
		profile := account.Profile
		if profile == "" {
			profile = p.cfg.CredentialsProfile
		}

		accountClient, err := aws.NewAccountClientWithLogging(profile, p.cfg.DefaultRegion, p.cfg.LogCommands)
		if err != nil {
			fmt.Printf("✗ Error creating AWS client for account %s with profile %s: %v\n", account.ID, profile, err)
			continue
		}

		discoverer := aws.NewRepositoryDiscoverer(accountClient)
		repos, err := discoverer.DiscoverRepositories(ctx, account.ID, p.cfg.Regions)
		if err != nil {
			fmt.Printf("✗ Error discovering repositories for account %s: %v\n", account.ID, err)
			continue
		}

		transport := resolveTransport(p.cfg, account)

		for _, repo := range repos {
			switch transport {
			case TransportGRC:
				// git-remote-codecommit authenticates with the AWS profile over
				// HTTPS; any SSH key path is ignored by the helper.
				repo.CloneURL = buildGRCURL(repo.Region, profile, repo.Name)
			default: // TransportSSH
				if account.SSHKeyID != "" {
					repo.SSHCloneURL = InjectSSHUser(repo.SSHCloneURL, account.SSHKeyID)
				}
				if account.SSHKeyFile != "" {
					repo.SSHKeyPath = account.SSHKeyFile
				}
				repo.CloneURL = repo.SSHCloneURL
			}
			repo.Provider = p.Name()
			allRepositories = append(allRepositories, repo)
		}
	}

	return allRepositories, nil
}

// InjectSSHUser injects the CodeCommit SSH user into an ssh:// URL.
// Exported for testing; non-ssh URLs are returned unchanged.
func InjectSSHUser(url, user string) string {
	if user == "" || !strings.HasPrefix(url, "ssh://") {
		return url
	}
	return "ssh://" + user + "@" + url[len("ssh://"):]
}

// resolveTransport returns the effective clone transport for an account:
// the account override if set, otherwise the codecommit-level default,
// otherwise ssh.
func resolveTransport(cfg *config.CodeCommitConfig, account config.CodeCommitAccount) string {
	if t := strings.ToLower(strings.TrimSpace(account.Transport)); t != "" {
		return t
	}
	if cfg != nil {
		if t := strings.ToLower(strings.TrimSpace(cfg.Transport)); t != "" {
			return t
		}
	}
	return TransportSSH
}

// buildGRCURL builds a git-remote-codecommit URL:
// codecommit::<region>://<profile>@<name> (region omitted when empty).
func buildGRCURL(region, profile, name string) string {
	if profile == "" {
		profile = "default"
	}
	if region == "" {
		return fmt.Sprintf("codecommit://%s@%s", profile, name)
	}
	return fmt.Sprintf("codecommit::%s://%s@%s", region, profile, name)
}

// injectSSHUser keeps backward compatibility with existing tests.
func injectSSHUser(url, user string) string {
	return InjectSSHUser(url, user)
}
