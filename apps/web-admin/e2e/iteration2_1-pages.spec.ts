import { expect, test } from '@playwright/test';

const webBaseURL = process.env.WEB_BASE_URL ?? 'http://127.0.0.1:3000';

async function expectStyledPage(page: import('@playwright/test').Page) {
  const shell = page.getByTestId('styled-page-shell');
  await expect(shell).toBeVisible();
  await expect(shell).toHaveCSS('display', 'grid');
  await expect(page.locator('.page-hero').first()).toHaveCSS('border-radius', /\d+px/);
  await expect(page.locator('.card').first()).toHaveCSS('background-color', 'rgb(255, 255, 255)');
  await expect(page.getByRole('navigation', { name: 'Iteration 2 navigation' })).toHaveCSS('display', 'flex');
  await expect(page.getByRole('button').first()).toHaveCSS('border-radius', /\d+px/);
}

test.describe('Iteration 2.1 Web Admin visual and functional acceptance', () => {
  test('article pack page stays stable when status API returns unregistered payload without defaults', async ({ page }) => {
    const pageErrors: string[] = [];
    page.on('pageerror', error => {
      pageErrors.push(error.message);
    });

    await page.route('**/api/v1/content-packs/article/status', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: { registered: false, default_metrics: null },
          error: null,
          request_id: 'req-article-pack-unregistered',
        }),
      });
    });

    await page.goto(`${webBaseURL}/article-pack`);

    await expect(page.getByRole('heading', { name: 'Article Pack 管理' })).toBeVisible();
    await expect(page.getByRole('status')).toContainText('req-article-pack-unregistered');
    await expect(page.getByText('content_pack_id=未生成')).toBeVisible();
    await expect(page.getByText('0 项')).toBeVisible();
    await expect(page.getByText('默认指标将在注册完成后显示。')).toBeVisible();
    expect(pageErrors).toEqual([]);
  });

  test('schedule page renders styled AppLayout and supports create, toggle, test-run, details, filters, and pagination', async ({ page }) => {
    await page.goto(`${webBaseURL}/workflow/schedules`, { waitUntil: 'domcontentloaded' });
    await page.waitForResponse((response) => new URL(response.url()).pathname === '/api/v1/workflow-schedules' && response.status() === 200);
    await expect(page.getByRole('heading', { name: '生产计划 / 调度管理' })).toBeVisible();
    await expectStyledPage(page);

    await page.getByLabel('状态筛选').selectOption('enabled');
    await expect(page.getByText(/第 \d+ \/ \d+ 页/)).toBeVisible();

    await page.getByRole('button', { name: '新建调度' }).click();
    await expect(page.getByRole('dialog', { name: '新建调度弹窗' })).toBeVisible();
    await page.getByLabel('daily_content_count').fill('6');
    const createResponse = page.waitForResponse((response) => new URL(response.url()).pathname === '/api/v1/workflow-schedules' && response.request().method() === 'POST');
    await page.getByRole('button', { name: '提交调度' }).click();
    await createResponse;
    await expect(page.getByRole('status')).toContainText('创建成功');

    const testRunResponse = page.waitForResponse((response) => new URL(response.url()).pathname.includes('/test-run') && response.status() < 500);
    await page.getByRole('button', { name: '试跑' }).click();
    await testRunResponse;
    await expect(page.getByRole('status')).toContainText('试跑已提交');

    const toggleResponse = page.waitForResponse((response) => /\/api\/v1\/workflow-schedules\/[^/]+\/(enable|disable)$/.test(new URL(response.url()).pathname));
    await page.getByRole('button', { name: /停用|启用/ }).first().click();
    await toggleResponse;
    await expect(page.getByRole('status')).toContainText('状态已更新');

    await page.getByRole('button', { name: '查看详情' }).first().click();
    await expect(page.getByTestId('schedule-detail')).toBeVisible();
    await expect(page.locator('.badge').first()).toBeVisible();
    await page.screenshot({ path: 'test-results/iteration2_1-schedules-visual.png', fullPage: true });
  });

  test('n8n page renders styled forms and supports provider and binding submission without plaintext token', async ({ page }) => {
    const providerSecret = `sk-live-red-contract-secret-${Date.now()}`;
    await page.goto(`${webBaseURL}/external-automation/n8n`, { waitUntil: 'domcontentloaded' });
    await page.waitForResponse((response) => new URL(response.url()).pathname === '/api/v1/external-automation/providers' && response.status() === 200);
    await expect(page.getByRole('heading', { name: '外部自动化 / n8n' })).toBeVisible();
    await expectStyledPage(page);

    await page.getByLabel('Base URL').fill(`https://n8n-${Date.now()}.example.invalid`);
    await page.getByLabel('Token').fill(providerSecret);
    const providerResponse = page.waitForResponse((response) => new URL(response.url()).pathname === '/api/v1/external-automation/providers' && response.request().method() === 'POST');
    await page.getByRole('button', { name: '保存 Provider' }).click();
    await providerResponse;
    await expect(page.getByRole('status')).toContainText('Provider 创建成功');
    await expect(page.locator('body')).not.toContainText(providerSecret);
    await expect(page.locator('.badge').filter({ hasText: /\*{2,}/ }).first()).toBeVisible();

    await page.getByLabel('事件筛选').selectOption('workflow_run.completed');
    const bindingResponse = page.waitForResponse((response) => new URL(response.url()).pathname === '/api/v1/external-automation/bindings' && response.request().method() === 'POST');
    await page.getByRole('button', { name: '保存 Binding' }).click();
    await bindingResponse;
    await expect(page.getByRole('status')).toContainText('Binding 创建成功');
    await expect(page.locator('.badge').first()).toBeVisible();
    await page.screenshot({ path: 'test-results/iteration2_1-n8n-visual.png', fullPage: true });
  });

  test('cost summary page renders styled metrics and supports refresh, filtering, and detail jump', async ({ page }) => {
    await page.goto(`${webBaseURL}/llm/cost-summary`, { waitUntil: 'domcontentloaded' });
    await page.waitForResponse((response) => new URL(response.url()).pathname === '/api/v1/llm-call-logs/summary' && response.status() === 200);
    await expect(page.getByRole('heading', { name: '成本汇总' })).toBeVisible();
    await expectStyledPage(page);
    await expect(page.locator('.metric').first()).toBeVisible();

    const refreshResponse = page.waitForResponse((response) => new URL(response.url()).pathname === '/api/v1/llm-call-logs/summary' && response.status() === 200);
    await page.getByRole('button', { name: '刷新汇总' }).click();
    await refreshResponse;
    await expect(page.getByRole('status')).toContainText('刷新完成');

    const options = await page.getByLabel('模型筛选').locator('option').count();
    if (options > 1) {
      await page.getByLabel('模型筛选').selectOption({ index: 1 });
    }
    await page.getByRole('button', { name: '查看详情' }).first().click();
    await expect(page.getByTestId('cost-detail')).toBeVisible();
    await page.screenshot({ path: 'test-results/iteration2_1-cost-visual.png', fullPage: true });
  });
});
