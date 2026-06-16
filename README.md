# HLS Converter

A Go HTTP service that converts videos to HLS (HTTP Live Streaming) format using FFmpeg. Supports multi-resolution output with adaptive bitrate master playlists.

## Requirements

- Go 1.26+
- FFmpeg installed and available in `$PATH`

## Setup

```bash
# Install dependencies
go mod download

# Run
go run main.go

# Build
go build -o hls_converter .
```

The server starts on `:8080`.

## Configuration

**`users.json`** — user registry with API keys and concurrency limits:

```json
[
  {
    "user_id": "user-001",
    "api_key": "secret-token-abc123",
    "max_concurrent_tasks": 2
  }
]
```

## Authentication

All endpoints require a Bearer token matching an `api_key` in `users.json`.

```
Authorization: Bearer secret-token-abc123
```

## API

### Convert a video

```
POST /api/v1/convert
```

**Request body:**

```json
{
  "video_url": "https://example.com/video.mp4",
  "resolutions": [720, 480],
  "chunk_duration": 10,
  "audio_channels": 2
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `video_url` | string | yes | URL of the source video |
| `resolutions` | []int | no | Target heights in pixels (e.g. `[1080, 720, 480]`). Omit for original quality. |
| `chunk_duration` | int | no | Segment length in seconds (default: 10) |
| `audio_channels` | int | no | Number of audio channels |

**Response:**

```json
{
  "task_id": "ae0f25e0-bfc8-4bed-af63-a1a8ab391a18",
  "message": "task created and queued successfully"
}
```

---

### Check status

```
GET /api/v1/status/{task_id}
```

**Response:**

```json
{
  "task_id": "ae0f25e0-...",
  "status": "Completed",
  "progress": 100,
  "current_step": "Completed",
  "retry_count": 0,
  "created_at": "2026-06-16T12:00:00Z",
  "updated_at": "2026-06-16T12:01:00Z",
  "qualities": [
    { "height": 720, "label": "720p", "playlist": "720p/output.m3u8" },
    { "height": 480, "label": "480p", "playlist": "480p/output.m3u8" }
  ],
  "master_playlist": "master.m3u8"
}
```

`qualities` and `master_playlist` are only present when `status` is `Completed`.

Possible statuses: `Pending`, `Processing`, `Completed`, `Failed`.

---

### Download output

```
GET /api/v1/download/{task_id}
```

Returns a `.zip` archive containing all HLS files, preserving the directory structure:

```
<task_id>.zip
├── master.m3u8         ← adaptive bitrate playlist (multi-resolution only)
├── 720p/
│   ├── output.m3u8
│   ├── segment_000.ts
│   └── ...
└── 480p/
    ├── output.m3u8
    └── ...
```

For single-resolution output, the zip contains `output.m3u8` and segment files at the root.

---

### Stream logs

```
GET /api/v1/logs/{task_id}
```

Returns the raw FFmpeg conversion log for the task.

## Storage layout

```
storage/
├── users/
│   └── <user_id>/
│       └── <task_id>/        ← HLS output files
└── logs/
    └── <task_id>.log         ← FFmpeg logs

tasks/
└── <task_id>.json            ← task state
```

## Concurrency & cleanup

- Worker pool size: 3 (configurable via `workerPool` in `main.go`)
- Task queue capacity: 1000
- Each user has a `max_concurrent_tasks` limit enforced at submission time
- Failed conversions are retried up to 4 times with exponential backoff
- Tasks and their files are automatically deleted after **24 hours**
- Pending/Processing tasks from before a server restart are deleted on startup
