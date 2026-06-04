import { existsSync } from 'node:fs';
import { defineConfig, devices } from '@playwright/test';

const webBaseURL = process.env.WEB_BASE_URL ?? 'http://127.0.0.1:3000';
const apiBaseURL = process.env.API_BASE_URL ?? 'http://127.0.0.1:18080';
const defaultChromiumPath = '/home/wangding/.cache/ms-playwright/chromium-1223/chrome-linux64/chrome';
const systemChromePath = '/usr/bin/google-chrome';
const chromiumExecutablePath = process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE
  ?? (existsSync(defaultChromiumPath) ? defaultChromiumPath : undefined)
  ?? (existsSync(systemChromePath) ? systemChromePath : undefined);

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  snapshotPathTemplate: '{testDir}/__screenshots__/{testFilePath}/{arg}{ext}',
  expect: {
    toHaveScreenshot: {
      maxDiffPixels: 100,
      threshold: 0.2,
      animations: 'disabled',
    },
    toMatchSnapshot: {
      maxDiffPixels: 100,
    },
  },
  use: {
    ...devices['Desktop Chrome'],
    baseURL: webBaseURL,
    viewport: { width: 1280, height: 720 },
    launchOptions: {
      executablePath: chromiumExecutablePath,
    },
    extraHTTPHeaders: {
      Authorization: 'Bearer dev',
    },
  },
  webServer: [
    {
      command: `HTTP_ADDR=:${new URL(apiBaseURL).port} go run ./apps/api-server/cmd/api`,
      url: `${apiBaseURL}/openapi.yaml`,
      cwd: '../..',
      env: { API_BEARER_TOKEN: 'dev' },
      reuseExistingServer: true,
      timeout: 20_000,
    },
    {
      command: `npm run dev -- --port ${new URL(webBaseURL).port}`,
      url: webBaseURL,
      cwd: '.',
      env: { API_PROXY_TARGET: apiBaseURL },
      reuseExistingServer: true,
      timeout: 20_000,
    },
  ],
});
