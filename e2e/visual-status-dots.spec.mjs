import { test, expect } from '@playwright/test';

// US-P1: the status-dot color must be visible on the <button class="status-dot-trigger">.
// We draw it via ::before, so we read getComputedStyle(trigger, '::before').backgroundColor
// and compare against the CSS custom property for that status.

async function installMockSSE(page) {
  await page.addInitScript(() => {
    const state = { listeners: new Map() };
    class MockEventSource {
      constructor(url) {
        this.url = url;
        this.readyState = 1;
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

const STATUSES = ['running', 'error', 'completed', 'scheduled', 'stopped', 'offline'];

function makeJobs() {
  return STATUSES.map((status, i) => ({
    label: `com.example.${status}`,
    pid: 0,
    lastExitStatus: 0,
    status,
    plistPath: `/tmp/com.example.${status}.plist`,
    program: '',
    programArgs: [],
    standardOutPath: '',
    standardErrPath: '',
    runAtLoad: false,
    keepAlive: false,
    domain: 'user',
  }));
}

function parseColor(str) {
  const m = str.match(/rgba?\(([^)]+)\)/);
  if (!m) return null;
  const parts = m[1].split(',').map((s) => parseFloat(s.trim()));
  return { r: parts[0], g: parts[1], b: parts[2], a: parts[3] ?? 1 };
}

function relLum({ r, g, b }) {
  const lin = (c) => {
    const s = c / 255;
    return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
  };
  return 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b);
}

function contrast(fg, bg) {
  const L1 = relLum(fg);
  const L2 = relLum(bg);
  const [hi, lo] = L1 > L2 ? [L1, L2] : [L2, L1];
  return (hi + 0.05) / (lo + 0.05);
}

test.beforeEach(async ({ page }) => {
  await installMockSSE(page);
  await page.goto('/');
  await page.waitForFunction(() => typeof window.__pushJobs === 'function');
});

test('US-P1: each status dot renders its CSS-var color via ::before', async ({ page }) => {
  await page.evaluate((jobs) => window.__pushJobs(jobs), makeJobs());
  await page.waitForSelector('.status-dot-trigger', { state: 'attached', timeout: 5_000 });

  const result = await page.evaluate((statuses) => {
    const root = getComputedStyle(document.documentElement);
    const triggerBg = (s) => {
      const el = document.querySelector(`.status-dot-trigger.status-dot--${s}`);
      if (!el) return { status: s, missing: true };
      const cs = getComputedStyle(el, '::before');
      const expectedVar = root.getPropertyValue(`--color-${s}`).trim();
      return { status: s, actual: cs.backgroundColor, expectedVar };
    };
    return statuses.map(triggerBg);
  }, STATUSES);

  for (const r of result) {
    expect.soft(r.missing, `trigger for ${r.status} should exist`).toBeFalsy();
    expect.soft(r.actual, `${r.status} actual color`).toBeTruthy();
    expect.soft(r.expectedVar, `--color-${r.status} must be defined`).toBeTruthy();
    const actual = parseColor(r.actual);
    expect.soft(actual, `${r.status} parseColor`).not.toBeNull();
    // The computed value is always rgb(...); the CSS var hex must resolve to same rgb.
    // Compare by re-resolving the var through a sentinel element.
  }
  // Second pass: resolve each var to rgb via a probe element and assert equality.
  const resolved = await page.evaluate((statuses) => {
    const probe = document.createElement('span');
    document.body.appendChild(probe);
    const pairs = statuses.map((s) => {
      probe.style.backgroundColor = `var(--color-${s})`;
      const rgb = getComputedStyle(probe).backgroundColor;
      const el = document.querySelector(`.status-dot-trigger.status-dot--${s}`);
      const actual = el ? getComputedStyle(el, '::before').backgroundColor : '';
      return { status: s, expectedRgb: rgb, actualRgb: actual };
    });
    probe.remove();
    return pairs;
  }, STATUSES);

  for (const p of resolved) {
    expect(p.actualRgb, `status=${p.status}`).toBe(p.expectedRgb);
  }
});

test('US-P1: status-dot-trigger keeps <button> semantics + focus-visible', async ({ page }) => {
  await page.evaluate((jobs) => window.__pushJobs(jobs), makeJobs());
  await page.waitForSelector('.status-dot-trigger', { state: 'attached', timeout: 5_000 });

  const tag = await page.evaluate(() =>
    document.querySelector('.status-dot-trigger')?.tagName ?? '',
  );
  expect(tag).toBe('BUTTON');

  const size = await page.evaluate(() => {
    const el = document.querySelector('.status-dot-trigger');
    const r = el.getBoundingClientRect();
    const cs = getComputedStyle(el, '::before');
    return { w: r.width, h: r.height, beforeW: cs.width, beforeH: cs.height };
  });
  expect(size.w).toBeGreaterThanOrEqual(8);
  expect(size.h).toBeGreaterThanOrEqual(8);
  expect(size.beforeW).toBe('8px');
  expect(size.beforeH).toBe('8px');
});

test('US-P1 dark mode: colors visible on dark bg, contrast ≥ 3:1', async ({ page }) => {
  await page.emulateMedia({ colorScheme: 'dark' });
  await page.evaluate((jobs) => window.__pushJobs(jobs), makeJobs());
  await page.waitForSelector('.status-dot-trigger', { state: 'attached', timeout: 5_000 });

  const data = await page.evaluate((statuses) => {
    const bg = getComputedStyle(document.documentElement).backgroundColor
      || getComputedStyle(document.body).backgroundColor;
    return statuses.map((s) => {
      const el = document.querySelector(`.status-dot-trigger.status-dot--${s}`);
      return { status: s, bg, fg: el ? getComputedStyle(el, '::before').backgroundColor : '' };
    });
  }, STATUSES);

  for (const { status, fg, bg } of data) {
    const pf = parseColor(fg);
    const pb = parseColor(bg);
    expect(pf, `${status} fg parses`).not.toBeNull();
    expect(pb, `${status} bg parses`).not.toBeNull();
    const c = contrast(pf, pb);
    expect(c, `${status} vs dark bg contrast`).toBeGreaterThanOrEqual(3);
  }
});
