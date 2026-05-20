import { expect, test } from '@playwright/test';

const webBaseURL = process.env.WEB_BASE_URL ?? 'http://127.0.0.1:3000';

test('project workspace exposes knowledge memory navigation and memory page states', async ({ page }) => {
  await page.goto(`${webBaseURL}/projects/seed-project/planning`);
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/api/v1/projects/seed-project/knowledge-memory')),
    page.getByRole('link', { name: '记忆上下文' }).click(),
  ]);
  await expect(page.getByRole('heading', { name: '记忆上下文' })).toBeVisible();
  await expect(page.getByText('RecentContentWindow')).toBeVisible();
  await page.getByLabel('内容单元筛选').fill('content-item-1');
  await Promise.all([
    page.waitForResponse(response => response.url().includes('content_item_id=content-item-1') && response.url().includes('sort=created_at') && response.url().includes('order=desc')),
    page.getByRole('button', { name: '筛选快照' }).click(),
  ]);
  await Promise.all([
    page.waitForResponse(response => response.url().includes('page=2')),
    page.getByRole('button', { name: '下一页' }).click(),
  ]);
  await expect(page.getByRole('button', { name: '纠偏 DynamicState' })).toBeVisible();
});

test('context preview supports preview without persistence and snapshot generation', async ({ page }) => {
  await page.goto(`${webBaseURL}/projects/seed-project/planning`);
  await page.getByRole('link', { name: '上下文预览' }).click();
  await page.getByLabel('用途').fill('draft_generation');
  await page.getByLabel('Token 预算').fill('1800');
  await page.getByLabel('内容单元 ID').fill('content-item-1');
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/knowledge-memory/context-preview') && response.url().includes('purpose=draft_generation') && response.url().includes('budget=1800') && response.url().includes('content_item_id=content-item-1')),
    page.getByRole('button', { name: '预览上下文' }).click(),
  ]);
  await expect(page.getByText('未落库')).toBeVisible();
  await Promise.all([
    page.waitForResponse(async response => {
      const request = response.request();
      if (!response.url().includes('/knowledge-memory/assemble-context')) return false;
      const body = request.postDataJSON() as { purpose?: string; budget?: number; content_item_id?: string };
      return body.purpose === 'draft_generation' && body.budget === 1800 && body.content_item_id === 'content-item-1';
    }),
    page.getByRole('button', { name: '生成上下文快照' }).click(),
  ]);
  await expect(page.getByText('已生成上下文快照')).toBeVisible();
});

test('consistency report list and detail expose structured issues and error states', async ({ page }) => {
  await page.goto(`${webBaseURL}/projects/seed-project/planning`);
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/api/v1/projects/seed-project/consistency-reports')),
    page.getByRole('link', { name: '一致性报告' }).click(),
  ]);
  await page.getByLabel('状态筛选').selectOption('completed');
  await page.getByLabel('排序字段').selectOption('created_at');
  await page.getByLabel('排序方向').selectOption('desc');
  await Promise.all([
    page.waitForResponse(response => response.url().includes('status=completed') && response.url().includes('sort=created_at') && response.url().includes('order=desc')),
    page.getByRole('button', { name: '筛选报告' }).click(),
  ]);
  await Promise.all([
    page.waitForResponse(response => response.url().includes('page=2')),
    page.getByRole('button', { name: '下一页' }).click(),
  ]);
  await page.getByRole('button', { name: '创建一致性报告' }).click();
  await expect(page.getByRole('status')).toContainText('报告已创建');
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/api/v1/projects/seed-project/consistency-reports/')),
    page.getByRole('link', { name: '查看详情' }).first().click(),
  ]);
  await expect(page.getByText('来源快照')).toBeVisible();
  await expect(page.getByText('受影响内容')).toBeVisible();

  await page.route('**/api/v1/projects/error-project/knowledge-memory', async route => {
    await route.fulfill({
      status: 500,
      contentType: 'application/json',
      body: JSON.stringify({ success: false, error: { code: 'INTERNAL_ERROR', message: '加载失败' }, request_id: 'req-error-1' }),
    });
  });
  await page.goto(`${webBaseURL}/projects/error-project/memory`);
  const errorAlert = page.getByRole('alert').filter({ hasText: '错误码' });
  await expect(errorAlert).toContainText('错误码');
  await expect(errorAlert).toContainText('错误信息');
  await expect(errorAlert).toContainText('req-error-1');

  await page.route('**/api/v1/projects/bad-json-project/knowledge-memory', async route => {
    await route.fulfill({ status: 502, contentType: 'text/html', body: '<html>bad gateway</html>' });
  });
  await page.goto(`${webBaseURL}/projects/bad-json-project/memory`);
  await expect(page.getByRole('alert').filter({ hasText: 'NETWORK_ERROR' })).toContainText('加载记忆上下文失败');
});
