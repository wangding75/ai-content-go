import { test, expect } from '@playwright/test';

test('iteration 4 skeleton pages are routable', async ({ page }) => {
  await page.goto('/projects/seed-project/production');
  await expect(page.getByRole('heading', { name: '内容生产' })).toBeVisible();
});
