package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/romayengineer/codecommit-mirror/pkg/exec"
	"github.com/romayengineer/codecommit-mirror/pkg/types"
)

type ghRepo struct {
	Name   string `json:"name"`
	SSHUrl string `json:"sshUrl"`
	Url    string `json:"url"`
}

type GitHubProvider struct {
	orgs      []string
	token     string
	commander exec.Commander
}

func NewGitHubProvider(orgs []string, token string) *GitHubProvider {
	return NewGitHubProviderWithLogging(orgs, token, false)
}

func NewGitHubProviderWithLogging(orgs []string, token string, logCommands bool) *GitHubProvider {
	return &GitHubProvider{
		orgs:      orgs,
		token:     token,
		commander: exec.NewCommanderWithLogging(logCommands),
	}
}

func (p *GitHubProvider) Name() string {
	return "github"
}

func (p *GitHubProvider) DiscoverRepositories(ctx context.Context) ([]*types.Repository, error) {
	var allRepositories []*types.Repository

	for _, org := range p.orgs {
		org = strings.TrimSpace(org)
		if org == "" {
			continue
		}

		repos, err := p.listOrgRepositories(ctx, org)
		if err != nil {
			fmt.Printf("✗ Error listing repositories for GitHub org %s: %v\n", org, err)
			continue
		}

		for _, repo := range repos {
			allRepositories = append(allRepositories, &types.Repository{
				Name:         repo.Name,
				SSHCloneURL:  repo.SSHUrl,
				HTTPCloneURL: repo.Url,
				AccountID:    "",
				AccountAlias: org,
				Region:       "",
				Provider:     p.Name(),
			})
		}
	}

	return allRepositories, nil
}

func (p *GitHubProvider) listOrgRepositories(ctx context.Context, org string) ([]ghRepo, error) {
	args := []string{"repo", "list", org, "--limit", "1000", "--json", "name,sshUrl,url"}
	cmd := p.commander.Command(ctx, "gh", args...)

	if p.token != "" {
		cmd.SetEnv(append(os.Environ(), fmt.Sprintf("GH_TOKEN=%s", p.token)))
	}

	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("gh command timed out for org %s", org)
		}
		return nil, fmt.Errorf("failed to execute gh repo list for org %s: %w", org, err)
	}

	if len(output) == 0 || string(output) == "[]" {
		return []ghRepo{}, nil
	}

	var repos []ghRepo
	if err := json.Unmarshal(output, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse gh output for org %s: %w", org, err)
	}

	return repos, nil
}

func CheckGHCLIAvailable() error {
	commander := exec.NewCommander()
	cmd := commander.Command(context.Background(), "gh", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitHub CLI (gh) is not available or not installed: %w", err)
	}
	return nil
}
