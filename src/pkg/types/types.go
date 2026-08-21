package types

type Repository struct {
	Name             string
	RepositoryURL    string
	HTTPCloneURL     string
	SSHCloneURL      string
	CloneURL         string // Canonical clone URL (codecommit:: for CodeCommit, SSH for GitHub)
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
	ForceReclone     bool
	GitHub           *GitHubConfig
	CodeCommit       *CodeCommitConfig
}

type GitHubConfig struct {
	Username   string   `yaml:"username"`
	Orgs       []string `yaml:"orgs"`
	Token      string   `yaml:"token"`
	SSHKeyPath string   `yaml:"ssh_key_path"`
}

type CodeCommitAccount struct {
	ID      string `yaml:"id"`
	Profile string `yaml:"profile"`
}

type CodeCommitConfig struct {
	Accounts           []CodeCommitAccount
	Regions            []string
	DefaultRegion      string
	CredentialsProfile string
	LogCommands        bool
}

type AccountInfo struct {
	ID    string
	Alias string
}

type RegionInfo struct {
	Name string
}
