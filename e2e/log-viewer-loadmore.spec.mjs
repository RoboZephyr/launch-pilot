import { test, expect } from '@playwright/test';
import { installMockSSE, pushJobs } from './helpers/mock-sse.mjs';

function makeJob(overrides = {}) {
  return {
    label: 'com.example.logs',
    pid: 0,
    lastExitStatus: 0,
    status: 'running',
    plistPath: '/tmp/com.example.logs.plist',
    program: '',
    programArgs: [],
    standardOutPath: '/tmp/logs.out',
    standardErrPath: '',
    runAtLoad: false,
    keepAlive: false,
    domain: 'user',
    ...overrides,
  };
}

function linesOf(n, prefix = 'l') {
  return Array.from({ length: n }, (_, i) => `${prefix}${i}`).join('\n');
}

test.beforeEach(async ({ page }) => {
  await installMockSSE(page);
});

test('US-P4 (a): fewer than requested → no Load more button, shows "Showing all N lines"', async ({ page }) => {
  await page.route('**/api/jobs/**/logs**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        stdoutAvailable: true,
        stdoutPath: '/tmp/logs.out',
        stdout: linesOf(150, 'o'),
        stderrAvailable: false,
        stderrPath: '',
        stderr: '',
        message: '',
      }),
    });
  });
  await page.goto('/');
  await pushJobs(page, [makeJob()]);
  await page.waitForSelector('.job-row', { timeout: 5_000 });

  await page.locator('button.btn:has-text("Logs")').first().click();
  await expect(page.locator('.log-viewer__content').first()).toBeVisible({ timeout: 5_000 });

  await expect(page.locator('.log-viewer__load-more')).toHaveCount(0);
  await expect(page.locator('.log-viewer__done')).toBeVisible();
  await expect(page.locator('.log-viewer__done')).toContainText('Showing all 150 lines');
});

test('US-P4 (b): each request returns full requested → button disappears at cap, "Showing all 10000 lines" appears', async ({ page }) => {
  await page.route('**/api/jobs/**/logs**', async (route) => {
    const url = new URL(route.request().url());
    const n = parseInt(url.searchParams.get('lines') || '200', 10);
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        stdoutAvailable: true,
        stdoutPath: '/tmp/logs.out',
        stdout: linesOf(n, 'o'),
        stderrAvailable: false,
        stderrPath: '',
        stderr: '',
        message: '',
      }),
    });
  });
  await page.goto('/');
  await pushJobs(page, [makeJob()]);
  await page.waitForSelector('.job-row', { timeout: 5_000 });

  await page.locator('button.btn:has-text("Logs")').first().click();
  await expect(page.locator('.log-viewer__load-more')).toBeVisible({ timeout: 5_000 });

  // 200 → 400 → 800 → 1600 → 3200 → 6400 → 10000 (capped): 6 clicks.
  for (let i = 0; i < 10; i++) {
    const btn = page.locator('.log-viewer__load-more');
    if ((await btn.count()) === 0) break;
    await btn.click();
    await page.waitForFunction(() => {
      const b = document.querySelector('.log-viewer__load-more');
      return !b || !b.disabled;
    }, null, { timeout: 10_000 });
  }

  await expect(page.locator('.log-viewer__load-more')).toHaveCount(0, { timeout: 5_000 });
  await expect(page.locator('.log-viewer__done')).toBeVisible();
  await expect(page.locator('.log-viewer__done')).toContainText('Showing all 10000 lines');
});
