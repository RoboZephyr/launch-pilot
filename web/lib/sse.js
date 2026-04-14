import { jobs } from './state.js';

/**
 * Connect to the SSE endpoint and update the jobs signal on each event.
 * EventSource handles reconnection automatically.
 * @returns {EventSource}
 */
export function connectSSE() {
  const es = new EventSource('/api/events');

  es.addEventListener('jobs', (e) => {
    try {
      jobs.value = JSON.parse(e.data);
    } catch (_) {
      // Malformed data — ignore, wait for next event.
    }
  });

  return es;
}
