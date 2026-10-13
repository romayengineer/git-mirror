module github.com/romayengineer/git-mirror-gitops

go 1.25.0

require (
	github.com/romayengineer/git-mirror-core v0.0.0
	github.com/romayengineer/git-mirror-exec v0.0.0
	github.com/romayengineer/git-mirror-path v0.0.0
)

// Local monorepo development. On publish, these replaces are dropped and
// the versioned requires above resolve from the proxy.
replace (
	github.com/romayengineer/git-mirror-core => ../core
	github.com/romayengineer/git-mirror-exec => ../exec
	github.com/romayengineer/git-mirror-path => ../path
)
