package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/romayengineer/git-mirror/pkg/types"
)

type Loader interface {
	Load() (*types.Config, error)
}

const defaultRegion = "us-east-1"

type envLoader struct {
	envFilePath string
}

func NewEnvLoader(envFilePath string) Loader {
	return &envLoader{envFilePath: envFilePath}
}

func (l *envLoader) Load() (*types.Config, error) {
	if err := godotenv.Load(l.envFilePath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	cfg := &types.Config{
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
		expandedPath, err := expandPath(sshKeyPath)
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

func parseGitHubConfig() *types.GitHubConfig {
	username := os.Getenv("GITHUB_USERNAME")
	orgsStr := os.Getenv("GITHUB_ORGS")
	if username == "" && orgsStr == "" {
		return nil
	}
	cfg := &types.GitHubConfig{
		Username: username,
		Token:    os.Getenv("GITHUB_TOKEN"),
	}
	if orgsStr != "" {
		cfg.Orgs = strings.Split(orgsStr, ",")
		for i, org := range cfg.Orgs {
			cfg.Orgs[i] = strings.TrimSpace(org)
		}
	}
	return cfg
}

func parseCodeCommitConfig() (*types.CodeCommitConfig, error) {
	accountIDs := os.Getenv("AWS_ACCOUNT_IDS")
	if accountIDs == "" {
		return nil, nil
	}
	cfg := &types.CodeCommitConfig{
		CredentialsProfile: "default",
		DefaultRegion:      defaultRegion,
	}
	cfg.AccountIDs = strings.Split(accountIDs, ",")
	for i, id := range cfg.AccountIDs {
		cfg.AccountIDs[i] = strings.TrimSpace(id)
	}
	if v := os.Getenv("AWS_REGIONS"); v != "" {
		cfg.Regions = strings.Split(v, ",")
		for i, r := range cfg.Regions {
			cfg.Regions[i] = strings.TrimSpace(r)
		}
	}
	if v := os.Getenv("AWS_DEFAULT_REGION"); v != "" {
		cfg.DefaultRegion = v
	}
	if v := os.Getenv("AWS_CREDENTIALS_PROFILE"); v != "" {
		cfg.CredentialsProfile = v
	}
	cfg.AccountProfiles = parseAccountProfiles(os.Getenv("AWS_ACCOUNT_PROFILES"))
	if logCmds, err := strconv.ParseBool(os.Getenv("LOG_COMMANDS")); err == nil {
		cfg.LogCommands = logCmds
	}

	return cfg, nil
}

func parseAccountProfiles(profilesStr string) map[string]string {
	profiles := make(map[string]string)
	if profilesStr == "" {
		return profiles
	}

	pairs := strings.Split(profilesStr, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) == 2 {
			accountID := strings.TrimSpace(parts[0])
			profileName := strings.TrimSpace(parts[1])
			profiles[accountID] = profileName
		}
	}

	return profiles
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, path[2:]), nil
	}
	return path, nil
}
