// Package orchestrator schedules provider discovery and parallel clone/update.
// It has no mpb dependency; progress is injected via ports.ProgressRenderer.
package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/romayengineer/git-mirror-config"
	core "github.com/romayengineer/git-mirror-core"
	"github.com/romayengineer/git-mirror-gitops/gitsync"
	"github.com/romayengineer/git-mirror-orchestrator/layout"
	"github.com/romayengineer/git-mirror-orchestrator/ports"
)

// Orchestrator runs the mirror workflow.
type Orchestrator interface {
	Mirror(ctx context.Context, config *config.Config) (*core.MirrorStats, error)
}

type orchestrator struct {
	cloner     gitsync.Cloner
	sshKeyPath string
	providers  []core.Provider
	progress   ports.ProgressRenderer
}

// NewOrchestrator creates an Orchestrator. progress may be nil (discard UI).
func NewOrchestrator(cloner gitsync.Cloner, sshKeyPath string, providers []core.Provider, progress ports.ProgressRenderer) Orchestrator {
	if progress == nil {
		progress = ports.DiscardRenderer{}
	}
	return &orchestrator{
		cloner:     cloner,
		sshKeyPath: sshKeyPath,
		providers:  providers,
		progress:   progress,
	}
}

func (o *orchestrator) Mirror(ctx context.Context, config *config.Config) (*core.MirrorStats, error) {
	stats := &core.MirrorStats{
		Errors: []string{},
	}

	concurrency := config.ConcurrentClones
	if concurrency <= 0 {
		concurrency = 5
	}

	var wg sync.WaitGroup
	resultsChan := make(chan *core.CloneResult, 100)
	semaphore := make(chan struct{}, concurrency)

	for _, provider := range o.providers {
		wg.Add(1)
		go func(p core.Provider) {
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
			name := ""
			if result.Repository != nil {
				name = result.Repository.Name
			}
			stats.Errors = append(stats.Errors, fmt.Sprintf("%s: %s", name, result.Message))
		}
		stats.TotalRepositories++
	}

	return stats, nil
}

func (o *orchestrator) processProvider(ctx context.Context, config *config.Config, provider core.Provider, resultsChan chan *core.CloneResult, semaphore chan struct{}) {
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

	handles, done := o.progress.Start(repos)
	defer done()

	var wg sync.WaitGroup
	for i, repo := range repos {
		wg.Add(1)
		go func(r *core.Repository, h ports.RepoProgress) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			r.LocalPath = layout.BuildLocalPath(config.MirrorOutputDir, r.Provider, r.AccountAlias, r.Region, r.Name)
			if r.SSHKeyPath == "" {
				r.SSHKeyPath = o.sshKeyPath
			}

			result := o.cloner.CloneRepository(r)
			if result.Success {
				h.Done(true, "✓ "+result.Message)
			} else {
				h.Done(false, "✗ "+result.Message)
			}

			resultsChan <- result
		}(repo, handles[i])
	}

	wg.Wait()
}
