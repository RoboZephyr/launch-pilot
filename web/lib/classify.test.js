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

  it('returns "thirdparty" for empty/missing domain (plist-less jobs)', () => {
    assert.equal(classifyJob({ label: 'com.example.noplist', domain: '' }), 'thirdparty');
    assert.equal(classifyJob({ label: 'com.example.noplist', domain: undefined }), 'thirdparty');
  });

  it('returns "system" for com.apple.* even with empty domain', () => {
    assert.equal(classifyJob({ label: 'com.apple.cfprefsd', domain: '' }), 'system');
  });

  it('returns "thirdparty" for missing/null label without crashing', () => {
    assert.equal(classifyJob({ domain: 'user' }), 'thirdparty');
    assert.equal(classifyJob({ label: null, domain: 'user' }), 'thirdparty');
    assert.equal(classifyJob({ label: '', domain: 'user' }), 'thirdparty');
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
  it('has 7 entries: all, running, scheduled, completed, stopped, error, offline', () => {
    assert.deepEqual(STATUS_KEYS, [
      'all', 'running', 'scheduled', 'completed', 'stopped', 'error', 'offline',
    ]);
    assert.equal(STATUS_KEYS.length, 7);
  });

  it('includes scheduled, completed, and offline', () => {
    assert.ok(STATUS_KEYS.includes('scheduled'));
    assert.ok(STATUS_KEYS.includes('completed'));
    assert.ok(STATUS_KEYS.includes('offline'));
  });
});
