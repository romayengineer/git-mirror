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

func TestEnvLoaderForceReclone(t *testing.T) {
	os.Setenv("AWS_ACCOUNT_IDS", "111111111111")
	os.Setenv("MIRROR_OUTPUT_DIR", "./mirror")
	os.Setenv("FORCE_RECLONE", "true")

	defer func() {
		os.Unsetenv("AWS_ACCOUNT_IDS")
		os.Unsetenv("MIRROR_OUTPUT_DIR")
		os.Unsetenv("FORCE_RECLONE")
	}()

	loader := NewEnvLoader("")
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !cfg.ForceReclone {
		t.Error("expected ForceReclone true")
	}
}

func TestEnvLoaderForceRecloneDefaultFalse(t *testing.T) {
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

	if cfg.ForceReclone {
		t.Error("expected ForceReclone false by default")
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

func TestYAMLLoaderFull(t *testing.T) {
	yamlContent := []byte(`
output_dir: ./mirror
concurrent_clones: 10
log_level: debug
log_commands: true
force_reclone: true

git:
  path: /usr/bin/git
  path_format: windows

ssh:
  default_key_path: ~/.ssh/id_rsa

codecommit:
  default_region: eu-west-1
  regions:
    - us-east-1
    - eu-west-1
  accounts:
    - id: "111111111111"
      profile: prod
    - id: "222222222222"
      profile: dev

github:
  username: testuser
  orgs:
    - my-org
  token: ghp_test123
  ssh_key_path: ~/.ssh/id_github
`)

	tmpFile, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpFile.Write(yamlContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	loader := NewYAMLLoader(tmpFile.Name())
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.MirrorOutputDir != "./mirror" {
		t.Errorf("expected output_dir './mirror', got '%s'", cfg.MirrorOutputDir)
	}
	if cfg.ConcurrentClones != 10 {
		t.Errorf("expected concurrent_clones 10, got %d", cfg.ConcurrentClones)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log_level 'debug', got '%s'", cfg.LogLevel)
	}
	if !cfg.LogCommands {
		t.Error("expected log_commands true")
	}
	if cfg.GitPath != "/usr/bin/git" {
		t.Errorf("expected git path '/usr/bin/git', got '%s'", cfg.GitPath)
	}
	if cfg.GitPathFormat != "windows" {
		t.Errorf("expected git path format 'windows', got '%s'", cfg.GitPathFormat)
	}
	home := os.Getenv("HOME")
	expectedKeyPath := home + "/.ssh/id_rsa"
	if cfg.SSHKeyPath != expectedKeyPath {
		t.Errorf("expected SSH key path '%s', got '%s'", expectedKeyPath, cfg.SSHKeyPath)
	}

	if cfg.CodeCommit == nil {
		t.Fatal("expected CodeCommit config")
	}
	if cfg.CodeCommit.DefaultRegion != "eu-west-1" {
		t.Errorf("expected default_region 'eu-west-1', got '%s'", cfg.CodeCommit.DefaultRegion)
	}
	if len(cfg.CodeCommit.Regions) != 2 {
		t.Errorf("expected 2 regions, got %d", len(cfg.CodeCommit.Regions))
	}
	if len(cfg.CodeCommit.Accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(cfg.CodeCommit.Accounts))
	}
	if cfg.CodeCommit.Accounts[0].ID != "111111111111" {
		t.Errorf("expected account ID '111111111111', got '%s'", cfg.CodeCommit.Accounts[0].ID)
	}
	if cfg.CodeCommit.Accounts[0].Profile != "prod" {
		t.Errorf("expected profile 'prod', got '%s'", cfg.CodeCommit.Accounts[0].Profile)
	}
	if !cfg.ForceReclone {
		t.Error("expected force_reclone true")
	}
	if cfg.CodeCommit.Accounts[1].Profile != "dev" {
		t.Errorf("expected profile 'dev', got '%s'", cfg.CodeCommit.Accounts[1].Profile)
	}

	if cfg.GitHub == nil {
		t.Fatal("expected GitHub config")
	}
	if cfg.GitHub.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", cfg.GitHub.Username)
	}
	if len(cfg.GitHub.Orgs) != 1 || cfg.GitHub.Orgs[0] != "my-org" {
		t.Errorf("expected orgs ['my-org'], got %v", cfg.GitHub.Orgs)
	}
	if cfg.GitHub.Token != "ghp_test123" {
		t.Errorf("expected token 'ghp_test123', got '%s'", cfg.GitHub.Token)
	}
	expectedGitHubKey := home + "/.ssh/id_github"
	if cfg.GitHub.SSHKeyPath != expectedGitHubKey {
		t.Errorf("expected GitHub SSH key path '%s', got '%s'", expectedGitHubKey, cfg.GitHub.SSHKeyPath)
	}
}

func TestYAMLLoaderMinimal(t *testing.T) {
	yamlContent := []byte(`
output_dir: ./mirror
codecommit:
  accounts:
    - id: "111111111111"
`)

	tmpFile, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpFile.Write(yamlContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	loader := NewYAMLLoader(tmpFile.Name())
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.ConcurrentClones != 5 {
		t.Errorf("expected default concurrent_clones 5, got %d", cfg.ConcurrentClones)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default log_level 'info', got '%s'", cfg.LogLevel)
	}
	if cfg.GitPath != "git" {
		t.Errorf("expected default git path 'git', got '%s'", cfg.GitPath)
	}
	if cfg.CodeCommit.DefaultRegion != "us-east-1" {
		t.Errorf("expected default region 'us-east-1', got '%s'", cfg.CodeCommit.DefaultRegion)
	}
}

func TestYAMLLoaderMissingOutputDir(t *testing.T) {
	yamlContent := []byte(`
codecommit:
  accounts:
    - id: "111111111111"
`)

	tmpFile, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpFile.Write(yamlContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	loader := NewYAMLLoader(tmpFile.Name())
	_, err = loader.Load()
	if err == nil {
		t.Fatal("expected error when output_dir is missing")
	}
}

func TestYAMLLoaderNoProviders(t *testing.T) {
	yamlContent := []byte(`
output_dir: ./mirror
`)

	tmpFile, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpFile.Write(yamlContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	loader := NewYAMLLoader(tmpFile.Name())
	_, err = loader.Load()
	if err == nil {
		t.Fatal("expected error when no providers configured")
	}
}

func TestYAMLLoaderGitHubOnly(t *testing.T) {
	yamlContent := []byte(`
output_dir: ./mirror
github:
  username: testuser
`)

	tmpFile, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpFile.Write(yamlContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	loader := NewYAMLLoader(tmpFile.Name())
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.GitHub == nil {
		t.Fatal("expected GitHub config")
	}
	if cfg.GitHub.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", cfg.GitHub.Username)
	}
	if cfg.CodeCommit != nil {
		t.Error("expected CodeCommit config to be nil")
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
