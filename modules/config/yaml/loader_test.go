package yaml

import (
	"os"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestYAMLLoaderFull(t *testing.T) {
	path := writeTemp(t, `
output_dir: ./mirror
concurrent_clones: 10
log_level: debug
log_commands: true
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
      ssh_key_id: AAAAAAA
      ssh_key_file: ~/.ssh/id_rsa
    - id: "222222222222"
      ssh_key_id: BBBBBBB
github:
  username: testuser
  orgs:
    - my-org
  token: ghp_test123
  ssh_key_path: ~/.ssh/id_github
`)
	loader := NewYAMLLoader(path)
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
	if cfg.GitPath != "/usr/bin/git" {
		t.Errorf("expected git path '/usr/bin/git', got '%s'", cfg.GitPath)
	}
	if cfg.CodeCommit == nil || len(cfg.CodeCommit.Accounts) != 2 {
		t.Fatalf("expected 2 codecommit accounts, got %+v", cfg.CodeCommit)
	}
	if cfg.GitHub == nil || cfg.GitHub.Username != "testuser" {
		t.Fatalf("expected github user testuser, got %+v", cfg.GitHub)
	}
}

func TestYAMLLoaderMinimal(t *testing.T) {
	path := writeTemp(t, `
output_dir: ./mirror
codecommit:
  accounts:
    - id: "111111111111"
`)
	loader := NewYAMLLoader(path)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.ConcurrentClones != 5 {
		t.Errorf("expected default concurrent_clones 5, got %d", cfg.ConcurrentClones)
	}
	if cfg.CodeCommit.DefaultRegion != "us-east-1" {
		t.Errorf("expected default region 'us-east-1', got '%s'", cfg.CodeCommit.DefaultRegion)
	}
}

func TestYAMLLoaderMissingOutputDir(t *testing.T) {
	path := writeTemp(t, `
codecommit:
  accounts:
    - id: "111111111111"
`)
	loader := NewYAMLLoader(path)
	if _, err := loader.Load(); err == nil {
		t.Fatal("expected error when output_dir is missing")
	}
}

func TestYAMLLoaderNoProviders(t *testing.T) {
	path := writeTemp(t, `
output_dir: ./mirror
`)
	loader := NewYAMLLoader(path)
	if _, err := loader.Load(); err == nil {
		t.Fatal("expected error when no providers configured")
	}
}

func TestYAMLLoaderGitHubOnly(t *testing.T) {
	path := writeTemp(t, `
output_dir: ./mirror
github:
  username: testuser
`)
	loader := NewYAMLLoader(path)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.GitHub == nil || cfg.GitHub.Username != "testuser" {
		t.Fatal("expected GitHub config")
	}
	if cfg.CodeCommit != nil {
		t.Error("expected CodeCommit config to be nil")
	}
}
