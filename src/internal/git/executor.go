package git

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/romayengineer/codecommit-mirror/pkg/exec"
)

type Executor interface {
	Clone(url, path, sshKeyPath string) error
	Pull(repoPath string) error
	GetOriginURL(repoPath string) (string, error)
	VerifyRepository(repoPath string) error
}

type gitCLIExecutor struct {
	gitPath   string
	commander exec.Commander
}

func NewExecutor(gitPath string) Executor {
	return NewExecutorWithCommander(gitPath, exec.NewCommander())
}

func NewExecutorWithCommander(gitPath string, commander exec.Commander) Executor {
	return &gitCLIExecutor{gitPath: gitPath, commander: commander}
}

func (e *gitCLIExecutor) Clone(url, path, sshKeyPath string) error {
	cmd := e.commander.Command(context.Background(), e.gitPath, "clone", url, path)
	cmd.SetEnv(append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no", sshKeyPath)))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone repository %s: %w\nOutput: %s", url, err, string(output))
	}

	return nil
}

func (e *gitCLIExecutor) Pull(repoPath string) error {
	cmd := e.commander.Command(context.Background(), e.gitPath, "-C", repoPath, "fetch", "--all")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch repository at %s: %w", repoPath, err)
	}

	return nil
}

func (e *gitCLIExecutor) GetOriginURL(repoPath string) (string, error) {
	cmd := e.commander.Command(context.Background(), e.gitPath, "-C", repoPath, "config", "--get", "remote.origin.url")

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
