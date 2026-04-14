/**
 * Perform a JSON API request.
 * @param {string} path  — URL path (e.g. "/api/jobs")
 * @param {RequestInit} [opts]
 * @returns {Promise<any>}
 */
export async function apiFetch(path, opts) {
  const res = await fetch(path, opts);
  const body = await res.json();
  if (!res.ok) {
    throw new Error(body.error || `HTTP ${res.status}`);
  }
  return body;
}

/** POST an action (reload/start/stop) on a job. */
export function postAction(label, action) {
  return apiFetch(`/api/jobs/${encodeURIComponent(label)}/${action}`, {
    method: 'POST',
  });
}
