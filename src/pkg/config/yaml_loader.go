package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/romayengineer/git-mirror/pkg/types"
)

type yamlConfig struct {
	OutputDir        string            `yaml:"output_dir"`
	ConcurrentClones int               `yaml:"concurrent_clones"`
	LogLevel         string            `yaml:"log_level"`
	LogCommands      bool              `yaml:"log_commands"`
	ForceReclone     bool              `yaml:"force_reclone"`
	Git              *yamlGitConfig    `yaml:"git"`
	SSH              *yamlSSHConfig    `yaml:"ssh"`
	GitHub           *yamlGitHubConfig `yaml:"github"`
	CodeCommit       *yamlCodeCommit   `yaml:"codecommit"`
}

type yamlGitConfig struct {
	Path       string `yaml:"path"`
	PathFormat string `yaml:"path_format"`
}

type yamlSSHConfig struct {
	DefaultKeyPath string `yaml:"default_key_path"`
}

type yamlGitHubConfig struct {
	Username   string   `yaml:"username"`
	Orgs       []string `yaml:"orgs"`
	Token      string   `yaml:"token"`
	SSHKeyPath string   `yaml:"ssh_key_path"`
}

type yamlCodeCommit struct {
	DefaultRegion string                   `yaml:"default_region"`
	Regions       []string                 `yaml:"regions"`
	Accounts      []types.CodeCommitAccount `yaml:"accounts"`
}

type yamlLoader struct {
	filePath string
}

func NewYAMLLoader(filePath string) Loader {
	return &yamlLoader{filePath: filePath}
}

func (l *yamlLoader) Load() (*types.Config, error) {
	data, err := os.ReadFile(l.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", l.filePath, err)
	}

	var yc yamlConfig
	if err := yaml.Unmarshal(data, &yc); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	cfg := &types.Config{
		ConcurrentClones: 5,
		LogLevel:         "info",
		GitPath:          "git",
		GitPathFormat:    "linux",
	}

	if yc.OutputDir == "" {
		return nil, fmt.Errorf("output_dir is required")
	}
	cfg.MirrorOutputDir = yc.OutputDir

	if yc.ConcurrentClones > 0 {
		cfg.ConcurrentClones = yc.ConcurrentClones
	}
	if yc.LogLevel != "" {
		cfg.LogLevel = yc.LogLevel
	}
	cfg.LogCommands = yc.LogCommands
	cfg.ForceReclone = yc.ForceReclone

	if yc.Git != nil {
		if yc.Git.Path != "" {
			cfg.GitPath = yc.Git.Path
		}
		if yc.Git.PathFormat != "" {
			cfg.GitPathFormat = yc.Git.PathFormat
		}
	}

	if yc.SSH != nil && yc.SSH.DefaultKeyPath != "" {
		expandedPath, err := expandPath(yc.SSH.DefaultKeyPath)
		if err == nil {
			cfg.SSHKeyPath = expandedPath
		}
	}

	if yc.GitHub != nil && (yc.GitHub.Username != "" || len(yc.GitHub.Orgs) > 0) {
		ghCfg := &types.GitHubConfig{
			Username: yc.GitHub.Username,
			Orgs:     yc.GitHub.Orgs,
			Token:    yc.GitHub.Token,
		}
		if yc.GitHub.SSHKeyPath != "" {
			expandedPath, err := expandPath(yc.GitHub.SSHKeyPath)
			if err == nil {
				ghCfg.SSHKeyPath = expandedPath
			}
		}
		cfg.GitHub = ghCfg
	}

	if yc.CodeCommit != nil && len(yc.CodeCommit.Accounts) > 0 {
		accounts := make([]types.CodeCommitAccount, len(yc.CodeCommit.Accounts))
		for i, a := range yc.CodeCommit.Accounts {
			accounts[i] = a
		}

		ccCfg := &types.CodeCommitConfig{
			Accounts:           accounts,
			Regions:            yc.CodeCommit.Regions,
			DefaultRegion:      "us-east-1",
			CredentialsProfile: "default",
			LogCommands:        yc.LogCommands,
		}
		if yc.CodeCommit.DefaultRegion != "" {
			ccCfg.DefaultRegion = yc.CodeCommit.DefaultRegion
		}
		cfg.CodeCommit = ccCfg
	}

	if cfg.CodeCommit == nil && cfg.GitHub == nil {
		return nil, fmt.Errorf("at least one provider must be configured: add a codecommit or github section to config.yaml")
	}

	return cfg, nil
}
