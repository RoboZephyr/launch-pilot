import { describe, it, beforeEach } from 'node:test';
import { strict as assert } from 'node:assert';
import {
  jobs,
  filteredJobs,
  categoryFilter,
  statusFilter,
  onlyMine,
  categoryCounts,
  statusCounts,
} from '../lib/state.js';
import { searchQuery } from '../lib/state.js';
import { CATEGORY_LABELS, CATEGORY_KEYS, STATUS_KEYS } from '../lib/classify.js';
import { STATUS_DISPLAY } from './filter-bar.js';
import { FIXTURES, resetSignals } from '../lib/test-fixtures.js';

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

  it('should have 7 status options including scheduled/completed/offline', () => {
    assert.equal(STATUS_KEYS.length, 7);
    assert.deepEqual(STATUS_KEYS, [
      'all', 'running', 'scheduled', 'completed', 'stopped', 'error', 'offline',
    ]);
  });

  it('STATUS_DISPLAY exports labels for all 7 statuses', () => {
    assert.equal(STATUS_DISPLAY.all, 'All');
    assert.equal(STATUS_DISPLAY.running, 'Running');
    assert.equal(STATUS_DISPLAY.scheduled, 'Scheduled');
    assert.equal(STATUS_DISPLAY.completed, 'Completed');
    assert.equal(STATUS_DISPLAY.stopped, 'Stopped');
    assert.equal(STATUS_DISPLAY.error, 'Error');
    assert.equal(STATUS_DISPLAY.offline, 'Offline');
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

  it('onlyMine=true forces categoryFilter to "mine" (via effect in state.js)', () => {
    jobs.value = FIXTURES;

    categoryFilter.value = 'system';
    assert.equal(categoryFilter.value, 'system');

    onlyMine.value = true;
    assert.equal(categoryFilter.value, 'mine');
  });

  it('turning onlyMine OFF resets categoryFilter to "all"', () => {
    jobs.value = FIXTURES;

    onlyMine.value = true;
    assert.equal(categoryFilter.value, 'mine');

    onlyMine.value = false;
    assert.equal(categoryFilter.value, 'all');
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
