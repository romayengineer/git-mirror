# git-mirror-orchestrator

Parallel discovery + clone scheduling. No mpb dependency.

- `orchestrator` — `Mirror()` fan-out, semaphore, stats
- `layout` — pure local-path builder
- `ports` — `ProgressRenderer` port (implemented by `cmd/git-mirror/presenter`)
