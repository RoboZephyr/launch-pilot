# Changelog

All notable changes to this project will be documented in this file.

## [0.3.0] — 2026-04-19

Status-dot fast tooltip: replaces browser UA `title` (500–1500 ms delay, unstyled) with a self-rendered, accessible overlay that shows in ≤200 ms and dismisses in ≤100 ms.

### Added

- **Single-instance `<StatusTooltip/>` overlay**: one fixed-position `role="tooltip"` node rendered at the `<App/>` root, driven by a single `tooltipTarget` signal (`{anchor, label} | null`) in `web/lib/state.js`. 100 rows share one overlay instead of 100 per-row instances
- **`<StatusDot/>` focusable trigger**: replaces the inline `<span title=…>` dot with a `<button type="button" class="status-dot-trigger">` carrying the full tooltip text as `aria-label` — keyboard-reachable via Tab, screen-reader readable, Esc-dismissible
- **Pure `buildStatusTooltipParts(job)`**: exposes the tooltip string as an ordered array so the overlay renders one line per part. Contract `buildStatusTooltip(job) === buildStatusTooltipParts(job).join(' — ')` holds — no duplicate formatting branches
- **Pure `placeTooltip(anchor, tip, viewport)`**: viewport-aware positioning — above the anchor when space allows, flips below when top-space < 4 px, clamps horizontally into `[4, viewport.w - tip.w - 4]`
- **Live content follow**: tooltip text re-derives from `jobs.value` on every SSE refresh (5 s cadence), so an open tooltip updates instead of showing stale data
- **Playwright E2E bootstrap**: `@playwright/test` 1.x as devDependency, `playwright.config.mjs` with `webServer` auto-starting `launch-pilot --no-open`, and `e2e/tooltip.spec.mjs` covering the 6 spec acceptance criteria (show ≤200 ms, hide ≤100 ms, Esc, Tab focus, SSE live update, continuous hover)

### Changed

- Tooltip state shape collapsed from 3 signals (`tooltipAnchor`, `tooltipLabel`, `tooltipVisible`) to 1 atomic signal `tooltipTarget`. `showTooltip(anchor, label)` now short-circuits when the target is unchanged
- `PRODUCT-STATE.md` slug/repo corrected to `launch-pilot` (matches actual GitHub remote)

### Unchanged (explicit guarantees)

- Go backend, SSE schema, and Job JSON fields: byte-identical
- `buildStatusTooltip(job)` output: byte-identical
- No new frontend runtime dependencies — `preact`, `@preact/signals`, `htm` already vendored

## [0.1.0] — 2026-04-18

Initial release of Launch Pilot (formerly Launchboard).

### Added

- **Job listing**: Parse `launchctl list` output and merge with plist metadata from `~/Library/LaunchAgents` and `/Library/LaunchAgents`
- **Time-aware status**: Six `JobStatus` values — `running`, `scheduled`, `completed`, `stopped`, `error`, `offline` — derived from PID, exit code, plist schedule shape, and a configurable recent-completion window
- **Next / last run fields**: Each job exposes optional `nextRunAt` (from `StartCalendarInterval` or `StartInterval`) and `lastRunAt` (newer of stdout / stderr file mtimes). Also surfaces `startInterval` and `startCalendarInterval` in the JSON
- **Offline merge**: Plists that exist on disk but are absent from `launchctl list` are appended as synthetic `offline` jobs so unloaded plists are visible
- **StartCalendarInterval parsing**: `howett.net/plist` decoder transparently accepts either a single `<dict>` or `<array>` of dicts via a custom `UnmarshalPlist` on `CalendarEntries`
- **Next-fire calculation**: Pure-function `NextCalendarFire` / `NextIntervalFire` using stdlib `time` only. 2-year horizon guard prevents infinite loops on impossible combinations (e.g. `Day=31, Month=2`). Rejects out-of-range fields up front
- **`--recent-window` CLI flag**: Accepts any Go duration string (`30m`, `1h30m`, `24h`). Validates `1m ≤ window ≤ 24h`; outside the range exits with a clear stderr message. Default `10m`
- **Real-time updates**: SSE endpoint (`/api/events`) pushes full job list every 5 seconds; frontend auto-updates via Preact Signals
- **Service control**: Start, stop, and reload LaunchAgents through REST API with confirmation dialogs in the UI
- **Log viewer**: Tail stdout / stderr log files in-browser, up to 10,000 lines per file
- **Diagnostics engine**: 6 automated health checks per job — exit code analysis, program existence, execute permission, plist ownership, plist permissions, log path verification
- **Search**: Real-time label substring filter
- **Job classification**: Automatic categorization into Mine / System / 3rd-party based on label prefix (`com.apple.*`) and domain (`user` / `global`)
- **Multi-dimensional filtering**: Category chips, 7-tab status filter (All / Running / Scheduled / Completed / Stopped / Error / Offline), and Only Mine toggle — all composable as AND with search
- **Status badge tooltips**: Frontend `buildStatusTooltip` renders `status — Next run: … — Last run: …`, falling back to `Last run: unknown (no log path configured)` when expected
- **Colored status dots**: CSS variables `--color-scheduled` / `--color-completed` / `--color-offline` drive distinct badge colors
- **Only Mine persistence**: Toggle state saved to localStorage, restored on page load
- **Plist caching**: mtime-based cache avoids redundant disk reads when plist files haven't changed
- **Label validation**: All labels validated against `[a-zA-Z0-9._-]+` before shell execution to prevent injection
- **Embedded frontend**: Preact + Signals + htm vendored as ESM modules, embedded in Go binary via `go:embed` — single binary, no runtime dependencies
- **Homebrew distribution**: `brew install A404coder/tap/launch-pilot` for macOS (amd64 + arm64)
- **CLI flags**: `--port` (default: random), `--no-open` (skip browser auto-open), `--recent-window` (default 10m), `--version`
