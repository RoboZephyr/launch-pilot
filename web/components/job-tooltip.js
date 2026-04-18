const STATUSES_EXPECTING_LAST_RUN = new Set(['completed', 'scheduled']);

/**
 * Build the hover-tooltip string for a job status badge. Pure function — no
 * DOM deps so `node --test` can verify it without resolving htm/preact.
 * @param {{status: string, nextRunAt?: string, lastRunAt?: string, standardOutPath?: string, standardErrPath?: string}} job
 * @returns {string}
 */
export function buildStatusTooltip(job) {
  const fmt = (iso) => new Date(iso).toLocaleString();
  const parts = [job.status];
  if (job.nextRunAt) parts.push(`Next run: ${fmt(job.nextRunAt)}`);
  if (job.lastRunAt) {
    parts.push(`Last run: ${fmt(job.lastRunAt)}`);
  } else if (STATUSES_EXPECTING_LAST_RUN.has(job.status)) {
    const hasLogPath = Boolean(job.standardOutPath) || Boolean(job.standardErrPath);
    parts.push(hasLogPath ? 'Last run: unknown' : 'Last run: unknown (no log path configured)');
  }
  return parts.join(' — ');
}
