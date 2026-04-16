# Changelog

All notable changes to this project will be documented in this file.

## [0.1.0] — 2026-04-16

Initial release of Launch Pilot (formerly Launchboard).

### Added

- **Job listing**: Parse `launchctl list` output and merge with plist metadata from `~/Library/LaunchAgents` and `/Library/LaunchAgents`
- **Real-time updates**: SSE endpoint (`/api/events`) pushes full job list every 5 seconds; frontend auto-updates via Preact Signals
- **Service control**: Start, stop, and reload LaunchAgents through REST API with confirmation dialogs in the UI
- **Log viewer**: Tail stdout/stderr log files in-browser, up to 10,000 lines per file
- **Diagnostics engine**: 6 automated health checks per job — exit code analysis, program existence, execute permission, plist ownership, plist permissions, log path verification
- **Search**: Real-time label substring filter
- **Job classification**: Automatic categorization into Mine / System / 3rd-party based on label prefix (`com.apple.*`) and domain (`user` / `global`)
- **Multi-dimensional filtering**: Category chips, status tabs (Running / Stopped / Error), and Only Mine toggle — all composable as AND with search
- **Only Mine persistence**: Toggle state saved to localStorage, restored on page load
- **Plist caching**: mtime-based cache avoids redundant disk reads when plist files haven't changed
- **Label validation**: All labels validated against `[a-zA-Z0-9._-]+` before shell execution to prevent injection
- **Embedded frontend**: Preact + Signals + htm vendored as ESM modules, embedded in Go binary via `go:embed` — single binary, no runtime dependencies
- **Homebrew distribution**: `brew install A404coder/tap/launch-pilot` for macOS (amd64 + arm64)
- **CLI flags**: `--port` (default: random), `--no-open` (skip browser auto-open), `--version`
