module github.com/romayengineer/git-mirror-config

go 1.25.0

require (
	github.com/joho/godotenv v1.5.1
	gopkg.in/yaml.v3 v3.0.1
)

// Local monorepo development. On publish, drop this replace.
replace github.com/romayengineer/git-mirror-core => ../core
