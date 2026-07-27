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
	MirrorOutputDir  string
	ConcurrentClones int
	SSHKeyPath       string
	LogLevel         string
	GitPath          string
	LogCommands      bool
	GitPathFormat    string
	GitHub           *GitHubConfig
	CodeCommit       *CodeCommitConfig
}

type GitHubConfig struct {
	Username   string
	Orgs       []string
	Token      string
	SSHKeyPath string
}

type CodeCommitConfig struct {
	AccountIDs            []string
	Regions               []string
	DefaultRegion         string
	CredentialsProfile    string
	AccountProfiles       map[string]string
	LogCommands           bool
	SSHKeyIDs             map[string]string
	SSHKeyFiles           map[string]string
}

type AccountInfo struct {
	ID    string
	Alias string
}

type RegionInfo struct {
	Name string
}
