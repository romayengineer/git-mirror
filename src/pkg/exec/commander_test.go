package exec

import (
	"context"
	"testing"
)

type mockCommand struct {
	combineOutputErr error
	combineOutput    []byte
	runErr           error
	outputErr        error
	output           []byte
	envSet           []string
}

func (m *mockCommand) CombinedOutput() ([]byte, error) {
	return m.combineOutput, m.combineOutputErr
}

func (m *mockCommand) Run() error {
	return m.runErr
}

func (m *mockCommand) Output() ([]byte, error) {
	return m.output, m.outputErr
}

func (m *mockCommand) SetEnv(env []string) {
	m.envSet = env
}

func TestOSCommanderCreatesCommand(t *testing.T) {
	commander := NewCommander()
	if commander == nil {
		t.Error("NewCommander returned nil")
	}
}

func TestOSCommanderCommandWithContext(t *testing.T) {
	commander := NewCommander()
	ctx := context.Background()
	cmd := commander.Command(ctx, "echo", "test")
	if cmd == nil {
		t.Error("Command returned nil")
	}
}

func TestOSCommandSetEnv(t *testing.T) {
	commander := NewCommander()
	ctx := context.Background()
	cmd := commander.Command(ctx, "echo", "test")
	env := []string{"TEST=value"}
	cmd.SetEnv(env)
	if len(env) == 0 {
		t.Error("environment not set")
	}
}

func TestOSCommandCombinedOutput(t *testing.T) {
	commander := NewCommander()
	ctx := context.Background()
	cmd := commander.Command(ctx, "echo", "test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output) == 0 {
		t.Error("expected output from echo command")
	}
}

func TestOSCommandRun(t *testing.T) {
	commander := NewCommander()
	ctx := context.Background()
	cmd := commander.Command(ctx, "true")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOSCommandOutput(t *testing.T) {
	commander := NewCommander()
	ctx := context.Background()
	cmd := commander.Command(ctx, "echo", "test")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(output) == 0 {
		t.Error("expected output from echo command")
	}
}

func TestMockCommand(t *testing.T) {
	m := &mockCommand{
		output: []byte("test output"),
	}
	output, err := m.Output()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(output) != "test output" {
		t.Errorf("expected 'test output', got '%s'", string(output))
	}
}
