package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/romayengineer/git-mirror/internal/aws"
	"github.com/romayengineer/git-mirror/internal/git"
	"github.com/romayengineer/git-mirror/internal/mirror"
	"github.com/romayengineer/git-mirror/internal/providers"
	codecommitprovider "github.com/romayengineer/git-mirror/internal/providers/codecommit"
	githubprovider "github.com/romayengineer/git-mirror/internal/providers/github"
	"github.com/romayengineer/git-mirror/pkg/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	envPath := findEnvFile()

	loader := config.NewEnvLoader(envPath)
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize git executor with support for Windows paths when using Windows Git from WSL
	executor := git.NewExecutorWithPathFormat(cfg.GitPath, cfg.LogCommands, cfg.GitPathFormat)
	cloner := git.NewCloner(executor)

	var providersSlice []providers.Provider

	if cfg.CodeCommit != nil {
		if err := aws.CheckAWSCLIAvailable(); err != nil {
			return fmt.Errorf("AWS CLI is required but not available: %w", err)
		}

		for _, profile := range cfg.CodeCommit.AccountProfiles {
			if profile != "" {
				os.Setenv("AWS_PROFILE", profile)
				break
			}
		}

		providersSlice = append(providersSlice, codecommitprovider.NewCodeCommitProvider(cfg.CodeCommit))
	}

	if cfg.GitHub != nil {
		providersSlice = append(providersSlice, githubprovider.NewGitHubProvider(cfg.GitHub))
	}

	orchestrator := mirror.NewOrchestrator(cloner, cfg.SSHKeyPath, providersSlice)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	fmt.Println("Starting repository mirror process...")
	fmt.Printf("Git path: %s\n", cfg.GitPath)
	if cfg.GitPathFormat == "windows" {
		fmt.Println("Git path format: WINDOWS (WSL paths will be converted to Windows format)")
	}
	if cfg.CodeCommit != nil {
		fmt.Printf("Accounts: %v\n", cfg.CodeCommit.AccountIDs)
		fmt.Printf("AWS Default Region: %s\n", cfg.CodeCommit.DefaultRegion)
		if len(cfg.CodeCommit.Regions) > 0 {
			fmt.Printf("Regions: %v\n", cfg.CodeCommit.Regions)
		} else {
			fmt.Println("Regions: All available regions will be discovered")
		}
		if len(cfg.CodeCommit.AccountProfiles) > 0 {
			fmt.Printf("Account Profiles: %v\n", cfg.CodeCommit.AccountProfiles)
		}
	}
	if cfg.GitHub != nil {
		if cfg.GitHub.Username != "" {
			fmt.Printf("GitHub User: %s\n", cfg.GitHub.Username)
		}
		if len(cfg.GitHub.Orgs) > 0 {
			fmt.Printf("GitHub Orgs: %v\n", cfg.GitHub.Orgs)
		}
	}
	fmt.Printf("Output directory: %s\n", cfg.MirrorOutputDir)
	fmt.Printf("Concurrent clones: %d\n", cfg.ConcurrentClones)
	if cfg.LogCommands {
		fmt.Println("Command logging: ENABLED")
	}

	startTime := time.Now()

	stats, err := orchestrator.Mirror(ctx, cfg)
	if err != nil {
		return fmt.Errorf("mirror operation failed: %w", err)
	}

	duration := time.Since(startTime)
	stats.Duration = duration.String()

	fmt.Println("\n=== Mirror Summary ===")
	fmt.Printf("Total repositories: %d\n", stats.TotalRepositories)
	fmt.Printf("Successfully processed: %d\n", stats.SuccessfulClones)
	fmt.Printf("Failed: %d\n", stats.FailedClones)
	fmt.Printf("Duration: %s\n", stats.Duration)

	if len(stats.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, errMsg := range stats.Errors {
			fmt.Printf("  - %s\n", errMsg)
		}
	}

	if stats.FailedClones > 0 {
		return fmt.Errorf("mirror operation completed with %d errors", stats.FailedClones)
	}

	return nil
}

func findEnvFile() string {
	homeDir, _ := os.UserHomeDir()
	candidates := []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join(homeDir, ".config", "git-mirror", ".env"),
		filepath.Join(homeDir, ".git-mirror"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ".env"
}
