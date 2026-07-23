package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/codecommit"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type AccountClient interface {
	GetAccountAlias(ctx context.Context, accountID string) (string, error)
	ListRegions(ctx context.Context) ([]string, error)
	ListRepositories(ctx context.Context, region string) ([]*Repository, error)
}

type Repository struct {
	Name          string
	SSHCloneURL   string
	HTTPCloneURL  string
	ARN           string
	Status        string
	RepositoryURL string
}

type awsAccountClient struct {
	cfg            aws.Config
	credentialsProfile string
}

func NewAccountClient(credentialsProfile string) (AccountClient, error) {
	opts := []func(*config.LoadOptions) error{}

	if credentialsProfile != "" && credentialsProfile != "default" {
		opts = append(opts, config.WithSharedConfigProfile(credentialsProfile))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &awsAccountClient{
		cfg:                cfg,
		credentialsProfile: credentialsProfile,
	}, nil
}

func (c *awsAccountClient) GetAccountAlias(ctx context.Context, accountID string) (string, error) {
	client := iam.NewFromConfig(c.cfg)

	result, err := client.ListAccountAliases(ctx, &iam.ListAccountAliasesInput{})
	if err != nil {
		return "", fmt.Errorf("failed to list account aliases: %w", err)
	}

	if len(result.AccountAliases) > 0 {
		return result.AccountAliases[0], nil
	}

	return accountID, nil
}

func (c *awsAccountClient) ListRegions(ctx context.Context) ([]string, error) {
	client := ec2.NewFromConfig(c.cfg)

	result, err := client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list regions: %w", err)
	}

	regions := make([]string, 0, len(result.Regions))
	for _, region := range result.Regions {
		if region.RegionName != nil {
			regions = append(regions, *region.RegionName)
		}
	}

	return regions, nil
}

func (c *awsAccountClient) ListRepositories(ctx context.Context, region string) ([]*Repository, error) {
	cfg := c.cfg.Copy()
	cfg.Region = region

	client := codecommit.NewFromConfig(cfg)

	var repositories []*Repository
	var repositoryNames []string

	paginator := codecommit.NewListRepositoriesPaginator(client, &codecommit.ListRepositoriesInput{})

	for paginator.HasMorePages() {
		result, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories in region %s: %w", region, err)
		}

		for _, repo := range result.Repositories {
			if repo.RepositoryName != nil {
				repositoryNames = append(repositoryNames, *repo.RepositoryName)
			}
		}
	}

	if len(repositoryNames) == 0 {
		return repositories, nil
	}

	batchResult, err := client.BatchGetRepositories(ctx, &codecommit.BatchGetRepositoriesInput{
		RepositoryNames: repositoryNames,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to batch get repositories in region %s: %w", region, err)
	}

	for _, repo := range batchResult.Repositories {
		repositories = append(repositories, &Repository{
			Name:        aws.ToString(repo.RepositoryName),
			SSHCloneURL: aws.ToString(repo.CloneUrlSsh),
			HTTPCloneURL: aws.ToString(repo.CloneUrlHttp),
			ARN:         aws.ToString(repo.Arn),
		})
	}

	return repositories, nil
}
