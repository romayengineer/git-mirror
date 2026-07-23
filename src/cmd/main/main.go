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
	"github.com/romayengineer/codecommit-mirror/pkg/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if err := aws.CheckAWSCLIAvailable(); err != nil {
		return fmt.Errorf("AWS CLI is required but not available: %w", err)
	}

	envPath := findEnvFile()

	loader := config.NewEnvLoader(envPath)
	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if len(cfg.AccountProfiles) > 0 {
		for _, profile := range cfg.AccountProfiles {
			if profile != "" {
				os.Setenv("AWS_PROFILE", profile)
				break
			}
		}
	}

	executor := git.NewExecutor()
	cloner := git.NewCloner(executor)

	codecommitProvider := codecommitprovider.NewCodeCommitProvider(
		cfg.AccountIDs,
		cfg.AWSRegions,
		cfg.AWSCredentialsProfile,
		cfg.AccountProfiles,
		cfg.AWSDefaultRegion,
	)

	providersSlice := []providers.Provider{codecommitProvider}

	orchestrator := mirror.NewOrchestrator(cloner, cfg.SSHKeyPath, providersSlice)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	fmt.Println("Starting repository mirror process...")
	fmt.Printf("Accounts: %v\n", cfg.AccountIDs)
	fmt.Printf("AWS Default Region: %s\n", cfg.AWSDefaultRegion)
	if len(cfg.AWSRegions) > 0 {
		fmt.Printf("Regions: %v\n", cfg.AWSRegions)
	} else {
		fmt.Println("Regions: All available regions will be discovered")
	}
	fmt.Printf("Output directory: %s\n", cfg.MirrorOutputDir)
	fmt.Printf("Concurrent clones: %d\n", cfg.ConcurrentClones)
	if len(cfg.AccountProfiles) > 0 {
		fmt.Printf("Account Profiles: %v\n", cfg.AccountProfiles)
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
