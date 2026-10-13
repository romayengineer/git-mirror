// Package report prints startup info and the final mirror summary.
package report

import (
	"fmt"

	"github.com/romayengineer/git-mirror-config"
	core "github.com/romayengineer/git-mirror-core"
)

// PrintStartup prints configuration summary before mirroring.
func PrintStartup(cfg *config.Config) {
	fmt.Println("Starting repository mirror process...")
	fmt.Printf("Git path: %s\n", cfg.GitPath)
	if cfg.GitPathFormat == "windows" {
		fmt.Println("Git path format: WINDOWS (WSL paths will be converted to Windows format)")
	}
	if cfg.CodeCommit != nil {
		accountIDs := make([]string, len(cfg.CodeCommit.Accounts))
		for i, a := range cfg.CodeCommit.Accounts {
			accountIDs[i] = a.ID
		}
		fmt.Printf("Accounts: %v\n", accountIDs)
		fmt.Printf("AWS Default Region: %s\n", cfg.CodeCommit.DefaultRegion)
		if len(cfg.CodeCommit.Regions) > 0 {
			fmt.Printf("Regions: %v\n", cfg.CodeCommit.Regions)
		} else {
			fmt.Println("Regions: All available regions will be discovered")
		}
		hasProfiles := false
		for _, a := range cfg.CodeCommit.Accounts {
			if a.Profile != "" {
				hasProfiles = true
				break
			}
		}
		if hasProfiles {
			profileMap := make(map[string]string)
			for _, a := range cfg.CodeCommit.Accounts {
				if a.Profile != "" {
					profileMap[a.ID] = a.Profile
				}
			}
			fmt.Printf("Account Profiles: %v\n", profileMap)
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
}

// PrintSummary prints the final stats table.
func PrintSummary(stats *core.MirrorStats) {
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
}
