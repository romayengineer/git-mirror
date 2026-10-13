module github.com/romayengineer/git-mirror-provider-github

go 1.25.0

require (
	github.com/romayengineer/git-mirror-config v0.0.0
	github.com/romayengineer/git-mirror-core v0.0.0
)

// Local monorepo development. On publish, drop these replaces.
replace (
	github.com/romayengineer/git-mirror-config => ../config
	github.com/romayengineer/git-mirror-core => ../core
)
