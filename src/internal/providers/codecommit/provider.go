package codecommit

import (
	"context"
	"fmt"

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
			// Build codecommit:: URL with profile embedded (git-remote-codecommit).
			repo.CloneURL = buildCodeCommitURL(repo.Region, profile, repo.Name)
			repo.Provider = p.Name()
			allRepositories = append(allRepositories, repo)
		}
	}

	return allRepositories, nil
}

// buildCodeCommitURL constructs a git-remote-codecommit URL.
// Format: codecommit::region://profile@repoName  (with profile) or codecommit::region://repoName
// See: https://github.com/awslabs/git-remote-codecommit
func buildCodeCommitURL(region, profile, repoName string) string {
	if profile != "" {
		return fmt.Sprintf("codecommit::%s://%s@%s", region, profile, repoName)
	}
	return fmt.Sprintf("codecommit::%s://%s", region, repoName)
}
