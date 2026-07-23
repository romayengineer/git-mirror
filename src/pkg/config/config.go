package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/romayengineer/codecommit-mirror/pkg/types"
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

	cfg := &types.Config{}

	accountIDs := os.Getenv("AWS_ACCOUNT_IDS")
	if accountIDs == "" {
		return nil, fmt.Errorf("AWS_ACCOUNT_IDS is required")
	}
	cfg.AccountIDs = strings.Split(accountIDs, ",")
	for i, id := range cfg.AccountIDs {
		cfg.AccountIDs[i] = strings.TrimSpace(id)
	}

	sshKeyPath := os.Getenv("SSH_KEY_PATH")
	if sshKeyPath == "" {
		return nil, fmt.Errorf("SSH_KEY_PATH is required")
	}
	expandedPath, err := expandPath(sshKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to expand SSH_KEY_PATH: %w", err)
	}
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("SSH key file does not exist: %s", expandedPath)
	}
	cfg.SSHKeyPath = expandedPath

	mirrorDir := os.Getenv("MIRROR_OUTPUT_DIR")
	if mirrorDir == "" {
		return nil, fmt.Errorf("MIRROR_OUTPUT_DIR is required")
	}
	cfg.MirrorOutputDir = mirrorDir

	awsRegions := os.Getenv("AWS_REGIONS")
	if awsRegions != "" {
		cfg.AWSRegions = strings.Split(awsRegions, ",")
		for i, region := range cfg.AWSRegions {
			cfg.AWSRegions[i] = strings.TrimSpace(region)
		}
	}

	cfg.AWSCredentialsProfile = os.Getenv("AWS_CREDENTIALS_PROFILE")
	if cfg.AWSCredentialsProfile == "" {
		cfg.AWSCredentialsProfile = "default"
	}

	concurrentClonesStr := os.Getenv("CONCURRENT_CLONES")
	cfg.ConcurrentClones = 5
	if concurrentClonesStr != "" {
		if clones, err := strconv.Atoi(concurrentClonesStr); err == nil && clones > 0 {
			cfg.ConcurrentClones = clones
		}
	}

	cfg.LogLevel = os.Getenv("LOG_LEVEL")
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	cfg.AWSDefaultRegion = os.Getenv("AWS_DEFAULT_REGION")
	if cfg.AWSDefaultRegion == "" {
		cfg.AWSDefaultRegion = defaultRegion
	}

	cfg.AccountProfiles = parseAccountProfiles(os.Getenv("AWS_ACCOUNT_PROFILES"))

	cfg.GitPath = os.Getenv("GIT_PATH")
	if cfg.GitPath == "" {
		cfg.GitPath = "git"
	}

	logCommandsStr := os.Getenv("LOG_COMMANDS")
	cfg.LogCommands = false
	if logCommandsStr != "" {
		if logCmds, err := strconv.ParseBool(logCommandsStr); err == nil {
			cfg.LogCommands = logCmds
		}
	}

	// GitPathFormat controls path conversion for Windows Git from WSL.
	// Set to "windows" to use Windows paths (e.g., C:\path) with wgit or other Windows Git.
	cfg.GitPathFormat = os.Getenv("GIT_PATH_FORMAT")
	if cfg.GitPathFormat == "" {
		cfg.GitPathFormat = "linux"
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
