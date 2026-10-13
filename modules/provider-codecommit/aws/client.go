// Package aws isolates AWS SDK auth from AWS CLI execution and discovery.
package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// AccountClient abstracts account-level AWS operations.
type AccountClient interface {
	GetAccountAlias(ctx context.Context, accountID string) (string, error)
	ListRegions(ctx context.Context) ([]string, error)
	ListRepositories(ctx context.Context, region string) ([]*Repository, error)
}

// Repository is the raw AWS-side repo record (mapped to core.Repository by discovery).
type Repository struct {
	Name          string
	SSHCloneURL   string
	HTTPCloneURL  string
	ARN           string
	Status        string
	RepositoryURL string
}

type awsAccountClient struct {
	cfg                aws.Config
	credentialsProfile string
	cliExecutor        *CLIExecutor
}

// NewAccountClient creates an AccountClient without command logging.
func NewAccountClient(credentialsProfile, region string) (AccountClient, error) {
	return NewAccountClientWithLogging(credentialsProfile, region, false)
}

// NewAccountClientWithLogging validates SDK credentials, then delegates to AWS CLI.
func NewAccountClientWithLogging(credentialsProfile, region string, logCommands bool) (AccountClient, error) {
	ctx := context.Background()

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config in region '%s': %w", region, err)
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve AWS credentials for profile '%s': %w", credentialsProfile, err)
	}

	if creds.AccessKeyID == "" {
		return nil, fmt.Errorf("AWS credentials for profile '%s' are empty - profile may not exist or have valid credentials", credentialsProfile)
	}

	if len(creds.AccessKeyID) > 4 {
		fmt.Printf("DEBUG: AWS client initialized - Profile: %s, Region: %s, AccessKeyID: %s...\n", credentialsProfile, cfg.Region, creds.AccessKeyID[:4])
	}

	return &awsAccountClient{
		cfg:                cfg,
		credentialsProfile: credentialsProfile,
		cliExecutor:        NewCLIExecutorWithLogging(credentialsProfile, logCommands),
	}, nil
}

func (c *awsAccountClient) GetAccountAlias(ctx context.Context, accountID string) (string, error) {
	return c.cliExecutor.GetAccountAlias(ctx, accountID)
}

func (c *awsAccountClient) ListRegions(ctx context.Context) ([]string, error) {
	return c.cliExecutor.ListRegions(ctx)
}

func (c *awsAccountClient) ListRepositories(ctx context.Context, region string) ([]*Repository, error) {
	return c.cliExecutor.ListRepositories(ctx, region)
}
