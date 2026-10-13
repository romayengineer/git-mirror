package aws

import (
	"context"
	"fmt"

	core "github.com/romayengineer/git-mirror-core"
)

// RepositoryDiscoverer discovers CodeCommit repos across regions.
type RepositoryDiscoverer interface {
	DiscoverRepositories(ctx context.Context, accountID string, regions []string) ([]*core.Repository, error)
}

type repositoryDiscoverer struct {
	client AccountClient
}

// NewRepositoryDiscoverer creates a RepositoryDiscoverer.
func NewRepositoryDiscoverer(client AccountClient) RepositoryDiscoverer {
	return &repositoryDiscoverer{client: client}
}

func (d *repositoryDiscoverer) DiscoverRepositories(ctx context.Context, accountID string, regions []string) ([]*core.Repository, error) {
	accountAlias, err := d.client.GetAccountAlias(ctx, accountID)
	if err != nil {
		fmt.Printf("Warning: could not get account alias (using account ID instead): %v\n", err)
		accountAlias = accountID
	}

	var regionsToSearch []string
	if len(regions) == 0 {
		discoveredRegions, err := d.client.ListRegions(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to discover regions: %w", err)
		}
		regionsToSearch = discoveredRegions
	} else {
		regionsToSearch = regions
	}

	var allRepositories []*core.Repository

	for _, region := range regionsToSearch {
		repos, err := d.client.ListRepositories(ctx, region)
		if err != nil {
			fmt.Printf("Warning: failed to list repositories in region %s: %v\n", region, err)
			continue
		}

		for _, repo := range repos {
			allRepositories = append(allRepositories, &core.Repository{
				Name:          repo.Name,
				RepositoryURL: repo.RepositoryURL,
				HTTPCloneURL:  repo.HTTPCloneURL,
				SSHCloneURL:   repo.SSHCloneURL,
				ARN:           repo.ARN,
				AccountID:     accountID,
				AccountAlias:  accountAlias,
				Region:        region,
			})
		}
	}

	return allRepositories, nil
}
