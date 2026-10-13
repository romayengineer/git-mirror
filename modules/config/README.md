# git-mirror-config

Application config DTOs (`Config`, `GitHubConfig`, `CodeCommitConfig`) + loaders:

- `env` — `.env` / environment variables (godotenv)
- `yaml` — `config.yaml` (gopkg.in/yaml.v3)
- `pathutil` — `~` expansion + `k:v` pair parsing
