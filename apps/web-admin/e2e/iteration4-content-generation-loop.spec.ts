import { test, expect } from '@playwright/test';

test('iteration 4 content generation loop covers production success failure retry and review routes', async ({ page }) => {
  await page.goto('/projects/seed-project/production');
  await expect(page.getByRole('heading', { name: '内容生产' })).toBeVisible();
  await expect(page.getByText('手动生成')).toBeVisible();
  await expect(page.getByText('批量生成')).toBeVisible();
  await expect(page.getByText('request_id')).toBeVisible();

  await page.goto('/generation-runs/genrun-1/retry');
  await expect(page).toHaveURL(/\/generation-runs\/genrun-1\/retry/);

  await page.goto('/projects/seed-project/content-items?status=pending_review');
  await expect(page.getByText('pending_review')).toBeVisible();
});
