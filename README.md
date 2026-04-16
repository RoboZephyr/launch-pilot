# Launch Pilot

Visual control console for macOS launchd services. View status, read logs, diagnose issues, and manage user-domain launch agents — all from your browser.

## Features

**Real-time monitoring** — Job status (running / stopped / error), PID, and exit codes update automatically via SSE every 5 seconds, no manual refresh needed.

**Service control** — Start, stop, and reload LaunchAgents with one click. Confirmation dialogs prevent accidental operations.

**Log viewer** — Tail stdout/stderr log files directly in the browser. Load up to 10,000 lines per file.

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

**Multi-dimensional filtering** — Three filter dimensions that compose as AND:

- **Category chips** — All / Mine / System / 3rd-party, each showing count badge
- **Status tabs** — All / Running / Stopped / Error, each showing count in parentheses
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
launch-pilot              # random port, auto-opens browser
launch-pilot --port 8080  # listen on explicit port
launch-pilot --no-open    # start server without opening browser
launch-pilot --version    # print version and exit
```

The server binds to `127.0.0.1` (localhost only). Press `Ctrl+C` to shut down gracefully (5-second timeout).

## API

All endpoints return JSON. Labels must match `[a-zA-Z0-9._-]+`.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/jobs` | List all jobs with status, PID, exit code, plist metadata |
| GET | `/api/jobs/{label}` | Get single job details |
| POST | `/api/jobs/{label}/start` | Start a stopped job |
| POST | `/api/jobs/{label}/stop` | Send SIGTERM to a running job |
| POST | `/api/jobs/{label}/reload` | Unload and reload (bootout + bootstrap) |
| GET | `/api/jobs/{label}/logs?lines=200` | Tail stdout/stderr logs (max 10,000 lines) |
| GET | `/api/jobs/{label}/diagnose` | Run 6 diagnostic checks |
| GET | `/api/events` | SSE stream — pushes full job list every 5s |

## Architecture

```
Browser (Preact + Signals, no build step)
    │
    ├── SSE (/api/events)     ← real-time job status push
    ├── REST (/api/jobs/...)  ← actions, logs, diagnostics
    │
    ▼
Go HTTP Server (net/http, embedded frontend via go:embed)
    │
    ├── Service layer         ← merges launchctl list + plist data
    ├── Diagnose engine       ← 6 read-only health checks
    │
    ▼
launchctl CLI + plist files
    ├── ~/Library/LaunchAgents     (user domain)
    └── /Library/LaunchAgents      (global domain)
```

**Frontend stack**: Preact + Signals + htm, vendored as ESM modules via import map. No bundler, no transpiler — browser-native ES modules.

**Plist scanning**: Reads `~/Library/LaunchAgents` and `/Library/LaunchAgents` with mtime-based caching to avoid redundant disk reads.

**Single binary**: All frontend assets (HTML, JS, CSS) are embedded in the Go binary via `go:embed`. No external files needed at runtime.

**Client-side filtering**: All filter/search logic runs in the browser using Preact Signals `computed()`. The backend pushes the full job list (~50 KB for 300 jobs) via SSE; the 4-stage filter pipeline (onlyMine → category → status → search) runs in < 1ms on 300 items.

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
```

The test loader maps bare specifiers (`@preact/signals`) to vendored ESM files for Node.js compatibility.

### Project structure

```
cmd/launch-pilot/       Go entrypoint (CLI flags, server startup)
internal/
  launchd/              Job model, launchctl parser, service layer
  diagnose/             6-check diagnostic engine
  plist/                Plist reader with mtime cache
  server/               HTTP router, REST handlers, SSE handler
web/
  app.js                Preact app root
  index.html            HTML shell with import map
  components/           UI components (JobRow, FilterBar, SearchBar, etc.)
  lib/                  State signals, classification logic, SSE client, API client
  styles/               CSS (single main.css, CSS variables for theming)
  vendor/               Vendored ESM: Preact, htm, Signals
```

## License

MIT
