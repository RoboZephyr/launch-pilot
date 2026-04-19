import { describe, it } from 'node:test';
import { strict as assert } from 'node:assert';
import { classifyJob, CATEGORY_LABELS } from '../lib/classify.js';
import { buildStatusTooltip, buildStatusTooltipParts, placeTooltip } from './job-tooltip.js';

/**
 * S04 — Category Badge in Job Rows
 *
 * job-row.js renders badge class and text as:
 *   class = 'category-badge category-badge--' + classifyJob(job)
 *   text  = CATEGORY_LABELS[classifyJob(job)]
 *
 * These tests verify the contract between classify.js output
 * and the expected badge attributes for each job type.
 */

describe('JobRow category badge derivation', () => {
  const badgeClass = (job) => 'category-badge category-badge--' + classifyJob(job);
  const badgeText = (job) => CATEGORY_LABELS[classifyJob(job)];

  it('com.apple.* job → System badge with --system class', () => {
    const job = { label: 'com.apple.spotlight', domain: 'user' };
    assert.equal(badgeClass(job), 'category-badge category-badge--system');
    assert.equal(badgeText(job), 'System');
  });

  it('com.apple.* job in global domain → still System badge', () => {
    const job = { label: 'com.apple.WindowServer', domain: 'global' };
    assert.equal(badgeClass(job), 'category-badge category-badge--system');
    assert.equal(badgeText(job), 'System');
  });

  it('user-domain non-apple job → Mine badge with --mine class', () => {
    const job = { label: 'com.example.myapp', domain: 'user' };
    assert.equal(badgeClass(job), 'category-badge category-badge--mine');
    assert.equal(badgeText(job), 'Mine');
  });

  it('global-domain non-apple job → 3rd-party badge with --thirdparty class', () => {
    const job = { label: 'com.docker.vmnetd', domain: 'global' };
    assert.equal(badgeClass(job), 'category-badge category-badge--thirdparty');
    assert.equal(badgeText(job), '3rd-party');
  });

  it('every category key maps to a defined CATEGORY_LABELS entry', () => {
    // Ensure no undefined badge text for any classifyJob output
    const jobs = [
      { label: 'com.apple.x', domain: 'user' },
      { label: 'com.foo.bar', domain: 'user' },
      { label: 'com.foo.bar', domain: 'global' },
    ];
    for (const job of jobs) {
      const cat = classifyJob(job);
      assert.ok(CATEGORY_LABELS[cat] !== undefined,
        `CATEGORY_LABELS missing key "${cat}" for job ${job.label}`);
    }
  });
});

describe('buildStatusTooltip', () => {
  it('returns just status when no schedule fields are set', () => {
    const tip = buildStatusTooltip({ status: 'running' });
    assert.equal(tip, 'running');
  });

  it('includes Next run when nextRunAt is set', () => {
    const tip = buildStatusTooltip({
      status: 'scheduled',
      nextRunAt: '2026-04-17T09:00:00Z',
    });
    assert.ok(tip.startsWith('scheduled'));
    assert.ok(tip.includes('Next run:'));
  });

  it('includes Last run when lastRunAt is set', () => {
    const tip = buildStatusTooltip({
      status: 'completed',
      lastRunAt: '2026-04-17T08:55:00Z',
    });
    assert.ok(tip.includes('Last run:'));
  });

  it('shows unknown fallback for completed/scheduled without lastRunAt', () => {
    const tipCompleted = buildStatusTooltip({ status: 'completed' });
    assert.ok(tipCompleted.includes('Last run: unknown'));

    const tipScheduled = buildStatusTooltip({ status: 'scheduled' });
    assert.ok(tipScheduled.includes('Last run: unknown'));
  });

  it('does not show unknown fallback for stopped/running/error without lastRunAt', () => {
    assert.ok(!buildStatusTooltip({ status: 'stopped' }).includes('unknown'));
    assert.ok(!buildStatusTooltip({ status: 'running' }).includes('unknown'));
    assert.ok(!buildStatusTooltip({ status: 'error' }).includes('unknown'));
  });

  it('joins parts with em-dash-like separator', () => {
    const tip = buildStatusTooltip({
      status: 'scheduled',
      nextRunAt: '2026-04-17T09:00:00Z',
      lastRunAt: '2026-04-17T08:00:00Z',
    });
    // Three parts: status, Next run, Last run
    assert.equal(tip.split(' — ').length, 3);
  });
});

describe('buildStatusTooltipParts parity with buildStatusTooltip', () => {
  const cases = [
    { name: 'running no schedule',
      job: { status: 'running' } },
    { name: 'error with nextRunAt',
      job: { status: 'error', nextRunAt: '2026-04-17T09:00:00Z' } },
    { name: 'completed with lastRunAt',
      job: { status: 'completed', lastRunAt: '2026-04-17T08:00:00Z' } },
    { name: 'scheduled no lastRun with log path',
      job: { status: 'scheduled', nextRunAt: '2026-04-18T10:00:00Z',
             standardOutPath: '/tmp/a.log' } },
    { name: 'scheduled no lastRun no log path',
      job: { status: 'scheduled', nextRunAt: '2026-04-18T10:00:00Z' } },
    { name: 'completed no lastRun no log path',
      job: { status: 'completed' } },
    { name: 'stopped',
      job: { status: 'stopped' } },
    { name: 'offline with lastRunAt',
      job: { status: 'offline', lastRunAt: '2026-04-15T08:00:00Z' } },
  ];
  for (const { name, job } of cases) {
    it(`${name}: parts.join(' — ') === buildStatusTooltip`, () => {
      assert.equal(buildStatusTooltipParts(job).join(' — '), buildStatusTooltip(job));
    });
    it(`${name}: parts is a non-empty array of strings`, () => {
      const parts = buildStatusTooltipParts(job);
      assert.ok(Array.isArray(parts));
      assert.ok(parts.length >= 1);
      for (const p of parts) assert.equal(typeof p, 'string');
    });
  }
});

describe('placeTooltip', () => {
  const vp = { width: 1024, height: 768 };
  it('above anchor when top-space available', () => {
    const anchor = { top: 500, left: 500, width: 8, height: 8, bottom: 508, right: 508 };
    const tip = { width: 200, height: 40 };
    const pos = placeTooltip(anchor, tip, vp);
    assert.ok(pos.top < 500); // above
    assert.equal(pos.left, 500 + 4 - 100); // anchor center - tip width/2
  });
  it('flips below when top-space insufficient', () => {
    const anchor = { top: 4, left: 500, width: 8, height: 8, bottom: 12, right: 508 };
    const tip = { width: 200, height: 40 };
    const pos = placeTooltip(anchor, tip, vp);
    assert.ok(pos.top > 12); // below
  });
  it('clamps left edge to 4', () => {
    const anchor = { top: 500, left: 0, width: 8, height: 8, bottom: 508, right: 8 };
    const tip = { width: 200, height: 40 };
    const pos = placeTooltip(anchor, tip, vp);
    assert.equal(pos.left, 4);
  });
  it('clamps right edge to viewportW - tipW - 4', () => {
    const anchor = { top: 500, left: 1020, width: 8, height: 8, bottom: 508, right: 1028 };
    const tip = { width: 200, height: 40 };
    const pos = placeTooltip(anchor, tip, vp);
    assert.equal(pos.left, 1024 - 200 - 4);
  });
  it('centers horizontally over small anchor', () => {
    const anchor = { top: 500, left: 500, width: 8, height: 8, bottom: 508, right: 508 };
    const tip = { width: 100, height: 30 };
    const pos = placeTooltip(anchor, tip, vp);
    assert.equal(pos.left, 504 - 50);
  });
});
