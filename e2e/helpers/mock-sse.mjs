// Install before navigation so tests never connect to the real /api/events.
export async function installMockSSE(page) {
  await page.addInitScript(() => {
    const state = { instance: null, listeners: new Map() };
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
    window.__jobsReady = () => state.instance?.readyState === 1
      && (state.listeners.get('jobs')?.length ?? 0) > 0;
    window.__pushJobs = (jobs) => {
      const evt = { data: JSON.stringify(jobs) };
      for (const cb of state.listeners.get('jobs') || []) cb(evt);
    };
  });
}

export async function pushJobs(page, jobs) {
  // Navigation can finish before the app's effect attaches its jobs listener.
  await page.waitForFunction(() => window.__jobsReady?.());
  await page.evaluate((snapshot) => window.__pushJobs(snapshot), jobs);
}
