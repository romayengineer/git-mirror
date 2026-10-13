// Package ports owns the presentation port. The orchestrator depends on
// ProgressRenderer, and cmd/git-mirror/presenter provides the implementation.
// This keeps UI concerns (e.g. mpb) out of the orchestrator module.
package ports

import core "github.com/romayengineer/git-mirror-core"

// CompletionReporter reports the outcome of a single repository as it
// finishes. Done is called once per repository, after its clone/update
// completes, so implementations can print results one after the other.
type CompletionReporter interface {
	// Done reports one finished repository with its result message.
	Done(repo *core.Repository, success bool, message string)
}

// ProgressRenderer creates a reporter for a discovery batch. Implementations
// must be safe for concurrent Done calls and for zero repositories.
type ProgressRenderer interface {
	// Start returns a reporter for total repositories plus a cleanup func.
	Start(provider string, total int) (CompletionReporter, func())
}

// DiscardRenderer is a no-op ProgressRenderer for tests and --no-progress mode.
type DiscardRenderer struct{}

type discardReporter struct{}

func (discardReporter) Done(*core.Repository, bool, string) {}

func (DiscardRenderer) Start(string, int) (CompletionReporter, func()) {
	return discardReporter{}, func() {}
}
