// Package gitexec runs git CLI commands. Infrastructure only: no clone-vs-pull policy.
package gitexec

import (
	"context"
	"fmt"
	"os"
	"strings"

	gmexec "github.com/romayengineer/git-mirror-exec"
	gmpath "github.com/romayengineer/git-mirror-path"
)

// Executor abstracts git CLI operations.
type Executor interface {
	Clone(url, clonePath, sshKeyPath string) error
	Pull(repoPath, sshKeyPath string) error
	GetOriginURL(repoPath string) (string, error)
	VerifyRepository(repoPath string) error
}

type gitCLIExecutor struct {
	gitPath       string
	commander     gmexec.Commander
	pathConverter gmpath.PathConverter // Converts paths between WSL and Windows formats as needed
}

// NewExecutor creates a git executor with default settings.
func NewExecutor(gitPath string) Executor {
	return NewExecutorWithLogging(gitPath, false)
}

// NewExecutorWithLogging creates a git executor with optional command logging.
func NewExecutorWithLogging(gitPath string, logCommands bool) Executor {
	return NewExecutorWithPathFormat(gitPath, logCommands, "linux")
}

// NewExecutorWithPathFormat creates a git executor with path format conversion support.
// pathFormat: "windows" for Windows Git from WSL, "linux" (default) for WSL Git.
func NewExecutorWithPathFormat(gitPath string, logCommands bool, pathFormat string) Executor {
	return NewExecutorWithCommander(gitPath, gmexec.NewCommanderWithLogging(logCommands), gmpath.NewPathConverter(pathFormat))
}

// NewExecutorWithCommander creates a git executor with injected dependencies (for tests).
func NewExecutorWithCommander(gitPath string, commander gmexec.Commander, pathConverter gmpath.PathConverter) Executor {
	return &gitCLIExecutor{gitPath: gitPath, commander: commander, pathConverter: pathConverter}
}

func (e *gitCLIExecutor) Clone(url, clonePath, sshKeyPath string) error {
	convertedPath := e.pathConverter.Convert(clonePath)
	cmd := e.commander.Command(context.Background(), e.gitPath, "clone", url, convertedPath)
	if sshKeyPath != "" {
		cmd.SetEnv(append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no", sshKeyPath)))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone repository %s: %w\nOutput: %s", url, err, string(output))
	}

	return nil
}

func (e *gitCLIExecutor) Pull(repoPath, sshKeyPath string) error {
	convertedPath := e.pathConverter.Convert(repoPath)
	cmd := e.commander.Command(context.Background(), e.gitPath, "-C", convertedPath, "fetch", "--all")

	if sshKeyPath != "" {
		cmd.SetEnv(append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no", sshKeyPath)))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch repository at %s: %w", repoPath, err)
	}

	return nil
}

func (e *gitCLIExecutor) GetOriginURL(repoPath string) (string, error) {
	// Convert path format if needed (e.g., /mnt/c/path -> C:\path for Windows Git)
	convertedPath := e.pathConverter.Convert(repoPath)
	cmd := e.commander.Command(context.Background(), e.gitPath, "-C", convertedPath, "config", "--get", "remote.origin.url")

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
