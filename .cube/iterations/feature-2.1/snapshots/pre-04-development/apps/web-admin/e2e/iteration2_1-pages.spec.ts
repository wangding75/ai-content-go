import { expect, test } from '@playwright/test';

const webBaseURL = process.env.WEB_BASE_URL ?? 'http://127.0.0.1:3000';

const routes = [
  {
    label: '生产计划 / 调度管理',
    path: '/workflow/schedules',
    heading: '生产计划 / 调度管理',
    endpoint: '/api/v1/workflow-schedules',
    action: '试跑',
  },
  {
    label: '外部自动化 / n8n',
    path: '/external-automation/n8n',
    heading: '外部自动化 / n8n',
    endpoint: '/api/v1/external-automation/providers',
    action: '保存 Provider',
  },
  {
    label: '成本汇总',
    path: '/llm/cost-summary',
    heading: '成本汇总',
    endpoint: '/api/v1/llm-call-logs/summary',
    action: '刷新汇总',
  },
];

test.describe('Iteration 2.1 Web Admin pages', () => {
  for (const route of routes) {
    test(`navigates to ${route.path}, loads data, and exposes primary action feedback`, async ({ page }) => {
      await page.goto(webBaseURL, { waitUntil: 'domcontentloaded' });
      const responsePromise = page.waitForResponse((response) => new URL(response.url()).pathname === route.endpoint);
      await page.getByRole('navigation', { name: 'Iteration 2 navigation' }).getByRole('link', { name: route.label }).click();
      await responsePromise;

      await expect(page).toHaveURL(new RegExp(`${route.path}$`));
      await expect(page.getByRole('heading', { name: route.heading })).toBeVisible();
      await expect(page.getByRole('navigation', { name: 'Iteration 2 navigation' }).getByRole('link', { name: route.label })).toHaveAttribute('aria-current', 'page');
      await expect(page.getByRole('button', { name: route.action }).or(page.getByText(route.action))).toBeVisible();
    });
  }

  test('external automation page never renders plaintext provider secret', async ({ page }) => {
    const providerSecret = `sk-live-red-contract-secret-${Date.now()}`;
    await page.goto(`${webBaseURL}/external-automation/n8n`, { waitUntil: 'domcontentloaded' });
    await page.getByLabel('Token').fill(providerSecret);
    const responsePromise = page.waitForResponse((response) => new URL(response.url()).pathname === '/api/v1/external-automation/providers');
    await page.getByRole('button', { name: '保存 Provider' }).click();
    await responsePromise;

    await expect(page.locator('body')).not.toContainText(providerSecret);
    await expect(page.getByText(/\*{2,}/)).toBeVisible();
  });
});
