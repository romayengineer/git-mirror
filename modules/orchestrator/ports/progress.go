// Package ports owns the presentation port. The orchestrator depends on
// ProgressRenderer, and cmd/git-mirror/presenter provides the mpb implementation.
// This keeps mpb out of the orchestrator module.
package ports

import core "github.com/romayengineer/git-mirror-core"

// RepoProgress is a per-repository progress handle.
type RepoProgress interface {
	// Done marks one repository finished with message (e.g. "✓ cloned").
	Done(success bool, message string)
}

// ProgressRenderer creates progress UI for a discovery batch.
type ProgressRenderer interface {
	// Start creates per-repo handles plus a total counter. Must be safe for 0 repos.
	Start(repos []*core.Repository) ([]RepoProgress, func())
}

// DiscardRenderer is a no-op ProgressRenderer for tests and --no-progress mode.
type DiscardRenderer struct{}

type discardHandle struct{}

func (discardHandle) Done(bool, string) {}

func (DiscardRenderer) Start(repos []*core.Repository) ([]RepoProgress, func()) {
	handles := make([]RepoProgress, len(repos))
	for i := range handles {
		handles[i] = discardHandle{}
	}
	return handles, func() {}
}
