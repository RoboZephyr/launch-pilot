import { html } from 'htm/preact';
import { toasts, removeToast } from '../lib/state.js';

/**
 * Toast container — renders active toast notifications.
 * Green for success (ok=true), red for failure (ok=false).
 * Each toast auto-dismisses after 3 seconds (handled in state.js).
 */
export function ToastContainer() {
  const list = toasts.value;
  if (list.length === 0) return null;

  return html`
    <div class="toast-container">
      ${list.map(t => html`
        <div key=${t.id} class="toast toast--${t.ok ? 'success' : 'error'}">
          <span class="toast__message">${t.message}</span>
          <button class="toast__close" onClick=${() => removeToast(t.id)}>\u00d7</button>
        </div>
      `)}
    </div>
  `;
}
