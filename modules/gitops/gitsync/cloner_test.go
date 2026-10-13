package gitsync

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	core "github.com/romayengineer/git-mirror-core"
)

type mockExecutor struct {
	cloneErr    error
	pullErr     error
	getURLErr   error
	urlToReturn string
	verifyErr   error
}

func (m *mockExecutor) Clone(url, path, sshKeyPath string) error {
	if m.cloneErr != nil {
		return m.cloneErr
	}
	os.MkdirAll(path, 0755)
	os.MkdirAll(filepath.Join(path, ".git"), 0755)
	return nil
}

func (m *mockExecutor) Pull(repoPath, sshKeyPath string) error {
	return m.pullErr
}

func (m *mockExecutor) GetOriginURL(repoPath string) (string, error) {
	if m.getURLErr != nil {
		return "", m.getURLErr
	}
	return m.urlToReturn, nil
}

func (m *mockExecutor) VerifyRepository(repoPath string) error {
	return m.verifyErr
}

func TestClonerNewRepository(t *testing.T) {
	tmpDir := t.TempDir()
	repo := &core.Repository{
		Name:        "test-repo",
		SSHCloneURL: "git@codecommit:test-repo.git",
		LocalPath:   filepath.Join(tmpDir, "test-repo"),
		SSHKeyPath:  "/path/to/key",
	}

	mockExec := &mockExecutor{}
	cloner := NewCloner(mockExec)

	result := cloner.CloneRepository(repo)

	if !result.Success {
		t.Errorf("expected clone to succeed, got error: %v", result.Error)
	}
}

func TestClonerExistingRepositoryURLMatch(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)

	repo := &core.Repository{
		Name:        "test-repo",
		SSHCloneURL: "git@codecommit:test-repo.git",
		LocalPath:   repoPath,
		SSHKeyPath:  "/path/to/key",
	}

	mockExec := &mockExecutor{
		urlToReturn: "git@codecommit:test-repo.git",
	}
	cloner := NewCloner(mockExec)

	result := cloner.CloneRepository(repo)

	if !result.Success {
		t.Errorf("expected pull to succeed, got error: %v", result.Error)
	}
}

func TestClonerExistingRepositoryURLMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)

	repo := &core.Repository{
		Name:        "test-repo",
		SSHCloneURL: "git@codecommit:test-repo.git",
		LocalPath:   repoPath,
		SSHKeyPath:  "/path/to/key",
	}

	mockExec := &mockExecutor{
		urlToReturn: "git@codecommit:different-repo.git",
	}
	cloner := NewCloner(mockExec)

	result := cloner.CloneRepository(repo)

	if result.Success {
		t.Error("expected clone to fail due to URL mismatch")
	}

	if result.Error == nil {
		t.Error("expected error for URL mismatch")
	}
}

func TestClonerPullError(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)

	repo := &core.Repository{
		Name:        "test-repo",
		SSHCloneURL: "git@codecommit:test-repo.git",
		LocalPath:   repoPath,
		SSHKeyPath:  "/path/to/key",
	}

	mockExec := &mockExecutor{
		urlToReturn: "git@codecommit:test-repo.git",
		pullErr:     fmt.Errorf("pull failed"),
	}
	cloner := NewCloner(mockExec)

	result := cloner.CloneRepository(repo)

	if result.Success {
		t.Error("expected clone to fail due to pull error")
	}
}
