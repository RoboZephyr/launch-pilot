import { html } from 'htm/preact';
import { useEffect, useRef } from 'preact/hooks';

/**
 * Confirmation dialog for destructive job actions (reload/start/stop).
 * Uses <dialog> element for native modal behavior and keyboard handling.
 *
 * @param {{ open: boolean, label: string, action: string, onConfirm: () => void, onCancel: () => void }} props
 */
export function ConfirmDialog({ open, label, action, onConfirm, onCancel }) {
  const ref = useRef(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    if (open && !el.open) {
      el.showModal();
    } else if (!open && el.open) {
      el.close();
    }
  }, [open]);

  if (!open) return null;

  const actionLabel = action.charAt(0).toUpperCase() + action.slice(1);

  return html`
    <dialog ref=${ref} class="confirm-dialog" onCancel=${onCancel}>
      <div class="confirm-dialog__body">
        <p class="confirm-dialog__title">${actionLabel} job?</p>
        <p class="confirm-dialog__label"><code>${label}</code></p>
        <div class="confirm-dialog__actions">
          <button class="btn btn--secondary" onClick=${onCancel}>Cancel</button>
          <button class="btn btn--danger" onClick=${onConfirm}>${actionLabel}</button>
        </div>
      </div>
    </dialog>
  `;
}
