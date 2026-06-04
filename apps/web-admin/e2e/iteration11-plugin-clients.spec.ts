import { expect, test } from '@playwright/test';

const webBaseURL = process.env.WEB_BASE_URL ?? 'http://127.0.0.1:3000';
const listRoute = '**/api/v1/plugin-clients?page=1&page_size=20';

const listEnvelope = (items: unknown, requestId: string) => ({
  success: true,
  data: {
    items,
    pagination: { page: 1, page_size: 20, total: 0, has_next: false },
  },
  error: null,
  request_id: requestId,
});

function trackPageErrors(page: import('@playwright/test').Page) {
  const errors: string[] = [];
  page.on('pageerror', error => {
    errors.push(error.message);
  });
  return errors;
}

// @Test
test('plugin clients page renders empty state when API returns null items', async ({ page }) => {
  const pageErrors = trackPageErrors(page);

  await page.route(listRoute, async route => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(listEnvelope(null, 'req-list-null-initial')),
    });
  });

  await page.goto(`${webBaseURL}/plugin-clients`);

  await expect(page.getByRole('heading', { name: '插件客户端管理' })).toBeVisible();
  await expect(page.getByText('暂无')).toBeVisible();
  expect(pageErrors).toEqual([]);
});

// @Test
test('plugin clients page stays stable when create refresh returns null items', async ({ page }) => {
  const pageErrors = trackPageErrors(page);
  let listCount = 0;

  await page.route(listRoute, async route => {
    listCount += 1;
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(listEnvelope(listCount === 1 ? [] : null, `req-list-${listCount}`)),
    });
  });

  await page.route('**/api/v1/plugin-clients', async route => {
    const request = route.request();
    expect(request.method()).toBe('POST');
    expect(JSON.parse(request.postData() ?? '{}')).toEqual({
      name: 'client-a',
      client_type: 'chrome_extension',
      scopes: ['publish'],
    });

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        success: true,
        data: { client_id: 'client-1', api_key_masked: 'pk_once_123' },
        error: null,
        request_id: 'req-create-1',
      }),
    });
  });

  await page.goto(`${webBaseURL}/plugin-clients`);
  await page.getByRole('button', { name: '注册客户端' }).click();
  await page.getByLabel('名称').fill('client-a');
  await page.getByLabel('客户端类型').fill('chrome_extension');
  await page.getByLabel('权限范围').fill('publish');
  await page.getByRole('button', { name: '注册', exact: true }).click();

  await expect(page.getByRole('heading', { name: '插件客户端管理' })).toBeVisible();
  await expect(page.getByText('api_key_once: pk_once_123')).toBeVisible();
  await expect(page.getByText('暂无')).toBeVisible();
  expect(listCount).toBe(2);
  expect(pageErrors).toEqual([]);
});

// @Test
test('plugin clients page stays stable when update refresh returns null items', async ({ page }) => {
  const pageErrors = trackPageErrors(page);
  let listCount = 0;

  await page.route(listRoute, async route => {
    listCount += 1;
    const firstList = [
      {
        id: 'client-1',
        name: 'client-a',
        client_type: 'chrome_extension',
        version: 'v1',
        scopes: ['publish'],
        status: 'active',
        api_key_masked: 'pk_live_123',
      },
    ];

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(listEnvelope(listCount === 1 ? firstList : null, `req-update-list-${listCount}`)),
    });
  });

  await page.route('**/api/v1/plugin-clients/client-1', async route => {
    const request = route.request();
    expect(request.method()).toBe('PATCH');
    expect(JSON.parse(request.postData() ?? '{}')).toEqual({
      name: 'client-a-updated',
      client_type: 'chrome_extension',
      scopes: ['publish', 'metrics'],
      status: 'disabled',
    });

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        success: true,
        data: {
          client_id: 'client-1',
          api_key_masked: 'pk_live_123',
          operation_log_id: 'op-1',
        },
        error: null,
        request_id: 'req-update-1',
      }),
    });
  });

  await page.goto(`${webBaseURL}/plugin-clients`);
  await expect(page.getByText('client-a (chrome_extension) - active')).toBeVisible();

  await page.getByRole('button', { name: '编辑' }).click();
  await page.getByLabel('名称').fill('client-a-updated');
  await page.getByLabel('客户端类型').fill('chrome_extension');
  await page.getByLabel('权限范围').fill('publish,metrics');
  await page.getByLabel('状态').fill('disabled');
  await page.getByRole('button', { name: '保存' }).click();

  await expect(page.getByText('暂无')).toBeVisible();
  expect(listCount).toBe(2);
  expect(pageErrors).toEqual([]);
});

// @Test
test('plugin clients page rotates key without runtime errors', async ({ page }) => {
  const pageErrors = trackPageErrors(page);

  await page.route(listRoute, async route => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(listEnvelope([
        {
          id: 'client-1',
          name: 'client-a',
          client_type: 'chrome_extension',
          version: 'v1',
          scopes: ['publish'],
          status: 'active',
          api_key_masked: 'pk_live_123',
        },
      ], 'req-rotate-list-1')),
    });
  });

  await page.route('**/api/v1/plugin-clients/client-1/rotate-key', async route => {
    const request = route.request();
    expect(request.method()).toBe('POST');

    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        success: true,
        data: {
          client_id: 'client-1',
          api_key_masked: 'pk_rotated_123',
          operation_log_id: 'op-rotate-1',
        },
        error: null,
        request_id: 'req-rotate-1',
      }),
    });
  });

  await page.goto(`${webBaseURL}/plugin-clients`);
  await page.getByRole('button', { name: '轮换密钥' }).click();

  await expect(page.getByText('api_key_once: pk_rotated_123')).toBeVisible();
  expect(pageErrors).toEqual([]);
});
