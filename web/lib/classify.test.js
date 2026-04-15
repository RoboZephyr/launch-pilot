import { describe, it } from 'node:test';
import { strict as assert } from 'node:assert';
import { classifyJob, CATEGORY_LABELS, CATEGORY_KEYS, STATUS_KEYS } from './classify.js';

describe('classifyJob', () => {
  it('returns "system" for com.apple.* labels regardless of domain', () => {
    assert.equal(classifyJob({ label: 'com.apple.spotlight', domain: 'user' }), 'system');
    assert.equal(classifyJob({ label: 'com.apple.spotlight', domain: 'global' }), 'system');
    assert.equal(classifyJob({ label: 'com.apple.WindowServer', domain: 'global' }), 'system');
  });

  it('returns "mine" for user-domain non-apple labels', () => {
    assert.equal(classifyJob({ label: 'com.example.myapp', domain: 'user' }), 'mine');
    assert.equal(classifyJob({ label: 'org.homebrew.mxcl.redis', domain: 'user' }), 'mine');
  });

  it('returns "thirdparty" for global-domain non-apple labels', () => {
    assert.equal(classifyJob({ label: 'com.docker.vmnetd', domain: 'global' }), 'thirdparty');
    assert.equal(classifyJob({ label: 'com.microsoft.autoupdate', domain: 'global' }), 'thirdparty');
  });
});

describe('CATEGORY_LABELS', () => {
  it('exports Mine/System/3rd-party display strings', () => {
    assert.equal(CATEGORY_LABELS.mine, 'Mine');
    assert.equal(CATEGORY_LABELS.system, 'System');
    assert.equal(CATEGORY_LABELS.thirdparty, '3rd-party');
  });
});

describe('CATEGORY_KEYS', () => {
  it('has 4 entries: all, mine, system, thirdparty', () => {
    assert.deepEqual(CATEGORY_KEYS, ['all', 'mine', 'system', 'thirdparty']);
  });
});

describe('STATUS_KEYS', () => {
  it('has 4 entries: all, running, stopped, error', () => {
    assert.deepEqual(STATUS_KEYS, ['all', 'running', 'stopped', 'error']);
  });
});
