import {
  jobs,
  searchQuery,
  categoryFilter,
  statusFilter,
  onlyMine,
} from './state.js';

export const FIXTURES = [
  { label: 'com.apple.spotlight',       domain: 'user',   status: 'running' },
  { label: 'com.apple.WindowServer',    domain: 'global', status: 'running' },
  { label: 'com.example.myapp',         domain: 'user',   status: 'stopped' },
  { label: 'org.homebrew.mxcl.redis',   domain: 'user',   status: 'running' },
  { label: 'com.docker.vmnetd',         domain: 'global', status: 'error'   },
  { label: 'com.microsoft.autoupdate',  domain: 'global', status: 'stopped' },
  { label: 'com.myco.agent',            domain: 'user',   status: 'error'   },
];

export function resetSignals() {
  jobs.value = [];
  searchQuery.value = '';
  categoryFilter.value = 'all';
  statusFilter.value = 'all';
  onlyMine.value = false;
}
