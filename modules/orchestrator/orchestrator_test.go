package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/romayengineer/git-mirror-config"
	core "github.com/romayengineer/git-mirror-core"
	"github.com/romayengineer/git-mirror-orchestrator/ports"
)

type fakeProvider struct {
	name  string
	repos []*core.Repository
	err   error
}

func (f *fakeProvider) Name() string { return f.name }
func (f *fakeProvider) DiscoverRepositories(context.Context) ([]*core.Repository, error) {
	return f.repos, f.err
}

type fakeReporter struct {
	mu      sync.Mutex
	reports []string
}

func (r *fakeReporter) Done(repo *core.Repository, success bool, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reports = append(r.reports, fmt.Sprintf("%s:%v:%s", repo.Name, success, message))
}

type fakeRenderer struct {
	reporter *fakeReporter
	total    int
}

func (f *fakeRenderer) Start(_ string, total int) (ports.CompletionReporter, func()) {
	f.total = total
	return f.reporter, func() {}
}

type fakeCloner struct {
	calls int
	fail  map[string]bool
}

func (f *fakeCloner) CloneRepository(r *core.Repository) *core.CloneResult {
	f.calls++
	if f.fail[r.Name] {
		return &core.CloneResult{Repository: r, Success: false, Message: "boom", Error: fmt.Errorf("boom")}
	}
	return &core.CloneResult{Repository: r, Success: true, Message: "cloned"}
}

func TestMirrorAggregatesStats(t *testing.T) {
	provider := &fakeProvider{name: "github", repos: []*core.Repository{
		{Name: "a", Provider: "github", AccountAlias: "org"},
		{Name: "b", Provider: "github", AccountAlias: "org"},
		{Name: "c", Provider: "github", AccountAlias: "org"},
	}}
	cloner := &fakeCloner{fail: map[string]bool{"b": true}}
	o := NewOrchestrator(cloner, "", []core.Provider{provider}, ports.DiscardRenderer{})
	stats, err := o.Mirror(context.Background(), &config.Config{MirrorOutputDir: t.TempDir(), ConcurrentClones: 2})
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalRepositories != 3 || stats.SuccessfulClones != 2 || stats.FailedClones != 1 {
		t.Errorf("unexpected stats: %+v", stats)
	}
	if cloner.calls != 3 {
		t.Errorf("expected 3 clone calls, got %d", cloner.calls)
	}
}

func TestMirrorReportsEachCompletion(t *testing.T) {
	provider := &fakeProvider{name: "github", repos: []*core.Repository{
		{Name: "a", Provider: "github", AccountAlias: "org"},
		{Name: "b", Provider: "github", AccountAlias: "org"},
	}}
	reporter := &fakeReporter{}
	renderer := &fakeRenderer{reporter: reporter}
	cloner := &fakeCloner{fail: map[string]bool{"b": true}}
	o := NewOrchestrator(cloner, "", []core.Provider{provider}, renderer)

	if _, err := o.Mirror(context.Background(), &config.Config{MirrorOutputDir: t.TempDir(), ConcurrentClones: 2}); err != nil {
		t.Fatal(err)
	}
	if renderer.total != 2 {
		t.Errorf("expected total 2, got %d", renderer.total)
	}
	if len(reporter.reports) != 2 {
		t.Fatalf("expected 2 completion reports, got %d", len(reporter.reports))
	}
	got := map[string]bool{}
	for _, r := range reporter.reports {
		got[r] = true
	}
	if !got["a:true:cloned"] || !got["b:false:boom"] {
		t.Errorf("unexpected reports: %v", reporter.reports)
	}
}

func TestMirrorNoProviders(t *testing.T) {
	o := NewOrchestrator(&fakeCloner{}, "", nil, ports.DiscardRenderer{})
	stats, err := o.Mirror(context.Background(), &config.Config{MirrorOutputDir: t.TempDir(), ConcurrentClones: 2})
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalRepositories != 0 {
		t.Errorf("expected 0 repos, got %+v", stats)
	}
}
