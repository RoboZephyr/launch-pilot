import { signal, computed } from '@preact/signals';

/** @type {import('@preact/signals').Signal<Array>} Full job list from SSE */
export const jobs = signal([]);

/** @type {import('@preact/signals').Signal<string>} Current search/filter text */
export const searchQuery = signal('');

/** @type {import('@preact/signals').Signal<object|null>} Currently selected job */
export const selectedJob = signal(null);

/** @type {import('@preact/signals').Signal<Array<{id:number, message:string, ok:boolean}>>} Active toast notifications */
export const toasts = signal([]);

/** @type {import('@preact/signals').Signal<string|null>} Label of job whose detail panel is open */
export const expandedJob = signal(null);

/** @type {import('@preact/signals').Signal<'logs'|'diagnose'|null>} Which panel is active */
export const activePanel = signal(null);

/** Filtered jobs — recomputes when jobs or searchQuery change */
export const filteredJobs = computed(() => {
  const q = searchQuery.value.toLowerCase();
  if (!q) return jobs.value;
  return jobs.value.filter(j =>
    j.label.toLowerCase().includes(q)
  );
});

let _toastId = 0;

/** Add a toast notification that auto-dismisses after 3 seconds. */
export function addToast(message, ok) {
  const id = ++_toastId;
  toasts.value = [...toasts.value, { id, message, ok }];
  setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== id);
  }, 3000);
}

/** Remove a specific toast by id. */
export function removeToast(id) {
  toasts.value = toasts.value.filter(t => t.id !== id);
}
