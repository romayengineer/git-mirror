// Package presenter provides the console implementation of orchestrator's
// ports.ProgressRenderer. Repositories are printed one after the other, only
// once their clone/update has completed.
package presenter

import (
	"fmt"
	"sync"

	core "github.com/romayengineer/git-mirror-core"
	"github.com/romayengineer/git-mirror-orchestrator/ports"
)

// ConsoleRenderer prints one line per repository as it finishes.
type ConsoleRenderer struct{}

type consoleReporter struct {
	mu sync.Mutex
}

func (r *consoleReporter) Done(repo *core.Repository, success bool, message string) {
	symbol := "✓"
	if !success {
		symbol = "✗"
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	fmt.Printf("%s %s: %s\n", symbol, displayName(repo), message)
}

func displayName(repo *core.Repository) string {
	if repo == nil {
		return ""
	}
	if repo.Region != "" {
		return fmt.Sprintf("%s/%s/%s", repo.AccountAlias, repo.Region, repo.Name)
	}
	return fmt.Sprintf("%s/%s", repo.AccountAlias, repo.Name)
}

// Start returns a reporter that serializes completion output.
func (ConsoleRenderer) Start(string, int) (ports.CompletionReporter, func()) {
	return &consoleReporter{}, func() {}
}
