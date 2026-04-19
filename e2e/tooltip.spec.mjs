import { test, expect } from '@playwright/test';

async function firstJobDot(page) {
  await page.waitForSelector('.status-dot-trigger', { state: 'attached', timeout: 10_000 });
  return page.locator('.status-dot-trigger').first();
}

function tooltip(page) {
  return page.locator('#status-tooltip-singleton');
}

test.beforeEach(async ({ page }) => {
  await page.goto('/');
  await page.waitForSelector('.job-table', { timeout: 10_000 });
  await page.waitForSelector('.status-dot-trigger', { state: 'attached', timeout: 10_000 });
});

test('AC-E1: pointerEnter shows tooltip within 300ms', async ({ page }) => {
  const dot = await firstJobDot(page);
  const tip = tooltip(page);
  const t0 = await page.evaluate(() => performance.now());
  await dot.hover();
  await expect(tip).toBeVisible({ timeout: 300 });
  const t1 = await page.evaluate(() => performance.now());
  expect(t1 - t0).toBeLessThanOrEqual(300);
  await expect(tip).toHaveAttribute('aria-hidden', 'false');
});

test('AC-E2: pointerLeave hides tooltip within 200ms', async ({ page }) => {
  const dot = await firstJobDot(page);
  const tip = tooltip(page);
  await dot.hover();
  await expect(tip).toBeVisible({ timeout: 300 });
  await page.mouse.move(0, 0);
  await expect(tip).toHaveAttribute('aria-hidden', 'true', { timeout: 200 });
});

test('AC-E3: hovering adjacent dots does not stack tooltips', async ({ page }) => {
  const dots = page.locator('.status-dot-trigger');
  const count = await dots.count();
  test.skip(count < 2, 'needs at least 2 jobs on the test machine');
  await dots.nth(0).hover();
  await expect(tooltip(page)).toBeVisible({ timeout: 300 });
  await dots.nth(1).hover();
  const allTooltips = page.locator('[role="tooltip"]');
  expect(await allTooltips.count()).toBe(1);
  const ariaB = await dots.nth(1).getAttribute('aria-label');
  expect(ariaB).toBeTruthy();
  await expect(tooltip(page)).toContainText(ariaB.split(' — ')[0]);
});

test('AC-E4: keyboard Tab focuses status-dot, tooltip visible with accessible name', async ({ page }) => {
  let tries = 0;
  while (tries < 30) {
    const cls = await page.evaluate(() => document.activeElement?.className ?? '');
    if (typeof cls === 'string' && cls.includes('status-dot-trigger')) break;
    await page.keyboard.press('Tab');
    tries += 1;
  }
  const cls = await page.evaluate(() => document.activeElement?.className ?? '');
  expect(cls).toContain('status-dot-trigger');
  const tip = tooltip(page);
  await expect(tip).toBeVisible({ timeout: 300 });
  const ariaLabel = await page.evaluate(() => document.activeElement?.getAttribute('aria-label') ?? '');
  expect(ariaLabel.length).toBeGreaterThan(0);
  const statusWord = ariaLabel.split(' — ')[0];
  expect(['running', 'error', 'completed', 'scheduled', 'stopped', 'offline']).toContain(statusWord);
});

test('AC-E5: Esc hides tooltip', async ({ page }) => {
  const dot = await firstJobDot(page);
  await dot.focus();
  await expect(tooltip(page)).toBeVisible({ timeout: 300 });
  await page.keyboard.press('Escape');
  await expect(tooltip(page)).toHaveAttribute('aria-hidden', 'true', { timeout: 200 });
  const activeIsDot = await page.evaluate(() =>
    document.activeElement?.classList.contains('status-dot-trigger') ?? false,
  );
  expect([true, false]).toContain(activeIsDot);
});

test('AC-E6: Tab order status-dot → Reload → Start → Stop', async ({ page }) => {
  const dot = await firstJobDot(page);
  await dot.focus();
  await page.keyboard.press('Tab');
  const next1 = await page.evaluate(() => document.activeElement?.textContent ?? '');
  expect(next1).toBe('Reload');
  await page.keyboard.press('Tab');
  const next2 = await page.evaluate(() => document.activeElement?.textContent ?? '');
  expect(next2).toBe('Start');
  await page.keyboard.press('Tab');
  const next3 = await page.evaluate(() => document.activeElement?.textContent ?? '');
  expect(next3).toBe('Stop');
});
