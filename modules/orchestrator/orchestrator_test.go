package orchestrator

import (
	"context"
	"fmt"
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
