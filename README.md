# Launch Pilot

Visual control console for macOS launchd services. View status, read logs, diagnose issues, and manage user-domain launch agents — all from your browser.

## Features

**Time-aware job status** — Six status values distinguish scheduled jobs waiting for their next trigger, recently-completed runs, truly offline plists, and classic running / stopped / error states. The backend derives these from PID, exit code, plist schedule shape, and a configurable recent-completion window.

| Status | Meaning |
|--------|---------|
| running | PID > 0 |
| scheduled | PID = 0, clean exit, has `StartInterval` / `StartCalendarInterval` / `RunAtLoad`, no recent log mtime |
| completed | PID = 0, clean exit, log mtime within `--recent-window` (default 10m) |
| stopped | PID = 0, clean exit, no schedule and no recent log mtime |
| error | Non-zero last exit status |
| offline | plist file exists but `launchctl list` does not return the label |

**Next / last run heuristics** — Each job carries optional `nextRunAt` (computed from `StartCalendarInterval` or `StartInterval`) and `lastRunAt` (newer of stdout / stderr mtimes). Status badges show both on hover.

**Real-time monitoring** — SSE (`/api/events`) pushes the full job list every 5 seconds; the frontend refreshes via Preact Signals without a manual reload.

**Service control** — Start, stop, and reload LaunchAgents with one click. Confirmation dialogs prevent accidental operations.

**Log viewer** — Tail stdout / stderr log files directly in the browser. Load up to 10,000 lines per file.

**Diagnostics** — 6 automated health checks per job:

| Check | What it verifies |
|-------|-----------------|
| Exit Code | Maps exit code to human-readable explanation (e.g. 127 = command not found) |
| Program Exists | Executable path exists on disk |
| Program Executable | File has execute permission |
| Plist Owner | Plist owned by current user |
| Plist Permissions | No group/world write bits (security) |
| Log Path | Parent directories for log files exist |

**Job classification** — Each job is automatically categorized based on label prefix and domain:

| Category | Rule | Badge color |
|----------|------|-------------|
| Mine | `domain=user` and label does not start with `com.apple.` | Blue |
| System | Label starts with `com.apple.` | Gray |
| 3rd-party | `domain=global` and label does not start with `com.apple.` | Purple |

**Multi-dimensional filtering** — Four filter dimensions that compose as AND:

- **Category chips** — All / Mine / System / 3rd-party, each showing count badge
- **Status tabs** — All / Running / Scheduled / Completed / Stopped / Error / Offline, each showing count
- **Only Mine toggle** — One-click shortcut to show only user-created jobs (persisted to localStorage)
- **Search** — Label substring filter, composable with all above

## Install

### Homebrew

```bash
brew install A404coder/tap/launch-pilot
```

### Build from source

```bash
git clone git@github.com:A404coder/launch-pilot.git
cd launch-pilot
make build
```

This produces a `launch-pilot` binary in the project root. The frontend is embedded in the binary — no separate build step or runtime dependencies needed.

## Usage

```bash
launch-pilot                          # random port, auto-opens browser
launch-pilot --port 8080              # listen on explicit port
launch-pilot --no-open                # start server without opening browser
launch-pilot --recent-window 30m      # mark jobs as "completed" if they ran in the last 30m
launch-pilot --version                # print version and exit
```

The server binds to `127.0.0.1` (localhost only). Press `Ctrl+C` to shut down gracefully (5-second timeout).

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--port` | int | `0` (random) | Listen port on `127.0.0.1` |
| `--no-open` | bool | `false` | Skip auto-opening the browser |
| `--recent-window` | duration | `10m` | How long after `lastRunAt` a job still shows as `completed`. Valid range: `1m`–`24h`. Accepts any Go duration string (`30m`, `1h30m`, `24h`). Outside the range exits with a clear stderr message. |
| `--version` | bool | `false` | Print version and exit |

## API

All endpoints return JSON. Labels must match `[a-zA-Z0-9._-]+`.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/jobs` | List all jobs including offline plists |
| GET | `/api/jobs/{label}` | Get single job details |
| POST | `/api/jobs/{label}/start` | `launchctl kickstart` |
| POST | `/api/jobs/{label}/stop` | `launchctl kill SIGTERM` |
| POST | `/api/jobs/{label}/reload` | `launchctl bootout` + `bootstrap` |
| GET | `/api/jobs/{label}/logs?lines=200` | Tail stdout / stderr (max 10,000 lines) |
| GET | `/api/jobs/{label}/diagnose` | Run the 6 diagnostic checks |
| GET | `/api/events` | SSE stream — pushes full job list every 5s |

### Job JSON shape

```json
{
  "label": "com.example.backup",
  "pid": 0,
  "lastExitStatus": 0,
  "status": "scheduled",
  "plistPath": "/Users/you/Library/LaunchAgents/com.example.backup.plist",
  "program": "/usr/local/bin/backup",
  "programArgs": ["/usr/local/bin/backup", "--daily"],
  "standardOutPath": "/tmp/backup.out",
  "standardErrPath": "/tmp/backup.err",
  "runAtLoad": false,
  "keepAlive": false,
  "domain": "user",
  "nextRunAt": "2026-04-19T03:00:00Z",
  "lastRunAt": "2026-04-18T03:00:04Z",
  "startInterval": 0,
  "startCalendarInterval": [{ "Hour": 3, "Minute": 0 }]
}
```

`nextRunAt`, `lastRunAt`, `startInterval`, and `startCalendarInterval` are optional (omitted when empty). Old clients that ignore these fields continue to work.

## Architecture

```
Browser (Preact + Signals, no build step)
    |
    +-- SSE (/api/events)     <- real-time job status push
    +-- REST (/api/jobs/...)  <- actions, logs, diagnostics
    |
    v
Go HTTP Server (net/http, embedded frontend via go:embed)
    |
    +-- Service layer         <- merges launchctl list + plist data
    |     NextCalendarFire / NextIntervalFire / LastRunAt
    |     DeriveStatus(pid, exit, plist, lastRunAt, now, window)
    +-- Diagnose engine       <- 6 read-only health checks
    |
    v
launchctl CLI + plist files
    +-- ~/Library/LaunchAgents     (user domain)
    +-- /Library/LaunchAgents      (global domain)
```

**Frontend stack**: Preact + Signals + htm, vendored as ESM modules via import map. No bundler, no transpiler — browser-native ES modules.

**Plist scanning**: Reads `~/Library/LaunchAgents` and `/Library/LaunchAgents` with mtime-based caching. Plists that exist on disk but are absent from `launchctl list` are appended as synthetic `offline` jobs so the UI can surface unloaded plists.

**Single binary**: All frontend assets (HTML, JS, CSS) are embedded in the Go binary via `go:embed`. No external files needed at runtime.

**Client-side filtering**: All filter/search logic runs in the browser using Preact Signals `computed()`. The backend pushes the full job list (~50 KB for 300 jobs) via SSE; the filter pipeline (onlyMine → category → status → search) runs in < 1ms on 300 items.

**Zero new Go deps**: calendar / interval next-fire calculation uses the standard-library `time` package only. No cron library.

## Development

### Prerequisites

- macOS (uses `launchctl`)
- Go 1.22+

### Commands

```bash
make build    # compile binary with version from git tag
make test     # run Go tests (go test ./... -count=1)
make run      # build + run
make clean    # remove binary
```

### Frontend tests

Frontend modules are tested with Node.js built-in test runner:

```bash
node --test web/lib/classify.test.js
node --loader web/lib/test-loader.mjs --test web/lib/state.test.js
node --loader web/lib/test-loader.mjs --test web/components/filter-bar.test.js
node --loader web/lib/test-loader.mjs --test web/components/job-row.test.js
```

The test loader maps bare specifiers (`@preact/signals`) to vendored ESM files for Node.js compatibility.

### Project structure

```
cmd/launch-pilot/       Go entrypoint (CLI flags, server startup)
internal/
  launchd/              Job model, launchctl parser, service layer, DeriveStatus
  diagnose/             6-check diagnostic engine
  plist/                Plist reader + mtime cache + NextCalendarFire / NextIntervalFire / LastRunAt
  server/               HTTP router, REST handlers, SSE handler
web/
  app.js                Preact app root
  index.html            HTML shell with import map
  components/           UI components (JobRow, FilterBar, SearchBar, job-tooltip, etc.)
  lib/                  State signals, classification logic, SSE client, API client
  styles/               CSS (single main.css, CSS variables for theming)
  vendor/               Vendored ESM: Preact, htm, Signals
```

## License

MIT
