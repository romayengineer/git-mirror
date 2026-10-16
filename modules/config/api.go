// Package config owns application configuration DTOs and Loader implementations.
// Config structs were moved out of core per the layered SRP decision:
// core keeps domain entities, config owns input DTOs.
package config

// Config is the application configuration.
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

// GitHubConfig configures the GitHub provider.
type GitHubConfig struct {
	Username   string   `yaml:"username"`
	Orgs       []string `yaml:"orgs"`
	Token      string   `yaml:"token"`
	SSHKeyPath string   `yaml:"ssh_key_path"`
}

// CodeCommitAccount configures a single AWS account.
type CodeCommitAccount struct {
	ID         string `yaml:"id"`
	Profile    string `yaml:"profile"`
	SSHKeyID   string `yaml:"ssh_key_id"`
	SSHKeyFile string `yaml:"ssh_key_file"`
	// Transport overrides the codecommit-level transport for this account.
	// One of "ssh" (default) or "grc".
	Transport string `yaml:"transport"`
}

// CodeCommitConfig configures the CodeCommit provider.
type CodeCommitConfig struct {
	Accounts           []CodeCommitAccount
	Regions            []string
	DefaultRegion      string
	CredentialsProfile string
	LogCommands        bool
	// Transport selects how repositories are cloned: "ssh" (default) uses
	// ssh://git-codecommit.<region>.amazonaws.com URLs with an SSH key, "grc"
	// uses git-remote-codecommit (codecommit::<region>://<profile>@<name>) with
	// the account's AWS profile. Individual accounts may override this.
	Transport string
}
