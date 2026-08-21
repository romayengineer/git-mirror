package mirror

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/romayengineer/git-mirror/internal/git"
	"github.com/romayengineer/git-mirror/internal/providers"
	"github.com/romayengineer/git-mirror/pkg/types"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type Orchestrator interface {
	Mirror(ctx context.Context, config *types.Config) (*types.MirrorStats, error)
}

type orchestrator struct {
	cloner     git.Cloner
	sshKeyPath string
	providers  []providers.Provider
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

	container := mpb.New()

	type barInfo struct {
		bar    *mpb.Bar
		result string
		mu     sync.Mutex
	}
	infos := make([]*barInfo, len(repos))

	for i, repo := range repos {
		bi := &barInfo{}
		infos[i] = bi

		name := fmt.Sprintf("%s/%-30s", repo.AccountAlias, repo.Name)
		if repo.Region != "" {
			name = fmt.Sprintf("%s/%-30s", repo.AccountAlias+"/"+repo.Region, repo.Name)
		}

		bi.bar = container.AddBar(1,
			mpb.BarFillerClearOnComplete(),
			mpb.PrependDecorators(
				decor.Name(name, decor.WC{W: len(name)}),
			),
			mpb.AppendDecorators(
				decor.OnComplete(
					decor.Spinner([]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}),
					""),
				decor.Any(func(s decor.Statistics) string {
					bi.mu.Lock()
					r := bi.result
					bi.mu.Unlock()
					return r
				}),
			),
		)
	}

	totalBar := container.AddBar(int64(len(repos)),
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name("  "),
		),
		mpb.AppendDecorators(
			decor.CountersNoUnit("%d / %d"),
		),
	)

	var barWg sync.WaitGroup
	for i, repo := range repos {
		barWg.Add(1)
		bi := infos[i]
		go func(r *types.Repository, info *barInfo) {
			defer barWg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			r.LocalPath = buildLocalPath(config.MirrorOutputDir, r.Provider, r.AccountAlias, r.Region, r.Name)
			// Only apply global SSH key for non-CodeCommit providers (CodeCommit uses git-remote-codecommit, no SSH)
			if r.Provider != "codecommit" && r.SSHKeyPath == "" {
				r.SSHKeyPath = o.sshKeyPath
			}

			result := o.cloner.CloneRepository(r)
			if result.Success {
				info.mu.Lock()
				info.result = "✓ " + result.Message
				info.mu.Unlock()
			} else {
				info.mu.Lock()
				info.result = "✗ " + result.Message
				info.mu.Unlock()
			}
			info.bar.Increment()
			totalBar.Increment()

			resultsChan <- result
		}(repo, bi)
	}

	barWg.Wait()
	container.Wait()
}

func buildLocalPath(baseDir, provider, accountAlias, region, repoName string) string {
	return filepath.Join(baseDir, provider, accountAlias, region, repoName)
}
