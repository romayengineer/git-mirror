// Package configfiles locates config files on disk. No parsing here.
package configfiles

import (
	"os"
	"path/filepath"
)

// FindConfigFile returns the first existing config.yaml candidate, or "".
func FindConfigFile() string {
	homeDir, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(homeDir, ".config", "git-mirror", "config.yaml"),
		"config.yaml",
		filepath.Join("..", "config.yaml"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// FindEnvFile returns the first existing .env candidate, or ".env" as default.
func FindEnvFile() string {
	homeDir, _ := os.UserHomeDir()
	candidates := []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join(homeDir, ".config", "git-mirror", ".env"),
		filepath.Join(homeDir, ".git-mirror"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ".env"
}
