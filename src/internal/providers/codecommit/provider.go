package codecommit

import (
	"context"
	"fmt"
	"strings"

	"github.com/romayengineer/git-mirror/internal/aws"
	"github.com/romayengineer/git-mirror/pkg/types"
)

type CodeCommitProvider struct {
	cfg *types.CodeCommitConfig
}

func NewCodeCommitProvider(cfg *types.CodeCommitConfig) *CodeCommitProvider {
	return &CodeCommitProvider{cfg: cfg}
}

func (p *CodeCommitProvider) Name() string {
	return "codecommit"
}

func (p *CodeCommitProvider) DiscoverRepositories(ctx context.Context) ([]*types.Repository, error) {
	var allRepositories []*types.Repository

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
				repo.SSHCloneURL = injectSSHUser(repo.SSHCloneURL, account.SSHKeyID)
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

func injectSSHUser(url, user string) string {
	if user == "" || !strings.HasPrefix(url, "ssh://") {
		return url
	}
	return "ssh://" + user + "@" + url[len("ssh://"):]
}
