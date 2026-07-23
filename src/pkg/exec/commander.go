package exec

import (
	"context"
	"os/exec"
)

type Command interface {
	CombinedOutput() ([]byte, error)
	Run() error
	Output() ([]byte, error)
	SetEnv(env []string)
}

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

type osCommander struct{}

func (c *osCommander) Command(ctx context.Context, name string, args ...string) Command {
	cmd := exec.CommandContext(ctx, name, args...)
	return &osCommand{cmd: cmd}
}

func NewCommander() Commander {
	return &osCommander{}
}
