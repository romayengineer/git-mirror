package types

type Repository struct {
	Name             string
	RepositoryURL    string
	HTTPCloneURL     string
	SSHCloneURL      string
	ARN              string
	AccountID        string
	AccountAlias     string
	Region           string
	LocalPath        string
	SSHKeyPath       string
	RepositoryStatus string
	Provider         string // Name of the provider (e.g., "codecommit", "github", "gitlab")
}

type CloneResult struct {
	Repository *Repository
	Success    bool
	Error      error
	Message    string
}

type DiscoveryResult struct {
	AccountID    string
	AccountAlias string
	Region       string
	Repositories []*Repository
	Error        error
	TotalCount   int
}

type MirrorStats struct {
	TotalRepositories   int
	SuccessfulClones    int
	FailedClones        int
	UpdatedRepositories int
	Duration            string
	Errors              []string
}

type Config struct {
	AccountIDs            []string
	AWSRegions            []string
	AWSDefaultRegion      string
	SSHKeyPath            string
	MirrorOutputDir       string
	AWSCredentialsProfile string
	AccountProfiles       map[string]string // Maps account ID to AWS profile
	ConcurrentClones      int
	LogLevel              string
	GitPath               string
	LogCommands           bool
	// GitPathFormat controls path conversion for git commands.
	// "linux": Use WSL paths (e.g., /mnt/c/path) - works with WSL git
	// "windows": Convert to Windows paths (e.g., C:\path) - required for Windows Git from WSL
	GitPathFormat string
	// GitHubOrgs is a list of GitHub organizations to mirror.
	// If set, the GitHub provider is activated automatically.
	GitHubOrgs  []string
	GitHubToken string
}

type AccountInfo struct {
	ID    string
	Alias string
}

type RegionInfo struct {
	Name string
}
