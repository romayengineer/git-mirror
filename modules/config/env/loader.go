// Package env loads Config from environment variables / .env files.
package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/romayengineer/git-mirror-config"
	"github.com/romayengineer/git-mirror-config/pathutil"
)

const defaultRegion = "us-east-1"

type envLoader struct {
	envFilePath string
}

// NewEnvLoader creates a Loader reading from envFilePath (.env) plus process env.
func NewEnvLoader(envFilePath string) config.Loader {
	return &envLoader{envFilePath: envFilePath}
}

func (l *envLoader) Load() (*config.Config, error) {
	if err := godotenv.Load(l.envFilePath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	cfg := &config.Config{
		ConcurrentClones: 5,
		LogLevel:         "info",
		GitPath:          "git",
		GitPathFormat:    "linux",
	}

	if clones, err := strconv.Atoi(os.Getenv("CONCURRENT_CLONES")); err == nil && clones > 0 {
		cfg.ConcurrentClones = clones
	}

	if logCmds, err := strconv.ParseBool(os.Getenv("LOG_COMMANDS")); err == nil {
		cfg.LogCommands = logCmds
	}

	if v := os.Getenv("GIT_PATH"); v != "" {
		cfg.GitPath = v
	}
	if v := os.Getenv("GIT_PATH_FORMAT"); v != "" {
		cfg.GitPathFormat = v
	}

	mirrorDir := os.Getenv("MIRROR_OUTPUT_DIR")
	if mirrorDir == "" {
		return nil, fmt.Errorf("MIRROR_OUTPUT_DIR is required")
	}
	cfg.MirrorOutputDir = mirrorDir

	sshKeyPath := os.Getenv("SSH_KEY_PATH")
	if sshKeyPath != "" {
		expandedPath, err := pathutil.ExpandPath(sshKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to expand SSH_KEY_PATH: %w", err)
		}
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("SSH key file does not exist: %s", expandedPath)
		}
		cfg.SSHKeyPath = expandedPath
	}

	cfg.GitHub = parseGitHubConfig()

	ccCfg, err := parseCodeCommitConfig()
	if err != nil {
		return nil, err
	}
	cfg.CodeCommit = ccCfg

	if cfg.CodeCommit == nil && cfg.GitHub == nil {
		return nil, fmt.Errorf("at least one provider must be configured: set AWS_ACCOUNT_IDS, GITHUB_USERNAME, or GITHUB_ORGS")
	}

	return cfg, nil
}

func parseGitHubConfig() *config.GitHubConfig {
	username := os.Getenv("GITHUB_USERNAME")
	orgsStr := os.Getenv("GITHUB_ORGS")
	if username == "" && orgsStr == "" {
		return nil
	}
	cfg := &config.GitHubConfig{
		Username: username,
		Token:    os.Getenv("GITHUB_TOKEN"),
	}
	if orgsStr != "" {
		cfg.Orgs = strings.Split(orgsStr, ",")
		for i, org := range cfg.Orgs {
			cfg.Orgs[i] = strings.TrimSpace(org)
		}
	}
	if v := os.Getenv("GITHUB_SSH_KEY_PATH"); v != "" {
		expandedPath, err := pathutil.ExpandPath(v)
		if err == nil {
			cfg.SSHKeyPath = expandedPath
		}
	}
	return cfg
}

func parseCodeCommitConfig() (*config.CodeCommitConfig, error) {
	accountIDs := os.Getenv("AWS_ACCOUNT_IDS")
	if accountIDs == "" {
		return nil, nil
	}
	cfg := &config.CodeCommitConfig{
		CredentialsProfile: "default",
		DefaultRegion:      defaultRegion,
	}
	ids := strings.Split(accountIDs, ",")
	profiles := pathutil.ParseKeyValuePairs(os.Getenv("AWS_ACCOUNT_PROFILES"))
	sshKeyIDs := pathutil.ParseKeyValuePairs(os.Getenv("CODECOMMIT_SSH_KEY_IDS"))
	sshKeyFiles := pathutil.ParseKeyValuePairs(os.Getenv("CODECOMMIT_SSH_KEY_FILES"))
	transports := pathutil.ParseKeyValuePairs(os.Getenv("CODECOMMIT_TRANSPORTS"))

	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		account := config.CodeCommitAccount{ID: id}
		if p, ok := profiles[id]; ok {
			account.Profile = p
		}
		if k, ok := sshKeyIDs[id]; ok {
			account.SSHKeyID = k
		}
		if t, ok := transports[id]; ok {
			account.Transport = t
		}
		f, fileSet := sshKeyFiles[id]
		if fileSet {
			expandedPath, err := pathutil.ExpandPath(f)
			if err == nil {
				account.SSHKeyFile = expandedPath
			} else {
				account.SSHKeyFile = f
			}
		}
		cfg.Accounts = append(cfg.Accounts, account)
	}

	if v := os.Getenv("AWS_REGIONS"); v != "" {
		cfg.Regions = strings.Split(v, ",")
		for i, r := range cfg.Regions {
			cfg.Regions[i] = strings.TrimSpace(r)
		}
	}
	if v := os.Getenv("CODECOMMIT_TRANSPORT"); v != "" {
		cfg.Transport = v
	}
	if v := os.Getenv("AWS_DEFAULT_REGION"); v != "" {
		cfg.DefaultRegion = v
	}
	if v := os.Getenv("AWS_CREDENTIALS_PROFILE"); v != "" {
		cfg.CredentialsProfile = v
	}
	if logCmds, err := strconv.ParseBool(os.Getenv("LOG_COMMANDS")); err == nil {
		cfg.LogCommands = logCmds
	}

	return cfg, nil
}
