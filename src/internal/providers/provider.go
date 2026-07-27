package providers

import (
	"context"

	"github.com/romayengineer/git-mirror/pkg/types"
)

type Provider interface {
	Name() string
	DiscoverRepositories(ctx context.Context) ([]*types.Repository, error)
}
