package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Executor interface {
	Clone(url, path, sshKeyPath string) error
	Pull(repoPath string) error
	GetOriginURL(repoPath string) (string, error)
	VerifyRepository(repoPath string) error
}

type gitCLIExecutor struct {
	gitPath string
}

func NewExecutor(gitPath string) Executor {
	return &gitCLIExecutor{gitPath: gitPath}
}

func (e *gitCLIExecutor) Clone(url, path, sshKeyPath string) error {
	cmd := exec.Command(e.gitPath, "clone", url, path)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no", sshKeyPath))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone repository %s: %w\nOutput: %s", url, err, string(output))
	}

	return nil
}

func (e *gitCLIExecutor) Pull(repoPath string) error {
	cmd := exec.Command(e.gitPath, "-C", repoPath, "fetch", "--all")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch repository at %s: %w", repoPath, err)
	}

	return nil
}

func (e *gitCLIExecutor) GetOriginURL(repoPath string) (string, error) {
	cmd := exec.Command(e.gitPath, "-C", repoPath, "config", "--get", "remote.origin.url")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get origin URL from %s: %w", repoPath, err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (e *gitCLIExecutor) VerifyRepository(repoPath string) error {
	gitDir := fmt.Sprintf("%s/.git", repoPath)
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("not a valid git repository: %s", repoPath)
	}

	return nil
}
