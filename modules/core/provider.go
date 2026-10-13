package core

import "context"

// Provider is the port every git provider must implement.
// It lives in core so orchestrator depends on the abstraction,
// and infrastructure adapters (github, codecommit, ...) implement it.
type Provider interface {
	Name() string
	DiscoverRepositories(ctx context.Context) ([]*Repository, error)
}
