package codecommit

import (
	"testing"
)

func TestBuildCodeCommitURL(t *testing.T) {
	tests := []struct {
		name     string
		region   string
		profile  string
		repoName string
		expected string
	}{
		{
			name:     "with profile",
			region:   "us-east-1",
			profile:  "my-profile",
			repoName: "my-repo",
			expected: "codecommit::us-east-1://my-profile@my-repo",
		},
		{
			name:     "with profile different region",
			region:   "eu-west-1",
			profile:  "prod",
			repoName: "other-repo",
			expected: "codecommit::eu-west-1://prod@other-repo",
		},
		{
			name:     "without profile",
			region:   "us-east-1",
			profile:  "",
			repoName: "my-repo",
			expected: "codecommit::us-east-1://my-repo",
		},
		{
			name:     "without profile us-west-2",
			region:   "us-west-2",
			profile:  "",
			repoName: "test",
			expected: "codecommit::us-west-2://test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCodeCommitURL(tt.region, tt.profile, tt.repoName)
			if result != tt.expected {
				t.Errorf("buildCodeCommitURL(%q, %q, %q) = %q, want %q", tt.region, tt.profile, tt.repoName, result, tt.expected)
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
