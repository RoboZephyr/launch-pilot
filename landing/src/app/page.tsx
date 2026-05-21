"use client";

import { useState } from "react";
import { motion } from "framer-motion";
import { Card, StatusDot } from "@/kit";
import { colors, type ServiceStatus } from "@/kit/tokens";
import { cn } from "@/kit/utils/cn";
import { content } from "@/content";

const fadeUp = {
  hidden: { opacity: 0, y: 16 },
  show: { opacity: 1, y: 0, transition: { duration: 0.45, ease: [0.22, 1, 0.36, 1] as const } },
};

const STATUS_DEFS: Array<{ status: ServiceStatus; rule: string }> = [
  { status: "running", rule: "PID > 0" },
  { status: "scheduled", rule: "PID = 0, clean exit, schedule defined, no recent log" },
  { status: "completed", rule: "PID = 0, clean exit, log mtime inside --recent-window" },
  { status: "stopped", rule: "PID = 0, clean exit, no schedule, no recent log" },
  { status: "error", rule: "Last exit status non-zero" },
  { status: "offline", rule: "plist file present, launchctl list does not return label" },
];

const DIAGNOSTIC_CHECKS = [
  { name: "Exit Code", body: "Decodes the last exit status (e.g. 127 = command not found, 139 = segfault)." },
  { name: "Program Exists", body: "The executable path in the plist actually exists on disk." },
  { name: "Program Executable", body: "The file has the execute bit set for the loading user." },
  { name: "Plist Owner", body: "The plist is owned by the current user (no cross-user surprises)." },
  { name: "Plist Permissions", body: "No group / world write bits — launchd will silently refuse otherwise." },
  { name: "Log Path", body: "Parent directories for stdout / stderr log paths exist and are writable." },
];

const API_ENDPOINTS = [
  { method: "GET", path: "/api/jobs", note: "List every job, including offline plists" },
  { method: "GET", path: "/api/jobs/{label}", note: "Single-job snapshot" },
  { method: "POST", path: "/api/jobs/{label}/start", note: "launchctl kickstart" },
  { method: "POST", path: "/api/jobs/{label}/stop", note: "launchctl kill SIGTERM" },
  { method: "POST", path: "/api/jobs/{label}/reload", note: "launchctl bootout + bootstrap" },
  { method: "GET", path: "/api/jobs/{label}/logs?lines=200", note: "Tail stdout + stderr (max 10,000)" },
  { method: "GET", path: "/api/jobs/{label}/diagnose", note: "Run the six health checks" },
  { method: "GET", path: "/api/events", note: "SSE stream — full job list pushed every 5s" },
];

export default function HomePage() {
  return (
    <>
      <a
        href="#main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:rounded focus:bg-[#22D3EE] focus:px-3 focus:py-1.5 focus:text-[#05070D] focus:text-sm"
      >
        Skip to content
      </a>
      <SiteHeader />
      <main id="main" className="relative min-h-dvh bg-[#05070D] text-[#F4F6FB] antialiased">
        <GridBackdrop />
        <Hero />
        <StatusReference />
        <DiagnosticChecks />
        <ApiSurface />
        <InstallBlock />
        <SiteFooter />
      </main>
    </>
  );
}

function SiteHeader() {
  return (
    <header className="sticky top-0 z-40 border-b border-white/[0.05] bg-[#05070D]/85 backdrop-blur-md">
      <div className="mx-auto flex h-14 w-full max-w-[1180px] items-center justify-between px-6">
        <a href="/" className="flex items-baseline gap-2 transition-opacity hover:opacity-80">
          <span aria-hidden className="text-[14px] text-[#67E8F9]">▸</span>
          <span className="text-[14px] font-medium tracking-tight">{content.brand.name}</span>
          <span className="ml-1 font-mono text-[11px] text-white/35">{content.brand.version}</span>
        </a>
        <nav aria-label="Primary" className="hidden items-center gap-6 text-[13px] text-white/60 md:flex">
          {content.nav.map((l) => (
            <a key={l.href} href={l.href} className="transition-colors hover:text-white">
              {l.label}
            </a>
          ))}
          <a
            href={content.brand.githubUrl}
            target="_blank"
            rel="noreferrer"
            className="transition-colors hover:text-white"
          >
            GitHub
          </a>
        </nav>
      </div>
    </header>
  );
}

function GridBackdrop() {
  return (
    <div
      aria-hidden
      className="pointer-events-none absolute inset-x-0 top-0 -z-0 h-[900px]"
      style={{
        backgroundImage:
          "linear-gradient(to right, rgba(255,255,255,0.04) 1px, transparent 1px), linear-gradient(to bottom, rgba(255,255,255,0.04) 1px, transparent 1px)",
        backgroundSize: "48px 48px",
        maskImage: "radial-gradient(ellipse 70% 55% at 50% 0%, #000 30%, transparent 80%)",
      }}
    />
  );
}

function Hero() {
  return (
    <section className="relative z-10 mx-auto w-full max-w-[1180px] px-6 pb-24 pt-20 lg:pt-28">
      <div className="grid grid-cols-1 gap-12 lg:grid-cols-12 lg:gap-10">
        <motion.div
          variants={fadeUp}
          initial="hidden"
          animate="show"
          className="lg:col-span-6"
        >
          <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-white/45">
            {content.hero.eyebrow}
          </p>
          <h1 className="mt-5 text-[clamp(2.25rem,4.4vw,3.5rem)] font-semibold leading-[1.05] tracking-[-0.02em]">
            Every launchd job on your Mac,
            <br />
            <span className="text-[#67E8F9]">in one browser tab.</span>
          </h1>
          <p className="mt-6 max-w-[58ch] text-[15.5px] leading-[1.65] text-white/65">
            {content.hero.body}
          </p>

          <div className="mt-8 max-w-[460px]">
            <InstallSnippet command={content.brand.installCommand} />
          </div>

          <div className="mt-6 flex flex-wrap items-center gap-x-5 gap-y-2 text-[12.5px] text-white/55">
            {content.hero.metaRow.map((m) => (
              <span key={m.key} className="flex items-center gap-2">
                <span className="h-1 w-1 rounded-full bg-white/30" />
                <span className="font-mono text-white/40">{m.key}</span>
                <span>{m.value}</span>
              </span>
            ))}
          </div>
        </motion.div>

        <motion.div
          variants={fadeUp}
          initial="hidden"
          animate="show"
          transition={{ delay: 0.08 }}
          className="lg:col-span-6"
        >
          <JobConsole />
        </motion.div>
      </div>
    </section>
  );
}

function InstallSnippet({ command }: { command: string }) {
  const [copied, setCopied] = useState(false);
  const copy = () => {
    if (typeof navigator === "undefined") return;
    navigator.clipboard?.writeText(command).catch(() => {});
    setCopied(true);
    setTimeout(() => setCopied(false), 1400);
  };
  return (
    <div className="flex items-center gap-3 rounded-lg border border-white/[0.08] bg-[#03040A] px-4 py-3 font-mono text-[13px]">
      <span className="text-white/30">$</span>
      <span className="flex-1 truncate text-white/90">{command}</span>
      <button
        type="button"
        onClick={copy}
        className="rounded-md border border-white/10 bg-white/[0.04] px-2 py-1 text-[10px] uppercase tracking-[0.18em] text-white/65 transition-colors hover:border-[#22D3EE]/50 hover:text-[#67E8F9]"
      >
        {copied ? "copied" : "copy"}
      </button>
    </div>
  );
}

function JobConsole() {
  const services = content.livePanel.services;
  const counts = services.reduce<Record<string, number>>((acc, s) => {
    acc[s.status] = (acc[s.status] || 0) + 1;
    return acc;
  }, {});

  return (
    <Card tone="raised" className="overflow-hidden border-white/[0.08] p-0">
      <header className="flex items-center justify-between border-b border-white/[0.06] px-4 py-2.5">
        <div className="flex items-center gap-2.5">
          <span className="flex gap-1">
            <span className="h-2 w-2 rounded-full bg-[#FF5F57]" />
            <span className="h-2 w-2 rounded-full bg-[#FEBC2E]" />
            <span className="h-2 w-2 rounded-full bg-[#28C840]" />
          </span>
          <span className="font-mono text-[11.5px] text-white/55">
            127.0.0.1:54231 · Launch Pilot
          </span>
        </div>
        <span className="flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-[0.2em] text-[#34D399]">
          <span className="relative inline-flex h-1.5 w-1.5">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-[#34D399] opacity-70" />
            <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-[#34D399]" />
          </span>
          SSE · live
        </span>
      </header>

      <div className="flex items-center gap-1.5 border-b border-white/[0.05] px-4 py-2 font-mono text-[10.5px] uppercase tracking-[0.16em]">
        <FilterChip active>All ({services.length})</FilterChip>
        <FilterChip>Mine</FilterChip>
        <FilterChip>System</FilterChip>
        <FilterChip>3rd-party</FilterChip>
        <span className="ml-auto flex items-center gap-1.5 text-white/35">
          <span aria-hidden>⌕</span>
          <span className="normal-case tracking-normal text-[11px]">search</span>
        </span>
      </div>

      <div className="flex flex-wrap items-center gap-x-4 gap-y-1.5 border-b border-white/[0.05] bg-white/[0.015] px-4 py-2 font-mono text-[11px] tracking-tight">
        {STATUS_DEFS.filter(({ status }) => counts[status]).map(({ status }) => (
          <span key={status} className="inline-flex items-center gap-1.5 text-white/75">
            <span
              className="h-1.5 w-1.5 shrink-0 rounded-full"
              style={{ backgroundColor: colors.status[status].fg }}
            />
            <span style={{ color: colors.status[status].fg }}>{colors.status[status].label}</span>
            <span className="tabular-nums text-white/50">{counts[status]}</span>
          </span>
        ))}
      </div>

      <ul className="divide-y divide-white/[0.04]">
        {services.map((svc) => {
          const meta = colors.status[svc.status];
          return (
            <li
              key={svc.name}
              className="flex items-center gap-3 px-4 py-2.5 transition-colors hover:bg-white/[0.02]"
            >
              <StatusDot status={svc.status} pulse={svc.status === "running"} />
              <span className="flex-1 truncate font-mono text-[12.5px] text-white/85">
                {svc.name}
              </span>
              <span
                className="font-mono text-[11px] tabular-nums"
                style={{ color: meta.fg }}
              >
                {svc.pid ? `pid ${svc.pid}` : meta.label.toLowerCase()}
              </span>
              <span className="w-[120px] text-right font-mono text-[10.5px] tabular-nums text-white/40">
                {svc.uptime}
              </span>
            </li>
          );
        })}
      </ul>

      <footer className="flex items-center justify-between border-t border-white/[0.06] bg-white/[0.015] px-4 py-2 font-mono text-[10px] uppercase tracking-[0.18em] text-white/40">
        <span>GET /api/events</span>
        <span>text/event-stream</span>
        <span>5s push</span>
      </footer>
    </Card>
  );
}

function FilterChip({
  active,
  children,
}: {
  active?: boolean;
  children: React.ReactNode;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-md border px-2 py-0.5",
        active
          ? "border-[#22D3EE]/40 bg-[#22D3EE]/10 text-[#67E8F9]"
          : "border-white/8 bg-white/[0.02] text-white/55",
      )}
    >
      {children}
    </span>
  );
}

function StatusReference() {
  return (
    <section id="features" className="relative z-10 mx-auto w-full max-w-[1180px] px-6 py-24">
      <SectionHeader
        eyebrow="Statuses"
        title="Six values, derived from the raw launchctl table."
        body="Most tools collapse to running / stopped / error. Launch Pilot keeps the distinctions launchctl actually exposes — so you can tell a job that's waiting for its schedule from a job that's missing in action."
      />
      <div className="mt-12 grid grid-cols-1 gap-px overflow-hidden rounded-xl border border-white/[0.07] bg-white/[0.04] sm:grid-cols-2 lg:grid-cols-3">
        {STATUS_DEFS.map(({ status, rule }) => {
          const meta = colors.status[status];
          return (
            <div key={status} className="bg-[#05070D] p-6">
              <div className="flex items-center gap-2.5">
                <StatusDot status={status} />
                <span
                  className="font-mono text-[12px] uppercase tracking-[0.18em]"
                  style={{ color: meta.fg }}
                >
                  {meta.label}
                </span>
              </div>
              <p className="mt-4 font-mono text-[12px] leading-[1.6] text-white/70">
                {rule}
              </p>
            </div>
          );
        })}
      </div>
    </section>
  );
}

function DiagnosticChecks() {
  return (
    <section className="relative z-10 mx-auto w-full max-w-[1180px] px-6 py-24">
      <SectionHeader
        eyebrow="Diagnostics"
        title="Six read-only checks per job."
        body="When a job is failing, the diagnose endpoint walks the most common breakage modes and gives you a typed answer — not a wall of stderr."
      />
      <div className="mt-12 overflow-hidden rounded-xl border border-white/[0.07]">
        {DIAGNOSTIC_CHECKS.map((c, i) => (
          <div
            key={c.name}
            className={cn(
              "grid grid-cols-1 gap-1 px-6 py-4 md:grid-cols-[180px_1fr] md:items-baseline md:gap-6",
              i > 0 && "border-t border-white/[0.06]",
            )}
          >
            <span className="font-mono text-[12.5px] tracking-tight text-white/85">
              {c.name}
            </span>
            <span className="text-[14px] leading-[1.65] text-white/60">{c.body}</span>
          </div>
        ))}
      </div>
    </section>
  );
}

function ApiSurface() {
  return (
    <section id="how" className="relative z-10 mx-auto w-full max-w-[1180px] px-6 py-24">
      <SectionHeader
        eyebrow="API"
        title="Everything the UI uses, also available to your scripts."
        body="The web UI is one consumer of /api/jobs and /api/events. So is curl. Bind the server, point at the endpoints, parse the JSON."
      />
      <div className="mt-12 overflow-hidden rounded-xl border border-white/[0.07] bg-[#03040A]">
        <div className="grid grid-cols-[64px_1fr_1.4fr] gap-0 border-b border-white/[0.06] bg-white/[0.025] px-6 py-2.5 font-mono text-[10.5px] uppercase tracking-[0.16em] text-white/40">
          <span>verb</span>
          <span>path</span>
          <span>description</span>
        </div>
        {API_ENDPOINTS.map((e, i) => (
          <div
            key={e.path + e.method}
            className={cn(
              "grid grid-cols-[64px_1fr_1.4fr] gap-0 px-6 py-3 items-baseline",
              i > 0 && "border-t border-white/[0.04]",
            )}
          >
            <span
              className={cn(
                "font-mono text-[11px] uppercase tracking-[0.14em]",
                e.method === "GET" ? "text-[#67E8F9]" : "text-[#A78BFA]",
              )}
            >
              {e.method}
            </span>
            <span className="font-mono text-[12.5px] text-white/85">{e.path}</span>
            <span className="text-[13px] text-white/55">{e.note}</span>
          </div>
        ))}
      </div>
    </section>
  );
}

function InstallBlock() {
  return (
    <section id="install" className="relative z-10 mx-auto w-full max-w-[1180px] px-6 py-24">
      <SectionHeader
        eyebrow="Install"
        title="One brew install. One binary. One tab."
        body="The Homebrew tap is public; the formula is auditable. The binary is a single static Go executable with the Preact frontend embedded via go:embed — nothing else to install or serve."
      />
      <div className="mt-12 grid grid-cols-1 gap-4 lg:grid-cols-3">
        <Card className="border-white/[0.07] bg-white/[0.02] p-6">
          <span className="font-mono text-[10.5px] uppercase tracking-[0.18em] text-white/40">
            01 · install
          </span>
          <pre className="mt-5 overflow-x-auto rounded-md border border-white/[0.07] bg-[#03040A] px-3 py-2.5 font-mono text-[12px] text-[#67E8F9]">
            <span className="text-white/35">$ </span>brew install RoboZephyr/tap/launch-pilot
          </pre>
          <p className="mt-4 text-[13px] leading-[1.65] text-white/55">
            Adds the public tap and pours the formula. Both arm64 and amd64 are
            built per release.
          </p>
        </Card>
        <Card className="border-white/[0.07] bg-white/[0.02] p-6">
          <span className="font-mono text-[10.5px] uppercase tracking-[0.18em] text-white/40">
            02 · run
          </span>
          <pre className="mt-5 overflow-x-auto rounded-md border border-white/[0.07] bg-[#03040A] px-3 py-2.5 font-mono text-[12px] text-[#67E8F9]">
            <span className="text-white/35">$ </span>launch-pilot
          </pre>
          <p className="mt-4 text-[13px] leading-[1.65] text-white/55">
            Picks a free port on 127.0.0.1 and opens your default browser. Pin
            with <code className="font-mono text-white/70">--port</code>, skip
            the browser with <code className="font-mono text-white/70">--no-open</code>.
          </p>
        </Card>
        <Card tone="glow" className="p-6">
          <span className="font-mono text-[10.5px] uppercase tracking-[0.18em] text-[#67E8F9]">
            03 · investigate
          </span>
          <p className="mt-5 text-[13px] leading-[1.65] text-white/70">
            Filter by category (Mine / System / 3rd-party), drill into a job to
            see status reasoning, tail logs up to 10,000 lines, run the six
            diagnostic checks, and start / stop / reload via launchctl — no
            terminal context-switch.
          </p>
        </Card>
      </div>
    </section>
  );
}

function SectionHeader({
  eyebrow,
  title,
  body,
}: {
  eyebrow: string;
  title: string;
  body: string;
}) {
  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
      <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-white/40 lg:col-span-2 lg:pt-2">
        {eyebrow}
      </p>
      <div className="lg:col-span-10">
        <h2 className="max-w-[26ch] text-[clamp(1.75rem,3vw,2.5rem)] font-semibold leading-[1.1] tracking-[-0.02em]">
          {title}
        </h2>
        <p className="mt-5 max-w-[68ch] text-[15.5px] leading-[1.65] text-white/60">
          {body}
        </p>
      </div>
    </div>
  );
}

function SiteFooter() {
  return (
    <footer className="relative z-10 border-t border-white/[0.06]">
      <div className="mx-auto flex w-full max-w-[1180px] flex-col items-start justify-between gap-6 px-6 py-10 md:flex-row md:items-center">
        <div className="flex items-baseline gap-2">
          <span aria-hidden className="text-[13px] text-[#67E8F9]">▸</span>
          <span className="font-mono text-[12px] text-white/40">
            {content.footer.copyline}
          </span>
        </div>
        <ul className="flex flex-wrap gap-x-6 gap-y-2 text-[13px] text-white/55">
          {content.footer.links.map((l) => (
            <li key={l.href}>
              <a href={l.href} target="_blank" rel="noreferrer" className="transition-colors hover:text-white">
                {l.label}
              </a>
            </li>
          ))}
        </ul>
      </div>
    </footer>
  );
}

