import { expect, test } from '@playwright/test';

const WEB = 'http://127.0.0.1:3100';
const SCREENSHOT_DIR = 'test-results/manual-article-admin';

test.describe('Article admin pages — real-browser verification', () => {
  test('Page 1: /article-pack — registered status, metrics, content_pack_id, workflow', async ({ page }) => {
    const pageErrors: string[] = [];
    page.on('pageerror', (err) => pageErrors.push(err.message));

    await page.goto(`${WEB}/article-pack`, { waitUntil: 'networkidle' });
    await page.waitForSelector('[role="status"]', { timeout: 10000 });

    // Verify heading
    await expect(page.getByRole('heading', { name: 'Article Pack 管理' })).toBeVisible();

    // Verify AppLayout shell (styled page check)
    const shell = page.locator('.page-shell');
    await expect(shell).toBeVisible();

    // Verify registered status content
    await expect(page.getByText(/content_pack_id/)).toBeVisible();
    await expect(page.getByText(/cp-content-type-new/)).toBeVisible();

    // Verify 5 metrics
    await expect(page.getByText('avg_read_time')).toBeVisible();
    await expect(page.getByText('comments')).toBeVisible();
    await expect(page.getByText('likes')).toBeVisible();
    await expect(page.getByText('shares')).toBeVisible();
    await expect(page.getByText('views')).toBeVisible();

    // Verify workflow template
    await expect(page.getByText(/Article Generation/)).toBeVisible();

    // Interaction: click refresh
    const refreshBtn = page.getByRole('button', { name: /刷新/ });
    await refreshBtn.click();
    await page.waitForSelector('[role="status"]');
    await expect(page.getByRole('status')).toContainText('状态已刷新');

    // Screenshot
    await page.screenshot({ path: `${SCREENSHOT_DIR}/page1-article-pack.png`, fullPage: true });

    // No page errors
    expect(pageErrors).toEqual([]);
  });

  test('Page 2: /projects/project-1/article — config, generation runs, retry', async ({ page }) => {
    const pageErrors: string[] = [];
    page.on('pageerror', (err) => pageErrors.push(err.message));

    await page.goto(`${WEB}/projects/project-1/article`, { waitUntil: 'networkidle' });
    await page.waitForSelector('[role="status"]', { timeout: 10000 });

    // Verify heading
    await expect(page.getByRole('heading', { name: /Article/ })).toBeVisible();

    // Verify AppLayout
    const shell = page.locator('.page-shell');
    await expect(shell).toBeVisible();

    // Verify config section: value "tech" is populated in topic_style input
    const topicInput = page.locator('input[value="tech"]');
    await expect(topicInput).toBeVisible();

    // Verify generation runs list (at least 1 run)
    await expect(page.getByText(/agr-1/)).toBeVisible();
    await expect(page.getByText(/agr-2/)).toBeVisible();

    // Interaction: click save/resave button
    const saveBtn = page.getByRole('button', { name: /保存配置/ });
    if (await saveBtn.isVisible()) {
      await saveBtn.click();
      await page.waitForResponse(
        (r) => r.url().includes('/article/config') && r.status() < 500,
        { timeout: 10000 },
      );
      await expect(page.getByRole('status')).toContainText(/保存|成功/);
    }

    await page.screenshot({ path: `${SCREENSHOT_DIR}/page2-project-article.png`, fullPage: true });
    expect(pageErrors).toEqual([]);
  });

  test('Page 3: /projects/project-1/article/metrics — metrics list with enabled codes', async ({ page }) => {
    const pageErrors: string[] = [];
    page.on('pageerror', (err) => pageErrors.push(err.message));

    await page.goto(`${WEB}/projects/project-1/article/metrics`, { waitUntil: 'networkidle' });
    await page.waitForSelector('[role="status"]', { timeout: 10000 });

    // Verify heading (use first() to avoid strict mode: h1 + h2 both match /指标/)
    await expect(page.getByRole('heading', { name: /指标/ }).first()).toBeVisible();

    // Verify AppLayout
    const shell = page.locator('.page-shell');
    await expect(shell).toBeVisible();

    // Verify metrics table has views and likes rows showing enabled status
    const metricsTable = page.locator('table').last();
    await expect(metricsTable.locator('tr', { hasText: 'views' })).toContainText('启用');
    await expect(metricsTable.locator('tr', { hasText: 'likes' })).toContainText('启用');

    // Interaction: save or refresh
    const refreshBtn = page.getByRole('button', { name: /刷新|保存/ }).first();
    if (await refreshBtn.isVisible()) {
      await refreshBtn.click();
      await page.waitForResponse(
        (r) => r.url().includes('/article/metrics') && r.status() < 500,
        { timeout: 10000 },
      );
    }

    await page.screenshot({ path: `${SCREENSHOT_DIR}/page3-project-article-metrics.png`, fullPage: true });
    expect(pageErrors).toEqual([]);
  });
});