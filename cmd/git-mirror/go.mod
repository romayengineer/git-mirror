module github.com/romayengineer/git-mirror

go 1.25.0

require (
	github.com/romayengineer/git-mirror-config v0.0.0
	github.com/romayengineer/git-mirror-core v0.0.0
	github.com/romayengineer/git-mirror-gitops v0.0.0
	github.com/romayengineer/git-mirror-orchestrator v0.0.0
	github.com/romayengineer/git-mirror-provider-codecommit v0.0.0
	github.com/romayengineer/git-mirror-provider-github v0.0.0
	github.com/vbauerster/mpb/v8 v8.14.0
)

require (
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d // indirect
	github.com/aws/aws-sdk-go-v2 v1.24.0 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.26.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.16.12 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.2.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.5.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.10.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.18.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.21.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.26.5 // indirect
	github.com/aws/smithy-go v1.19.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/mattn/go-runewidth v0.0.27 // indirect
	github.com/romayengineer/git-mirror-exec v0.0.0 // indirect
	github.com/romayengineer/git-mirror-path v0.0.0 // indirect
	github.com/vbauerster/cupwriter v0.0.4 // indirect
	golang.org/x/sys v0.47.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Local monorepo development. On publish, drop these replaces.
replace (
	github.com/romayengineer/git-mirror-config => ../../modules/config
	github.com/romayengineer/git-mirror-core => ../../modules/core
	github.com/romayengineer/git-mirror-exec => ../../modules/exec
	github.com/romayengineer/git-mirror-gitops => ../../modules/gitops
	github.com/romayengineer/git-mirror-orchestrator => ../../modules/orchestrator
	github.com/romayengineer/git-mirror-path => ../../modules/path
	github.com/romayengineer/git-mirror-provider-codecommit => ../../modules/provider-codecommit
	github.com/romayengineer/git-mirror-provider-github => ../../modules/provider-github
)
