package codecommit

import (
	"testing"

	"github.com/romayengineer/git-mirror-config"
)

func TestInjectSSHUser(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		user     string
		expected string
	}{
		{
			name:     "injects user into ssh URL",
			url:      "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/my-repo",
			user:     "APKA4YBIPDVQB6DUKUSZ",
			expected: "ssh://APKA4YBIPDVQB6DUKUSZ@git-codecommit.us-east-1.amazonaws.com/v1/repos/my-repo",
		},
		{
			name:     "empty user returns URL unchanged",
			url:      "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/my-repo",
			user:     "",
			expected: "ssh://git-codecommit.us-east-1.amazonaws.com/v1/repos/my-repo",
		},
		{
			name:     "non-ssh URL returns unchanged",
			url:      "git@github.com:org/repo.git",
			user:     "someuser",
			expected: "git@github.com:org/repo.git",
		},
		{
			name:     "empty URL with non-empty user",
			url:      "",
			user:     "someuser",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := injectSSHUser(tt.url, tt.user)
			if result != tt.expected {
				t.Errorf("injectSSHUser(%q, %q) = %q, want %q", tt.url, tt.user, result, tt.expected)
			}
		})
	}
}

func TestProviderName(t *testing.T) {
	provider := NewCodeCommitProvider(nil)
	if provider.Name() != "codecommit" {
		t.Errorf("expected name 'codecommit', got %s", provider.Name())
	}
}

func TestBuildGRCURL(t *testing.T) {
	tests := []struct {
		name     string
		region   string
		profile  string
		repo     string
		expected string
	}{
		{
			name:     "region and profile",
			region:   "us-east-1",
			profile:  "valtira-shared",
			repo:     "Valtira-Simplicis-Dev",
			expected: "codecommit::us-east-1://valtira-shared@Valtira-Simplicis-Dev",
		},
		{
			name:     "ca-central-1",
			region:   "ca-central-1",
			profile:  "valtira-sparq",
			repo:     "sparqpoint-infra",
			expected: "codecommit::ca-central-1://valtira-sparq@sparqpoint-infra",
		},
		{
			name:     "empty profile falls back to default",
			region:   "us-east-1",
			profile:  "",
			repo:     "my-repo",
			expected: "codecommit::us-east-1://default@my-repo",
		},
		{
			name:     "empty region omits region segment",
			region:   "",
			profile:  "burningtorch",
			repo:     "rust_miner",
			expected: "codecommit://burningtorch@rust_miner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildGRCURL(tt.region, tt.profile, tt.repo); got != tt.expected {
				t.Errorf("buildGRCURL(%q, %q, %q) = %q, want %q", tt.region, tt.profile, tt.repo, got, tt.expected)
			}
		})
	}
}

func TestResolveTransport(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.CodeCommitConfig
		account config.CodeCommitAccount
		want    string
	}{
		{
			name:    "defaults to ssh when nothing set",
			cfg:     &config.CodeCommitConfig{},
			account: config.CodeCommitAccount{ID: "1"},
			want:    TransportSSH,
		},
		{
			name:    "nil cfg defaults to ssh",
			cfg:     nil,
			account: config.CodeCommitAccount{ID: "1"},
			want:    TransportSSH,
		},
		{
			name:    "codecommit-level grc",
			cfg:     &config.CodeCommitConfig{Transport: "grc"},
			account: config.CodeCommitAccount{ID: "1"},
			want:    TransportGRC,
		},
		{
			name:    "account override wins over codecommit level",
			cfg:     &config.CodeCommitConfig{Transport: "grc"},
			account: config.CodeCommitAccount{ID: "1", Transport: "ssh"},
			want:    TransportSSH,
		},
		{
			name:    "transport is case and whitespace insensitive",
			cfg:     nil,
			account: config.CodeCommitAccount{ID: "1", Transport: "  GRC  "},
			want:    TransportGRC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveTransport(tt.cfg, tt.account); got != tt.want {
				t.Errorf("resolveTransport() = %q, want %q", got, tt.want)
			}
		})
	}
}
