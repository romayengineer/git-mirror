package codecommit

import (
	"testing"
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
			name:     "injects user with region in URL",
			url:      "ssh://git-codecommit.eu-west-1.amazonaws.com/v1/repos/other-repo",
			user:     "APKAX4ENFON4GCH7RIJY",
			expected: "ssh://APKAX4ENFON4GCH7RIJY@git-codecommit.eu-west-1.amazonaws.com/v1/repos/other-repo",
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
			name:     "https URL returns unchanged",
			url:      "https://github.com/org/repo.git",
			user:     "someuser",
			expected: "https://github.com/org/repo.git",
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
