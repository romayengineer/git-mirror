package mirror

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/romayengineer/codecommit-mirror/internal/git"
	"github.com/romayengineer/codecommit-mirror/internal/providers"
	"github.com/romayengineer/codecommit-mirror/pkg/types"
)

type Orchestrator interface {
	Mirror(ctx context.Context, config *types.Config) (*types.MirrorStats, error)
}

type orchestrator struct {
	cloner    git.Cloner
	sshKeyPath string
	providers []providers.Provider
}

func NewOrchestrator(cloner git.Cloner, sshKeyPath string, providers []providers.Provider) Orchestrator {
	return &orchestrator{
		cloner:     cloner,
		sshKeyPath: sshKeyPath,
		providers:  providers,
	}
}

func (o *orchestrator) Mirror(ctx context.Context, config *types.Config) (*types.MirrorStats, error) {
	stats := &types.MirrorStats{
		Errors: []string{},
	}

	var wg sync.WaitGroup
	resultsChan := make(chan *types.CloneResult, 100)
	semaphore := make(chan struct{}, config.ConcurrentClones)

	for _, provider := range o.providers {
		wg.Add(1)
		go func(p providers.Provider) {
			defer wg.Done()
			o.processProvider(ctx, config, p, resultsChan, semaphore)
		}(provider)
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

func (o *orchestrator) processProvider(ctx context.Context, config *types.Config, provider providers.Provider, resultsChan chan *types.CloneResult, semaphore chan struct{}) {
	repos, err := provider.DiscoverRepositories(ctx)
	if err != nil {
		fmt.Printf("✗ Error discovering repositories from provider %s: %v\n", provider.Name(), err)
		return
	}

	if len(repos) == 0 {
		fmt.Printf("📦 Provider %s: Found 0 repositories\n", provider.Name())
		return
	}

	fmt.Printf("\n📦 Provider %s: Found %d repositories\n", provider.Name(), len(repos))
	fmt.Printf("═══════════════════════════════════════════════════════════════════════\n\n")

	var wg sync.WaitGroup
	for _, repo := range repos {
		wg.Add(1)
		go func(r *types.Repository) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			r.LocalPath = buildLocalPath(config.MirrorOutputDir, r.Provider, r.AccountAlias, r.Region, r.Name)
			r.SSHKeyPath = o.sshKeyPath

			result := o.cloner.CloneRepository(r)
			resultsChan <- result
		}(repo)
	}

	wg.Wait()
}

func buildLocalPath(baseDir, provider, accountAlias, region, repoName string) string {
	return filepath.Join(baseDir, provider, accountAlias, region, repoName)
}
