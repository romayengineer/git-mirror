package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

type CLIExecutor struct {
	profile string
}

func NewCLIExecutor(profile string) *CLIExecutor {
	return &CLIExecutor{profile: profile}
}

type cliRepositoryOutput struct {
	Repositories []cliRepository `json:"repositories"`
}

type cliRepository struct {
	RepositoryName string `json:"repositoryName"`
	RepositoryID   string `json:"repositoryId"`
	CloneURLSsh    string `json:"cloneUrlSsh"`
	CloneURLHttp   string `json:"cloneUrlHttp"`
	Arn            string `json:"repositoryArn"`
}

func (c *CLIExecutor) ListRepositories(ctx context.Context, region string) ([]*Repository, error) {
	listArgs := []string{
		"codecommit",
		"list-repositories",
		"--region", region,
		"--output", "json",
	}

	if c.profile != "" && c.profile != "default" {
		listArgs = append(listArgs, "--profile", c.profile)
	}

	cmd := exec.CommandContext(ctx, "aws", listArgs...)
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("AWS CLI command timed out for region %s", region)
		}
		return nil, fmt.Errorf("failed to execute AWS CLI list-repositories for region %s: %w", region, err)
	}

	var listResult cliRepositoryOutput
	if err := json.Unmarshal(output, &listResult); err != nil {
		return nil, fmt.Errorf("failed to parse AWS CLI list-repositories output for region %s: %w", region, err)
	}

	if len(listResult.Repositories) == 0 {
		return []*Repository{}, nil
	}

	repoNames := make([]string, len(listResult.Repositories))
	for i, repo := range listResult.Repositories {
		repoNames[i] = repo.RepositoryName
	}

	repositories := make([]*Repository, 0, len(repoNames))
	const batchSize = 25

	for i := 0; i < len(repoNames); i += batchSize {
		end := i + batchSize
		if end > len(repoNames) {
			end = len(repoNames)
		}

		batch := repoNames[i:end]

		batchArgs := []string{
			"codecommit",
			"batch-get-repositories",
			"--region", region,
			"--repository-names",
		}
		batchArgs = append(batchArgs, batch...)
		batchArgs = append(batchArgs, "--output", "json")

		if c.profile != "" && c.profile != "default" {
			batchArgs = append(batchArgs, "--profile", c.profile)
		}

		cmd = exec.CommandContext(ctx, "aws", batchArgs...)
		output, err = cmd.Output()
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("AWS CLI command timed out for region %s", region)
			}
			return nil, fmt.Errorf("failed to execute AWS CLI batch-get-repositories (batch %d-%d) for region %s: %w", i+1, end, region, err)
		}

		var batchResult struct {
			Repositories []cliRepository `json:"repositories"`
		}
		if err := json.Unmarshal(output, &batchResult); err != nil {
			return nil, fmt.Errorf("failed to parse AWS CLI batch-get-repositories output for region %s: %w", region, err)
		}

		for _, repo := range batchResult.Repositories {
			repositories = append(repositories, &Repository{
				Name:        repo.RepositoryName,
				SSHCloneURL: repo.CloneURLSsh,
				HTTPCloneURL: repo.CloneURLHttp,
				ARN:         repo.Arn,
			})
		}
	}

	return repositories, nil
}

func (c *CLIExecutor) GetAccountAlias(ctx context.Context, accountID string) (string, error) {
	args := []string{
		"iam",
		"list-account-aliases",
		"--output", "json",
	}

	if c.profile != "" && c.profile != "default" {
		args = append(args, "--profile", c.profile)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)

	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("AWS CLI command timed out getting account alias")
		}
		return "", fmt.Errorf("failed to execute AWS CLI list-account-aliases: %w", err)
	}

	var result struct {
		AccountAliases []string `json:"AccountAliases"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("failed to parse AWS CLI output: %w", err)
	}

	if len(result.AccountAliases) > 0 {
		return result.AccountAliases[0], nil
	}

	return accountID, nil
}

func (c *CLIExecutor) ListRegions(ctx context.Context) ([]string, error) {
	args := []string{
		"ec2",
		"describe-regions",
		"--output", "json",
	}

	if c.profile != "" && c.profile != "default" {
		args = append(args, "--profile", c.profile)
	}

	cmd := exec.CommandContext(ctx, "aws", args...)

	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("AWS CLI command timed out listing regions")
		}
		return nil, fmt.Errorf("failed to execute AWS CLI describe-regions: %w", err)
	}

	var result struct {
		Regions []struct {
			RegionName string `json:"RegionName"`
		} `json:"Regions"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AWS CLI output: %w", err)
	}

	regions := make([]string, 0, len(result.Regions))
	for _, region := range result.Regions {
		regions = append(regions, region.RegionName)
	}

	return regions, nil
}

// CheckAWSCLIAvailable checks if the AWS CLI is available and accessible
func CheckAWSCLIAvailable() error {
	cmd := exec.Command("aws", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("AWS CLI is not available or not installed: %w", err)
	}
	return nil
}
