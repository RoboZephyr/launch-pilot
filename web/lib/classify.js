const APPLE_PREFIX = 'com.apple.';

/**
 * Classify a job into a category based on label prefix and domain.
 * @param {{ label: string, domain: string }} job
 * @returns {'mine'|'system'|'thirdparty'}
 */
export function classifyJob(job) {
  if (!job.label) return 'thirdparty';
  if (job.label.startsWith(APPLE_PREFIX)) return 'system';
  if (job.domain === 'user') return 'mine';
  return 'thirdparty';
}

/** Display labels for each category. */
export const CATEGORY_LABELS = {
  mine: 'Mine',
  system: 'System',
  thirdparty: '3rd-party',
};

/** All category filter keys including "all". */
export const CATEGORY_KEYS = ['all', 'mine', 'system', 'thirdparty'];

/** All status filter keys including "all". */
export const STATUS_KEYS = [
  'all',
  'running',
  'scheduled',
  'completed',
  'stopped',
  'error',
  'offline',
];
