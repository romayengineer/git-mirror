// Package codecommit discovers AWS CodeCommit repositories.
// Single responsibility: per-account discovery + SSH mapping. No cloning.
package codecommit

import (
	"context"
	"fmt"
	"strings"

	"github.com/romayengineer/git-mirror-config"
	core "github.com/romayengineer/git-mirror-core"
	"github.com/romayengineer/git-mirror-provider-codecommit/aws"
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

		for _, repo := range repos {
			if account.SSHKeyID != "" {
				repo.SSHCloneURL = InjectSSHUser(repo.SSHCloneURL, account.SSHKeyID)
			}
			if account.SSHKeyFile != "" {
				repo.SSHKeyPath = account.SSHKeyFile
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

// injectSSHUser keeps backward compatibility with existing tests.
func injectSSHUser(url, user string) string {
	return InjectSSHUser(url, user)
}
