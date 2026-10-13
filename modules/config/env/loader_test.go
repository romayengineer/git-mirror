package env

import (
	"os"
	"testing"
)

func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		t.Setenv(k, v)
	}
	// Ensure unrelated provider vars are cleared.
	for _, k := range []string{"AWS_ACCOUNT_IDS", "GITHUB_USERNAME", "GITHUB_ORGS", "GITHUB_TOKEN", "GITHUB_SSH_KEY_PATH", "SSH_KEY_PATH", "MIRROR_OUTPUT_DIR", "CONCURRENT_CLONES", "CODECOMMIT_SSH_KEY_IDS", "CODECOMMIT_SSH_KEY_FILES", "AWS_REGIONS", "AWS_DEFAULT_REGION", "AWS_CREDENTIALS_PROFILE", "AWS_ACCOUNT_PROFILES", "GIT_PATH", "GIT_PATH_FORMAT", "LOG_COMMANDS"} {
		if _, ok := kv[k]; !ok {
			os.Unsetenv(k)
		}
	}
}

func TestEnvLoaderLoadSuccess(t *testing.T) {
	setEnv(t, map[string]string{
		"AWS_ACCOUNT_IDS":   "123456789012,210987654321",
		"SSH_KEY_PATH":      os.TempDir(),
		"MIRROR_OUTPUT_DIR": "./mirror",
		"CONCURRENT_CLONES": "10",
	})

	loader := NewEnvLoader("")
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.CodeCommit.Accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(cfg.CodeCommit.Accounts))
	}
	if cfg.CodeCommit.Accounts[0].ID != "123456789012" {
		t.Errorf("expected first account ID to be 123456789012, got %s", cfg.CodeCommit.Accounts[0].ID)
	}
	if cfg.ConcurrentClones != 10 {
		t.Errorf("expected concurrent clones to be 10, got %d", cfg.ConcurrentClones)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected log level to be info, got %s", cfg.LogLevel)
	}
}

func TestParseCodeCommitSSHKeyIDs(t *testing.T) {
	setEnv(t, map[string]string{
		"AWS_ACCOUNT_IDS":        "111111111111,222222222222",
		"MIRROR_OUTPUT_DIR":      "./mirror",
		"CODECOMMIT_SSH_KEY_IDS": "111111111111:AAAAAAA,222222222222:BBBBBBB",
	})

	loader := NewEnvLoader("")
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.CodeCommit == nil {
		t.Fatal("expected CodeCommit config to be set")
	}
	if cfg.CodeCommit.Accounts[0].SSHKeyID != "AAAAAAA" {
		t.Errorf("expected SSH key ID AAAAAAA, got %s", cfg.CodeCommit.Accounts[0].SSHKeyID)
	}
	if cfg.CodeCommit.Accounts[1].SSHKeyID != "BBBBBBB" {
		t.Errorf("expected SSH key ID BBBBBBB, got %s", cfg.CodeCommit.Accounts[1].SSHKeyID)
	}
}

func TestEnvLoaderMissingAllProviders(t *testing.T) {
	setEnv(t, map[string]string{
		"MIRROR_OUTPUT_DIR": "./mirror",
	})

	loader := NewEnvLoader("")
	if _, err := loader.Load(); err == nil {
		t.Fatal("expected error when no providers configured")
	}
}

func TestParseGitHubSSHKeyPath(t *testing.T) {
	setEnv(t, map[string]string{
		"GITHUB_USERNAME":    "testuser",
		"MIRROR_OUTPUT_DIR":  "./mirror",
		"GITHUB_SSH_KEY_PATH": "~/.ssh/id_github_test",
	})

	loader := NewEnvLoader("")
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.GitHub == nil {
		t.Fatal("expected GitHub config to be set")
	}
	expected := os.Getenv("HOME") + "/.ssh/id_github_test"
	if cfg.GitHub.SSHKeyPath != expected {
		t.Errorf("expected SSH key path %s, got %s", expected, cfg.GitHub.SSHKeyPath)
	}
}
