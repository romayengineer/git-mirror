module github.com/romayengineer/git-mirror-orchestrator

go 1.25.0

require (
	github.com/romayengineer/git-mirror-config v0.0.0
	github.com/romayengineer/git-mirror-core v0.0.0
	github.com/romayengineer/git-mirror-gitops v0.0.0
)

require (
	github.com/romayengineer/git-mirror-exec v0.0.0 // indirect
	github.com/romayengineer/git-mirror-path v0.0.0 // indirect
)

// Local monorepo development. On publish, drop these replaces.
// exec/path are transitive via gitops but needed for `go mod tidy` graph.
replace (
	github.com/romayengineer/git-mirror-config => ../config
	github.com/romayengineer/git-mirror-core => ../core
	github.com/romayengineer/git-mirror-exec => ../exec
	github.com/romayengineer/git-mirror-gitops => ../gitops
	github.com/romayengineer/git-mirror-path => ../path
)
