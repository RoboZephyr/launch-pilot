import { html } from 'htm/preact';
import { useState, useEffect } from 'preact/hooks';
import { signal } from '@preact/signals';
import { apiFetch } from '../lib/api.js';

/** Local search filter for log content */
const logSearch = signal('');

/**
 * Log viewer panel — fetches and displays stdout/stderr for a job.
 * Includes keyword search that filters displayed log lines.
 * Supports "Load More" to fetch additional lines (up to 10000).
 *
 * @param {{ label: string }} props
 */
export function LogViewer({ label }) {
  const [logs, setLogs] = useState(null);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState(null);
  const [lines, setLines] = useState(200);

  const fetchLogs = (n) => {
    return apiFetch(`/api/jobs/${encodeURIComponent(label)}/logs?lines=${n}`);
  };

  useEffect(() => {
    logSearch.value = '';
    setLines(200);
    setLoading(true);
    setError(null);
    fetchLogs(200)
      .then(data => { setLogs(data); setLoading(false); })
      .catch(err => { setError(err.message); setLoading(false); });
  }, [label]);

  const loadMore = () => {
    const next = Math.min(lines * 2, 10000);
    if (next === lines) return;
    setLoadingMore(true);
    fetchLogs(next)
      .then(data => { setLogs(data); setLines(next); setLoadingMore(false); })
      .catch(err => { setError(err.message); setLoadingMore(false); });
  };

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

  const query = logSearch.value.toLowerCase();

  const filterLines = (text) => {
    if (!text) return '';
    if (!query) return text;
    return text.split('\n').filter(line => line.toLowerCase().includes(query)).join('\n');
  };

  return html`
    <div class="log-viewer">
      <div class="log-viewer__search">
        <input
          type="text"
          class="log-viewer__search-input"
          placeholder="Filter logs\u2026"
          value=${logSearch}
          onInput=${(e) => { logSearch.value = e.target.value; }}
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
      ${lines < 10000 && html`
        <button class="btn btn--sm btn--outline log-viewer__load-more" onClick=${loadMore} disabled=${loadingMore}>
          ${loadingMore ? 'Loading\u2026' : `Load more (currently ${lines} lines)`}
        </button>
      `}
    </div>
  `;
}
