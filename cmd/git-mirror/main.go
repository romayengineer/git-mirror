// Command git-mirror is the thin composition root. It wires config,
// providers, gitops, orchestrator and presenter, then runs the mirror.
// All business logic lives in the modules; main only handles wiring,
// startup/report printing and exit codes.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/romayengineer/git-mirror-config"
	"github.com/romayengineer/git-mirror-config/env"
	yamlloader "github.com/romayengineer/git-mirror-config/yaml"
	core "github.com/romayengineer/git-mirror-core"
	"github.com/romayengineer/git-mirror-gitops/gitexec"
	"github.com/romayengineer/git-mirror-gitops/gitsync"
	"github.com/romayengineer/git-mirror-orchestrator"
	codecommitprovider "github.com/romayengineer/git-mirror-provider-codecommit"
	"github.com/romayengineer/git-mirror-provider-codecommit/aws"
	githubprovider "github.com/romayengineer/git-mirror-provider-github"
	"github.com/romayengineer/git-mirror/configfiles"
	"github.com/romayengineer/git-mirror/presenter"
	"github.com/romayengineer/git-mirror/report"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var loader config.Loader

	if configPath := configfiles.FindConfigFile(); configPath != "" {
		loader = yamlloader.NewYAMLLoader(configPath)
	} else {
		envPath := configfiles.FindEnvFile()
		loader = env.NewEnvLoader(envPath)
	}

	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Composition: git executor -> cloner -> providers -> orchestrator.
	// main is the only place that knows all concrete types.
	executor := gitexec.NewExecutorWithPathFormat(cfg.GitPath, cfg.LogCommands, cfg.GitPathFormat)
	cloner := gitsync.NewCloner(executor)

	var providersSlice []core.Provider

	if cfg.CodeCommit != nil {
		if err := aws.CheckAWSCLIAvailable(); err != nil {
			return fmt.Errorf("AWS CLI is required but not available: %w", err)
		}

		for _, account := range cfg.CodeCommit.Accounts {
			if account.Profile != "" {
				os.Setenv("AWS_PROFILE", account.Profile)
				break
			}
		}

		providersSlice = append(providersSlice, codecommitprovider.NewCodeCommitProvider(cfg.CodeCommit))
	}

	if cfg.GitHub != nil {
		providersSlice = append(providersSlice, githubprovider.NewGitHubProvider(cfg.GitHub))
	}

	orch := orchestrator.NewOrchestrator(cloner, cfg.SSHKeyPath, providersSlice, presenter.ConsoleRenderer{})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	report.PrintStartup(cfg)

	startTime := time.Now()

	stats, err := orch.Mirror(ctx, cfg)
	if err != nil {
		return fmt.Errorf("mirror operation failed: %w", err)
	}

	duration := time.Since(startTime)
	stats.Duration = duration.String()

	report.PrintSummary(stats)

	if stats.FailedClones > 0 {
		return fmt.Errorf("mirror operation completed with %d errors", stats.FailedClones)
	}

	return nil
}
