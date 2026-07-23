package codecommit

import (
	"context"
	"fmt"

	"github.com/romayengineer/codecommit-mirror/internal/aws"
	"github.com/romayengineer/codecommit-mirror/pkg/types"
)

type CodeCommitProvider struct {
	accountIDs       []string
	regions          []string
	discoverer       aws.RepositoryDiscoverer
	credentialsProfile string
	accountProfiles  map[string]string
	defaultRegion    string
}

func NewCodeCommitProvider(
	accountIDs []string,
	regions []string,
	credentialsProfile string,
	accountProfiles map[string]string,
	defaultRegion string,
) *CodeCommitProvider {
	return &CodeCommitProvider{
		accountIDs:        accountIDs,
		regions:           regions,
		credentialsProfile: credentialsProfile,
		accountProfiles:   accountProfiles,
		defaultRegion:     defaultRegion,
	}
}

func (p *CodeCommitProvider) Name() string {
	return "codecommit"
}

func (p *CodeCommitProvider) DiscoverRepositories(ctx context.Context) ([]*types.Repository, error) {
	var allRepositories []*types.Repository

	for _, accountID := range p.accountIDs {
		profile := p.getProfileForAccount(accountID)

		accountClient, err := aws.NewAccountClient(profile, p.defaultRegion)
		if err != nil {
			fmt.Printf("✗ Error creating AWS client for account %s with profile %s: %v\n", accountID, profile, err)
			continue
		}

		discoverer := aws.NewRepositoryDiscoverer(accountClient)
		repos, err := discoverer.DiscoverRepositories(ctx, accountID, p.regions)
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
	if profile, exists := p.accountProfiles[accountID]; exists {
		return profile
	}
	return p.credentialsProfile
}
