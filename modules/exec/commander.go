// Package exec abstracts OS command execution for testability and logging.
package exec

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Command abstracts a single OS process.
type Command interface {
	CombinedOutput() ([]byte, error)
	Run() error
	Output() ([]byte, error)
	SetEnv(env []string)
}

// Commander creates Commands.
type Commander interface {
	Command(ctx context.Context, name string, args ...string) Command
}

type osCommand struct {
	cmd *exec.Cmd
}

func (c *osCommand) CombinedOutput() ([]byte, error) {
	return c.cmd.CombinedOutput()
}

func (c *osCommand) Run() error {
	return c.cmd.Run()
}

func (c *osCommand) Output() ([]byte, error) {
	return c.cmd.Output()
}

func (c *osCommand) SetEnv(env []string) {
	c.cmd.Env = env
}

type osCommander struct {
	logCommands bool
}

func (c *osCommander) Command(ctx context.Context, name string, args ...string) Command {
	if c.logCommands {
		cmdString := name
		if len(args) > 0 {
			cmdString += " " + strings.Join(args, " ")
		}
		fmt.Printf("Executing: %s\n", cmdString)
	}

	cmd := exec.CommandContext(ctx, name, args...)
	return &osCommand{cmd: cmd}
}

// NewCommander creates a Commander without command logging.
func NewCommander() Commander {
	return NewCommanderWithLogging(false)
}

// NewCommanderWithLogging creates a Commander with optional command logging.
func NewCommanderWithLogging(logCommands bool) Commander {
	return &osCommander{logCommands: logCommands}
}
