import { signal, computed, effect } from '@preact/signals';
import { classifyJob } from './classify.js';

const ONLY_MINE_KEY = 'launch-pilot:only-mine';

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

/** @type {import('@preact/signals').Signal<'all'|'mine'|'system'|'thirdparty'>} */
export const categoryFilter = signal('all');

/** @type {import('@preact/signals').Signal<'all'|'running'|'scheduled'|'completed'|'stopped'|'error'|'offline'>} */
export const statusFilter = signal('all');

/** @type {import('@preact/signals').Signal<boolean>} Quick toggle — persisted to localStorage */
export const onlyMine = signal(
  (() => { try { return localStorage.getItem(ONLY_MINE_KEY) === 'true'; } catch { return false; } })()
);

// Persist onlyMine to localStorage
effect(() => {
  try { localStorage.setItem(ONLY_MINE_KEY, String(onlyMine.value)); }
  catch { /* private browsing */ }
});

// onlyMine ON → force categoryFilter to 'mine'; OFF → reset to 'all'
let _prevOnlyMine = onlyMine.peek();
effect(() => {
  const current = onlyMine.value;
  if (current) {
    if (categoryFilter.peek() !== 'mine') categoryFilter.value = 'mine';
  } else if (_prevOnlyMine) {
    categoryFilter.value = 'all';
  }
  _prevOnlyMine = current;
});

/** Filtered jobs — 4-stage pipeline: onlyMine → category → status → search */
export const filteredJobs = computed(() => {
  let list = jobs.value;

  if (onlyMine.value) {
    list = list.filter(j => classifyJob(j) === 'mine');
  }

  const cat = categoryFilter.value;
  if (cat !== 'all') {
    list = list.filter(j => classifyJob(j) === cat);
  }

  const st = statusFilter.value;
  if (st !== 'all') {
    list = list.filter(j => j.status === st);
  }

  const q = searchQuery.value.toLowerCase();
  if (q) {
    list = list.filter(j => j.label.toLowerCase().includes(q));
  }

  return list;
});

/** Category counts from full job list (not filtered). */
export const categoryCounts = computed(() => {
  const list = jobs.value;
  const counts = { all: list.length, mine: 0, system: 0, thirdparty: 0 };
  for (const j of list) counts[classifyJob(j)]++;
  return counts;
});

/** Status counts from full job list (not filtered). */
export const statusCounts = computed(() => {
  const list = jobs.value;
  const counts = {
    all: list.length,
    running: 0,
    scheduled: 0,
    completed: 0,
    stopped: 0,
    error: 0,
    offline: 0,
  };
  for (const j of list) {
    const s = j.status;
    if (s in counts) counts[s]++;
  }
  return counts;
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
