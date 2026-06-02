import { expect, test, type Locator, type Page } from '@playwright/test';

const webBaseURL = process.env.WEB_BASE_URL ?? 'http://127.0.0.1:3000';

function dynamicTextMasks(page: Page): Locator[] {
  return [
    page.locator('text=/request_id=/'),
    page.locator('text=/req-/'),
    page.locator('text=/summary_snapshot_id=/'),
  ];
}

async function expectStableScreenshot(page: Page, name: string) {
  await expect(page).toHaveScreenshot(name, {
    fullPage: true,
    mask: dynamicTextMasks(page),
  });
}

test.describe('Visual Regression — Core Pages', () => {
  test('dashboard page', async ({ page }) => {
    await page.goto(webBaseURL);
    await page.waitForResponse((response) =>
      response.url().includes('/api/v1/dashboard/summary') && response.status() === 200
    );
    await expectStableScreenshot(page, 'dashboard.png');
  });

  test('projects page', async ({ page }) => {
    await page.goto(`${webBaseURL}/?view=projects`);
    await page.waitForResponse((response) =>
      response.url().includes('/api/v1/projects') && response.status() === 200
    );
    await expectStableScreenshot(page, 'projects.png');
  });

  test('workflow schedules page', async ({ page }) => {
    await page.goto(`${webBaseURL}/workflow/schedules`);
    await page.waitForResponse((response) =>
      response.url().includes('/api/v1/workflow-schedules') && response.status() === 200
    );
    await expectStableScreenshot(page, 'workflow-schedules.png');
  });

  test('metrics dashboard page', async ({ page }) => {
    await page.goto(`${webBaseURL}/projects/seed-project/metrics`);
    await Promise.all([
      page.waitForResponse((response) =>
        response.url().includes('/api/v1/projects/seed-project/metrics/summary') && response.status() === 200
      ),
      page.waitForResponse((response) =>
        response.url().includes('/api/v1/metric-records') && response.status() === 200
      ),
      page.waitForResponse((response) =>
        response.url().includes('/api/v1/projects/seed-project/metrics/missing-dates') && response.status() === 200
      ),
    ]);
    await expectStableScreenshot(page, 'metrics-dashboard.png');
  });

  test('publish queue page', async ({ page }) => {
    await page.goto(`${webBaseURL}/projects/seed-project/planning`);
    await Promise.all([
      page.waitForResponse((response) =>
        response.url().includes('/api/v1/projects/seed-project/publish-jobs') && response.status() === 200
      ),
      page.getByRole('link', { name: '发布队列' }).click(),
    ]);
    await expectStableScreenshot(page, 'publish-queue.png');
  });
});
