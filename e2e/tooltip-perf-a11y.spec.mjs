import { test, expect } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/');
  await page.waitForSelector('.job-table', { timeout: 10_000 });
  await page.waitForSelector('.status-dot-trigger', { state: 'attached', timeout: 10_000 });
});

test('US-P5: enter p95 ≤ 200ms, leave p95 ≤ 100ms over 20 cycles', async ({ page }) => {
  const results = await page.evaluate(async () => {
    const dot = document.querySelector('.status-dot-trigger');
    const tip = document.getElementById('status-tooltip-singleton');
    const waitAria = (val) => new Promise((res) => {
      if (tip.getAttribute('aria-hidden') === val) return res();
      const mo = new MutationObserver(() => {
        if (tip.getAttribute('aria-hidden') === val) { mo.disconnect(); res(); }
      });
      mo.observe(tip, { attributes: true, attributeFilter: ['aria-hidden'] });
    });

    // warm-up
    dot.dispatchEvent(new PointerEvent('pointerenter', { bubbles: true }));
    await waitAria('false');
    dot.dispatchEvent(new PointerEvent('pointerleave', { bubbles: true }));
    await waitAria('true');

    const enter = [];
    const leave = [];
    for (let i = 0; i < 20; i++) {
      const e0 = performance.now();
      dot.dispatchEvent(new PointerEvent('pointerenter', { bubbles: true }));
      await waitAria('false');
      enter.push(performance.now() - e0);

      const l0 = performance.now();
      dot.dispatchEvent(new PointerEvent('pointerleave', { bubbles: true }));
      await waitAria('true');
      leave.push(performance.now() - l0);
    }
    const pct = (arr, p) => {
      const s = [...arr].sort((a, b) => a - b);
      return s[Math.min(s.length - 1, Math.floor(s.length * p))];
    };
    return { enterP95: pct(enter, 0.95), leaveP95: pct(leave, 0.95) };
  });

  expect(results.enterP95).toBeLessThanOrEqual(200);
  expect(results.leaveP95).toBeLessThanOrEqual(100);
});

test('US-P6: focus-then-Esc refocuses the triggering dot', async ({ page }) => {
  const dot = page.locator('.status-dot-trigger').first();
  await dot.focus();
  const tip = page.locator('#status-tooltip-singleton');
  await expect(tip).toBeVisible({ timeout: 500 });

  await page.keyboard.press('Escape');
  await expect(tip).toHaveAttribute('aria-hidden', 'true', { timeout: 300 });

  const sameDot = await page.evaluate(() => {
    const ae = document.activeElement;
    const first = document.querySelector('.status-dot-trigger');
    return ae === first;
  });
  expect(sameDot).toBe(true);
});

test('US-P6: hover-then-Esc does not move activeElement onto the dot', async ({ page }) => {
  // Ensure nothing is focused initially
  await page.evaluate(() => { if (document.activeElement instanceof HTMLElement) document.activeElement.blur(); });

  const dot = page.locator('.status-dot-trigger').first();
  await dot.hover();
  const tip = page.locator('#status-tooltip-singleton');
  await expect(tip).toBeVisible({ timeout: 500 });

  const beforeIsDot = await page.evaluate(() =>
    document.activeElement?.classList.contains('status-dot-trigger') ?? false,
  );
  expect(beforeIsDot).toBe(false);

  await page.keyboard.press('Escape');
  await expect(tip).toHaveAttribute('aria-hidden', 'true', { timeout: 300 });

  const afterIsDot = await page.evaluate(() =>
    document.activeElement?.classList.contains('status-dot-trigger') ?? false,
  );
  expect(afterIsDot).toBe(false);
});
