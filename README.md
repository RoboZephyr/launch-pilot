# Launch Pilot

Visual control console for macOS launchd services. View status, read logs, diagnose issues, and manage user-domain launch agents — all from your browser.

## Features

**Real-time monitoring** — Job status (running / stopped / error), PID, and exit codes update automatically via SSE, no manual refresh needed.

**Service control** — Start, stop, and reload LaunchAgents with one click. Confirmation dialogs prevent accidental operations.

**Log viewer** — Tail stdout/stderr log files directly in the browser. Search within logs and load up to 10,000 lines.

**Diagnostics** — 6 automated health checks per job:

| Check | What it verifies |
|-------|-----------------|
| Exit Code | Maps exit code to human-readable explanation (e.g. 127 = not found) |
| Program Exists | Executable path exists on disk |
| Program Executable | File has execute permission |
| Plist Owner | Plist owned by current user |
| Plist Permissions | No group/world write bits (security) |
| Log Path | Parent directories for log files exist |

**Search & filter** — Filter jobs by label in real time.

## Install

```bash
brew install A404coder/tap/launch-pilot
```

Or build from source:

```bash
git clone git@github.com:A404coder/launch-pilot.git
cd launch-pilot
make build
```

## Usage

```bash
launch-pilot              # random port, opens browser
launch-pilot --port 8080  # explicit port
launch-pilot --no-open    # skip auto-opening browser
launch-pilot --version    # print version
```

## API

All endpoints return JSON.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/jobs` | List all jobs with status, PID, exit code, plist metadata |
| GET | `/api/jobs/{label}` | Get single job details |
| POST | `/api/jobs/{label}/start` | Kick-start a job |
| POST | `/api/jobs/{label}/stop` | Send SIGTERM to a running job |
| POST | `/api/jobs/{label}/reload` | Unload and reload (bootout + bootstrap) |
| GET | `/api/jobs/{label}/logs?lines=200` | Tail stdout/stderr logs (max 10,000 lines) |
| GET | `/api/jobs/{label}/diagnose` | Run 6 diagnostic checks |
| GET | `/api/events` | SSE stream, pushes job list every 5s |

## Architecture

```
Browser (Preact + Signals)
    │
    ├── SSE (/api/events)     ← real-time job status
    ├── REST (/api/jobs/...)  ← actions, logs, diagnostics
    │
    ▼
Go HTTP Server (embedded frontend)
    │
    ├── Service layer         ← merges launchctl + plist data
    ├── Diagnose engine       ← 6 read-only health checks
    │
    ▼
launchctl CLI + ~/Library/LaunchAgents plist files
```

- **Plist scanning**: reads `~/Library/LaunchAgents` and `/Library/LaunchAgents`, with mtime-based caching
- **Label validation**: all labels validated against `[a-zA-Z0-9._-]+` before shell execution
- **Single binary**: frontend assets embedded in the Go binary, no external dependencies at runtime

## Requirements

- macOS (uses launchctl)
- Go 1.22+ (build only)

## License

MIT
