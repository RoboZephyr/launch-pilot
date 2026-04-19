import { html } from 'htm/preact';
import { useRef, useLayoutEffect, useEffect } from 'preact/hooks';
import { useComputed, useSignalEffect } from '@preact/signals';
import { jobs, tooltipTarget } from '../lib/state.js';

const STATUSES_EXPECTING_LAST_RUN = new Set(['completed', 'scheduled']);

export function buildStatusTooltip(job) {
  return buildStatusTooltipParts(job).join(' — ');
}

// Contract: buildStatusTooltip(job) === buildStatusTooltipParts(job).join(' — ').
export function buildStatusTooltipParts(job) {
  const fmt = (iso) => new Date(iso).toLocaleString();
  const parts = [job.status];
  if (job.nextRunAt) parts.push(`Next run: ${fmt(job.nextRunAt)}`);
  if (job.lastRunAt) {
    parts.push(`Last run: ${fmt(job.lastRunAt)}`);
  } else if (STATUSES_EXPECTING_LAST_RUN.has(job.status)) {
    const hasLogPath = Boolean(job.standardOutPath) || Boolean(job.standardErrPath);
    parts.push(hasLogPath ? 'Last run: unknown' : 'Last run: unknown (no log path configured)');
  }
  return parts;
}

// Above anchor, centered. Flips below when top-space < 4.
// Clamps left into [4, viewport.width - tip.width - 4].
export function placeTooltip(anchor, tip, viewport, margin = 6) {
  let top = anchor.top - tip.height - margin;
  if (top < 4) top = anchor.bottom + margin;
  const anchorCenter = anchor.left + anchor.width / 2;
  let left = anchorCenter - tip.width / 2;
  const maxLeft = viewport.width - tip.width - 4;
  if (left < 4) left = 4;
  else if (left > maxLeft) left = maxLeft;
  return { top, left };
}

export function showTooltip(anchor, label, enteredVia = 'hover') {
  const cur = tooltipTarget.peek();
  if (cur && cur.anchor === anchor && cur.label === label && cur.enteredVia === enteredVia) return;
  tooltipTarget.value = { anchor, label, enteredVia };
}

export function hideTooltip() {
  if (tooltipTarget.peek() === null) return;
  tooltipTarget.value = null;
}

export function StatusDot({ job }) {
  const ariaLabel = buildStatusTooltip(job);
  const dotClass = `status-dot status-dot--${job.status} status-dot-trigger`;
  const show = (via) => (e) => showTooltip(e.currentTarget, job.label, via);
  return html`
    <button
      type="button"
      class=${dotClass}
      aria-label=${ariaLabel}
      data-label=${job.label}
      onPointerEnter=${show('hover')}
      onPointerLeave=${hideTooltip}
      onFocus=${show('focus')}
      onBlur=${hideTooltip}
    ></button>
  `;
}

// Singleton overlay — mount once at <App> root.
export function StatusTooltip() {
  const ref = useRef(null);

  const jobSignal = useComputed(() => {
    const t = tooltipTarget.value;
    if (!t) return null;
    return jobs.value.find(j => j.label === t.label) ?? null;
  });

  const target = tooltipTarget.value;
  const job = jobSignal.value;
  const visible = target !== null && job !== null;

  // Close immediately if the label vanishes from the snapshot or anchor detaches.
  useSignalEffect(() => {
    const t = tooltipTarget.value;
    if (!t) return;
    if (jobSignal.value === null || !t.anchor.isConnected) hideTooltip();
  });

  useEffect(() => {
    if (!visible) return undefined;
    const dismiss = () => hideTooltip();
    const onKey = (e) => {
      if (e.key !== 'Escape') return;
      const t = tooltipTarget.peek();
      if (t && t.enteredVia === 'focus' && t.anchor && t.anchor.isConnected) {
        t.anchor.focus();
      }
      hideTooltip();
    };
    window.addEventListener('scroll', dismiss, true);
    window.addEventListener('resize', dismiss);
    window.addEventListener('keydown', onKey);
    return () => {
      window.removeEventListener('scroll', dismiss, true);
      window.removeEventListener('resize', dismiss);
      window.removeEventListener('keydown', onKey);
    };
  }, [visible]);

  useLayoutEffect(() => {
    if (!visible || !ref.current) return;
    const a = target.anchor.getBoundingClientRect();
    const t = ref.current.getBoundingClientRect();
    const pos = placeTooltip(a, t, { width: window.innerWidth, height: window.innerHeight });
    ref.current.style.top = `${pos.top}px`;
    ref.current.style.left = `${pos.left}px`;
  }, [target, job]);

  const parts = job ? buildStatusTooltipParts(job) : [];
  const cls = `status-tooltip${visible ? ' status-tooltip--visible' : ''}`;

  return html`
    <div
      id="status-tooltip-singleton"
      ref=${ref}
      class=${cls}
      role="tooltip"
      aria-hidden=${!visible}
    >
      ${parts.map((line, i) => html`<div class="status-tooltip__line" key=${i}>${line}</div>`)}
    </div>
  `;
}
