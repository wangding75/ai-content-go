import { test, expect } from '@playwright/test';

const BASE = 'http://localhost:13000';

test.describe('Portfolio Pages', () => {
  test('list page renders with AppLayout and portfolio elements', async ({ page }) => {
    await page.goto(`${BASE}/portfolios`);
    await page.waitForLoadState('networkidle');
    // AppLayout: nav exists
    await expect(page.locator('nav')).toBeVisible({ timeout: 10000 });
    // Page heading
    const heading = page.locator('h1, h2');
    await expect(heading.first()).toBeVisible();
    await page.screenshot({ path: 'e2e/__screenshots__/portfolio-list.png', fullPage: true });
  });

  test('detail page renders with portfolioId route param', async ({ page }) => {
    await page.goto(`${BASE}/portfolios/pf-test1`);
    await page.waitForLoadState('networkidle');
    await expect(page.locator('nav')).toBeVisible({ timeout: 10000 });
    await page.screenshot({ path: 'e2e/__screenshots__/portfolio-detail.png', fullPage: true });
  });

  test('projects page renders', async ({ page }) => {
    await page.goto(`${BASE}/portfolios/pf-test1/projects`);
    await page.waitForLoadState('networkidle');
    await expect(page.locator('nav')).toBeVisible({ timeout: 10000 });
    await page.screenshot({ path: 'e2e/__screenshots__/portfolio-projects.png', fullPage: true });
  });

  test('health page renders', async ({ page }) => {
    await page.goto(`${BASE}/portfolios/pf-test1/health`);
    await page.waitForLoadState('networkidle');
    await expect(page.locator('nav')).toBeVisible({ timeout: 10000 });
    await page.screenshot({ path: 'e2e/__screenshots__/portfolio-health.png', fullPage: true });
  });

  test('global nav has Portfolio entry', async ({ page }) => {
    await page.goto(`${BASE}/`);
    await page.waitForLoadState('networkidle');
    const navLink = page.locator('a[href="/portfolios"], a[href="/portfolios/"]');
    await expect(navLink).toBeVisible({ timeout: 10000 });
    await page.screenshot({ path: 'e2e/__screenshots__/portfolio-nav.png', fullPage: true });
  });
});
