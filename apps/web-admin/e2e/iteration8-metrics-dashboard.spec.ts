import { expect, test } from '@playwright/test';

const webBaseURL = process.env.WEB_BASE_URL ?? 'http://127.0.0.1:3000';

// @Test
test('metrics dashboard renders summary records snapshot and navigation entries', async ({ page }) => {
  await page.goto(`${webBaseURL}/projects/seed-project/metrics`);
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/api/v1/projects/seed-project/metrics/summary')),
    page.waitForResponse(response => response.url().includes('/api/v1/metric-records')),
    page.waitForResponse(response => response.url().includes('/api/v1/projects/seed-project/metrics/missing-dates')),
  ]);
  await expect(page.getByRole('heading', { name: '指标表现' })).toBeVisible();
  await expect(page.getByText('summary_snapshot_id')).toBeVisible();
  await expect(page.getByText('source_record_count')).toBeVisible();
  await expect(page.getByRole('link', { name: '趋势图' })).toBeVisible();
  await expect(page.getByRole('link', { name: '缺失提醒' })).toBeVisible();
});

// @Test
test('metric input page creates templates records and batch imports with row errors', async ({ page }) => {
  await page.goto(`${webBaseURL}/projects/seed-project/metrics/input`);
  await page.waitForResponse(response => response.url().includes('/api/v1/metric-templates'));
  await expect(page.getByRole('heading', { name: '指标录入' })).toBeVisible();
  await page.getByLabel('指标编码').fill('views');
  await page.getByLabel('指标名称').fill('阅读量');
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/api/v1/metric-templates') && response.request().method() === 'POST'),
    page.getByRole('button', { name: '创建模板' }).click(),
  ]);
  await expect(page.getByRole('status')).toContainText('模板已创建');
  await page.getByLabel('原始值').fill('1200');
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/api/v1/metric-records') && response.request().method() === 'POST'),
    page.getByRole('button', { name: '保存指标' }).click(),
  ]);
  await expect(page.getByRole('status')).toContainText('指标已保存');
  await page.getByLabel('批量导入').fill('views,2026-05-25,1200\nlikes,bad-date,oops');
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/api/v1/metric-records/batch')),
    page.getByRole('button', { name: '批量导入' }).click(),
  ]);
  await expect(page.getByText('逐条错误')).toBeVisible();
});

// @Test
test('metric trends page displays bucket aggregation signature and missing points', async ({ page }) => {
  await page.goto(`${webBaseURL}/projects/seed-project/metrics/trends`);
  await page.getByLabel('指标编码').fill('views');
  await page.getByLabel('分桶').selectOption('day');
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/api/v1/projects/seed-project/metrics/trends') && response.url().includes('target_id=')),
    page.getByRole('button', { name: '查询趋势' }).click(),
  ]);
  await expect(page.getByRole('heading', { name: '趋势图' })).toBeVisible();
  await expect(page.getByText('aggregation_method')).toBeVisible();
  await expect(page.getByText('query_signature')).toBeVisible();
  await expect(page.getByText('缺失点')).toBeVisible();
});

// @Test
test('missing metrics page lists reasons and deep links to backfill context', async ({ page }) => {
  await page.goto(`${webBaseURL}/projects/seed-project/metrics/missing`);
  await Promise.all([
    page.waitForResponse(response => response.url().includes('/api/v1/projects/seed-project/metrics/missing-dates')),
    page.getByRole('button', { name: '查询缺失' }).click(),
  ]);
  await expect(page.getByRole('heading', { name: '缺失提醒' })).toBeVisible();
  await expect(page.getByText('required_metric_missing')).toBeVisible();
  const backfillLink = page.getByRole('link', { name: '补录' }).first();
  await expect(backfillLink).toHaveAttribute('href', /metrics\/input.*metric_code=.*metric_date=/);
});
