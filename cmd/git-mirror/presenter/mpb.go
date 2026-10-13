// Package presenter provides the mpb implementation of orchestrator's
// ports.ProgressRenderer. Only cmd links mpb; orchestrator stays UI-free.
package presenter

import (
	"fmt"
	"sync"

	core "github.com/romayengineer/git-mirror-core"
	"github.com/romayengineer/git-mirror-orchestrator/ports"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// MpbRenderer renders live per-repo progress bars.
type MpbRenderer struct{}

type mpbHandle struct {
	bar *mpb.Bar
	mu  *sync.Mutex
	res  *string
}

func (h *mpbHandle) Done(success bool, message string) {
	_ = success
	h.mu.Lock()
	*h.res = message
	h.mu.Unlock()
	h.bar.Increment()
}

type countingHandle struct {
	inner ports.RepoProgress
	total *mpb.Bar
}

func (h *countingHandle) Done(success bool, message string) {
	h.inner.Done(success, message)
	h.total.Increment()
}

// Start creates one bar per repo plus a total counter.
func (MpbRenderer) Start(repos []*core.Repository) ([]ports.RepoProgress, func()) {
	container := mpb.New()
	handles := make([]ports.RepoProgress, len(repos))

	for i, repo := range repos {
		var mu sync.Mutex
		var result string

		name := fmt.Sprintf("%s/%-30s", repo.AccountAlias, repo.Name)
		if repo.Region != "" {
			name = fmt.Sprintf("%s/%-30s", repo.AccountAlias+"/"+repo.Region, repo.Name)
		}

		bar := container.AddBar(1,
			mpb.BarFillerClearOnComplete(),
			mpb.PrependDecorators(
				decor.Name(name, decor.WC{W: len(name)}),
			),
			mpb.AppendDecorators(
				decor.OnComplete(
					decor.Spinner([]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}),
					""),
				decor.Any(func(s decor.Statistics) string {
					mu.Lock()
					r := result
					mu.Unlock()
					return r
				}),
			),
		)
		handles[i] = &mpbHandle{bar: bar, mu: &mu, res: &result}
	}

	totalBar := container.AddBar(int64(len(repos)),
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(decor.Name("  ")),
		mpb.AppendDecorators(decor.CountersNoUnit("%d / %d")),
	)

	wrapped := make([]ports.RepoProgress, len(handles))
	for i, h := range handles {
		wrapped[i] = &countingHandle{inner: h, total: totalBar}
	}

	return wrapped, func() { container.Wait() }
}
