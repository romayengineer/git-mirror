package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/romayengineer/git-mirror/pkg/types"
)

func newTestProvider(cfg *types.GitHubConfig, serverURL string, client *http.Client) *GitHubProvider {
	return &GitHubProvider{
		cfg:     cfg,
		client:  client,
		baseURL: serverURL,
	}
}

func TestDiscoverRepositoriesUserRepos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/testuser/repos" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("per_page") != "100" {
			t.Fatalf("expected per_page=100, got %s", r.URL.Query().Get("per_page"))
		}
		if r.URL.Query().Get("type") != "owner" {
			t.Fatalf("expected type=owner, got %s", r.URL.Query().Get("type"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]ghRepo{
			{Name: "repo1", SSHUrl: "git@github.com:testuser/repo1.git", URL: "https://github.com/testuser/repo1"},
			{Name: "repo2", SSHUrl: "git@github.com:testuser/repo2.git", URL: "https://github.com/testuser/repo2"},
		})
	}))
	defer server.Close()

	provider := newTestProvider(&types.GitHubConfig{Username: "testuser"}, server.URL, server.Client())
	repos, err := provider.DiscoverRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
	if repos[0].Name != "repo1" || repos[0].AccountAlias != "testuser" || repos[0].Provider != "github" {
		t.Errorf("unexpected repo fields: %+v", repos[0])
	}
	if repos[1].SSHCloneURL != "git@github.com:testuser/repo2.git" {
		t.Errorf("unexpected SSH URL: %s", repos[1].SSHCloneURL)
	}
}

func TestDiscoverRepositoriesOrgRepos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orgs/testorg/repos" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]ghRepo{
			{Name: "org-repo1", SSHUrl: "git@github.com:testorg/org-repo1.git", URL: "https://github.com/testorg/org-repo1"},
		})
	}))
	defer server.Close()

	provider := newTestProvider(&types.GitHubConfig{Orgs: []string{"testorg"}}, server.URL, server.Client())
	repos, err := provider.DiscoverRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Name != "org-repo1" || repos[0].AccountAlias != "testorg" {
		t.Errorf("unexpected repo fields: %+v", repos[0])
	}
}

func TestDiscoverRepositoriesUserAndOrg(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/users/testuser/repos":
			json.NewEncoder(w).Encode([]ghRepo{
				{Name: "user-repo", SSHUrl: "git@github.com:testuser/user-repo.git"},
			})
		case "/orgs/testorg/repos":
			json.NewEncoder(w).Encode([]ghRepo{
				{Name: "org-repo", SSHUrl: "git@github.com:testorg/org-repo.git"},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	provider := newTestProvider(&types.GitHubConfig{Username: "testuser", Orgs: []string{"testorg"}}, server.URL, server.Client())
	repos, err := provider.DiscoverRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestDiscoverRepositoriesPagination(t *testing.T) {
	page := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		w.Header().Set("Content-Type", "application/json")
		if page == 1 {
			repos := make([]ghRepo, 100)
			for i := 0; i < 100; i++ {
				repos[i] = ghRepo{Name: fmt.Sprintf("repo-%d", i+1)}
			}
			json.NewEncoder(w).Encode(repos)
		} else {
			json.NewEncoder(w).Encode([]ghRepo{
				{Name: "repo-101"},
			})
		}
	}))
	defer server.Close()

	provider := newTestProvider(&types.GitHubConfig{Username: "testuser"}, server.URL, server.Client())
	repos, err := provider.DiscoverRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repos) != 101 {
		t.Fatalf("expected 101 repos, got %d", len(repos))
	}
	if page != 2 {
		t.Errorf("expected 2 pages, got %d", page)
	}
}

func TestDiscoverRepositoriesEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]ghRepo{})
	}))
	defer server.Close()

	provider := newTestProvider(&types.GitHubConfig{Username: "testuser"}, server.URL, server.Client())
	repos, err := provider.DiscoverRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repos) != 0 {
		t.Fatalf("expected 0 repos, got %d", len(repos))
	}
}

func TestDiscoverRepositoriesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"rate limit exceeded"}`))
	}))
	defer server.Close()

	provider := newTestProvider(&types.GitHubConfig{Username: "testuser"}, server.URL, server.Client())
	repos, err := provider.DiscoverRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected no error from DiscoverRepositories (error is logged, not returned), got %v", err)
	}
	if len(repos) != 0 {
		t.Fatalf("expected 0 repos on API error, got %d", len(repos))
	}
}

func TestDiscoverRepositoriesNoUserOrOrg(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make any API calls")
	}))
	defer server.Close()

	provider := newTestProvider(&types.GitHubConfig{}, server.URL, server.Client())
	repos, err := provider.DiscoverRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repos) != 0 {
		t.Fatalf("expected 0 repos, got %d", len(repos))
	}
}

func TestDiscoverRepositoriesTokenSent(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]ghRepo{{Name: "repo1"}})
	}))
	defer server.Close()

	provider := newTestProvider(&types.GitHubConfig{Username: "testuser", Token: "ghp_test123"}, server.URL, server.Client())
	_, err := provider.DiscoverRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if authHeader != "Bearer ghp_test123" {
		t.Errorf("expected Bearer token, got %s", authHeader)
	}
}

func TestProviderName(t *testing.T) {
	provider := NewGitHubProvider(&types.GitHubConfig{Username: "testuser"})
	if provider.Name() != "github" {
		t.Errorf("expected name 'github', got %s", provider.Name())
	}
}

func TestDiscoverRepositoriesOrgWithWhitespace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]ghRepo{
			{Name: "org-repo", SSHUrl: "git@github.com:myorg/org-repo.git"},
		})
	}))
	defer server.Close()

	provider := newTestProvider(&types.GitHubConfig{Orgs: []string{"  myorg  "}}, server.URL, server.Client())
	repos, err := provider.DiscoverRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].AccountAlias != "myorg" {
		t.Errorf("expected trimmed org name 'myorg', got %s", repos[0].AccountAlias)
	}
}

func TestDiscoverRepositoriesWithSSHKeyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]ghRepo{
			{Name: "repo1", SSHUrl: "git@github.com:testuser/repo1.git"},
		})
	}))
	defer server.Close()

	provider := newTestProvider(&types.GitHubConfig{
		Username:   "testuser",
		SSHKeyPath: "~/.ssh/id_github_test",
	}, server.URL, server.Client())

	repos, err := provider.DiscoverRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}

	if repos[0].SSHKeyPath != "~/.ssh/id_github_test" {
		t.Errorf("expected SSHKeyPath '~/.ssh/id_github_test', got '%s'", repos[0].SSHKeyPath)
	}
}

func TestDiscoverRepositoriesWithoutSSHKeyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]ghRepo{
			{Name: "repo1", SSHUrl: "git@github.com:testuser/repo1.git"},
		})
	}))
	defer server.Close()

	provider := newTestProvider(&types.GitHubConfig{
		Username: "testuser",
	}, server.URL, server.Client())

	repos, err := provider.DiscoverRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}

	if repos[0].SSHKeyPath != "" {
		t.Errorf("expected empty SSHKeyPath, got '%s'", repos[0].SSHKeyPath)
	}
}

func TestDiscoverRepositoriesMultipleOrgs(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]ghRepo{
			{Name: fmt.Sprintf("repo-from-%s", r.URL.Path)},
		})
	}))
	defer server.Close()

	provider := newTestProvider(&types.GitHubConfig{Orgs: []string{"org1", "org2", "org3"}}, server.URL, server.Client())
	repos, err := provider.DiscoverRepositories(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(repos) != 3 {
		t.Fatalf("expected 3 repos (1 per org), got %d", len(repos))
	}
	if callCount != 3 {
		t.Errorf("expected 3 API calls, got %d", callCount)
	}
}
