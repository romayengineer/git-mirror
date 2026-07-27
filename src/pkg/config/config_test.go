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

func TestParseCodeCommitSSHKeyIDs(t *testing.T) {
	os.Setenv("AWS_ACCOUNT_IDS", "111111111111,222222222222")
	os.Setenv("MIRROR_OUTPUT_DIR", "./mirror")
	os.Setenv("CODECOMMIT_SSH_KEY_IDS", "111111111111:AAAAAAA,222222222222:BBBBBBB")

	defer func() {
		os.Unsetenv("AWS_ACCOUNT_IDS")
		os.Unsetenv("MIRROR_OUTPUT_DIR")
		os.Unsetenv("CODECOMMIT_SSH_KEY_IDS")
	}()

	loader := NewEnvLoader("")
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.CodeCommit == nil {
		t.Fatal("expected CodeCommit config to be set")
	}

	if len(cfg.CodeCommit.SSHKeyIDs) != 2 {
		t.Errorf("expected 2 SSH key ID mappings, got %d", len(cfg.CodeCommit.SSHKeyIDs))
	}

	if cfg.CodeCommit.SSHKeyIDs["111111111111"] != "AAAAAAA" {
		t.Errorf("expected SSH key ID AAAAAAA for account 111111111111, got %s", cfg.CodeCommit.SSHKeyIDs["111111111111"])
	}

	if cfg.CodeCommit.SSHKeyIDs["222222222222"] != "BBBBBBB" {
		t.Errorf("expected SSH key ID BBBBBBB for account 222222222222, got %s", cfg.CodeCommit.SSHKeyIDs["222222222222"])
	}
}

func TestParseCodeCommitSSHKeyIDsEmpty(t *testing.T) {
	os.Setenv("AWS_ACCOUNT_IDS", "111111111111")
	os.Setenv("MIRROR_OUTPUT_DIR", "./mirror")

	defer func() {
		os.Unsetenv("AWS_ACCOUNT_IDS")
		os.Unsetenv("MIRROR_OUTPUT_DIR")
	}()

	loader := NewEnvLoader("")
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.CodeCommit == nil {
		t.Fatal("expected CodeCommit config to be set")
	}

	if len(cfg.CodeCommit.SSHKeyIDs) != 0 {
		t.Errorf("expected 0 SSH key ID mappings when env not set, got %d", len(cfg.CodeCommit.SSHKeyIDs))
	}
}

func TestParseCodeCommitSSHKeyFiles(t *testing.T) {
	os.Setenv("AWS_ACCOUNT_IDS", "111111111111,222222222222")
	os.Setenv("MIRROR_OUTPUT_DIR", "./mirror")
	os.Setenv("CODECOMMIT_SSH_KEY_FILES", "111111111111:~/.ssh/id_rsa,222222222222:~/.ssh/other_key")

	defer func() {
		os.Unsetenv("AWS_ACCOUNT_IDS")
		os.Unsetenv("MIRROR_OUTPUT_DIR")
		os.Unsetenv("CODECOMMIT_SSH_KEY_FILES")
	}()

	loader := NewEnvLoader("")
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.CodeCommit == nil {
		t.Fatal("expected CodeCommit config to be set")
	}

	if len(cfg.CodeCommit.SSHKeyFiles) != 2 {
		t.Errorf("expected 2 SSH key file mappings, got %d", len(cfg.CodeCommit.SSHKeyFiles))
	}

	expected1 := os.Getenv("HOME") + "/.ssh/id_rsa"
	if cfg.CodeCommit.SSHKeyFiles["111111111111"] != expected1 {
		t.Errorf("expected SSH key file %s for account 111111111111, got %s", expected1, cfg.CodeCommit.SSHKeyFiles["111111111111"])
	}

	expected2 := os.Getenv("HOME") + "/.ssh/other_key"
	if cfg.CodeCommit.SSHKeyFiles["222222222222"] != expected2 {
		t.Errorf("expected SSH key file %s for account 222222222222, got %s", expected2, cfg.CodeCommit.SSHKeyFiles["222222222222"])
	}
}

func TestParseGitHubSSHKeyPath(t *testing.T) {
	os.Setenv("GITHUB_USERNAME", "testuser")
	os.Setenv("MIRROR_OUTPUT_DIR", "./mirror")
	os.Setenv("GITHUB_SSH_KEY_PATH", "~/.ssh/id_github_test")

	defer func() {
		os.Unsetenv("GITHUB_USERNAME")
		os.Unsetenv("MIRROR_OUTPUT_DIR")
		os.Unsetenv("GITHUB_SSH_KEY_PATH")
	}()

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

func TestParseGitHubSSHKeyPathNotSet(t *testing.T) {
	os.Setenv("GITHUB_USERNAME", "testuser")
	os.Setenv("MIRROR_OUTPUT_DIR", "./mirror")

	defer func() {
		os.Unsetenv("GITHUB_USERNAME")
		os.Unsetenv("MIRROR_OUTPUT_DIR")
	}()

	loader := NewEnvLoader("")
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.GitHub == nil {
		t.Fatal("expected GitHub config to be set")
	}

	if cfg.GitHub.SSHKeyPath != "" {
		t.Errorf("expected empty SSH key path when env not set, got %s", cfg.GitHub.SSHKeyPath)
	}
}

func TestParseKeyValuePairs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:     "single pair",
			input:    "key1:value1",
			expected: map[string]string{"key1": "value1"},
		},
		{
			name:     "multiple pairs",
			input:    "key1:value1,key2:value2",
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:     "with whitespace",
			input:    "  key1:value1  ,  key2:value2  ",
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:     "malformed pair skipped",
			input:    "key1:value1,invalidpair,key2:value2",
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseKeyValuePairs(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d pairs, got %d", len(tt.expected), len(result))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("expected [%s]=%s, got [%s]=%s", k, v, k, result[k])
				}
			}
		})
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
