import { expect, test } from '@playwright/test';

const webBaseURL = process.env.WEB_BASE_URL ?? 'http://127.0.0.1:3000';

// @Test
test('publish queue navigation supports filtering creation and detail transitions', async ({ page }) => {
  await page.goto(`${webBaseURL}/projects/seed-project/planning`);
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/api/v1/projects/seed-project/publish-jobs')),
    page.getByRole('link', { name: '发布队列' }).click(),
  ]);
  await expect(page.getByRole('heading', { name: '发布队列' })).toBeVisible();
  await expect(page.getByLabel('状态筛选')).toBeVisible();
  await expect(page.getByLabel('目标筛选')).toBeVisible();
  await page.getByLabel('状态筛选').selectOption('queued');
  await Promise.all([
    page.waitForResponse(response => response.url().includes('status=queued') && response.url().includes('page=1')),
    page.getByRole('button', { name: '筛选' }).click(),
  ]);
  await page.getByRole('button', { name: '创建发布任务' }).click();
  await expect(page.getByRole('status')).toContainText('发布任务已入队');
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/api/v1/projects/seed-project/publish-jobs/')),
    page.getByRole('link', { name: '详情' }).first().click(),
  ]);
  await expect(page.getByRole('heading', { name: '发布详情' })).toBeVisible();
  await expect(page.getByText('行为摘要')).toBeVisible();
});

// @Test
test('copy preview records copy only after copy button is clicked', async ({ page }) => {
  await page.goto(`${webBaseURL}/publish-jobs/publish-job-1/copy?project_id=seed-project`);
  await expect(page.getByRole('heading', { name: '复制发布内容' })).toBeVisible();
  await expect(page.getByText('content_version_id=')).toBeVisible();
  await expect(page.getByText('payload_hash=')).toBeVisible();
  await Promise.all([
    page.waitForResponse(async response => {
      if (!response.url().includes('/api/v1/projects/seed-project/publish-jobs/publish-job-1/copy')) return false;
      const body = response.request().postDataJSON() as { copy_scope?: string };
      return body.copy_scope === 'full';
    }),
    page.getByRole('button', { name: '复制完整载荷' }).click(),
  ]);
  await expect(page.getByRole('status')).toContainText('复制已记录');
});

// @Test
test('backfill page displays unified error fields and supports requeue action', async ({ page }) => {
  await page.route('**/api/v1/projects/error-project/publish-jobs/job-error/mark-published', async route => {
    await route.fulfill({
      status: 409,
      contentType: 'application/json',
      body: JSON.stringify({ success: false, error: { code: 'CONFLICT', message: '状态不允许' }, request_id: 'req-publish-conflict' }),
    });
  });
  await page.goto(`${webBaseURL}/publish-jobs/job-error/backfill?project_id=error-project`);
  await page.getByLabel('外部链接').fill('https://example.com/post/1');
  await page.getByRole('button', { name: '标记已发布' }).click();
  const errorAlert = page.getByRole('alert').filter({ hasText: 'CONFLICT' });
  await expect(errorAlert).toContainText('状态不允许');
  await expect(errorAlert).toContainText('req-publish-conflict');

  await page.goto(`${webBaseURL}/publish-jobs/publish-job-1/backfill?project_id=seed-project`);
  await page.getByLabel('原因').fill('平台失败后重试');
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/api/v1/projects/seed-project/publish-jobs/publish-job-1/requeue')),
    page.getByRole('button', { name: '重新入队' }).click(),
  ]);
  await expect(page.getByRole('status')).toContainText('已重新入队');
});
