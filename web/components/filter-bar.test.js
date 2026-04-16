import { describe, it, beforeEach } from 'node:test';
import { strict as assert } from 'node:assert';
import { effect } from '@preact/signals';
import {
  jobs,
  searchQuery,
  filteredJobs,
  categoryFilter,
  statusFilter,
  onlyMine,
  categoryCounts,
  statusCounts,
} from '../lib/state.js';
import { CATEGORY_LABELS, CATEGORY_KEYS, STATUS_KEYS } from '../lib/classify.js';

// --- Test fixtures ---

const FIXTURES = [
  { label: 'com.apple.spotlight',       domain: 'user',   status: 'running' },
  { label: 'com.apple.WindowServer',    domain: 'global', status: 'running' },
  { label: 'com.example.myapp',         domain: 'user',   status: 'stopped' },
  { label: 'org.homebrew.mxcl.redis',   domain: 'user',   status: 'running' },
  { label: 'com.docker.vmnetd',         domain: 'global', status: 'error'   },
  { label: 'com.microsoft.autoupdate',  domain: 'global', status: 'stopped' },
  { label: 'com.myco.agent',            domain: 'user',   status: 'error'   },
];

function resetSignals() {
  jobs.value = [];
  searchQuery.value = '';
  categoryFilter.value = 'all';
  statusFilter.value = 'all';
  onlyMine.value = false;
}

// --- FilterBar behavioral tests ---

describe('FilterBar: category chips', () => {
  beforeEach(resetSignals);

  it('should have 4 category options: All, Mine, System, 3rd-party', () => {
    assert.equal(CATEGORY_KEYS.length, 4);
    assert.deepEqual(CATEGORY_KEYS, ['all', 'mine', 'system', 'thirdparty']);
    assert.equal(CATEGORY_LABELS.mine, 'Mine');
    assert.equal(CATEGORY_LABELS.system, 'System');
    assert.equal(CATEGORY_LABELS.thirdparty, '3rd-party');
  });

  it('categoryCounts reflects correct counts per category', () => {
    jobs.value = FIXTURES;
    const counts = categoryCounts.value;
    assert.equal(counts.all, 7);
    assert.equal(counts.mine, 3);
    assert.equal(counts.system, 2);
    assert.equal(counts.thirdparty, 2);
  });

  it('setting categoryFilter updates filteredJobs', () => {
    jobs.value = FIXTURES;
    categoryFilter.value = 'mine';
    assert.equal(filteredJobs.value.length, 3);
    categoryFilter.value = 'system';
    assert.equal(filteredJobs.value.length, 2);
    categoryFilter.value = 'thirdparty';
    assert.equal(filteredJobs.value.length, 2);
    categoryFilter.value = 'all';
    assert.equal(filteredJobs.value.length, 7);
  });
});

describe('FilterBar: status tabs', () => {
  beforeEach(resetSignals);

  it('should have 4 status options: all, running, stopped, error', () => {
    assert.equal(STATUS_KEYS.length, 4);
    assert.deepEqual(STATUS_KEYS, ['all', 'running', 'stopped', 'error']);
  });

  it('statusCounts reflects correct counts per status', () => {
    jobs.value = FIXTURES;
    const counts = statusCounts.value;
    assert.equal(counts.all, 7);
    assert.equal(counts.running, 3);
    assert.equal(counts.stopped, 2);
    assert.equal(counts.error, 2);
  });

  it('setting statusFilter updates filteredJobs', () => {
    jobs.value = FIXTURES;
    statusFilter.value = 'running';
    assert.equal(filteredJobs.value.length, 3);
    statusFilter.value = 'stopped';
    assert.equal(filteredJobs.value.length, 2);
    statusFilter.value = 'error';
    assert.equal(filteredJobs.value.length, 2);
    statusFilter.value = 'all';
    assert.equal(filteredJobs.value.length, 7);
  });
});

describe('FilterBar: Only Mine toggle', () => {
  beforeEach(resetSignals);

  it('onlyMine=true forces categoryFilter to "mine" (via effect)', () => {
    jobs.value = FIXTURES;

    // Simulate the effect that FilterBar sets up
    const dispose = effect(() => {
      if (onlyMine.value) {
        categoryFilter.value = 'mine';
      }
    });

    categoryFilter.value = 'system';
    assert.equal(categoryFilter.value, 'system');

    onlyMine.value = true;
    assert.equal(categoryFilter.value, 'mine');

    dispose();
  });

  it('turning onlyMine OFF leaves categoryFilter at "mine" (no jump to "all")', () => {
    jobs.value = FIXTURES;

    const dispose = effect(() => {
      if (onlyMine.value) {
        categoryFilter.value = 'mine';
      }
    });

    onlyMine.value = true;
    assert.equal(categoryFilter.value, 'mine');

    onlyMine.value = false;
    // Should stay at 'mine', not reset to 'all'
    assert.equal(categoryFilter.value, 'mine');

    dispose();
  });

  it('onlyMine + statusFilter compose correctly (intersection)', () => {
    jobs.value = FIXTURES;
    onlyMine.value = true;
    statusFilter.value = 'error';
    // Only user non-apple error jobs
    const labels = filteredJobs.value.map(j => j.label);
    assert.deepEqual(labels, ['com.myco.agent']);
  });
});

describe('FilterBar: filter composition', () => {
  beforeEach(resetSignals);

  it('category + status compose as AND (Mine + Error = intersection)', () => {
    jobs.value = FIXTURES;
    categoryFilter.value = 'mine';
    statusFilter.value = 'error';
    const labels = filteredJobs.value.map(j => j.label);
    assert.deepEqual(labels, ['com.myco.agent']);
  });

  it('category + status + search all compose as AND', () => {
    jobs.value = FIXTURES;
    categoryFilter.value = 'system';
    statusFilter.value = 'running';
    searchQuery.value = 'spotlight';
    const labels = filteredJobs.value.map(j => j.label);
    assert.deepEqual(labels, ['com.apple.spotlight']);
  });

  it('search is preserved across tab/chip switches', () => {
    jobs.value = FIXTURES;
    searchQuery.value = 'com.';

    categoryFilter.value = 'mine';
    for (const j of filteredJobs.value) {
      assert.ok(j.label.includes('com.'));
    }

    statusFilter.value = 'stopped';
    for (const j of filteredJobs.value) {
      assert.ok(j.label.includes('com.'));
      assert.equal(j.status, 'stopped');
    }

    // Switch back to all — search still active
    categoryFilter.value = 'all';
    statusFilter.value = 'all';
    for (const j of filteredJobs.value) {
      assert.ok(j.label.includes('com.'));
    }
  });
});
