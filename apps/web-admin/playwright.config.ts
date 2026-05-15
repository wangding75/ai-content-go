import { defineConfig, devices } from '@playwright/test';

const webBaseURL = process.env.WEB_BASE_URL ?? 'http://127.0.0.1:3000';
const apiBaseURL = process.env.API_BASE_URL ?? 'http://127.0.0.1:18080';

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  use: {
    ...devices['Desktop Chrome'],
    baseURL: webBaseURL,
    launchOptions: {
      executablePath: process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE ?? '/home/wangding/.cache/ms-playwright/chromium-1223/chrome-linux64/chrome',
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
      env: { NEXT_PUBLIC_API_BASE_URL: apiBaseURL, NEXT_PUBLIC_API_TOKEN: 'dev' },
      reuseExistingServer: true,
      timeout: 20_000,
    },
  ],
});
