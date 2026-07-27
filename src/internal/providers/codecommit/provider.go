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

	for _, accountID := range p.cfg.AccountIDs {
		profile := p.getProfileForAccount(accountID)

		accountClient, err := aws.NewAccountClientWithLogging(profile, p.cfg.DefaultRegion, p.cfg.LogCommands)
		if err != nil {
			fmt.Printf("✗ Error creating AWS client for account %s with profile %s: %v\n", accountID, profile, err)
			continue
		}

		discoverer := aws.NewRepositoryDiscoverer(accountClient)
		repos, err := discoverer.DiscoverRepositories(ctx, accountID, p.cfg.Regions)
		if err != nil {
			fmt.Printf("✗ Error discovering repositories for account %s: %v\n", accountID, err)
			continue
		}

		for _, repo := range repos {
			repo.Provider = p.Name()
			allRepositories = append(allRepositories, repo)
		}
	}

	return allRepositories, nil
}

func (p *CodeCommitProvider) getProfileForAccount(accountID string) string {
	if profile, exists := p.cfg.AccountProfiles[accountID]; exists {
		return profile
	}
	return p.cfg.CredentialsProfile
}
