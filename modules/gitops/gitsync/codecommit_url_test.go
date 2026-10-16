package gitsync

import (
	"os"
	"path/filepath"
	"testing"

	core "github.com/romayengineer/git-mirror-core"
)

func TestCodecommitIdentity(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantRegion string
		wantName   string
		wantOK     bool
	}{
		{
			name:       "grc with region and profile",
			url:        "codecommit::us-east-1://valtira-shared@Valtira-Simplicis-Dev",
			wantRegion: "us-east-1",
			wantName:   "Valtira-Simplicis-Dev",
			wantOK:     true,
		},
		{
			name:       "grc with region, no profile",
			url:        "codecommit::ca-central-1://sparqstream-hci",
			wantRegion: "ca-central-1",
			wantName:   "sparqstream-hci",
			wantOK:     true,
		},
		{
			name:       "grc without region",
			url:        "codecommit://valtira-bcbsma@bcbsma-systems",
			wantRegion: "",
			wantName:   "bcbsma-systems",
			wantOK:     true,
		},
		{
			name:       "grc strips .git suffix",
			url:        "codecommit::us-east-1://valtira-shared@Valtira-Simplicis-Dev.git",
			wantRegion: "us-east-1",
			wantName:   "Valtira-Simplicis-Dev",
			wantOK:     true,
		},
		{
			name:       "ssh url",
			url:        "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/Valtira-Simplicis-Dev",
			wantRegion: "us-east-1",
			wantName:   "Valtira-Simplicis-Dev",
			wantOK:     true,
		},
		{
			name:       "ssh url with injected key id",
			url:        "ssh://APKA4YBIPDVQB6DUKUSZ@git-codecommit.us-east-1.amazonaws.com/v1/repos/Valtira-Simplicis-Dev",
			wantRegion: "us-east-1",
			wantName:   "Valtira-Simplicis-Dev",
			wantOK:     true,
		},
		{
			name:       "region is lowercased",
			url:        "ssh://git-codecommit.US-EAST-1.amazonaws.com/v1/repos/Repo",
			wantRegion: "us-east-1",
			wantName:   "Repo",
			wantOK:     true,
		},
		{
			name:       "https url",
			url:        "https://git-codecommit.eu-west-1.amazonaws.com/v1/repos/my-repo",
			wantRegion: "eu-west-1",
			wantName:   "my-repo",
			wantOK:     true,
		},
		{
			name:   "non codecommit url",
			url:    "git@github.com:testuser/repo2.git",
			wantOK: false,
		},
		{
			name:   "plain path",
			url:    "/some/local/path",
			wantOK: false,
		},
		{
			name:   "malformed grc",
			url:    "codecommit:us-east-1/Repo",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region, name, ok := codecommitIdentity(tt.url)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if region != tt.wantRegion {
				t.Errorf("region = %q, want %q", region, tt.wantRegion)
			}
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
		})
	}
}

func TestURLsMatchCodeCommitTransports(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "grc vs ssh us-east-1",
			a:    "codecommit::us-east-1://valtira-shared@Valtira-Simplicis-Dev",
			b:    "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/Valtira-Simplicis-Dev",
			want: true,
		},
		{
			name: "grc vs ssh ca-central-1",
			a:    "codecommit::ca-central-1://valtira-sparq@sparq-festival-web",
			b:    "ssh://git-codecommit.ca-central-1.amazonaws.com/v1/repos/sparq-festival-web",
			want: true,
		},
		{
			name: "ssh with key id vs plain ssh",
			a:    "ssh://APKA4YBIPDVQB6DUKUSZ@git-codecommit.us-east-1.amazonaws.com/v1/repos/my-repo",
			b:    "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/my-repo",
			want: true,
		},
		{
			name: "different repo name",
			a:    "codecommit::us-east-1://valtira-shared@RepoA",
			b:    "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/RepoB",
			want: false,
		},
		{
			name: "different region",
			a:    "codecommit::us-east-1://valtira-shared@RepoA",
			b:    "ssh://git-codecommit.ca-central-1.amazonaws.com/v1/repos/RepoA",
			want: false,
		},
		{
			name: "grc without region matches any region",
			a:    "codecommit://valtira-shared@RepoA",
			b:    "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/RepoA",
			want: true,
		},
		{
			name: "genuinely foreign remote",
			a:    "https://github.com/other/repo.git",
			b:    "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/RepoA",
			want: false,
		},
		{
			name: "exact match still works",
			a:    "git@codecommit:test-repo.git",
			b:    "git@codecommit:test-repo.git",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := urlsMatch(tt.a, tt.b); got != tt.want {
				t.Errorf("urlsMatch(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestClonerExistingRepositoryCodeCommitGRCRemote(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "Valtira-Simplicis-Dev")
	if err := os.MkdirAll(filepath.Join(repoPath, ".git"), 0755); err != nil {
		t.Fatalf("failed to create repo dir: %v", err)
	}

	repo := &core.Repository{
		Name:        "Valtira-Simplicis-Dev",
		SSHCloneURL: "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/Valtira-Simplicis-Dev",
		LocalPath:   repoPath,
		SSHKeyPath:  "/path/to/key",
	}

	mockExec := &mockExecutor{
		urlToReturn: "codecommit::us-east-1://valtira-shared@Valtira-Simplicis-Dev",
	}
	cloner := NewCloner(mockExec)

	result := cloner.CloneRepository(repo)

	if !result.Success {
		t.Errorf("expected pull to succeed for equivalent GRC remote, got error: %v", result.Error)
	}
}
