package git

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/romayengineer/git-mirror/pkg/types"
)

type mockExecutor struct {
	cloneErr    error
	pullErr     error
	getURLErr   error
	urlToReturn string
	verifyErr   error
	lastCloneURL string
}

func (m *mockExecutor) Clone(url, path, sshKeyPath string) error {
	m.lastCloneURL = url
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
	repo := &types.Repository{
		Name:        "test-repo",
		CloneURL:    "codecommit::us-east-1://profile@test-repo",
		LocalPath:   filepath.Join(tmpDir, "test-repo"),
	}

	mockExec := &mockExecutor{}
	cloner := NewCloner(mockExec)

	result := cloner.CloneRepository(repo)

	if !result.Success {
		t.Errorf("expected clone to succeed, got error: %v", result.Error)
	}
	if mockExec.lastCloneURL != "codecommit::us-east-1://profile@test-repo" {
		t.Errorf("expected clone URL codecommit::..., got %s", mockExec.lastCloneURL)
	}
}

func TestClonerNewRepositoryFallbackSSHCloneURL(t *testing.T) {
	tmpDir := t.TempDir()
	repo := &types.Repository{
		Name:        "test-repo",
		SSHCloneURL: "git@github.com:org/repo.git",
		LocalPath:   filepath.Join(tmpDir, "test-repo"),
		SSHKeyPath:  "/path/to/key",
	}

	mockExec := &mockExecutor{}
	cloner := NewCloner(mockExec)

	result := cloner.CloneRepository(repo)

	if !result.Success {
		t.Errorf("expected clone to succeed, got error: %v", result.Error)
	}
	if mockExec.lastCloneURL != "git@github.com:org/repo.git" {
		t.Errorf("expected fallback SSH URL, got %s", mockExec.lastCloneURL)
	}
}

func TestClonerExistingRepositoryURLMatch(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)

	repo := &types.Repository{
		Name:      "test-repo",
		CloneURL:  "codecommit::us-east-1://profile@test-repo",
		LocalPath: repoPath,
	}

	mockExec := &mockExecutor{
		urlToReturn: "codecommit::us-east-1://profile@test-repo",
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

	repo := &types.Repository{
		Name:      "test-repo",
		CloneURL:  "codecommit::us-east-1://profile@test-repo",
		LocalPath: repoPath,
	}

	mockExec := &mockExecutor{
		urlToReturn: "codecommit::us-east-1://profile@different-repo",
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

func TestClonerExistingRepositoryURLMismatchForceReclone(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)

	repo := &types.Repository{
		Name:      "test-repo",
		CloneURL:  "codecommit::us-east-1://profile@test-repo",
		LocalPath: repoPath,
	}

	mockExec := &mockExecutor{
		urlToReturn: "ssh://old-url@test-repo",
	}
	cloner := NewClonerWithForceReclone(mockExec, true)

	result := cloner.CloneRepository(repo)

	if !result.Success {
		t.Errorf("expected re-clone to succeed with forceReclone, got error: %v", result.Error)
	}
	if result.Message != "re-cloned (URL mismatch)" {
		t.Errorf("expected re-cloned message, got %s", result.Message)
	}
}

func TestClonerExistingRepositorySSHToCodeCommitMismatchForceRecloneFalse(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)

	repo := &types.Repository{
		Name:      "test-repo",
		CloneURL:  "codecommit::us-east-1://profile@test-repo",
		LocalPath: repoPath,
	}

	mockExec := &mockExecutor{
		urlToReturn: "ssh://APKAX@git-codecommit.us-east-1.amazonaws.com/v1/repos/test-repo",
	}
	cloner := NewClonerWithForceReclone(mockExec, false)

	result := cloner.CloneRepository(repo)

	if result.Success {
		t.Error("expected failure without forceReclone")
	}
}

func TestClonerPullError(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	os.MkdirAll(filepath.Join(repoPath, ".git"), 0755)

	repo := &types.Repository{
		Name:      "test-repo",
		CloneURL:  "codecommit::us-east-1://profile@test-repo",
		LocalPath: repoPath,
	}

	mockExec := &mockExecutor{
		urlToReturn: "codecommit::us-east-1://profile@test-repo",
		pullErr:     fmt.Errorf("pull failed"),
	}
	cloner := NewCloner(mockExec)

	result := cloner.CloneRepository(repo)

	if result.Success {
		t.Error("expected clone to fail due to pull error")
	}
}
