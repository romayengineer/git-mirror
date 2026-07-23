package aws

import (
	"context"
	"fmt"
	"testing"
)

type mockAccountClient struct {
	aliasToReturn        string
	aliasErr             error
	regionsToReturn      []string
	regionsErr           error
	repositoriesToReturn []*Repository
	reposErr             error
}

func (m *mockAccountClient) GetAccountAlias(ctx context.Context, accountID string) (string, error) {
	if m.aliasErr != nil {
		return "", m.aliasErr
	}
	return m.aliasToReturn, nil
}

func (m *mockAccountClient) ListRegions(ctx context.Context) ([]string, error) {
	if m.regionsErr != nil {
		return nil, m.regionsErr
	}
	return m.regionsToReturn, nil
}

func (m *mockAccountClient) ListRepositories(ctx context.Context, region string) ([]*Repository, error) {
	if m.reposErr != nil {
		return nil, m.reposErr
	}
	return m.repositoriesToReturn, nil
}

func TestDiscovererWithSpecificRegions(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockAccountClient{
		aliasToReturn: "test-account",
		repositoriesToReturn: []*Repository{
			{
				Name:        "repo1",
				SSHCloneURL: "git@codecommit:repo1.git",
			},
			{
				Name:        "repo2",
				SSHCloneURL: "git@codecommit:repo2.git",
			},
		},
	}

	discoverer := NewRepositoryDiscoverer(mockClient)
	repos, err := discoverer.DiscoverRepositories(ctx, "123456789012", []string{"us-east-1"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(repos) != 2 {
		t.Errorf("expected 2 repositories, got %d", len(repos))
	}

	if repos[0].AccountAlias != "test-account" {
		t.Errorf("expected account alias to be 'test-account', got %s", repos[0].AccountAlias)
	}

	if repos[0].Region != "us-east-1" {
		t.Errorf("expected region to be 'us-east-1', got %s", repos[0].Region)
	}
}

func TestDiscovererAutoDiscoverRegions(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockAccountClient{
		aliasToReturn:   "test-account",
		regionsToReturn: []string{"us-east-1", "eu-west-1"},
		repositoriesToReturn: []*Repository{
			{
				Name:        "repo1",
				SSHCloneURL: "git@codecommit:repo1.git",
			},
		},
	}

	discoverer := NewRepositoryDiscoverer(mockClient)
	repos, err := discoverer.DiscoverRepositories(ctx, "123456789012", []string{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(repos) != 2 {
		t.Errorf("expected 2 repositories (1 per region), got %d", len(repos))
	}
}

func TestDiscovererAliasError(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockAccountClient{
		aliasErr: fmt.Errorf("failed to get alias"),
		repositoriesToReturn: []*Repository{
			{
				Name:        "repo1",
				SSHCloneURL: "git@codecommit:repo1.git",
			},
		},
	}

	discoverer := NewRepositoryDiscoverer(mockClient)
	repos, err := discoverer.DiscoverRepositories(ctx, "123456789012", []string{"us-east-1"})

	if err != nil {
		t.Fatalf("expected no error (alias errors are handled gracefully), got %v", err)
	}

	if len(repos) == 0 {
		t.Fatal("expected repositories to be discovered despite alias error")
	}

	if repos[0].AccountAlias != "123456789012" {
		t.Errorf("expected account alias to fall back to account ID '123456789012', got %s", repos[0].AccountAlias)
	}
}

func TestDiscovererRegionListError(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockAccountClient{
		aliasToReturn: "test-account",
		regionsErr:    fmt.Errorf("failed to list regions"),
	}

	discoverer := NewRepositoryDiscoverer(mockClient)
	_, err := discoverer.DiscoverRepositories(ctx, "123456789012", []string{})

	if err == nil {
		t.Fatal("expected error when listing regions fails")
	}
}
