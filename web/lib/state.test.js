import { describe, it, beforeEach } from 'node:test';
import { strict as assert } from 'node:assert';
import {
  filteredJobs,
  categoryFilter,
  statusFilter,
  onlyMine,
  categoryCounts,
  statusCounts,
} from './state.js';
import { jobs, searchQuery } from './state.js';
import { FIXTURES, resetSignals } from './test-fixtures.js';

describe('filteredJobs', () => {
  beforeEach(resetSignals);

  it('returns all jobs when no filter is active', () => {
    jobs.value = FIXTURES;
    assert.equal(filteredJobs.value.length, FIXTURES.length);
    assert.deepEqual(filteredJobs.value, FIXTURES);
  });

  it('categoryFilter="mine" → only user-domain non-apple jobs', () => {
    jobs.value = FIXTURES;
    categoryFilter.value = 'mine';
    const labels = filteredJobs.value.map(j => j.label);
    assert.deepEqual(labels, [
      'com.example.myapp',
      'org.homebrew.mxcl.redis',
      'com.myco.agent',
    ]);
  });

  it('categoryFilter="system" → only com.apple.* jobs', () => {
    jobs.value = FIXTURES;
    categoryFilter.value = 'system';
    const labels = filteredJobs.value.map(j => j.label);
    assert.deepEqual(labels, [
      'com.apple.spotlight',
      'com.apple.WindowServer',
    ]);
  });

  it('categoryFilter="thirdparty" → only global-domain non-apple jobs', () => {
    jobs.value = FIXTURES;
    categoryFilter.value = 'thirdparty';
    const labels = filteredJobs.value.map(j => j.label);
    assert.deepEqual(labels, [
      'com.docker.vmnetd',
      'com.microsoft.autoupdate',
    ]);
  });

  it('statusFilter="running" → only running jobs', () => {
    jobs.value = FIXTURES;
    statusFilter.value = 'running';
    const labels = filteredJobs.value.map(j => j.label);
    assert.deepEqual(labels, [
      'com.apple.spotlight',
      'com.apple.WindowServer',
      'org.homebrew.mxcl.redis',
    ]);
  });

  it('statusFilter="error" → only error jobs', () => {
    jobs.value = FIXTURES;
    statusFilter.value = 'error';
    const labels = filteredJobs.value.map(j => j.label);
    assert.deepEqual(labels, [
      'com.docker.vmnetd',
      'com.myco.agent',
    ]);
  });

  it('onlyMine=true → only user-domain non-apple jobs', () => {
    jobs.value = FIXTURES;
    onlyMine.value = true;
    const labels = filteredJobs.value.map(j => j.label);
    assert.deepEqual(labels, [
      'com.example.myapp',
      'org.homebrew.mxcl.redis',
      'com.myco.agent',
    ]);
  });

  it('combined: onlyMine + statusFilter="error" → user error jobs only', () => {
    jobs.value = FIXTURES;
    onlyMine.value = true;
    statusFilter.value = 'error';
    const labels = filteredJobs.value.map(j => j.label);
    assert.deepEqual(labels, ['com.myco.agent']);
  });

  it('combined: categoryFilter + searchQuery → intersection of both', () => {
    jobs.value = FIXTURES;
    categoryFilter.value = 'system';
    searchQuery.value = 'spotlight';
    const labels = filteredJobs.value.map(j => j.label);
    assert.deepEqual(labels, ['com.apple.spotlight']);
  });

  it('search query preserved across filter changes', () => {
    jobs.value = FIXTURES;
    searchQuery.value = 'com.';

    // All matching 'com.'
    const before = filteredJobs.value.length;
    assert.ok(before > 0);

    // Change category filter — search should still apply
    categoryFilter.value = 'mine';
    const after = filteredJobs.value;
    for (const j of after) {
      assert.ok(j.label.toLowerCase().includes('com.'));
    }

    // Change status filter — search still applies
    statusFilter.value = 'stopped';
    for (const j of filteredJobs.value) {
      assert.ok(j.label.toLowerCase().includes('com.'));
      assert.equal(j.status, 'stopped');
    }
  });
});

describe('categoryCounts', () => {
  beforeEach(resetSignals);

  it('computed from full jobs list, not filteredJobs', () => {
    jobs.value = FIXTURES;
    // Apply a status filter that reduces filteredJobs
    statusFilter.value = 'error';
    // categoryCounts should still reflect ALL jobs
    const counts = categoryCounts.value;
    assert.equal(counts.all, 7);
    assert.equal(counts.mine, 3);      // myapp, redis, myco.agent
    assert.equal(counts.system, 2);    // spotlight, WindowServer
    assert.equal(counts.thirdparty, 2); // docker, microsoft
  });
});

describe('statusCounts', () => {
  beforeEach(resetSignals);

  it('computed from full jobs list, not filteredJobs', () => {
    jobs.value = FIXTURES;
    // Apply a category filter that reduces filteredJobs
    categoryFilter.value = 'mine';
    // statusCounts should still reflect ALL jobs
    const counts = statusCounts.value;
    assert.equal(counts.all, 7);
    assert.equal(counts.running, 3);   // spotlight, WindowServer, redis
    assert.equal(counts.stopped, 2);   // myapp, microsoft
    assert.equal(counts.error, 2);     // docker, myco.agent
  });
});
