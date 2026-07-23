package mirror

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/romayengineer/codecommit-mirror/internal/aws"
	"github.com/romayengineer/codecommit-mirror/internal/git"
	"github.com/romayengineer/codecommit-mirror/pkg/types"
)

type Orchestrator interface {
	Mirror(ctx context.Context, config *types.Config) (*types.MirrorStats, error)
}

type orchestrator struct {
	cloner          git.Cloner
	sshKeyPath      string
	accountProfiles map[string]string
	defaultRegion   string
}

func NewOrchestrator(cloner git.Cloner, sshKeyPath string, accountProfiles map[string]string, defaultRegion string) Orchestrator {
	return &orchestrator{
		cloner:          cloner,
		sshKeyPath:      sshKeyPath,
		accountProfiles: accountProfiles,
		defaultRegion:   defaultRegion,
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
			profile := o.getProfileForAccount(accID, config.AWSCredentialsProfile)
			o.processAccount(ctx, config, accID, profile, resultsChan, semaphore)
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

func (o *orchestrator) getProfileForAccount(accountID, defaultProfile string) string {
	if profile, exists := o.accountProfiles[accountID]; exists {
		return profile
	}
	return defaultProfile
}

func (o *orchestrator) processAccount(ctx context.Context, config *types.Config, accountID, profile string, resultsChan chan *types.CloneResult, semaphore chan struct{}) {
	accountClient, err := aws.NewAccountClient(profile, o.defaultRegion)
	if err != nil {
		fmt.Printf("✗ Error creating AWS client for account %s with profile %s: %v\n", accountID, profile, err)
		return
	}

	discoverer := aws.NewRepositoryDiscoverer(accountClient)
	repos, err := discoverer.DiscoverRepositories(ctx, accountID, config.AWSRegions)
	if err != nil {
		fmt.Printf("✗ Error discovering repositories for account %s: %v\n", accountID, err)
		return
	}

	fmt.Printf("\n📦 Account %s (%s): Found %d repositories\n", accountID, repos[0].AccountAlias, len(repos))
	fmt.Printf("═══════════════════════════════════════════════════════════════════════\n\n")

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
