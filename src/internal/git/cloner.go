package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/romayengineer/git-mirror/pkg/types"
)

type Cloner interface {
	CloneRepository(repo *types.Repository) *types.CloneResult
}

type cloner struct {
	executor     Executor
	forceReclone bool
}

func NewCloner(executor Executor) Cloner {
	return &cloner{executor: executor}
}

func NewClonerWithForceReclone(executor Executor, forceReclone bool) Cloner {
	return &cloner{executor: executor, forceReclone: forceReclone}
}

func getCloneURL(repo *types.Repository) string {
	if repo.CloneURL != "" {
		return repo.CloneURL
	}
	return repo.SSHCloneURL
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

	cloneURL := getCloneURL(repo)
	if err := c.executor.Clone(cloneURL, repo.LocalPath, repo.SSHKeyPath); err != nil {
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
		Message:    "cloned",
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

	cloneURL := getCloneURL(repo)
	if !urlsMatch(currentURL, cloneURL) {
		if c.forceReclone {
			// Remove mismatched repo and re-clone
			if err := os.RemoveAll(repo.LocalPath); err != nil {
				return &types.CloneResult{
					Repository: repo,
					Success:    false,
					Error:      err,
					Message:    fmt.Sprintf("URL mismatch (existing: %s, expected: %s) - failed to remove for re-clone: %v", currentURL, cloneURL, err),
				}
			}
			if err := c.ensureDirectoryExists(repo.LocalPath); err != nil {
				return &types.CloneResult{
					Repository: repo,
					Success:    false,
					Error:      err,
					Message:    fmt.Sprintf("failed to create directory for re-clone: %v", err),
				}
			}
			if err := c.executor.Clone(cloneURL, repo.LocalPath, repo.SSHKeyPath); err != nil {
				os.RemoveAll(repo.LocalPath)
				return &types.CloneResult{
					Repository: repo,
					Success:    false,
					Error:      err,
					Message:    fmt.Sprintf("failed to re-clone after URL mismatch: %v", err),
				}
			}
			return &types.CloneResult{
				Repository: repo,
				Success:    true,
				Message:    "re-cloned (URL mismatch)",
			}
		}
		return &types.CloneResult{
			Repository: repo,
			Success:    false,
			Error:      fmt.Errorf("URL mismatch: existing=%s, expected=%s (hint: remove %s or set force_reclone=true)", currentURL, cloneURL, repo.LocalPath),
			Message:    fmt.Sprintf("URL mismatch (existing: %s, expected: %s) - remove %s or set force_reclone: true", currentURL, cloneURL, repo.LocalPath),
		}
	}

	if err := c.executor.Pull(repo.LocalPath, repo.SSHKeyPath); err != nil {
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
		Message:    "updated",
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
