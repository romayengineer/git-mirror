// Package pathutil holds path helpers shared by config loaders.
package pathutil

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandPath expands a leading ~/ to the user home directory.
func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, path[2:]), nil
	}
	return path, nil
}

// ParseKeyValuePairs parses "k1:v1,k2:v2" into a map, skipping malformed pairs.
func ParseKeyValuePairs(input string) map[string]string {
	pairs := make(map[string]string)
	if input == "" {
		return pairs
	}
	items := strings.Split(input, ",")
	for _, item := range items {
		parts := strings.Split(strings.TrimSpace(item), ":")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			pairs[key] = value
		}
	}
	return pairs
}
