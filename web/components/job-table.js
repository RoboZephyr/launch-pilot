import { html } from 'htm/preact';
import { filteredJobs } from '../lib/state.js';
import { JobRow } from './job-row.js';

export function JobTable() {
  const list = filteredJobs.value;

  if (list.length === 0) {
    return html`<p class="job-table__empty">No jobs found.</p>`;
  }

  return html`
    <table class="job-table">
      <thead>
        <tr>
          <th class="job-table__th--status"></th>
          <th>Label</th>
          <th>PID</th>
          <th>Exit Status</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        ${list.map(job => html`<${JobRow} key=${job.label} job=${job} />`)}
      </tbody>
    </table>
  `;
}
