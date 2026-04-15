import { html } from 'htm/preact';
import { useState } from 'preact/hooks';
import { expandedJob, activePanel, addToast } from '../lib/state.js';
import { postAction } from '../lib/api.js';
import { classifyJob, CATEGORY_LABELS } from '../lib/classify.js';
import { ConfirmDialog } from './confirm-dialog.js';
import { LogViewer } from './log-viewer.js';
import { DiagnosePanel } from './diagnose-panel.js';

/**
 * Single row in the job table with action buttons and expandable panels.
 * @param {{ job: object }} props
 */
export function JobRow({ job }) {
  const [confirm, setConfirm] = useState(null); // { action: string } | null

  const dotClass = `status-dot status-dot--${job.status}`;
  const isExpanded = expandedJob.value === job.label;
  const panel = activePanel.value;

  const togglePanel = (panelName) => {
    if (isExpanded && panel === panelName) {
      expandedJob.value = null;
      activePanel.value = null;
    } else {
      expandedJob.value = job.label;
      activePanel.value = panelName;
    }
  };

  const handleAction = async (action) => {
    setConfirm(null);
    try {
      await postAction(job.label, action);
      addToast(`${action} succeeded: ${job.label}`, true);
    } catch (err) {
      addToast(`${action} failed: ${err.message}`, false);
    }
  };

  return html`
    <tr class="job-row">
      <td class="job-row__status"><span class=${dotClass} title=${job.status}></span></td>
      <td class="job-row__label"><code>${job.label}</code><span class=${'category-badge category-badge--' + classifyJob(job)}>${CATEGORY_LABELS[classifyJob(job)]}</span></td>
      <td class="job-row__pid">${job.pid > 0 ? job.pid : '\u2014'}</td>
      <td class="job-row__exit">${job.lastExitStatus}</td>
      <td class="job-row__actions">
        <button class="btn btn--sm" onClick=${() => setConfirm({ action: 'reload' })} title="Reload">Reload</button>
        <button class="btn btn--sm" onClick=${() => setConfirm({ action: 'start' })} title="Start">Start</button>
        <button class="btn btn--sm" onClick=${() => setConfirm({ action: 'stop' })} title="Stop">Stop</button>
        <button class="btn btn--sm btn--outline ${isExpanded && panel === 'logs' ? 'btn--active' : ''}" onClick=${() => togglePanel('logs')}>Logs</button>
        <button class="btn btn--sm btn--outline ${isExpanded && panel === 'diagnose' ? 'btn--active' : ''}" onClick=${() => togglePanel('diagnose')}>Diagnose</button>
      </td>
    </tr>
    ${isExpanded && html`
      <tr class="job-row__detail">
        <td colspan="5">
          ${panel === 'logs' && html`<${LogViewer} label=${job.label} />`}
          ${panel === 'diagnose' && html`<${DiagnosePanel} label=${job.label} />`}
        </td>
      </tr>
    `}
    <${ConfirmDialog}
      open=${confirm !== null}
      label=${job.label}
      action=${confirm ? confirm.action : ''}
      onConfirm=${() => confirm && handleAction(confirm.action)}
      onCancel=${() => setConfirm(null)}
    />
  `;
}
