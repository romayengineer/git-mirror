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
	user    string
	orgs    []string
	token   string
	client  *http.Client
	baseURL string
}

func NewGitHubProvider(user string, orgs []string, token string) *GitHubProvider {
	return NewGitHubProviderWithClient(user, orgs, token, &http.Client{Timeout: 30 * time.Second})
}

func NewGitHubProviderWithClient(user string, orgs []string, token string, client *http.Client) *GitHubProvider {
	return &GitHubProvider{
		user:    user,
		orgs:    orgs,
		token:   token,
		client:  client,
		baseURL: "https://api.github.com",
	}
}

func (p *GitHubProvider) Name() string {
	return "github"
}

func (p *GitHubProvider) DiscoverRepositories(ctx context.Context) ([]*types.Repository, error) {
	var allRepositories []*types.Repository

	if p.user != "" {
		repos, err := p.listUserRepositories(ctx, p.user)
		if err != nil {
			fmt.Printf("✗ Error listing repositories for GitHub user %s: %v\n", p.user, err)
		} else {
			for _, repo := range repos {
				allRepositories = append(allRepositories, &types.Repository{
					Name:         repo.Name,
					SSHCloneURL:  repo.SSHUrl,
					HTTPCloneURL: repo.URL,
					AccountAlias: p.user,
					Region:       "",
					Provider:     p.Name(),
				})
			}
		}
	}

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

func (p *GitHubProvider) listUserRepositories(ctx context.Context, username string) ([]ghRepo, error) {
	var allRepos []ghRepo
	page := 1

	for {
		url := fmt.Sprintf("%s/users/%s/repos?per_page=100&page=%d&type=owner", p.baseURL, username, page)
		repos, err := p.fetchPage(ctx, url)
		if err != nil {
			return nil, err
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

func (p *GitHubProvider) listOrgRepositories(ctx context.Context, org string) ([]ghRepo, error) {
	var allRepos []ghRepo
	page := 1

	for {
		url := fmt.Sprintf("%s/orgs/%s/repos?per_page=100&page=%d", p.baseURL, org, page)
		repos, err := p.fetchPage(ctx, url)
		if err != nil {
			return nil, err
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

func (p *GitHubProvider) fetchPage(ctx context.Context, url string) ([]ghRepo, error) {
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
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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

	return repos, nil
}
