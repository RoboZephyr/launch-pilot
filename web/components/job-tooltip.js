import { html } from 'htm/preact';
import { useRef, useLayoutEffect, useEffect } from 'preact/hooks';
import { jobs, tooltipAnchor, tooltipLabel, tooltipVisible } from '../lib/state.js';

const STATUSES_EXPECTING_LAST_RUN = new Set(['completed', 'scheduled']);

/**
 * Build the hover-tooltip string for a job status badge. Pure function — no
 * DOM deps so `node --test` can verify it without resolving htm/preact.
 * @param {{status: string, nextRunAt?: string, lastRunAt?: string, standardOutPath?: string, standardErrPath?: string}} job
 * @returns {string}
 */
export function buildStatusTooltip(job) {
  return buildStatusTooltipParts(job).join(' — ');
}

/**
 * Pure: returns tooltip lines in order so renderers can format multi-line.
 * Contract: buildStatusTooltip(job) === buildStatusTooltipParts(job).join(' — ').
 * @param {object} job
 * @returns {string[]}
 */
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

/**
 * Pure: compute fixed-position top/left for tooltip given anchor + tooltip rects + viewport.
 * Preferred placement: above anchor, centered. Flips below when top-space < 4.
 * Clamps left into [4, viewport.width - tip.width - 4].
 * @param {{top:number,left:number,width:number,height:number,bottom:number,right:number}} anchor
 * @param {{width:number,height:number}} tip
 * @param {{width:number,height:number}} viewport
 * @param {number} [margin=6]
 * @returns {{top:number,left:number}}
 */
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

/** Imperative: set the tooltip target + show. */
export function showTooltip(anchorEl, label) {
  tooltipAnchor.value = anchorEl;
  tooltipLabel.value = label;
  tooltipVisible.value = true;
}

/** Imperative: hide the tooltip. Anchor/label kept to avoid transient null. */
export function hideTooltip() {
  tooltipVisible.value = false;
}

/**
 * Status badge trigger — renders a focusable button styled as the 8x8 dot
 * and wires pointer/focus/Esc handlers to the singleton tooltip signals.
 * @param {{ job: object }} props
 */
export function StatusDot({ job }) {
  const label = buildStatusTooltip(job);
  const dotClass = `status-dot status-dot--${job.status} status-dot-trigger`;
  const onEnter = (e) => showTooltip(e.currentTarget, job.label);
  const onLeave = () => hideTooltip();
  const onFocus = (e) => showTooltip(e.currentTarget, job.label);
  const onBlur = () => hideTooltip();
  const onKeyDown = (e) => {
    if (e.key === 'Escape') {
      hideTooltip();
      e.currentTarget.blur();
    }
  };
  return html`
    <button
      type="button"
      class=${dotClass}
      aria-label=${label}
      data-label=${job.label}
      onPointerEnter=${onEnter}
      onPointerLeave=${onLeave}
      onFocus=${onFocus}
      onBlur=${onBlur}
      onKeyDown=${onKeyDown}
    ></button>
  `;
}

/** Singleton overlay — mount once at <App> root. */
export function StatusTooltip() {
  const ref = useRef(null);
  const visible = tooltipVisible.value;
  const anchor = tooltipAnchor.value;
  const label = tooltipLabel.value;
  const job = label ? jobs.value.find(j => j.label === label) : null;

  useEffect(() => {
    if (visible && anchor && !anchor.isConnected) hideTooltip();
  });

  useEffect(() => {
    if (visible && label && !job) hideTooltip();
  });

  useEffect(() => {
    if (!visible) return undefined;
    const onScrollOrResize = () => hideTooltip();
    window.addEventListener('scroll', onScrollOrResize, true);
    window.addEventListener('resize', onScrollOrResize);
    return () => {
      window.removeEventListener('scroll', onScrollOrResize, true);
      window.removeEventListener('resize', onScrollOrResize);
    };
  }, [visible]);

  useEffect(() => {
    if (!visible) return undefined;
    const onKey = (e) => { if (e.key === 'Escape') hideTooltip(); };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [visible]);

  useLayoutEffect(() => {
    if (!visible || !anchor || !ref.current) return;
    const a = anchor.getBoundingClientRect();
    const t = ref.current.getBoundingClientRect();
    const pos = placeTooltip(a, t, { width: window.innerWidth, height: window.innerHeight });
    ref.current.style.top = `${pos.top}px`;
    ref.current.style.left = `${pos.left}px`;
  });

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
