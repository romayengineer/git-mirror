// Package github discovers GitHub repositories via the REST API.
// Single responsibility: discovery + mapping to core.Repository. No cloning, no config I/O.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	config "github.com/romayengineer/git-mirror-config"
	core "github.com/romayengineer/git-mirror-core"
)

type ghRepo struct {
	Name   string `json:"name"`
	SSHUrl string `json:"ssh_url"`
	URL    string `json:"html_url"`
}

// GitHubProvider discovers user + org repositories.
type GitHubProvider struct {
	cfg     *config.GitHubConfig
	client  *http.Client
	baseURL string
}

// NewGitHubProvider creates a provider with the default HTTP client.
func NewGitHubProvider(cfg *config.GitHubConfig) *GitHubProvider {
	return NewGitHubProviderWithClient(cfg, &http.Client{Timeout: 30 * time.Second}, "https://api.github.com")
}

// NewGitHubProviderWithClient creates a provider with injected HTTP dependencies (for tests).
func NewGitHubProviderWithClient(cfg *config.GitHubConfig, client *http.Client, baseURL string) *GitHubProvider {
	return &GitHubProvider{
		cfg:     cfg,
		client:  client,
		baseURL: baseURL,
	}
}

// Name returns the provider name used in output paths.
func (p *GitHubProvider) Name() string {
	return "github"
}

// DiscoverRepositories lists user repos then each org's repos. Per-repo errors are
// logged and skipped so one bad org does not fail the whole run.
func (p *GitHubProvider) DiscoverRepositories(ctx context.Context) ([]*core.Repository, error) {
	var allRepositories []*core.Repository

	if p.cfg.Username != "" {
		path := fmt.Sprintf("/users/%s/repos?type=owner", p.cfg.Username)
		repos, err := p.listRepos(ctx, path)
		if err != nil {
			fmt.Printf("✗ Error listing repositories for GitHub user %s: %v\n", p.cfg.Username, err)
		} else {
			for _, repo := range repos {
				allRepositories = append(allRepositories, p.makeRepo(repo, p.cfg.Username))
			}
		}
	}

	for _, org := range p.cfg.Orgs {
		org = strings.TrimSpace(org)
		if org == "" {
			continue
		}
		path := fmt.Sprintf("/orgs/%s/repos", org)
		repos, err := p.listRepos(ctx, path)
		if err != nil {
			fmt.Printf("✗ Error listing repositories for GitHub org %s: %v\n", org, err)
			continue
		}
		for _, repo := range repos {
			allRepositories = append(allRepositories, p.makeRepo(repo, org))
		}
	}

	return allRepositories, nil
}

func (p *GitHubProvider) listRepos(ctx context.Context, path string) ([]ghRepo, error) {
	var allRepos []ghRepo
	page := 1

	for {
		url := fmt.Sprintf("%s%s&per_page=100&page=%d", p.baseURL, path, page)
		if !strings.Contains(path, "?") {
			url = fmt.Sprintf("%s%s?per_page=100&page=%d", p.baseURL, path, page)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Accept", "application/vnd.github.v3+json")
		if p.cfg.Token != "" {
			req.Header.Set("Authorization", "Bearer "+p.cfg.Token)
		}

		resp, err := p.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
		}

		var repos []ghRepo
		if err := json.Unmarshal(body, &repos); err != nil {
			return nil, fmt.Errorf("failed to parse GitHub API response: %w", err)
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

func (p *GitHubProvider) makeRepo(r ghRepo, alias string) *core.Repository {
	return &core.Repository{
		Name:         r.Name,
		CloneURL:     r.SSHUrl,
		SSHCloneURL:  r.SSHUrl,
		HTTPCloneURL: r.URL,
		AccountAlias: alias,
		Provider:     p.Name(),
		SSHKeyPath:   p.cfg.SSHKeyPath,
	}
}
