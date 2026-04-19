import { html } from 'htm/preact';
import { useState, useEffect } from 'preact/hooks';
import { signal } from '@preact/signals';
import { apiFetch } from '../lib/api.js';

/** Local search filter for log content */
const logSearch = signal('');

/**
 * Infer whether more log lines may exist server-side without a backend
 * `hasMore` field. If `requested >= 10000` we've hit the hard cap → false.
 * Otherwise, if either stream returned ≥ requested lines, there may be more.
 *
 * Edge case: a stream whose actual line count equals `requested` yields a
 * one-time false positive; the next Load more fetches the same count and
 * clears hasMore. Acceptable per spec (AC does not require 0 false positives).
 *
 * @param {object|null} logs  /api/jobs/:label/logs response
 * @param {number} requested  lines query param sent to the API
 * @returns {boolean}
 */
export function computeHasMore(logs, requested) {
  if (requested >= 10000) return false;
  if (!logs) return false;
  const stdoutLines = logs.stdoutAvailable && logs.stdout
    ? logs.stdout.split('\n').length : 0;
  const stderrLines = logs.stderrAvailable && logs.stderr
    ? logs.stderr.split('\n').length : 0;
  return stdoutLines >= requested || stderrLines >= requested;
}

function countLines(text) {
  if (!text) return 0;
  return text.split('\n').length;
}

/**
 * Pure render body for LogViewer — receives resolved state as props so it can
 * be exercised by render-to-string tests without stubbing fetch.
 */
export function LogViewerView({
  logs, lines, loading, loadingMore, error, hasMore,
  search, onSearchInput, onLoadMore,
}) {
  if (loading) {
    return html`<div class="log-viewer"><p class="log-viewer__status">Loading logs\u2026</p></div>`;
  }
  if (error) {
    return html`<div class="log-viewer"><p class="log-viewer__status log-viewer__status--error">${error}</p></div>`;
  }
  if (!logs.stdoutAvailable && !logs.stderrAvailable) {
    return html`
      <div class="log-viewer">
        <p class="log-viewer__status">${logs.message || 'No log paths configured in plist'}</p>
      </div>
    `;
  }

  const query = (search || '').toLowerCase();
  const filterLines = (text) => {
    if (!text) return '';
    if (!query) return text;
    return text.split('\n').filter(line => line.toLowerCase().includes(query)).join('\n');
  };

  const totalLines = (logs.stdoutAvailable ? countLines(logs.stdout) : 0)
    + (logs.stderrAvailable ? countLines(logs.stderr) : 0);
  const showLoadMore = hasMore && lines < 10000;
  const showDone = !hasMore && lines >= 200;

  return html`
    <div class="log-viewer">
      <div class="log-viewer__search">
        <input
          type="text"
          class="log-viewer__search-input"
          placeholder="Filter logs\u2026"
          value=${search}
          onInput=${onSearchInput}
        />
      </div>
      ${logs.stdoutAvailable && html`
        <div class="log-viewer__section">
          <h4 class="log-viewer__heading">stdout <code class="log-viewer__path">${logs.stdoutPath}</code></h4>
          <pre class="log-viewer__content">${filterLines(logs.stdout) || '(empty)'}</pre>
        </div>
      `}
      ${logs.stderrAvailable && html`
        <div class="log-viewer__section">
          <h4 class="log-viewer__heading">stderr <code class="log-viewer__path">${logs.stderrPath}</code></h4>
          <pre class="log-viewer__content log-viewer__content--stderr">${filterLines(logs.stderr) || '(empty)'}</pre>
        </div>
      `}
      ${showLoadMore && html`
        <button class="btn btn--sm btn--outline log-viewer__load-more" onClick=${onLoadMore} disabled=${loadingMore}>
          ${loadingMore ? 'Loading\u2026' : `Load more (currently ${lines} lines)`}
        </button>
      `}
      ${showDone && html`
        <p class="log-viewer__done">Showing all ${totalLines} lines</p>
      `}
    </div>
  `;
}

/**
 * Log viewer panel — fetches and displays stdout/stderr for a job.
 * Includes keyword search that filters displayed log lines.
 * Supports "Load more" to fetch additional lines (up to 10000).
 *
 * @param {{ label: string }} props
 */
export function LogViewer({ label }) {
  const [logs, setLogs] = useState(null);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState(null);
  const [lines, setLines] = useState(200);
  const [hasMore, setHasMore] = useState(false);

  const fetchLogs = (n) => {
    return apiFetch(`/api/jobs/${encodeURIComponent(label)}/logs?lines=${n}`);
  };

  useEffect(() => {
    logSearch.value = '';
    setLines(200);
    setLoading(true);
    setError(null);
    setHasMore(false);
    fetchLogs(200)
      .then(data => {
        setLogs(data);
        setHasMore(computeHasMore(data, 200));
        setLoading(false);
      })
      .catch(err => { setError(err.message); setLoading(false); });
  }, [label]);

  const loadMore = () => {
    const next = Math.min(lines * 2, 10000);
    if (next === lines) return;
    setLoadingMore(true);
    fetchLogs(next)
      .then(data => {
        setLogs(data);
        setLines(next);
        setHasMore(computeHasMore(data, next));
        setLoadingMore(false);
      })
      .catch(err => { setError(err.message); setLoadingMore(false); });
  };

  return html`<${LogViewerView}
    logs=${logs}
    lines=${lines}
    loading=${loading}
    loadingMore=${loadingMore}
    error=${error}
    hasMore=${hasMore}
    search=${logSearch.value}
    onSearchInput=${(e) => { logSearch.value = e.target.value; }}
    onLoadMore=${loadMore}
  />`;
}
