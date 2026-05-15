import { expect, test } from '@playwright/test';

const apiBaseURL = process.env.API_BASE_URL ?? 'http://127.0.0.1:18081';
const webBaseURL = process.env.WEB_BASE_URL ?? 'http://127.0.0.1:3001';

const requiredEndpoints = [
  '/api/v1/dashboard/summary',
  '/api/v1/projects',
  '/api/v1/projects/seed-project/overview',
  '/api/v1/projects/seed-project/pause',
  '/api/v1/content-types',
  '/api/v1/content-types/seed-content-type/project-schema',
  '/api/v1/prompt-templates',
  '/api/v1/llm-providers',
];

test.describe('Iteration 1 Web Admin frontend UI roundtrip', () => {
  test('home dashboard, projects, detail shell, and content templates call real backend and handle UI states', async ({ page }) => {
    const apiHits: string[] = [];
    page.on('response', (response) => {
      const url = new URL(response.url());
      if (url.pathname.startsWith('/api/v1/')) {
        apiHits.push(url.pathname);
      }
    });

    await page.goto(webBaseURL);

    await expect(page.getByRole('heading', { name: '首页 / 系统大盘' })).toBeVisible();
    await expect(page.getByTestId('dashboard-loading')).toBeVisible();
    await page.waitForResponse((response) => response.url().includes('/api/v1/dashboard/summary') && response.status() === 200);
    await expect(page.getByTestId('dashboard-loading')).toBeHidden();
    await expect(page.getByTestId('dashboard-project-count')).toContainText(/\d+/);

    await page.getByRole('button', { name: '项目管理' }).click();
    await page.waitForResponse((response) => response.url().includes('/api/v1/projects') && response.status() === 200);
    await expect(page.getByRole('heading', { name: '项目管理' })).toBeVisible();
    await expect(page.getByTestId('projects-empty')).toBeVisible();

    await page.getByRole('button', { name: '新建项目' }).click();
    await page.getByLabel('项目名称').fill('RED roundtrip project');
    await page.getByLabel('项目模板').selectOption('seed-content-type');
    await page.getByRole('button', { name: '提交新建项目' }).click();
    await page.waitForResponse((response) => response.url().includes('/api/v1/projects') && response.request().method() === 'POST' && response.status() === 201);
    await expect(page.getByRole('status')).toContainText('创建成功');

    await page.getByRole('button', { name: '进入项目' }).click();
    await page.waitForResponse((response) => response.url().includes('/api/v1/projects/') && response.url().includes('/overview') && response.status() === 200);
    await expect(page.getByRole('heading', { name: /项目工作区/ })).toBeVisible();
    await expect(page.getByRole('tab', { name: '项目概览' })).toHaveAttribute('aria-selected', 'true');

    await page.getByRole('button', { name: '暂停项目' }).click();
    await page.getByLabel('暂停原因').fill('RED contract requires reason and note');
    await page.getByRole('button', { name: '确认暂停' }).click();
    await page.waitForResponse((response) => response.url().includes('/api/v1/projects/') && response.url().includes('/pause') && response.status() === 200);
    await expect(page.getByRole('status')).toContainText('已暂停');

    await page.getByRole('button', { name: '返回系统' }).click();
    await page.getByRole('button', { name: '项目模板管理' }).click();
    await page.waitForResponse((response) => response.url().includes('/api/v1/content-types') && response.status() === 200);
    await expect(page.getByRole('heading', { name: '项目模板管理' })).toBeVisible();

    await page.getByRole('button', { name: '查看 Schema' }).click();
    await page.waitForResponse((response) => response.url().includes('/api/v1/content-types/') && response.url().includes('/project-schema') && response.status() === 200);
    await expect(page.getByTestId('project-schema')).toContainText('project_schema');

    for (const endpoint of requiredEndpoints.slice(0, 6)) {
      expect(apiHits.some((hit) => hit === endpoint || hit.startsWith(endpoint.replace('seed-project', '')))).toBeTruthy();
    }
  });

  test('home dashboard renders request_id and retry action on backend error', async ({ page }) => {
    await page.route(`${apiBaseURL}/api/v1/dashboard/summary`, async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          data: null,
          error: { code: 'INTERNAL_ERROR', message: 'dashboard unavailable', details: [] },
          request_id: 'req-dashboard-error',
        }),
      });
    });

    await page.goto(webBaseURL);
    const app = page.locator('main');
    await expect(app.getByRole('alert')).toContainText('INTERNAL_ERROR');
    await expect(app.getByRole('alert')).toContainText('req-dashboard-error');
    await expect(page.getByRole('button', { name: '重试' })).toBeVisible();
  });

  test('prompt and provider pages load real data, mutate, and never expose plaintext provider api_key', async ({ page }) => {
    const apiHits: string[] = [];
    page.on('request', (request) => {
      const url = new URL(request.url());
      if (url.pathname.startsWith('/api/v1/')) {
        apiHits.push(`${request.method()} ${url.pathname}`);
      }
    });

    await page.goto(`${webBaseURL}/prompt`);
    await expect(page.getByRole('heading', { name: 'Prompt 模板管理' })).toBeVisible();
    await expect(page.getByTestId('prompt-loading')).toBeVisible();
    await page.waitForResponse((response) => response.url().includes('/api/v1/prompt-templates') && response.status() === 200);
    await expect(page.getByTestId('prompt-loading')).toBeHidden();

    await page.getByRole('button', { name: '新建 Prompt' }).click();
    await page.getByLabel('Prompt Code').fill('outline_red_contract');
    await page.getByLabel('Prompt 内容').fill('生成一个项目大纲：{{topic}}');
    await page.getByRole('button', { name: '提交 Prompt' }).click();
    await page.waitForResponse((response) => response.url().includes('/api/v1/prompt-templates') && response.request().method() === 'POST' && response.status() === 201);
    await expect(page.getByRole('status')).toContainText('Prompt 创建成功');

    await page.goto(`${webBaseURL}/provider`);
    await expect(page.getByRole('heading', { name: '模型 Provider 管理' })).toBeVisible();
    await expect(page.getByTestId('provider-loading')).toBeVisible();
    await page.waitForResponse((response) => response.url().includes('/api/v1/llm-providers') && response.status() === 200);
    await expect(page.getByTestId('provider-loading')).toBeHidden();
    await expect(page.locator('body')).not.toContainText('sk-live-red-contract-secret');
    await expect(page.getByTestId('provider-key-masked')).toContainText(/\*{2,}/);

    await page.getByRole('button', { name: '新增 Provider' }).click();
    await page.getByLabel('Provider 类型').fill('openai_compatible');
    await page.getByLabel('Base URL').fill('https://llm.example.invalid/v1');
    await page.getByLabel('API Key').fill('sk-live-red-contract-secret');
    await page.getByRole('button', { name: '提交 Provider' }).click();
    await page.waitForResponse((response) => response.url().includes('/api/v1/llm-providers') && response.request().method() === 'POST' && response.status() === 201);
    await expect(page.getByRole('status')).toContainText('Provider 创建成功');
    await expect(page.locator('body')).not.toContainText('sk-live-red-contract-secret');

    expect(apiHits).toEqual(expect.arrayContaining([
      'GET /api/v1/prompt-templates',
      'POST /api/v1/prompt-templates',
      'GET /api/v1/llm-providers',
      'POST /api/v1/llm-providers',
    ]));
  });

  test('prompt page renders request_id on validation failure and provider page has empty state', async ({ page }) => {
    await page.goto(`${webBaseURL}/provider?fixture=empty`);
    await expect(page.getByTestId('provider-empty')).toContainText('暂无 Provider');

    await page.goto(`${webBaseURL}/prompt`);
    await page.getByRole('button', { name: '新建 Prompt' }).click();
    await page.getByRole('button', { name: '提交 Prompt' }).click();
    const app = page.locator('main');
    await expect(app.getByRole('alert')).toContainText('VALIDATION_ERROR');
    await expect(app.getByRole('alert')).toContainText(/req-/);
  });
});
