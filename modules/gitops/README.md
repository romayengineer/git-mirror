# git-mirror-gitops

Git operations with SRP split:

- `gitexec` (infrastructure): runs `git clone` / `fetch --all` / `config --get` via exec+path.
- `gitsync` (application): clone-vs-pull policy, URL match, dir setup, cleanup on fail.
