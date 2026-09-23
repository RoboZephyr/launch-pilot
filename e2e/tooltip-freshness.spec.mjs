import { test, expect } from '@playwright/test';
import { installMockSSE, pushJobs } from './helpers/mock-sse.mjs';

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
});

test('US-P2: same label field change refreshes tooltip content within 500ms', async ({ page }) => {
  const job = makeJob({ label: 'com.example.foo', status: 'running' });
  await pushJobs(page, [job]);

  await page.waitForSelector('.status-dot-trigger', { state: 'attached', timeout: 5_000 });
  const dot = page.locator('.status-dot-trigger').first();
  const tip = page.locator('#status-tooltip-singleton');

  await dot.hover();
  await expect(tip).toBeVisible({ timeout: 500 });
  await expect(tip).toContainText('running');

  const t0 = await page.evaluate(() => performance.now());
  await pushJobs(page, [{ ...job, status: 'error' }]);

  await expect(tip).toContainText('error', { timeout: 500 });
  const t1 = await page.evaluate(() => performance.now());
  expect(t1 - t0).toBeLessThanOrEqual(500);
  await expect(tip).not.toContainText('running');
});

test('US-P2: label disappearing from snapshot closes the tooltip', async ({ page }) => {
  const job = makeJob({ label: 'com.example.foo', status: 'running' });
  await pushJobs(page, [job]);

  await page.waitForSelector('.status-dot-trigger', { state: 'attached', timeout: 5_000 });
  const dot = page.locator('.status-dot-trigger').first();
  const tip = page.locator('#status-tooltip-singleton');

  await dot.hover();
  await expect(tip).toBeVisible({ timeout: 500 });

  await pushJobs(page, []);
  await expect(tip).toHaveAttribute('aria-hidden', 'true', { timeout: 500 });
});
