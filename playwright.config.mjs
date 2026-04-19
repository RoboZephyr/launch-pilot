import { defineConfig } from '@playwright/test';

const PORT = 18080;

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  fullyParallel: false,
  workers: 1,
  retries: 0,
  use: {
    baseURL: `http://127.0.0.1:${PORT}`,
    trace: 'retain-on-failure',
    video: 'retain-on-failure',
  },
  webServer: {
    command: `go run ./cmd/launch-pilot --no-open --port=${PORT}`,
    url: `http://127.0.0.1:${PORT}/`,
    timeout: 60_000,
    reuseExistingServer: !process.env.CI,
    stdout: 'pipe',
    stderr: 'pipe',
  },
  projects: [
    { name: 'chromium', use: { channel: 'chromium' } },
  ],
});
