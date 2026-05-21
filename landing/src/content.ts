import type { ServiceStatus } from "@/kit/tokens";

export interface ServiceRow {
  name: string;
  status: ServiceStatus;
  pid: number | null;
  uptime: string;
  note?: string;
}

export interface LogLine {
  ts: string;
  level: "info" | "warn" | "error" | "stream";
  service: string;
  body: string;
}

export interface FeatureItem {
  index: string;
  title: string;
  body: string;
  accent?: "brand" | "accent" | "warning";
}

export interface StepItem {
  index: string;
  title: string;
  body: string;
  command?: string;
}

export interface Stat {
  value: string;
  label: string;
  caption?: string;
}

export interface FaqItem {
  question: string;
  answer: string;
}

export const content = {
  meta: {
    title: "Launch Pilot — every launchd job, in one browser tab",
    description:
      "A local Go binary that turns macOS launchctl into a real-time browser console. See every LaunchAgent, tail logs over SSE, run six diagnostic checks per job. Runs on 127.0.0.1, no cloud, no telemetry.",
  },
  brand: {
    name: "Launch Pilot",
    mark: "LP",
    tagline: "launchctl, finally legible",
    installCommand: "brew install RoboZephyr/tap/launch-pilot",
    launchCommand: "launch-pilot",
    github: "github.com/RoboZephyr/launch-pilot",
    githubUrl: "https://github.com/RoboZephyr/launch-pilot",
    tapUrl: "https://github.com/RoboZephyr/homebrew-tap",
    changelogUrl:
      "https://github.com/RoboZephyr/launch-pilot/blob/main/CHANGELOG.md",
    version: "v0.3.1",
    license: "MIT",
  },
  nav: [
    { label: "Why", href: "#why" },
    { label: "How", href: "#how" },
    { label: "Features", href: "#features" },
    { label: "Install", href: "#install" },
  ],
  hero: {
    eyebrow: "v0.3.1 · MIT · macOS",
    headline: ["Every launchd job", "on your Mac,", "in one browser tab."],
    body: "Launch Pilot is a single Go binary that turns launchctl into a live browser console. Six time-aware statuses, per-job log tail, six diagnostic checks. Binds to 127.0.0.1; no daemon, no telemetry, no cloud.",
    primaryCta: { label: "Install via Homebrew", href: "#install" },
    secondaryCta: { label: "Read the source", href: "https://github.com/RoboZephyr/launch-pilot" },
    metaRow: [
      { key: "runtime", value: "Local · 127.0.0.1 only" },
      { key: "scope", value: "user agents · system daemons" },
      { key: "data", value: "launchctl + plist · zero telemetry" },
    ],
  },
  livePanel: {
    title: "/api/jobs · live",
    subtitle: "SSE push every 5s",
    sseHeartbeat: "GET /api/events · event-stream",
    services: [
      { name: "com.docker.vmnetd", status: "running", pid: 849, uptime: "3d 12h" },
      { name: "homebrew.mxcl.postgresql@16", status: "running", pid: 3611, uptime: "1d 02h" },
      { name: "com.tailscale.tailscaled", status: "running", pid: 1204, uptime: "14h 03m" },
      { name: "com.user.nightly-backup", status: "error", pid: null, uptime: "exit 24 · 03:12" },
      { name: "com.user.weekly-cleanup", status: "scheduled", pid: null, uptime: "next · Sun 03:00" },
      { name: "homebrew.autoupdate", status: "completed", pid: null, uptime: "ran 6m ago" },
      { name: "com.zsa.wally.daemon", status: "offline", pid: null, uptime: "plist present" },
    ] satisfies ServiceRow[],
  },
  why: {
    eyebrow: "Why",
    headline: "launchctl tells you the truth — in the worst possible format.",
    paragraphs: [
      "Fresh macOS ships with hundreds of LaunchAgents and Daemons. Install Docker, Homebrew services, Tailscale, a CUDA toolkit, an old Python — each one drops a plist into your domain. When something breaks at 3am, the diagnosis ritual is the same every time: `launchctl list | grep`, open the plist in a text editor, hope Console.app still has the window.",
      "Launch Pilot replaces that ritual. Six status values that distinguish scheduled jobs from genuinely offline ones. Per-job log tail over Server-Sent Events. Six read-only health checks (exit code, program exists, plist permissions, log path). Everything in a tab that opens on first run.",
    ],
    contrast: {
      before: {
        label: "Without Launch Pilot",
        bullets: [
          "launchctl list | grep — every time you forget a label",
          "Plist XML in a text editor; squint at KeepAlive, StartInterval",
          "Console.app, then /var/log, then giving up",
          "No idea whether a job is offline or just waiting for its schedule",
        ],
      },
      after: {
        label: "With Launch Pilot",
        bullets: [
          "One browser tab. Every job. Sorted by status.",
          "Six statuses derived from PID, exit, plist shape, log mtime",
          "Per-job log tail up to 10,000 lines, filtered live",
          "Six health checks per job, plus a one-click start / stop / reload",
        ],
      },
    },
  },
  steps: [
    {
      index: "01",
      title: "Install",
      body: "One Homebrew formula in a public tap. The binary is a single static Go executable with the frontend embedded via go:embed — nothing else to deploy.",
      command: "brew install RoboZephyr/tap/launch-pilot",
    },
    {
      index: "02",
      title: "Run",
      body: "No subcommand, no config file. The server picks a free port on 127.0.0.1 and opens your default browser. Pass --port to pin it, --no-open to skip the browser, --recent-window to tune the completed-job window.",
      command: "launch-pilot",
    },
    {
      index: "03",
      title: "Investigate",
      body: "Filter by category (Mine / System / 3rd-party), drill into a job to see status reasoning, tail its logs, run six diagnostic checks, and start / stop / reload through launchctl without dropping back to a terminal.",
    },
  ] satisfies StepItem[],
  features: [
    {
      index: "01",
      title: "Six time-aware statuses, not three.",
      body: "running (PID > 0), scheduled (waiting on StartInterval or StartCalendarInterval), completed (recent clean exit), stopped (clean exit, no schedule), error (non-zero last exit), offline (plist present but launchctl list doesn't know it).",
      accent: "brand",
    },
    {
      index: "02",
      title: "Six diagnostic checks per job.",
      body: "Exit Code (decoded — e.g. 127 = command not found), Program Exists, Program Executable, Plist Owner, Plist Permissions, Log Path. All read-only — Launch Pilot never edits your plists.",
      accent: "warning",
    },
    {
      index: "03",
      title: "Real-time over SSE, not polling.",
      body: "GET /api/events pushes the full job list every 5 seconds. The frontend re-renders only what changed, using Preact Signals computed() — under 1ms for 300 jobs.",
      accent: "accent",
    },
    {
      index: "04",
      title: "Next-run heuristics.",
      body: "nextRunAt is derived from StartCalendarInterval / StartInterval using only the standard library time package. lastRunAt is the newer of stdout / stderr mtimes. Status badges show both on hover.",
    },
    {
      index: "05",
      title: "Filter four dimensions at once.",
      body: "Category chips (All / Mine / System / 3rd-party) × status tabs × Only Mine toggle × label substring search. Filtering runs entirely in the browser on the signals graph.",
    },
    {
      index: "06",
      title: "Zero cloud, zero telemetry, zero deps.",
      body: "Binds to 127.0.0.1. No analytics SDK, no crash reporter, no phone-home update check — updates arrive via Homebrew. The Go binary has no third-party deps for next-fire calculation either.",
    },
  ] satisfies FeatureItem[],
  logStream: [
    { ts: "14:02:11.402", level: "info", service: "com.tailscale.tailscaled", body: "accept conn 100.64.0.1:33214" },
    { ts: "14:02:11.889", level: "stream", service: "homebrew.mxcl.postgresql@16", body: "checkpoint complete: wrote 42 buffers (0.3%)" },
    { ts: "14:02:12.014", level: "warn", service: "com.user.nightly-backup", body: "rsync exit 24 (vanished source files) — KeepAlive triggers respawn" },
    { ts: "14:02:12.311", level: "info", service: "com.docker.vmnetd", body: "proxy: bound 127.0.0.1:6443 → gvproxy" },
    { ts: "14:02:12.744", level: "error", service: "com.user.nightly-backup", body: "respawned too quickly: 4 times in 12s → throttled to 10s" },
    { ts: "14:02:13.210", level: "info", service: "homebrew.mxcl.postgresql@16", body: "autovacuum: completed in 2.1s" },
    { ts: "14:02:13.558", level: "stream", service: "com.tailscale.tailscaled", body: "magicsock: derp-2 active · rtt 18ms" },
    { ts: "14:02:13.902", level: "info", service: "com.user.nightly-backup", body: "retry 1/3 — rsync resumed at offset 0x2fa00" },
  ] satisfies LogLine[],
  stream: {
    eyebrow: "Live stream",
    headline: "Logs land in the browser as launchd writes them.",
    body: "GET /api/events?label=<job> is a Server-Sent Events stream of the job's stdout and stderr, served straight from the file descriptors launchctl wrote them to. Backed by a 10,000-line tail buffer per file.",
    badges: ["Server-Sent Events", "Per-job tail", "Up to 10,000 lines", "Live filter"],
    endpoint: "GET /api/events",
    endpointStatus: "streaming",
  },
  demo: {
    eyebrow: "Install · seconds",
    title: "From brew install to a live console.",
    body: "One Homebrew formula. One binary. One browser tab. No daemon-for-the-daemons, no privileges beyond what launchctl already has.",
    session: [
      {
        prompt: "brew install RoboZephyr/tap/launch-pilot",
        stdout: "==> Tapping RoboZephyr/tap\n==> Fetching RoboZephyr/tap/launch-pilot\n==> Pouring launch-pilot--0.3.1.arm64.bottle.tar.gz\n  /opt/homebrew/Cellar/launch-pilot/0.3.1: 3 files, 8.6MB",
      },
      {
        prompt: "launch-pilot",
        stdout: "Launch Pilot running at http://127.0.0.1:54231\n(opening in default browser)",
      },
    ],
  },
  stats: [
    { value: "6", label: "time-aware status values", caption: "running · scheduled · completed · stopped · error · offline" },
    { value: "6", label: "diagnostic checks per job", caption: "exit code · program · plist · log path" },
    { value: "5s", label: "SSE push interval", caption: "GET /api/events, full job list each tick" },
    { value: "127.0.0.1", label: "the only address it binds to", caption: "no cloud, no telemetry, no update ping" },
  ] satisfies Stat[],
  faqs: [
    {
      question: "Does Launch Pilot need root or a kernel extension?",
      answer:
        "No kernel extensions, no setuid binary. The privilege model mirrors launchctl: jobs in your user domain (~/Library/LaunchAgents) can be started, stopped, and reloaded. System-level daemons appear in a read-only overview.",
    },
    {
      question: "Does any data leave my machine?",
      answer:
        "No. The HTTP server binds to 127.0.0.1 only — there's no public listener. No analytics SDK, no crash reporter, no phone-home update check. Updates come through Homebrew on your schedule.",
    },
    {
      question: "Does it work on Intel Macs?",
      answer:
        "Yes. Every release ships both arm64 and amd64 .tar.gz artifacts on GitHub, and the Homebrew tap formula resolves the right one automatically.",
    },
    {
      question: "Can I script against it?",
      answer:
        "Yes. The same surface the UI consumes is documented in the README: GET /api/jobs, GET /api/jobs/{label}, POST /api/jobs/{label}/start | stop | reload, GET /api/jobs/{label}/logs?lines=200, GET /api/jobs/{label}/diagnose, and the GET /api/events SSE stream.",
    },
    {
      question: "How does it compare to LaunchControl or lingon?",
      answer:
        "Those are paid plist editors. Launch Pilot is free, open-source under MIT, and focused on the observability side — six time-aware statuses, per-job log tail, six diagnostic checks. Editing plists is not a feature; the tool is strictly read + control (start / stop / reload).",
    },
    {
      question: "Can it edit my plist files?",
      answer:
        "Deliberately not. Launch Pilot is read-only on plist contents. It can ask launchctl to start, stop, or reload a label — but it will not rewrite a file on disk. If you want a plist editor, LaunchControl and lingon are the established options.",
    },
  ] satisfies FaqItem[],
  cta: {
    eyebrow: "Install",
    headline: "Open the hood on launchd.",
    body: "One brew install, one binary, one browser tab. The next daemon that dies at 3am will announce itself in the morning, with its exit code already decoded.",
    command: "brew install RoboZephyr/tap/launch-pilot",
  },
  footer: {
    copyline: "© 2026 Launch Pilot — single Go binary, MIT licensed",
    links: [
      { label: "GitHub", href: "https://github.com/RoboZephyr/launch-pilot" },
      { label: "Changelog", href: "https://github.com/RoboZephyr/launch-pilot/blob/main/CHANGELOG.md" },
      { label: "Homebrew tap", href: "https://github.com/RoboZephyr/homebrew-tap" },
      { label: "launchd(8)", href: "https://keith.github.io/xcode-man-pages/launchd.8.html" },
    ],
  },
} as const;

export type Content = typeof content;
