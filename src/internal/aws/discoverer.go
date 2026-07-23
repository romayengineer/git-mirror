package aws

import (
	"context"
	"fmt"

	"github.com/maximiliano-romay-figueroa/codecommit-mirror/pkg/types"
)

type RepositoryDiscoverer interface {
	DiscoverRepositories(ctx context.Context, accountID string, regions []string) ([]*types.Repository, error)
}

type repositoryDiscoverer struct {
	client AccountClient
}

func NewRepositoryDiscoverer(client AccountClient) RepositoryDiscoverer {
	return &repositoryDiscoverer{client: client}
}

func (d *repositoryDiscoverer) DiscoverRepositories(ctx context.Context, accountID string, regions []string) ([]*types.Repository, error) {
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

	var allRepositories []*types.Repository

	for _, region := range regionsToSearch {
		repos, err := d.client.ListRepositories(ctx, region)
		if err != nil {
			fmt.Printf("Warning: failed to list repositories in region %s: %v\n", region, err)
			continue
		}

		for _, repo := range repos {
			allRepositories = append(allRepositories, &types.Repository{
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
