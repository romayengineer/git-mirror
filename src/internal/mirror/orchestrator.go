package mirror

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/maximiliano-romay-figueroa/codecommit-mirror/internal/aws"
	"github.com/maximiliano-romay-figueroa/codecommit-mirror/internal/git"
	"github.com/maximiliano-romay-figueroa/codecommit-mirror/pkg/types"
)

type Orchestrator interface {
	Mirror(ctx context.Context, config *types.Config) (*types.MirrorStats, error)
}

type orchestrator struct {
	discoverer aws.RepositoryDiscoverer
	cloner     git.Cloner
	sshKeyPath string
}

func NewOrchestrator(discoverer aws.RepositoryDiscoverer, cloner git.Cloner, sshKeyPath string) Orchestrator {
	return &orchestrator{
		discoverer: discoverer,
		cloner:     cloner,
		sshKeyPath: sshKeyPath,
	}
}

func (o *orchestrator) Mirror(ctx context.Context, config *types.Config) (*types.MirrorStats, error) {
	stats := &types.MirrorStats{
		Errors: []string{},
	}

	var wg sync.WaitGroup
	resultsChan := make(chan *types.CloneResult, 100)

	semaphore := make(chan struct{}, config.ConcurrentClones)

	for _, accountID := range config.AccountIDs {
		wg.Add(1)
		go func(accID string) {
			defer wg.Done()
			o.processAccount(ctx, config, accID, resultsChan, semaphore)
		}(accountID)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		if result.Success {
			stats.SuccessfulClones++
		} else {
			stats.FailedClones++
			stats.Errors = append(stats.Errors, fmt.Sprintf("%s: %s", result.Repository.Name, result.Message))
		}
		stats.TotalRepositories++
	}

	return stats, nil
}

func (o *orchestrator) processAccount(ctx context.Context, config *types.Config, accountID string, resultsChan chan *types.CloneResult, semaphore chan struct{}) {
	repos, err := o.discoverer.DiscoverRepositories(ctx, accountID, config.AWSRegions)
	if err != nil {
		fmt.Printf("Error discovering repositories for account %s: %v\n", accountID, err)
		return
	}

	var wg sync.WaitGroup
	for _, repo := range repos {
		wg.Add(1)
		go func(r *types.Repository) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			r.LocalPath = buildLocalPath(config.MirrorOutputDir, r.AccountAlias, r.Region, r.Name)
			r.SSHKeyPath = o.sshKeyPath

			result := o.cloner.CloneRepository(r)
			resultsChan <- result
		}(repo)
	}

	wg.Wait()
}

func buildLocalPath(baseDir, accountAlias, region, repoName string) string {
	return filepath.Join(baseDir, accountAlias, region, repoName)
}
