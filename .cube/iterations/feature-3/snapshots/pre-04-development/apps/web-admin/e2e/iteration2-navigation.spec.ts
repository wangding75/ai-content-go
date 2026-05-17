import { expect, test } from '@playwright/test';

const webBaseURL = process.env.WEB_BASE_URL ?? 'http://127.0.0.1:3000';

const routes = [
  {
    label: '工作流模板',
    path: '/workflow/templates',
    heading: 'Workflow Templates',
    endpoint: '/api/v1/workflow-templates',
    emptyTestId: 'workflow-templates-empty',
  },
  {
    label: '运行记录',
    path: '/workflow/runs',
    heading: 'Workflow Runs',
    endpoint: '/api/v1/workflow-runs',
    emptyTestId: 'workflow-runs-empty',
  },
  {
    label: 'Agent 管理',
    path: '/agent/tasks',
    heading: 'Agent Tasks',
    endpoint: '/api/v1/agent-tasks',
    emptyTestId: 'agent-tasks-empty',
  },
  {
    label: 'LLM Logs',
    path: '/llm/logs',
    heading: 'LLM Call Logs',
    endpoint: '/api/v1/llm-call-logs',
    emptyTestId: 'llm-logs-empty',
  },
];

test.describe('Iteration 2 Web Admin navigation', () => {
  for (const route of routes) {
    test(`homepage links to ${route.path} with active navigation and data state`, async ({ page }) => {
      const apiHits: string[] = [];
      page.on('request', (request) => {
        const url = new URL(request.url());
        if (url.pathname.startsWith('/api/v1/')) {
          apiHits.push(url.pathname);
        }
      });

      await page.goto(webBaseURL, { waitUntil: 'domcontentloaded' });
      const responsePromise = page.waitForResponse((response) => new URL(response.url()).pathname === route.endpoint && response.status() === 200);
      await page.getByRole('navigation', { name: 'Iteration 2 navigation' }).getByRole('link', { name: route.label }).click();
      await responsePromise;

      await expect(page).toHaveURL(new RegExp(`${route.path}$`));
      await expect(page.getByRole('heading', { name: route.heading })).toBeVisible();
      await expect(page.getByRole('navigation', { name: 'Iteration 2 navigation' }).getByRole('link', { name: route.label })).toHaveAttribute('aria-current', 'page');
      await expect(page.getByTestId(route.emptyTestId).or(page.locator('tbody tr').first())).toBeVisible();
      expect(apiHits).toContain(route.endpoint);
    });
  }

  test('Iteration 2 routes are directly refreshable and keep highlighted navigation', async ({ page }) => {
    for (const route of routes) {
      await page.goto(`${webBaseURL}${route.path}`, { waitUntil: 'domcontentloaded' });
      await page.reload({ waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: route.heading })).toBeVisible();
      await expect(page.getByRole('navigation', { name: 'Iteration 2 navigation' }).getByRole('link', { name: route.label })).toHaveAttribute('aria-current', 'page');
    }
  });
});
