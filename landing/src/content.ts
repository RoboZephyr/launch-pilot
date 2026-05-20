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
    title: "Launch Pilot — the visual console for macOS launchd",
    description:
      "Launch Pilot is a local-first control console for macOS launchd. Brew install it, watch every LaunchAgent, tail logs over SSE, diagnose the daemon that keeps crashing while you sleep.",
  },
  brand: {
    name: "Launch Pilot",
    mark: "LP",
    tagline: "launchd, visualized",
    installCommand: "brew install launch-pilot/tap/launch-pilot",
    launchCommand: "launch-pilot open",
    github: "github.com/launch-pilot/launch-pilot",
    version: "v0.3.1",
  },
  nav: [
    { label: "Why", href: "#why" },
    { label: "How it works", href: "#how" },
    { label: "Features", href: "#features" },
    { label: "Install", href: "#install" },
  ],
  hero: {
    kicker: "For macOS developers who inherited a plist",
    headline: ["Your daemons", "have been running", "in the dark."],
    body: "Launch Pilot is a local-first control console for macOS launchd. Every LaunchAgent, every restart loop, every log line — streaming to a browser tab over SSE, running on your laptop, touching no cloud, asking no permissions you haven't already given launchctl.",
    primaryCta: { label: "brew install launch-pilot", href: "#install" },
    secondaryCta: { label: "See a service recover", href: "#demo" },
    metaRow: [
      { key: "runtime", value: "Local · Go + SSE" },
      { key: "scope", value: "per-user agents · system daemons" },
      { key: "install time", value: "< 4s via Homebrew" },
    ],
  },
  livePanel: {
    title: "/launchctl/status",
    subtitle: "live · streaming over /api/events",
    sseHeartbeat: "event: heartbeat · 1 Hz",
    services: [
      { name: "com.docker.vmnetd", status: "running", pid: 849, uptime: "3d 12h 04m" },
      { name: "homebrew.mxcl.postgresql@16", status: "running", pid: 3611, uptime: "1d 02h" },
      { name: "com.tailscale.tailscaled", status: "running", pid: 1204, uptime: "14h 03m" },
      { name: "com.user.nightly-backup", status: "warning", pid: 9321, uptime: "restarted ×3" },
      { name: "com.zsa.wally.daemon", status: "failed", pid: null, uptime: "crashed · exit 139" },
      { name: "com.vercel.cli.updater", status: "idle", pid: null, uptime: "next run · 03:00" },
      { name: "homebrew.autoupdate", status: "loaded", pid: null, uptime: "loaded, not run" },
    ] satisfies ServiceRow[],
  },
  why: {
    eyebrow: "Why",
    headline: "launchctl is a power tool with no dashboard.",
    paragraphs: [
      "Three hundred services run on a fresh macOS install. You inherited thirty more when you said yes to Docker, Homebrew, Rectangle, and that Python installer from 2021.",
      "When one fails, you read plist XML, grep Console.app, and pray /var/log/system.log still has the window. Launch Pilot replaces that ritual with a real-time console — one tab, every agent, full log tail, zero kernel extensions.",
    ],
    contrast: {
      before: {
        label: "Without Launch Pilot",
        bullets: [
          "launchctl list | grep | head, every time",
          "plist XML open in BBEdit, fingers crossed",
          "Console.app with ten filters you rebuilt last week",
          "No history when a daemon crashes at 3am",
        ],
      },
      after: {
        label: "With Launch Pilot",
        bullets: [
          "One browser tab, every service, sorted by health",
          "Plist rendered as a form, not a tag soup",
          "Per-service log tail, filtered server-side, streamed live",
          "A circular restart buffer that remembers the 3am crash",
        ],
      },
    },
  },
  steps: [
    {
      index: "01",
      title: "Install from Homebrew.",
      body: "A single Go binary, no Electron, no background services it didn't ship with. The formula is public and auditable — read the bottle, read the source.",
      command: "brew install launch-pilot/tap/launch-pilot",
    },
    {
      index: "02",
      title: "Run launch-pilot open.",
      body: "The CLI boots a local web server on 127.0.0.1, opens a browser tab, and subscribes to launchctl events through the native bootstrap APIs. Nothing leaves your machine.",
      command: "launch-pilot open --port 7331",
    },
    {
      index: "03",
      title: "Watch, diagnose, fix.",
      body: "Every service is a row. Green means healthy, amber means thrashing, red means bring-it-back-up. Click a row for tail logs, exit codes, the plist that defined it, and a one-click restart button.",
    },
  ] satisfies StepItem[],
  features: [
    {
      index: "01",
      title: "Real-time SSE, not polling.",
      body: "A launchctl subscription fans out over Server-Sent Events. Status changes land in the browser in under 50ms — faster than you can alt-tab to a terminal.",
      accent: "brand",
    },
    {
      index: "02",
      title: "Crash-loop detective.",
      body: "Exit codes, signal names, and a rolling buffer of the last 500 log lines per service. When something dies at 3am, the evidence is waiting when you wake up.",
      accent: "warning",
    },
    {
      index: "03",
      title: "Plist, rendered sanely.",
      body: "LaunchAgent plists are XML pretending to be dictionaries. Launch Pilot parses them into typed forms — ProgramArguments, KeepAlive, StartInterval — with inline validation and diff-on-save.",
      accent: "accent",
    },
    {
      index: "04",
      title: "Log tail that survives restart.",
      body: "Standard out, standard error, and launchd's own emission — merged, timestamp-aligned, searchable. Filter by level without losing context.",
    },
    {
      index: "05",
      title: "Zero cloud, zero telemetry.",
      body: "Everything runs on 127.0.0.1. No analytics, no crash reporter, no call-home update check. You'd hear about a new release from Homebrew, like every other local tool.",
    },
    {
      index: "06",
      title: "Keyboard-first, everywhere.",
      body: "⌘K opens command palette. j/k moves between services. r restarts. l opens logs. Esc closes. Designed for someone who types faster than they click.",
    },
  ] satisfies FeatureItem[],
  logStream: [
    { ts: "14:02:11.402", level: "info", service: "com.tailscale.tailscaled", body: "accepted inbound conn from 100.64.0.1" },
    { ts: "14:02:11.889", level: "stream", service: "homebrew.mxcl.postgresql@16", body: "autovacuum: completed on idea_business.public.events (2.1s)" },
    { ts: "14:02:12.014", level: "warn", service: "com.user.nightly-backup", body: "rsync exited 24 (vanished source files) — scheduling retry" },
    { ts: "14:02:12.311", level: "info", service: "com.docker.vmnetd", body: "proxy: bound 127.0.0.1:6443 → gvproxy" },
    { ts: "14:02:12.744", level: "error", service: "com.zsa.wally.daemon", body: "exit status 139 (segmentation fault) after 842ms — KeepAlive triggers respawn" },
    { ts: "14:02:13.001", level: "warn", service: "com.zsa.wally.daemon", body: "respawned too quickly: 4 times in 12s, throttling to 10s" },
    { ts: "14:02:13.210", level: "info", service: "homebrew.mxcl.postgresql@16", body: "checkpoint complete, wal=47MB kept" },
    { ts: "14:02:13.558", level: "stream", service: "com.tailscale.tailscaled", body: "magicsock: derp-2 active · rtt 18ms" },
    { ts: "14:02:13.902", level: "info", service: "com.user.nightly-backup", body: "retry 1/3 — rsync resumed from offset 0x2fa00" },
  ] satisfies LogLine[],
  demo: {
    eyebrow: "Install · 4 seconds",
    title: "From zero to a live daemon console.",
    body: "One tap, one binary, one browser tab. No daemon-for-the-daemons, no elevated privileges beyond what launchctl already has.",
    session: [
      { prompt: "brew install launch-pilot/tap/launch-pilot", stdout: "==> Fetching launch-pilot\n==> Pouring launch-pilot-0.3.1.arm64_sonoma.bottle.tar.gz\n  🚀  /opt/homebrew/Cellar/launch-pilot/0.3.1: 11 files, 8.9MB" },
      { prompt: "launch-pilot open", stdout: "▶ launch-pilot v0.3.1\n✓ launchctl subscription (313 services indexed)\n✓ SSE bus on 127.0.0.1:7331\n✓ opening http://127.0.0.1:7331/ in your browser" },
    ],
  },
  stats: [
    { value: "313", label: "services indexed on a fresh macOS 15 install", caption: "on this M2 MacBook Air, right now" },
    { value: "< 50ms", label: "SSE round-trip from launchctl event to DOM update", caption: "95th percentile, local loopback" },
    { value: "0", label: "System Preference panes opened to diagnose a crash", caption: "this has been a System Preference-free quarter" },
    { value: "8.9MB", label: "single static Go binary, arm64 + x86_64", caption: "smaller than the average launch-screen video" },
  ] satisfies Stat[],
  faqs: [
    {
      question: "Does this need root or a kernel extension?",
      answer:
        "No kernel extensions. System-level daemons appear in a read-only overview; agent-level services in your own user domain can be started, stopped, and restarted. The privilege model mirrors launchctl — if launchctl can do it, so can Launch Pilot.",
    },
    {
      question: "Is any data sent off my machine?",
      answer:
        "Nothing. The binary binds only to 127.0.0.1, there is no analytics SDK, no crash reporter, no phone-home update check. Updates come through Homebrew on your schedule.",
    },
    {
      question: "Does it work on Intel Macs?",
      answer:
        "Yes. The Homebrew bottle ships both arm64 and x86_64 slices, built against macOS 12+ SDK. Sonoma and Sequoia are the primary targets; Ventura works and is tested in CI.",
    },
    {
      question: "Can I script against it?",
      answer:
        "Yes. The same SSE stream the UI consumes is available at /api/events, and there is a JSON REST surface under /api/services for list/start/stop/restart. Use it with curl, Bruno, or whatever you script with.",
    },
    {
      question: "How does this compare to LaunchControl or lingon?",
      answer:
        "Those are paid macOS apps for editing plists. Launch Pilot is a free, open, local web console focused on the observability side — real-time status, logs, crash-loop detection — with plist editing as a secondary surface, not the headline.",
    },
  ] satisfies FaqItem[],
  cta: {
    eyebrow: "Install",
    headline: "Take the hood off launchd.",
    body: "Four seconds of Homebrew and one browser tab. That's the whole setup. The next daemon that dies will announce itself.",
    command: "brew install launch-pilot/tap/launch-pilot",
  },
  footer: {
    copyline: "© 2026 Launch Pilot — a single binary, shipped under MIT",
    links: [
      { label: "GitHub", href: "https://github.com/launch-pilot/launch-pilot" },
      { label: "Changelog", href: "/changelog" },
      { label: "Homebrew tap", href: "https://github.com/launch-pilot/homebrew-tap" },
      { label: "launchd(8)", href: "https://keith.github.io/xcode-man-pages/launchd.8.html" },
    ],
  },
} as const;

export type Content = typeof content;
