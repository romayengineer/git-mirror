package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/romayengineer/codecommit-mirror/pkg/types"
)

type ghRepo struct {
	Name   string `json:"name"`
	SSHUrl string `json:"ssh_url"`
	URL    string `json:"html_url"`
}

type GitHubProvider struct {
	orgs  []string
	token string
	client *http.Client
}

func NewGitHubProvider(orgs []string, token string) *GitHubProvider {
	return &GitHubProvider{
		orgs:  orgs,
		token: token,
		client: &http.Client{Timeout: 30 * time.Second},
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
				HTTPCloneURL: repo.URL,
				AccountAlias: org,
				Region:       "",
				Provider:     p.Name(),
			})
		}
	}

	return allRepositories, nil
}

func (p *GitHubProvider) listOrgRepositories(ctx context.Context, org string) ([]ghRepo, error) {
	var allRepos []ghRepo
	page := 1

	for {
		url := fmt.Sprintf("https://api.github.com/orgs/%s/repos?per_page=100&page=%d", org, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Accept", "application/vnd.github.v3+json")
		if p.token != "" {
			req.Header.Set("Authorization", "Bearer "+p.token)
		}

		resp, err := p.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch repos for org %s: %w", org, err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body for org %s: %w", org, err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API returned %d for org %s: %s", resp.StatusCode, org, string(body))
		}

		var repos []ghRepo
		if err := json.Unmarshal(body, &repos); err != nil {
			return nil, fmt.Errorf("failed to parse GitHub API response for org %s: %w", org, err)
		}

		if len(repos) == 0 {
			break
		}

		allRepos = append(allRepos, repos...)

		if len(repos) < 100 {
			break
		}

		page++
	}

	return allRepos, nil
}
