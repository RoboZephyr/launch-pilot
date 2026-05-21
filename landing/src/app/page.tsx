"use client";

import {
  motion,
  useReducedMotion,
  useScroll,
  useTransform,
  type Variants,
} from "framer-motion";
import { useEffect, useMemo, useRef, useState } from "react";
import { Badge, Card, StatusDot, Terminal } from "@/kit";
import { colors, type ServiceStatus } from "@/kit/tokens";
import { cn } from "@/kit/utils/cn";
import { content } from "@/content";

const SECTION_EASE = [0.22, 1, 0.36, 1] as const;

const fadeUp: Variants = {
  hidden: { opacity: 0, y: 28 },
  show: { opacity: 1, y: 0, transition: { duration: 0.8, ease: SECTION_EASE } },
};

const heroStagger: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.12, delayChildren: 0.05 } },
};

const heroItem: Variants = {
  hidden: { opacity: 0, y: 36 },
  show: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.85, ease: SECTION_EASE },
  },
};

const logLevelStyle: Record<"info" | "warn" | "error" | "stream", string> = {
  info: "text-[#67E8F9]",
  warn: "text-[#FBBF24]",
  error: "text-[#F87171]",
  stream: "text-[#A78BFA]",
};

const featureAccent: Partial<
  Record<NonNullable<(typeof content.features)[number]["accent"]>, string>
> = {
  brand: "from-[#22D3EE]/30 via-[#22D3EE]/5 to-transparent",
  accent: "from-[#7C5CFF]/30 via-[#7C5CFF]/5 to-transparent",
  warning: "from-[#FBBF24]/25 via-[#FBBF24]/5 to-transparent",
};

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
      <main
        id="main"
        className="relative min-h-dvh overflow-hidden bg-[#05070D] text-[#F4F6FB]"
      >
        <BackdropGrid />
        <Hero />
        <WhySection />
        <HowSection />
        <StreamSection />
        <FeatureSection />
        <DemoSection />
        <StatsSection />
        <FaqSection />
        <CtaSection />
        <SiteFooter />
      </main>
    </>
  );
}

function SiteHeader() {
  const [copied, setCopied] = useState(false);
  const onCopy = () => {
    if (typeof navigator === "undefined") return;
    navigator.clipboard?.writeText(content.brand.installCommand).catch(() => {});
    setCopied(true);
    setTimeout(() => setCopied(false), 1600);
  };

  return (
    <header className="fixed inset-x-0 top-0 z-40 flex items-center justify-between px-6 py-4 sm:px-10">
      <div className="flex items-center gap-3">
        <span
          aria-hidden
          className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/15 bg-[#0B0F1A] font-mono text-[13px] tracking-tight text-[#67E8F9] shadow-[0_0_0_1px_rgba(34,211,238,0.22),0_8px_28px_rgba(34,211,238,0.16)]"
        >
          {content.brand.mark}
        </span>
        <span className="flex flex-col leading-tight">
          <span className="text-[13px] font-semibold tracking-tight">
            {content.brand.name}
          </span>
          <span className="font-mono text-[10px] uppercase tracking-[0.2em] text-white/45">
            {content.brand.version}
          </span>
        </span>
      </div>
      <nav
        aria-label="Primary"
        className="hidden items-center gap-7 text-[13px] text-white/55 md:flex"
      >
        {content.nav.map((link) => (
          <a
            key={link.href}
            href={link.href}
            className="transition-colors hover:text-white"
          >
            {link.label}
          </a>
        ))}
      </nav>
      <button
        type="button"
        onClick={onCopy}
        className={cn(
          "group flex items-center gap-2 rounded-full border border-white/10 bg-white/[0.04] px-3.5 py-1.5 font-mono text-[12px] text-white/80 transition-colors",
          "hover:border-[#22D3EE]/50 hover:text-[#67E8F9]",
        )}
        aria-label="Copy install command"
      >
        <span className="text-white/35">$</span>
        <span>{content.brand.installCommand}</span>
        <span className="text-[10px] uppercase tracking-[0.2em] text-white/35 group-hover:text-[#67E8F9]/70">
          {copied ? "copied" : "copy"}
        </span>
      </button>
    </header>
  );
}

function BackdropGrid() {
  const { scrollYProgress } = useScroll();
  const y = useTransform(scrollYProgress, [0, 1], ["0%", "-18%"]);
  return (
    <motion.div
      aria-hidden
      style={{ y }}
      className="pointer-events-none absolute inset-x-0 top-0 h-[1800px] -z-0"
    >
      <div
        className="absolute inset-0 opacity-[0.35]"
        style={{
          backgroundImage:
            "linear-gradient(to right, rgba(255,255,255,0.045) 1px, transparent 1px), linear-gradient(to bottom, rgba(255,255,255,0.045) 1px, transparent 1px)",
          backgroundSize: "64px 64px",
          maskImage:
            "radial-gradient(ellipse 80% 60% at 50% 18%, #000 40%, transparent 75%)",
        }}
      />
      <div
        className="absolute left-[-15%] top-[-10%] h-[720px] w-[720px] rounded-full blur-[140px] opacity-60"
        style={{ background: "radial-gradient(closest-side, rgba(34,211,238,0.28), transparent 75%)" }}
      />
      <div
        className="absolute right-[-20%] top-[30%] h-[640px] w-[640px] rounded-full blur-[160px] opacity-50"
        style={{ background: "radial-gradient(closest-side, rgba(124,92,255,0.24), transparent 75%)" }}
      />
    </motion.div>
  );
}

function Hero() {
  return (
    <section className="relative z-10 flex min-h-[110vh] flex-col justify-center px-6 pb-16 pt-40 sm:px-10 lg:pt-48">
      <motion.div
        variants={heroStagger}
        initial="hidden"
        animate="show"
        className="mx-auto grid w-full max-w-[1320px] grid-cols-1 gap-y-16 lg:grid-cols-12 lg:gap-x-12"
      >
        <div className="lg:col-span-7 xl:col-span-7">
          <motion.p
            variants={heroItem}
            className="mb-10 font-mono text-[11px] uppercase tracking-[0.28em] text-white/45"
          >
            {content.hero.eyebrow}
          </motion.p>
          <h1 className="max-w-[18ch] font-semibold tracking-[-0.04em] text-white">
            {content.hero.headline.map((line, i) => (
              <motion.span
                key={i}
                variants={heroItem}
                className={cn(
                  "block text-[clamp(2.75rem,7.6vw,7rem)] leading-[0.95]",
                  i === 1 && "text-white/85",
                  i === 2 && "text-[#67E8F9]",
                )}
              >
                {line}
              </motion.span>
            ))}
          </h1>
          <motion.p
            variants={heroItem}
            className="mt-10 max-w-[58ch] text-balance text-[17px] leading-[1.65] text-white/65 sm:text-[18px]"
          >
            {content.hero.body}
          </motion.p>
          <motion.div
            variants={heroItem}
            className="mt-12 flex flex-wrap items-center gap-x-4 gap-y-3"
          >
            <a
              href={content.hero.primaryCta.href}
              className="inline-flex h-12 items-center gap-2 rounded-lg bg-[#22D3EE] px-6 font-medium tracking-tight text-[#05070D] shadow-[0_0_0_1px_rgba(34,211,238,0.4),0_8px_32px_rgba(34,211,238,0.28)] transition-colors hover:bg-[#67E8F9]"
            >
              {content.hero.primaryCta.label}
            </a>
            <a
              href={content.hero.secondaryCta.href}
              target="_blank"
              rel="noreferrer"
              className="inline-flex h-12 items-center gap-2 rounded-lg bg-transparent px-6 font-medium tracking-tight text-white/80 transition-colors hover:bg-white/5 hover:text-white"
            >
              {content.hero.secondaryCta.label}
              <span aria-hidden className="text-white/40">→</span>
            </a>
          </motion.div>
          <motion.dl
            variants={heroItem}
            className="mt-16 grid max-w-xl grid-cols-1 gap-6 border-t border-white/[0.06] pt-8 sm:grid-cols-3"
          >
            {content.hero.metaRow.map((row) => (
              <div key={row.key}>
                <dt className="font-mono text-[10px] uppercase tracking-[0.22em] text-white/40">
                  {row.key}
                </dt>
                <dd className="mt-1.5 text-[13px] text-white/80">{row.value}</dd>
              </div>
            ))}
          </motion.dl>
        </div>

        <motion.div
          variants={heroItem}
          className="lg:col-span-5 xl:col-span-5"
        >
          <LivePanel />
        </motion.div>
      </motion.div>

      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 1.2, duration: 0.8 }}
        className="mx-auto mt-24 flex items-center gap-3 font-mono text-[11px] uppercase tracking-[0.28em] text-white/35"
      >
        <span className="h-px w-12 bg-white/15" />
        scroll to inspect
        <span className="h-px w-12 bg-white/15" />
      </motion.div>
    </section>
  );
}

function LivePanel() {
  const [clock, setClock] = useState("—");
  const reduced = useReducedMotion();
  useEffect(() => {
    const tick = () => {
      const d = new Date();
      const pad = (n: number) => n.toString().padStart(2, "0");
      setClock(
        `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`,
      );
    };
    tick();
    const id = window.setInterval(tick, 1000);
    return () => window.clearInterval(id);
  }, []);

  return (
    <Card
      tone="raised"
      className="relative overflow-hidden border-white/[0.08] p-0"
    >
      <div
        aria-hidden
        className="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-[#22D3EE]/60 to-transparent"
      />
      <header className="flex items-center justify-between border-b border-white/[0.06] px-5 py-3.5">
        <div className="flex items-center gap-3">
          <span className="flex gap-1.5">
            <span className="h-2.5 w-2.5 rounded-full bg-[#FF5F57]" />
            <span className="h-2.5 w-2.5 rounded-full bg-[#FEBC2E]" />
            <span className="h-2.5 w-2.5 rounded-full bg-[#28C840]" />
          </span>
          <span className="font-mono text-[12px] tracking-tight text-white/65">
            {content.livePanel.title}
          </span>
        </div>
        <span className="flex items-center gap-2 font-mono text-[10px] uppercase tracking-[0.22em] text-white/40">
          <span className="relative inline-flex h-1.5 w-1.5">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-[#34D399] opacity-80" />
            <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-[#34D399]" />
          </span>
          {content.livePanel.subtitle}
        </span>
      </header>
      <ul className="divide-y divide-white/[0.05]">
        {content.livePanel.services.map((svc, i) => {
          const statusMeta = colors.status[svc.status];
          return (
            <motion.li
              key={svc.name}
              initial={{ opacity: 0, x: 24 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{
                delay: 0.4 + i * 0.08,
                duration: 0.55,
                ease: SECTION_EASE,
              }}
              className="group flex items-center justify-between gap-4 px-5 py-3.5 transition-colors hover:bg-white/[0.03]"
            >
              <div className="flex min-w-0 items-center gap-3">
                <StatusDot
                  status={svc.status}
                  pulse={!reduced && (svc.status === "running" || svc.status === "scheduled")}
                />
                <span className="truncate font-mono text-[13px] text-white/85">
                  {svc.name}
                </span>
              </div>
              <div className="flex shrink-0 items-center gap-4">
                <span
                  className="font-mono text-[11px] tabular-nums"
                  style={{ color: statusMeta.fg }}
                >
                  {svc.pid ? `pid ${svc.pid}` : "—"}
                </span>
                <span className="w-[84px] text-right font-mono text-[11px] tabular-nums text-white/45">
                  {svc.uptime}
                </span>
              </div>
            </motion.li>
          );
        })}
      </ul>
      <footer className="flex items-center justify-between border-t border-white/[0.06] bg-white/[0.02] px-5 py-3">
        <span className="font-mono text-[11px] text-white/40">
          {content.livePanel.sseHeartbeat}
        </span>
        <span className="flex items-center gap-2 font-mono text-[11px] text-white/60">
          <span className="h-1 w-1 animate-pulse rounded-full bg-[#22D3EE]" />
          <span className="tabular-nums">{clock}</span>
        </span>
      </footer>
    </Card>
  );
}

function WhySection() {
  return (
    <section
      id="why"
      className="relative z-10 mx-auto grid w-full max-w-[1320px] grid-cols-1 gap-12 px-6 py-36 sm:px-10 lg:grid-cols-12"
    >
      <motion.div
        variants={fadeUp}
        initial="hidden"
        whileInView="show"
        viewport={{ once: true, amount: 0.3 }}
        className="lg:col-span-5"
      >
        <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-white/40">
          {content.why.eyebrow}
        </p>
        <h2 className="mt-6 text-[clamp(2rem,4.4vw,3.75rem)] font-semibold leading-[1.02] tracking-[-0.03em]">
          {content.why.headline}
        </h2>
        <div className="mt-10 space-y-5 text-[16px] leading-[1.7] text-white/65">
          {content.why.paragraphs.map((p, i) => (
            <p key={i}>{p}</p>
          ))}
        </div>
      </motion.div>

      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:col-span-7">
        <motion.div
          variants={fadeUp}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.2 }}
          transition={{ delay: 0.1 }}
        >
          <Card className="h-full border-white/[0.06] bg-white/[0.02] p-7">
            <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-[#F87171]/90">
              {content.why.contrast.before.label}
            </p>
            <ul className="mt-6 space-y-4">
              {content.why.contrast.before.bullets.map((b) => (
                <li key={b} className="flex gap-3 text-[15px] text-white/60">
                  <span aria-hidden className="mt-2 h-px w-4 shrink-0 bg-[#F87171]/60" />
                  <span className="line-through decoration-white/15">{b}</span>
                </li>
              ))}
            </ul>
          </Card>
        </motion.div>
        <motion.div
          variants={fadeUp}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.2 }}
          transition={{ delay: 0.2 }}
        >
          <Card
            tone="glow"
            className="h-full p-7"
          >
            <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-[#67E8F9]">
              {content.why.contrast.after.label}
            </p>
            <ul className="mt-6 space-y-4">
              {content.why.contrast.after.bullets.map((b) => (
                <li key={b} className="flex gap-3 text-[15px] text-white/85">
                  <span aria-hidden className="mt-2 h-px w-4 shrink-0 bg-[#22D3EE]" />
                  <span>{b}</span>
                </li>
              ))}
            </ul>
          </Card>
        </motion.div>
      </div>
    </section>
  );
}

function HowSection() {
  return (
    <section id="how" className="relative z-10 mx-auto w-full max-w-[1320px] px-6 py-36 sm:px-10">
      <motion.div
        variants={fadeUp}
        initial="hidden"
        whileInView="show"
        viewport={{ once: true, amount: 0.3 }}
        className="mb-20 flex items-end justify-between gap-8"
      >
        <div>
          <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-white/40">
            How
          </p>
          <h2 className="mt-6 max-w-[24ch] text-[clamp(2rem,4.2vw,3.5rem)] font-semibold leading-[1.04] tracking-[-0.03em]">
            Three steps from <span className="font-mono text-[#67E8F9]">brew install</span> to a live console.
          </h2>
        </div>
      </motion.div>

      <ol className="grid grid-cols-1 gap-10 md:grid-cols-3">
        {content.steps.map((step, i) => (
          <motion.li
            key={step.index}
            variants={fadeUp}
            initial="hidden"
            whileInView="show"
            viewport={{ once: true, amount: 0.2 }}
            transition={{ delay: 0.08 * i }}
            className="relative"
          >
            <div className="flex items-baseline gap-4">
              <span className="font-mono text-[12px] tracking-[0.2em] text-white/35">
                {step.index}
              </span>
              <span className="h-px flex-1 bg-white/10" />
            </div>
            <h3 className="mt-6 text-[22px] font-medium tracking-[-0.02em] text-white">
              {step.title}
            </h3>
            <p className="mt-4 text-[15px] leading-[1.65] text-white/60">
              {step.body}
            </p>
            {step.command && (
              <pre className="mt-6 overflow-x-auto rounded-lg border border-white/[0.06] bg-[#03040A] px-4 py-3 font-mono text-[12.5px] text-[#67E8F9]">
                <span className="text-white/35">$ </span>
                {step.command}
              </pre>
            )}
          </motion.li>
        ))}
      </ol>
    </section>
  );
}

function StreamSection() {
  const reduced = useReducedMotion();
  const loop = useMemo(
    () => [...content.logStream, ...content.logStream],
    [],
  );

  return (
    <section
      id="stream"
      className="relative z-10 mx-auto w-full max-w-[1320px] px-6 py-36 sm:px-10"
    >
      <div className="grid grid-cols-1 gap-14 lg:grid-cols-12 lg:items-center">
        <motion.div
          variants={fadeUp}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.3 }}
          className="lg:col-span-5"
        >
          <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-white/40">
            {content.stream.eyebrow}
          </p>
          <h2 className="mt-6 text-[clamp(2rem,4.2vw,3.5rem)] font-semibold leading-[1.04] tracking-[-0.03em]">
            {content.stream.headline}
          </h2>
          <p className="mt-8 max-w-[46ch] text-[16px] leading-[1.7] text-white/60">
            {content.stream.body}
          </p>
          <div className="mt-10 flex flex-wrap gap-2">
            {content.stream.badges.map((b, i) => (
              <Badge key={b} variant={i === 0 ? "brand" : i === 1 ? "accent" : "neutral"}>
                {b}
              </Badge>
            ))}
          </div>
        </motion.div>

        <motion.div
          variants={fadeUp}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.2 }}
          className="lg:col-span-7"
        >
          <Card
            tone="raised"
            className="relative overflow-hidden border-white/[0.08] p-0"
          >
            <header className="flex items-center justify-between border-b border-white/[0.06] px-5 py-3">
              <span className="font-mono text-[12px] text-white/65">
                {content.stream.endpoint}
              </span>
              <span className="flex items-center gap-2 font-mono text-[10px] uppercase tracking-[0.22em] text-[#34D399]">
                <span className="relative inline-flex h-1.5 w-1.5">
                  <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-[#34D399] opacity-80" />
                  <span className="relative inline-flex h-1.5 w-1.5 rounded-full bg-[#34D399]" />
                </span>
                {content.stream.endpointStatus}
              </span>
            </header>
            <div className="relative h-[440px] overflow-hidden bg-[#03040A]">
              <div
                aria-hidden
                className="pointer-events-none absolute inset-x-0 top-0 z-10 h-20 bg-gradient-to-b from-[#03040A] to-transparent"
              />
              <div
                aria-hidden
                className="pointer-events-none absolute inset-x-0 bottom-0 z-10 h-20 bg-gradient-to-t from-[#03040A] to-transparent"
              />
              <motion.ul
                className="px-5 py-4 font-mono text-[12.5px] leading-[1.75]"
                animate={reduced ? undefined : { y: ["0%", "-50%"] }}
                transition={
                  reduced
                    ? undefined
                    : { duration: 22, ease: "linear", repeat: Infinity }
                }
              >
                {loop.map((line, i) => (
                  <li
                    key={`${line.ts}-${i}`}
                    className="flex items-start gap-3 py-[3px]"
                  >
                    <span className="w-[86px] shrink-0 text-white/30 tabular-nums">
                      {line.ts}
                    </span>
                    <span
                      className={cn(
                        "w-[60px] shrink-0 uppercase tracking-[0.18em] text-[10px] pt-[3px]",
                        logLevelStyle[line.level],
                      )}
                    >
                      {line.level}
                    </span>
                    <span className="w-[220px] shrink-0 truncate text-white/55">
                      {line.service}
                    </span>
                    <span className="flex-1 text-white/80">{line.body}</span>
                  </li>
                ))}
              </motion.ul>
            </div>
            <footer className="flex items-center justify-between border-t border-white/[0.06] bg-white/[0.02] px-5 py-2.5 font-mono text-[11px] text-white/45">
              <span>text/event-stream</span>
              <span className="hidden sm:inline">auto-reconnect on disconnect</span>
              <span>tail · 10,000 lines</span>
            </footer>
          </Card>
        </motion.div>
      </div>
    </section>
  );
}

function FeatureSection() {
  return (
    <section
      id="features"
      className="relative z-10 mx-auto w-full max-w-[1320px] px-6 py-36 sm:px-10"
    >
      <motion.div
        variants={fadeUp}
        initial="hidden"
        whileInView="show"
        viewport={{ once: true, amount: 0.3 }}
        className="mb-20 max-w-[44ch]"
      >
        <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-white/40">
          Features
        </p>
        <h2 className="mt-6 text-[clamp(2rem,4.4vw,3.75rem)] font-semibold leading-[1.02] tracking-[-0.03em]">
          A launchctl wrapper, not a SaaS.
        </h2>
      </motion.div>

      <div className="grid grid-cols-1 gap-[1px] bg-white/[0.05] md:grid-cols-6">
        {content.features.map((feat, i) => {
          // Grid-breaking layout: some cards span wider
          const span =
            i === 0
              ? "md:col-span-4"
              : i === 1
                ? "md:col-span-2"
                : i === 2
                  ? "md:col-span-2"
                  : i === 3
                    ? "md:col-span-4"
                    : "md:col-span-3";
          return (
            <motion.article
              key={feat.index}
              variants={fadeUp}
              initial="hidden"
              whileInView="show"
              viewport={{ once: true, amount: 0.2 }}
              transition={{ delay: (i % 3) * 0.08 }}
              className={cn(
                "relative overflow-hidden bg-[#05070D] p-8 sm:p-10",
                span,
              )}
            >
              {feat.accent && featureAccent[feat.accent] && (
                <div
                  aria-hidden
                  className={cn(
                    "pointer-events-none absolute -right-20 -top-20 h-64 w-64 rounded-full bg-gradient-to-br blur-2xl",
                    featureAccent[feat.accent],
                  )}
                />
              )}
              <div className="relative flex flex-col gap-5">
                <span className="font-mono text-[11px] uppercase tracking-[0.22em] text-white/35">
                  {feat.index}
                </span>
                <h3 className="text-[22px] font-medium tracking-[-0.02em] text-white">
                  {feat.title}
                </h3>
                <p className="text-[15px] leading-[1.65] text-white/60">
                  {feat.body}
                </p>
              </div>
            </motion.article>
          );
        })}
      </div>
    </section>
  );
}

function DemoSection() {
  return (
    <section
      id="demo"
      className="relative z-10 mx-auto w-full max-w-[1320px] px-6 py-36 sm:px-10"
    >
      <div className="grid grid-cols-1 gap-14 lg:grid-cols-12">
        <motion.div
          variants={fadeUp}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.3 }}
          className="lg:col-span-5"
        >
          <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-[#67E8F9]">
            {content.demo.eyebrow}
          </p>
          <h2 className="mt-6 text-[clamp(2rem,4.4vw,3.5rem)] font-semibold leading-[1.02] tracking-[-0.03em]">
            {content.demo.title}
          </h2>
          <p className="mt-8 max-w-[44ch] text-[16px] leading-[1.7] text-white/60">
            {content.demo.body}
          </p>
          <div className="mt-10 flex flex-wrap gap-3">
            <StatusDot status="running" label pulse />
            <StatusDot status="scheduled" label />
            <StatusDot status="completed" label />
            <StatusDot status="stopped" label />
            <StatusDot status="error" label />
            <StatusDot status="offline" label />
          </div>
        </motion.div>
        <motion.div
          variants={fadeUp}
          initial="hidden"
          whileInView="show"
          viewport={{ once: true, amount: 0.2 }}
          className="space-y-6 lg:col-span-7"
        >
          {content.demo.session.map((step, i) => (
            <Terminal key={i} prompt={step.prompt}>
              <span className="block whitespace-pre-wrap text-white/70">
                {step.stdout}
              </span>
              {i === content.demo.session.length - 1 && (
                <span className="mt-3 inline-flex items-center gap-2 text-white/55">
                  <span className="text-white/35">$</span>
                  <span aria-hidden className="inline-block h-4 w-1.5 translate-y-[2px] animate-pulse bg-[#22D3EE]" />
                </span>
              )}
            </Terminal>
          ))}
        </motion.div>
      </div>
    </section>
  );
}

function StatsSection() {
  return (
    <section
      id="stats"
      className="relative z-10 border-y border-white/[0.06] bg-[#03040A]/40 py-28"
    >
      <div className="mx-auto grid w-full max-w-[1320px] grid-cols-1 gap-14 px-6 sm:px-10 md:grid-cols-2 lg:grid-cols-4">
        {content.stats.map((s, i) => (
          <motion.div
            key={s.label}
            variants={fadeUp}
            initial="hidden"
            whileInView="show"
            viewport={{ once: true, amount: 0.4 }}
            transition={{ delay: 0.08 * i }}
          >
            <p className="font-mono text-[10px] uppercase tracking-[0.28em] text-white/35">
              /stat {String(i + 1).padStart(2, "0")}
            </p>
            <p className="mt-6 text-[clamp(2.5rem,5vw,4.25rem)] font-semibold leading-[0.95] tracking-[-0.04em] text-white">
              {s.value}
            </p>
            <p className="mt-5 text-[14px] leading-[1.55] text-white/70">
              {s.label}
            </p>
            {s.caption && (
              <p className="mt-2 font-mono text-[11px] text-white/35">
                {s.caption}
              </p>
            )}
          </motion.div>
        ))}
      </div>
    </section>
  );
}

function FaqSection() {
  const [open, setOpen] = useState<number | null>(0);
  return (
    <section
      id="faq"
      className="relative z-10 mx-auto w-full max-w-[1080px] px-6 py-36 sm:px-10"
    >
      <motion.div
        variants={fadeUp}
        initial="hidden"
        whileInView="show"
        viewport={{ once: true, amount: 0.3 }}
        className="mb-16"
      >
        <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-white/40">
          FAQ
        </p>
        <h2 className="mt-6 text-[clamp(2rem,4.2vw,3.25rem)] font-semibold leading-[1.04] tracking-[-0.03em]">
          Questions worth asking before brew install.
        </h2>
      </motion.div>
      <ul className="divide-y divide-white/[0.06] border-y border-white/[0.06]">
        {content.faqs.map((f, i) => {
          const isOpen = open === i;
          return (
            <li key={f.question}>
              <button
                type="button"
                aria-expanded={isOpen}
                onClick={() => setOpen(isOpen ? null : i)}
                className="flex w-full items-start justify-between gap-8 py-6 text-left transition-colors hover:text-white"
              >
                <span className="flex items-baseline gap-4">
                  <span className="font-mono text-[11px] text-white/35">
                    Q{String(i + 1).padStart(2, "0")}
                  </span>
                  <span className="text-[18px] font-medium tracking-[-0.01em] text-white/90">
                    {f.question}
                  </span>
                </span>
                <span
                  aria-hidden
                  className={cn(
                    "mt-1 inline-block h-5 w-5 shrink-0 rounded-full border border-white/20 text-center text-[12px] leading-[18px] transition-transform",
                    isOpen && "rotate-45 border-[#22D3EE] text-[#22D3EE]",
                  )}
                >
                  +
                </span>
              </button>
              <motion.div
                initial={false}
                animate={{
                  height: isOpen ? "auto" : 0,
                  opacity: isOpen ? 1 : 0,
                }}
                transition={{ duration: 0.35, ease: SECTION_EASE }}
                className="overflow-hidden"
              >
                <p className="max-w-[70ch] pb-7 pl-[60px] pr-8 text-[15px] leading-[1.7] text-white/60">
                  {f.answer}
                </p>
              </motion.div>
            </li>
          );
        })}
      </ul>
    </section>
  );
}

function CtaSection() {
  const ref = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ["start end", "end start"],
  });
  const glow = useTransform(scrollYProgress, [0, 0.5, 1], [0.2, 1, 0.3]);
  return (
    <section
      id="install"
      className="relative z-10 mx-auto w-full max-w-[1320px] px-6 py-36 sm:px-10"
    >
      <motion.div
        ref={ref}
        variants={fadeUp}
        initial="hidden"
        whileInView="show"
        viewport={{ once: true, amount: 0.3 }}
        className="relative overflow-hidden rounded-[28px] border border-white/[0.08] bg-[#0B0F1A] px-8 py-20 text-center sm:px-16 sm:py-28"
      >
        <motion.div
          aria-hidden
          style={{ opacity: glow }}
          className="pointer-events-none absolute inset-0"
        >
          <div className="absolute left-1/2 top-1/2 h-[620px] w-[620px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-[radial-gradient(closest-side,rgba(34,211,238,0.32),transparent)]" />
          <div className="absolute left-1/3 top-1/3 h-[380px] w-[380px] rounded-full bg-[radial-gradient(closest-side,rgba(124,92,255,0.28),transparent)]" />
        </motion.div>
        <div className="relative">
          <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-[#67E8F9]">
            {content.cta.eyebrow}
          </p>
          <h2 className="mx-auto mt-8 max-w-[18ch] text-[clamp(2.5rem,7vw,6rem)] font-semibold leading-[0.96] tracking-[-0.04em]">
            {content.cta.headline}
          </h2>
          <p className="mx-auto mt-8 max-w-[48ch] text-balance text-[17px] leading-[1.65] text-white/65">
            {content.cta.body}
          </p>
          <div className="mx-auto mt-12 inline-flex items-center gap-3 rounded-2xl border border-white/[0.1] bg-[#03040A] px-5 py-4 font-mono text-[15px] shadow-[0_24px_64px_rgba(0,0,0,0.55)]">
            <span className="text-white/35">$</span>
            <span className="text-white/90">{content.cta.command}</span>
            <CopyButton text={content.cta.command} />
          </div>
          <p className="mt-8 font-mono text-[11px] text-white/35">
            then: <span className="text-white/60">{content.brand.launchCommand}</span>
          </p>
        </div>
      </motion.div>
    </section>
  );
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <button
      type="button"
      onClick={() => {
        if (typeof navigator === "undefined") return;
        navigator.clipboard?.writeText(text).catch(() => {});
        setCopied(true);
        setTimeout(() => setCopied(false), 1600);
      }}
      className="ml-2 rounded-md border border-white/15 bg-white/[0.04] px-2.5 py-1 text-[10px] uppercase tracking-[0.2em] text-white/65 transition-colors hover:border-[#22D3EE]/60 hover:text-[#67E8F9]"
      aria-label="Copy command"
    >
      {copied ? "copied" : "copy"}
    </button>
  );
}

function SiteFooter() {
  return (
    <footer className="relative z-10 mx-auto w-full max-w-[1320px] px-6 pb-16 pt-8 sm:px-10">
      <div className="flex flex-col items-start justify-between gap-8 border-t border-white/[0.06] pt-10 md:flex-row md:items-center">
        <div className="flex items-center gap-3">
          <span className="flex h-8 w-8 items-center justify-center rounded-md border border-white/15 bg-[#0B0F1A] font-mono text-[12px] text-[#67E8F9]">
            {content.brand.mark}
          </span>
          <span className="font-mono text-[12px] text-white/45">
            {content.footer.copyline}
          </span>
        </div>
        <ul className="flex flex-wrap gap-x-7 gap-y-2 text-[13px] text-white/55">
          {content.footer.links.map((l) => (
            <li key={l.href}>
              <a
                href={l.href}
                className="transition-colors hover:text-white"
              >
                {l.label}
              </a>
            </li>
          ))}
        </ul>
      </div>
    </footer>
  );
}
