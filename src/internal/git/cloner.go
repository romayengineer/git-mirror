package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/maximiliano-romay-figueroa/codecommit-mirror/pkg/types"
)

type Cloner interface {
	CloneRepository(repo *types.Repository) *types.CloneResult
}

type cloner struct {
	executor Executor
}

func NewCloner(executor Executor) Cloner {
	return &cloner{executor: executor}
}

func (c *cloner) CloneRepository(repo *types.Repository) *types.CloneResult {
	if repo.LocalPath == "" {
		return &types.CloneResult{
			Repository: repo,
			Success:    false,
			Error:      fmt.Errorf("local path not set"),
			Message:    "local path not set for repository",
		}
	}

	if err := c.ensureDirectoryExists(repo.LocalPath); err != nil {
		return &types.CloneResult{
			Repository: repo,
			Success:    false,
			Error:      err,
			Message:    fmt.Sprintf("failed to create directory: %v", err),
		}
	}

	if c.repositoryExists(repo.LocalPath) {
		return c.handleExistingRepository(repo)
	}

	if err := c.executor.Clone(repo.SSHCloneURL, repo.LocalPath, repo.SSHKeyPath); err != nil {
		os.RemoveAll(repo.LocalPath)
		return &types.CloneResult{
			Repository: repo,
			Success:    false,
			Error:      err,
			Message:    fmt.Sprintf("failed to clone repository: %v", err),
		}
	}

	return &types.CloneResult{
		Repository: repo,
		Success:    true,
		Message:    "repository cloned successfully",
	}
}

func (c *cloner) handleExistingRepository(repo *types.Repository) *types.CloneResult {
	currentURL, err := c.executor.GetOriginURL(repo.LocalPath)
	if err != nil {
		return &types.CloneResult{
			Repository: repo,
			Success:    false,
			Error:      err,
			Message:    fmt.Sprintf("failed to verify existing repository: %v", err),
		}
	}

	if !urlsMatch(currentURL, repo.SSHCloneURL) {
		return &types.CloneResult{
			Repository: repo,
			Success:    false,
			Error:      fmt.Errorf("URL mismatch: existing=%s, expected=%s", currentURL, repo.SSHCloneURL),
			Message:    fmt.Sprintf("repository URL mismatch (existing: %s, expected: %s)", currentURL, repo.SSHCloneURL),
		}
	}

	if err := c.executor.Pull(repo.LocalPath); err != nil {
		return &types.CloneResult{
			Repository: repo,
			Success:    false,
			Error:      err,
			Message:    fmt.Sprintf("failed to pull latest changes: %v", err),
		}
	}

	return &types.CloneResult{
		Repository: repo,
		Success:    true,
		Message:    "repository updated with latest changes",
	}
}

func (c *cloner) ensureDirectoryExists(path string) error {
	parentDir := filepath.Dir(path)
	return os.MkdirAll(parentDir, 0755)
}

func (c *cloner) repositoryExists(path string) bool {
	gitDir := filepath.Join(path, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

func urlsMatch(url1, url2 string) bool {
	url1 = normalizeURL(url1)
	url2 = normalizeURL(url2)
	return url1 == url2
}

func normalizeURL(url string) string {
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimSuffix(url, "/")
	return url
}
