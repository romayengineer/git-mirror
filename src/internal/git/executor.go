package git

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/romayengineer/codecommit-mirror/pkg/exec"
	"github.com/romayengineer/codecommit-mirror/pkg/path"
)

type Executor interface {
	Clone(url, clonePath, sshKeyPath string) error
	Pull(repoPath string) error
	GetOriginURL(repoPath string) (string, error)
	VerifyRepository(repoPath string) error
}

type gitCLIExecutor struct {
	gitPath       string
	commander     exec.Commander
	pathConverter path.PathConverter // Converts paths between WSL and Windows formats as needed
}

func NewExecutor(gitPath string) Executor {
	return NewExecutorWithLogging(gitPath, false)
}

func NewExecutorWithLogging(gitPath string, logCommands bool) Executor {
	return NewExecutorWithPathFormat(gitPath, logCommands, "linux")
}

// NewExecutorWithPathFormat creates a git executor with path format conversion support.
// pathFormat: "windows" for Windows Git from WSL, "linux" (default) for WSL Git.
// This allows using Windows Git (e.g., wgit) from WSL by converting paths to Windows format.
func NewExecutorWithPathFormat(gitPath string, logCommands bool, pathFormat string) Executor {
	return NewExecutorWithCommander(gitPath, exec.NewCommanderWithLogging(logCommands), path.NewPathConverter(pathFormat))
}

func NewExecutorWithCommander(gitPath string, commander exec.Commander, pathConverter path.PathConverter) Executor {
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

func (e *gitCLIExecutor) Pull(repoPath string) error {
	// Convert path format if needed (e.g., /mnt/c/path -> C:\path for Windows Git)
	convertedPath := e.pathConverter.Convert(repoPath)
	cmd := e.commander.Command(context.Background(), e.gitPath, "-C", convertedPath, "fetch", "--all")

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
