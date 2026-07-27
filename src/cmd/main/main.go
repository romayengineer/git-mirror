package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/romayengineer/codecommit-mirror/internal/aws"
	"github.com/romayengineer/codecommit-mirror/internal/git"
	"github.com/romayengineer/codecommit-mirror/internal/mirror"
	"github.com/romayengineer/codecommit-mirror/internal/providers"
	codecommitprovider "github.com/romayengineer/codecommit-mirror/internal/providers/codecommit"
	githubprovider "github.com/romayengineer/codecommit-mirror/internal/providers/github"
	"github.com/romayengineer/codecommit-mirror/pkg/config"
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

	if len(cfg.AccountIDs) > 0 {
		if err := aws.CheckAWSCLIAvailable(); err != nil {
			return fmt.Errorf("AWS CLI is required but not available: %w", err)
		}

		if len(cfg.AccountProfiles) > 0 {
			for _, profile := range cfg.AccountProfiles {
				if profile != "" {
					os.Setenv("AWS_PROFILE", profile)
					break
				}
			}
		}

		codecommitProvider := codecommitprovider.NewCodeCommitProviderWithLogging(
			cfg.AccountIDs,
			cfg.AWSRegions,
			cfg.AWSCredentialsProfile,
			cfg.AccountProfiles,
			cfg.AWSDefaultRegion,
			cfg.LogCommands,
		)
		providersSlice = append(providersSlice, codecommitProvider)
	}

	if cfg.GitHubUsername != "" || len(cfg.GitHubOrgs) > 0 {
		githubProvider := githubprovider.NewGitHubProvider(
			cfg.GitHubUsername,
			cfg.GitHubOrgs,
			cfg.GitHubToken,
		)
		providersSlice = append(providersSlice, githubProvider)
	}

	orchestrator := mirror.NewOrchestrator(cloner, cfg.SSHKeyPath, providersSlice)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	fmt.Println("Starting repository mirror process...")
	fmt.Printf("Git path: %s\n", cfg.GitPath)
	if cfg.GitPathFormat == "windows" {
		fmt.Println("Git path format: WINDOWS (WSL paths will be converted to Windows format)")
	}
	if len(cfg.AccountIDs) > 0 {
		fmt.Printf("Accounts: %v\n", cfg.AccountIDs)
		fmt.Printf("AWS Default Region: %s\n", cfg.AWSDefaultRegion)
		if len(cfg.AWSRegions) > 0 {
			fmt.Printf("Regions: %v\n", cfg.AWSRegions)
		} else {
			fmt.Println("Regions: All available regions will be discovered")
		}
		if len(cfg.AccountProfiles) > 0 {
			fmt.Printf("Account Profiles: %v\n", cfg.AccountProfiles)
		}
	}
	if cfg.GitHubUsername != "" {
		fmt.Printf("GitHub User: %s\n", cfg.GitHubUsername)
	}
	if len(cfg.GitHubOrgs) > 0 {
		fmt.Printf("GitHub Orgs: %v\n", cfg.GitHubOrgs)
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
		filepath.Join(homeDir, ".config", "codecommit-mirror", ".env"),
		filepath.Join(homeDir, ".codecommit-mirror"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ".env"
}
