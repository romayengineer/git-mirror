// Package core holds pure domain entities shared by all git-mirror modules.
// It has no third-party dependencies.
package core

// Repository is the normalized domain entity produced by every provider.
type Repository struct {
	Name          string
	RepositoryURL string
	// CloneURL is the transport-neutral URL used to clone and compare the
	// repository. It carries whatever the provider's configured transport
	// produced (e.g. an ssh:// or codecommit:: URL for CodeCommit).
	CloneURL     string
	HTTPCloneURL string
	// SSHCloneURL is the provider's native SSH clone URL. Kept for reference;
	// prefer CloneURL when deciding what to clone.
	SSHCloneURL      string
	ARN              string
	AccountID        string
	AccountAlias     string
	Region           string
	LocalPath        string
	SSHKeyPath       string
	RepositoryStatus string
	Provider         string // e.g. "codecommit", "github", "gitlab"
}

// CloneResult is the outcome of a single clone/update operation.
type CloneResult struct {
	Repository *Repository
	Success    bool
	Error      error
	Message    string
}

// DiscoveryResult groups discovery output per account/region.
type DiscoveryResult struct {
	AccountID    string
	AccountAlias string
	Region       string
	Repositories []*Repository
	Error        error
	TotalCount   int
}

// MirrorStats aggregates a full mirror run.
type MirrorStats struct {
	TotalRepositories   int
	SuccessfulClones    int
	FailedClones        int
	UpdatedRepositories int
	Duration            string
	Errors              []string
}

// AccountInfo identifies a cloud account.
type AccountInfo struct {
	ID    string
	Alias string
}

// RegionInfo identifies a cloud region.
type RegionInfo struct {
	Name string
}
