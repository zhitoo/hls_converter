# AGENTS.md — HLS Converter

Go HTTP service converting videos to HLS via FFmpeg. No database — all state is files on disk.

## Commands

```bash
go run main.go          # Run server on :8080
go build -o hls_converter .
go test ./...
docker compose up -d    # Recommended: Docker with FFmpeg pre-installed
```

## Architecture

Entry: `main.go` → wires packages → `http.Server` on `:8080`

| Package | Role |
|---|---|
| `internal/handler` | HTTP routes, request validation, JSON responses |
| `internal/auth` | Bearer token middleware, user lookup from `users.json` |
| `internal/queue` | In-memory job queue (capacity 1000), worker pool (3 goroutines) |
| `internal/converter` | FFmpeg wrapper — shells out to `ffmpeg` binary |
| `internal/task` | Task state persistence as `tasks/<uuid>.json` files |
| `internal/storage` | HLS output dirs under `storage/users/<uid>/<tid>/` |
| `internal/tasklog` | Per-task log files at `storage/logs/<tid>.log` |
| `internal/cleanup` | Deletes tasks older than 24h, drains stale tasks on startup |

## Key Gotchas

- **FFmpeg required**: Locally or inside Docker. The `converter` package shells out to `ffmpeg` — no Go bindings.
- **No test files exist**: The repo has zero `*_test.go` files. `go test ./...` passes vacuously.
- **Path traversal protection**: `storage.validatePathSegment` rejects `..`, `/`, `\` in user/task IDs. Don't bypass this.
- **Worker retry**: 4 retries with exponential backoff (2s × attempt) per resolution. Progress is tracked per-resolution across the worker loop.
- **Master playlist**: Only generated when `len(resolutions) > 1`. Single resolution → no `master.m3u8`.
- **User ownership**: `server.requireOwned` checks task ownership against the auth'd user. All `/api/v1/status|download|logs/{task_id}` routes use this.
- **Startup cleanup**: `cleaner.DrainStale()` deletes Pending/Processing tasks from before the last restart.

## Data Layout

```
tasks/<uuid>.json           # Task state (status, progress, config)
storage/users/<uid>/<tid>/  # HLS output files
storage/logs/<tid>.log      # Raw FFmpeg output
users.json                  # User registry (read at startup)
```

## Config

Hardcoded in `main.go`: `:8080`, worker pool 3, queue 1000. Change there, not via env vars.
