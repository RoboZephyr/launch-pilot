import { html } from 'htm/preact';
import { useState, useEffect } from 'preact/hooks';
import { apiFetch } from '../lib/api.js';

/** Severity icon mapping: ok=green circle, warning=yellow triangle, error=red cross */
const severityIcon = { ok: '\u2705', warning: '\u26a0\ufe0f', error: '\u274c' };

/**
 * Diagnostics panel — fetches and displays check results for a job.
 * Shows 6 checks with green/yellow/red severity indicators.
 * Displays suggestion text for non-ok checks.
 *
 * @param {{ label: string }} props
 */
export function DiagnosePanel({ label }) {
  const [report, setReport] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    apiFetch(`/api/jobs/${encodeURIComponent(label)}/diagnose`)
      .then(data => { setReport(data); setLoading(false); })
      .catch(err => { setError(err.message); setLoading(false); });
  }, [label]);

  if (loading) {
    return html`<div class="diagnose-panel"><p class="diagnose-panel__status">Running diagnostics\u2026</p></div>`;
  }

  if (error) {
    return html`<div class="diagnose-panel"><p class="diagnose-panel__status diagnose-panel__status--error">${error}</p></div>`;
  }

  if (!report || !report.checks || report.checks.length === 0) {
    return html`<div class="diagnose-panel"><p class="diagnose-panel__status">No diagnostic checks available.</p></div>`;
  }

  return html`
    <div class="diagnose-panel">
      <ul class="diagnose-panel__list">
        ${report.checks.map(check => html`
          <li key=${check.id} class="diagnose-panel__check">
            <div class="diagnose-panel__header">
              <span class="diagnose-panel__icon severity--${check.severity}">${severityIcon[check.severity] || '?'}</span>
              <span class="diagnose-panel__name">${check.name}</span>
            </div>
            <p class="diagnose-panel__message">${check.message}</p>
            ${check.severity !== 'ok' && check.suggestion && html`
              <p class="diagnose-panel__suggestion">${check.suggestion}</p>
            `}
          </li>
        `)}
      </ul>
    </div>
  `;
}
