// Package path provides utilities for converting between WSL and Windows path formats.
// This is useful when running Windows Git (e.g., wgit) from WSL, which expects Windows paths
// instead of WSL paths (e.g., C:\path\to\repo instead of /mnt/c/path/to/repo).
package path

import (
	"strings"
)

// PathConverter converts file paths to the appropriate format.
type PathConverter interface {
	// Convert transforms a path according to the converter's format.
	Convert(path string) string
}

// linuxConverter returns paths unchanged (WSL format: /mnt/c/path)
type linuxConverter struct{}

func (c *linuxConverter) Convert(path string) string {
	return path
}

// windowsConverter converts WSL paths to Windows format (C:\path)
// Used when running Windows Git from WSL to ensure git receives paths it understands.
type windowsConverter struct{}

func (c *windowsConverter) Convert(wslPath string) string {
	if !strings.HasPrefix(wslPath, "/mnt/") {
		return wslPath
	}

	parts := strings.Split(wslPath, "/")
	if len(parts) < 3 {
		return wslPath
	}

	// Convert /mnt/c/path -> C:\path
	drive := strings.ToUpper(parts[2])
	remaining := strings.Join(parts[3:], "\\")
	return drive + ":\\" + remaining
}

// NewPathConverter creates a PathConverter based on the format string.
// format: "windows" for Windows Git from WSL, "linux" (default) for WSL Git.
func NewPathConverter(pathFormat string) PathConverter {
	if strings.EqualFold(pathFormat, "windows") {
		return &windowsConverter{}
	}
	return &linuxConverter{}
}
