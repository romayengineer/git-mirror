package config

import (
	"os"
	"testing"
)

func TestEnvLoaderLoadSuccess(t *testing.T) {
	os.Setenv("AWS_ACCOUNT_IDS", "123456789012,210987654321")
	os.Setenv("SSH_KEY_PATH", os.TempDir())
	os.Setenv("MIRROR_OUTPUT_DIR", "./mirror")
	os.Setenv("CONCURRENT_CLONES", "10")

	defer func() {
		os.Unsetenv("AWS_ACCOUNT_IDS")
		os.Unsetenv("SSH_KEY_PATH")
		os.Unsetenv("MIRROR_OUTPUT_DIR")
		os.Unsetenv("CONCURRENT_CLONES")
	}()

	loader := NewEnvLoader("")
	cfg, err := loader.Load()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(cfg.CodeCommit.AccountIDs) != 2 {
		t.Errorf("expected 2 account IDs, got %d", len(cfg.CodeCommit.AccountIDs))
	}

	if cfg.CodeCommit.AccountIDs[0] != "123456789012" {
		t.Errorf("expected first account ID to be 123456789012, got %s", cfg.CodeCommit.AccountIDs[0])
	}

	if cfg.ConcurrentClones != 10 {
		t.Errorf("expected concurrent clones to be 10, got %d", cfg.ConcurrentClones)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("expected log level to be info, got %s", cfg.LogLevel)
	}
}

func TestEnvLoaderMissingAllProviders(t *testing.T) {
	os.Setenv("MIRROR_OUTPUT_DIR", "./mirror")

	defer func() {
		os.Unsetenv("MIRROR_OUTPUT_DIR")
	}()

	loader := NewEnvLoader("")
	_, err := loader.Load()

	if err == nil {
		t.Fatal("expected error when no providers configured")
	}
}
