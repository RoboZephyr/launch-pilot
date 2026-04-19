import { test, expect } from '@playwright/test';

// Inject a mock EventSource BEFORE any app script runs so that we control
// every `event: jobs` snapshot pushed to the browser. The real SSE at
// /api/events is never contacted.
async function installMockSSE(page) {
  await page.addInitScript(() => {
    const state = {
      instance: null,
      listeners: new Map(),
    };
    class MockEventSource {
      constructor(url) {
        this.url = url;
        this.readyState = 1;
        state.instance = this;
        state.listeners = new Map();
      }
      addEventListener(event, cb) {
        if (!state.listeners.has(event)) state.listeners.set(event, []);
        state.listeners.get(event).push(cb);
      }
      removeEventListener(event, cb) {
        const arr = state.listeners.get(event);
        if (!arr) return;
        const i = arr.indexOf(cb);
        if (i >= 0) arr.splice(i, 1);
      }
      close() { this.readyState = 2; }
    }
    window.EventSource = MockEventSource;
    window.__pushJobs = (arr) => {
      const ls = state.listeners.get('jobs') || [];
      const evt = { data: JSON.stringify(arr) };
      for (const cb of ls) cb(evt);
    };
  });
}

function makeJob(overrides) {
  return {
    label: 'com.example.foo',
    pid: 0,
    lastExitStatus: 0,
    status: 'running',
    plistPath: '/tmp/com.example.foo.plist',
    program: '',
    programArgs: [],
    standardOutPath: '',
    standardErrPath: '',
    runAtLoad: false,
    keepAlive: false,
    domain: 'user',
    ...overrides,
  };
}

test.beforeEach(async ({ page }) => {
  await installMockSSE(page);
  await page.goto('/');
  await page.waitForFunction(() => typeof window.__pushJobs === 'function');
});

test('US-P2: same label field change refreshes tooltip content within 500ms', async ({ page }) => {
  const job = makeJob({ label: 'com.example.foo', status: 'running' });
  await page.evaluate((j) => window.__pushJobs([j]), job);

  await page.waitForSelector('.status-dot-trigger', { state: 'attached', timeout: 5_000 });
  const dot = page.locator('.status-dot-trigger').first();
  const tip = page.locator('#status-tooltip-singleton');

  await dot.hover();
  await expect(tip).toBeVisible({ timeout: 500 });
  await expect(tip).toContainText('running');

  const t0 = await page.evaluate(() => performance.now());
  await page.evaluate((j) => window.__pushJobs([j]), { ...job, status: 'error' });

  await expect(tip).toContainText('error', { timeout: 500 });
  const t1 = await page.evaluate(() => performance.now());
  expect(t1 - t0).toBeLessThanOrEqual(500);
  await expect(tip).not.toContainText('running');
});

test('US-P2: label disappearing from snapshot closes the tooltip', async ({ page }) => {
  const job = makeJob({ label: 'com.example.foo', status: 'running' });
  await page.evaluate((j) => window.__pushJobs([j]), job);

  await page.waitForSelector('.status-dot-trigger', { state: 'attached', timeout: 5_000 });
  const dot = page.locator('.status-dot-trigger').first();
  const tip = page.locator('#status-tooltip-singleton');

  await dot.hover();
  await expect(tip).toBeVisible({ timeout: 500 });

  await page.evaluate(() => window.__pushJobs([]));
  await expect(tip).toHaveAttribute('aria-hidden', 'true', { timeout: 500 });
});
