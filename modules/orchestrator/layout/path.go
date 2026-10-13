// Package layout builds local mirror paths. Pure, no I/O.
package layout

import "path/filepath"

// BuildLocalPath returns baseDir/provider/accountAlias/[region/]repoName.
// Region may be empty (e.g. GitHub); filepath.Join skips it cleanly.
func BuildLocalPath(baseDir, provider, accountAlias, region, repoName string) string {
	if region == "" {
		return filepath.Join(baseDir, provider, accountAlias, repoName)
	}
	return filepath.Join(baseDir, provider, accountAlias, region, repoName)
}
