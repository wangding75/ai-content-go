# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: iteration1-ui.spec.ts >> Iteration 1 Web Admin frontend UI roundtrip >> prompt and provider pages load real data, mutate, and never expose plaintext provider api_key
- Location: e2e/iteration1-ui.spec.ts:95:7

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: page.waitForResponse: Test timeout of 30000ms exceeded.
```

# Page snapshot

```yaml
- generic [ref=e1]:
  - navigation "Iteration 2 navigation" [ref=e2]:
    - link "首页 / 系统大盘" [ref=e3] [cursor=pointer]:
      - /url: /
    - link "项目管理" [ref=e4] [cursor=pointer]:
      - /url: /?view=projects
    - link "项目模板管理" [ref=e5] [cursor=pointer]:
      - /url: /?view=content-types
    - link "工作流模板" [ref=e6] [cursor=pointer]:
      - /url: /workflow/templates
    - link "运行记录" [ref=e7] [cursor=pointer]:
      - /url: /workflow/runs
    - link "Agent 管理" [ref=e8] [cursor=pointer]:
      - /url: /agent/tasks
    - link "LLM Logs" [ref=e9] [cursor=pointer]:
      - /url: /llm/logs
  - main [ref=e10]:
    - heading "模型 Provider 管理" [level=1] [ref=e11]
    - alert [ref=e12]: "CONFLICT：create llm provider failed（request_id: LAPTOP-MUCR3O1G/yUOF98JuvS-000351）"
    - button "新增 Provider" [ref=e13]
    - list [ref=e14]:
      - listitem [ref=e15]: openai-compatible see****1234
      - listitem [ref=e16]: openai_compatible sk-****cret
    - generic [ref=e17]:
      - generic [ref=e18]:
        - text: Provider 类型
        - textbox "Provider 类型" [ref=e19]: openai_compatible
      - generic [ref=e20]:
        - text: Base URL
        - textbox "Base URL" [ref=e21]: https://llm.example.invalid/v1
      - generic [ref=e22]:
        - text: API Key
        - textbox "API Key" [ref=e23]: sk-live-red-contract-secret
      - button "提交 Provider" [active] [ref=e24]
  - button "Open Next.js Dev Tools" [ref=e30] [cursor=pointer]:
    - img [ref=e31]
  - alert [ref=e34]
```

# Test source

```ts
  31  |     await page.waitForResponse((response) => response.url().includes('/api/v1/dashboard/summary') && response.status() === 200);
  32  |     await expect(page.getByTestId('dashboard-loading')).toBeHidden();
  33  |     await expect(page.getByTestId('dashboard-project-count')).toContainText(/\d+/);
  34  | 
  35  |     await page.getByRole('navigation', { name: 'Iteration 2 navigation' }).getByRole('link', { name: '项目管理' }).click();
  36  |     await page.waitForResponse((response) => response.url().includes('/api/v1/projects') && response.status() === 200);
  37  |     await expect(page.getByRole('heading', { name: '项目管理' })).toBeVisible();
  38  |     await expect(page.getByTestId('projects-empty')).toBeVisible();
  39  | 
  40  |     await page.getByRole('button', { name: '新建项目' }).click();
  41  |     await page.getByLabel('项目名称').fill('RED roundtrip project');
  42  |     await page.getByLabel('项目模板').selectOption('seed-content-type');
  43  |     await page.getByRole('button', { name: '提交新建项目' }).click();
  44  |     await page.waitForResponse((response) => response.url().includes('/api/v1/projects') && response.request().method() === 'POST' && response.status() === 201);
  45  |     await expect(page.getByRole('status')).toContainText('创建成功');
  46  | 
  47  |     const projectOverviewResponse = page.waitForResponse((response) => response.url().includes('/api/v1/projects/') && response.url().includes('/overview') && response.status() === 200);
  48  |     await page.getByRole('listitem').filter({ hasText: 'RED roundtrip project' }).getByRole('button', { name: '进入项目' }).click();
  49  |     await projectOverviewResponse;
  50  |     await expect(page.getByRole('heading', { name: /项目工作区/ })).toBeVisible();
  51  |     await expect(page.getByRole('tab', { name: '项目概览' })).toHaveAttribute('aria-selected', 'true');
  52  | 
  53  |     await page.getByRole('button', { name: '暂停项目' }).click();
  54  |     await page.getByLabel('暂停原因').fill('RED contract requires reason and note');
  55  |     const pauseResponse = page.waitForResponse((response) => response.url().includes('/api/v1/projects/') && response.url().includes('/pause') && response.status() === 200);
  56  |     await page.getByRole('button', { name: '确认暂停' }).click();
  57  |     await pauseResponse;
  58  |     await expect(page.getByRole('status')).toContainText('已暂停');
  59  | 
  60  |     await page.getByRole('button', { name: '返回系统' }).click();
  61  |     await page.getByRole('navigation', { name: 'Iteration 2 navigation' }).getByRole('link', { name: '项目模板管理' }).click();
  62  |     await page.waitForResponse((response) => response.url().includes('/api/v1/content-types') && response.status() === 200);
  63  |     await expect(page.getByRole('heading', { name: '项目模板管理' })).toBeVisible();
  64  | 
  65  |     await page.getByRole('button', { name: '查看 Schema' }).click();
  66  |     await page.waitForResponse((response) => response.url().includes('/api/v1/content-types/') && response.url().includes('/project-schema') && response.status() === 200);
  67  |     await expect(page.getByTestId('project-schema')).toContainText('project_schema');
  68  | 
  69  |     for (const endpoint of requiredEndpoints.slice(0, 6)) {
  70  |       expect(apiHits.some((hit) => hit === endpoint || hit.startsWith(endpoint.replace('seed-project', '')))).toBeTruthy();
  71  |     }
  72  |   });
  73  | 
  74  |   test('home dashboard renders request_id and retry action on backend error', async ({ page }) => {
  75  |     await page.route('**/api/v1/dashboard/summary', async (route) => {
  76  |       await route.fulfill({
  77  |         status: 500,
  78  |         contentType: 'application/json',
  79  |         body: JSON.stringify({
  80  |           success: false,
  81  |           data: null,
  82  |           error: { code: 'INTERNAL_ERROR', message: 'dashboard unavailable', details: [] },
  83  |           request_id: 'req-dashboard-error',
  84  |         }),
  85  |       });
  86  |     });
  87  | 
  88  |     await page.goto(webBaseURL, { waitUntil: 'domcontentloaded' });
  89  |     const app = page.locator('main');
  90  |     await expect(app.getByRole('alert')).toContainText('INTERNAL_ERROR');
  91  |     await expect(app.getByRole('alert')).toContainText('req-dashboard-error');
  92  |     await expect(page.getByRole('button', { name: '重试' })).toBeVisible();
  93  |   });
  94  | 
  95  |   test('prompt and provider pages load real data, mutate, and never expose plaintext provider api_key', async ({ page }) => {
  96  |     const apiHits: string[] = [];
  97  |     page.on('request', (request) => {
  98  |       const url = new URL(request.url());
  99  |       if (url.pathname.startsWith('/api/v1/')) {
  100 |         apiHits.push(`${request.method()} ${url.pathname}`);
  101 |       }
  102 |     });
  103 | 
  104 |     await page.goto(`${webBaseURL}/prompt`);
  105 |     await expect(page.getByRole('heading', { name: 'Prompt 模板管理' })).toBeVisible();
  106 |     await expect(page.getByTestId('prompt-loading')).toBeVisible();
  107 |     await page.waitForResponse((response) => response.url().includes('/api/v1/prompt-templates') && response.status() === 200);
  108 |     await expect(page.getByTestId('prompt-loading')).toBeHidden();
  109 | 
  110 |     const promptCode = `outline_red_contract_${Date.now()}`;
  111 |     await page.getByRole('button', { name: '新建 Prompt' }).click();
  112 |     await page.getByLabel('Prompt Code').fill(promptCode);
  113 |     await page.getByLabel('Prompt 内容').fill('生成一个项目大纲：{{topic}}');
  114 |     const promptCreateResponse = page.waitForResponse((response) => response.url().includes('/api/v1/prompt-templates') && response.request().method() === 'POST' && response.status() === 201);
  115 |     await page.getByRole('button', { name: '提交 Prompt' }).click();
  116 |     await promptCreateResponse;
  117 |     await expect(page.getByRole('status')).toContainText('Prompt 创建成功');
  118 | 
  119 |     await page.goto(`${webBaseURL}/provider`);
  120 |     await expect(page.getByRole('heading', { name: '模型 Provider 管理' })).toBeVisible();
  121 |     await expect(page.getByTestId('provider-loading')).toBeVisible();
  122 |     await page.waitForResponse((response) => response.url().includes('/api/v1/llm-providers') && response.status() === 200);
  123 |     await expect(page.getByTestId('provider-loading')).toBeHidden();
  124 |     await expect(page.locator('body')).not.toContainText('sk-live-red-contract-secret');
  125 |     await expect(page.getByTestId('provider-key-masked').first()).toContainText(/\*{2,}/);
  126 | 
  127 |     await page.getByRole('button', { name: '新增 Provider' }).click();
  128 |     await page.getByLabel('Provider 类型').fill('openai_compatible');
  129 |     await page.getByLabel('Base URL').fill('https://llm.example.invalid/v1');
  130 |     await page.getByLabel('API Key').fill('sk-live-red-contract-secret');
> 131 |     const providerCreateResponse = page.waitForResponse((response) => response.url().includes('/api/v1/llm-providers') && response.request().method() === 'POST' && response.status() === 201);
      |                                         ^ Error: page.waitForResponse: Test timeout of 30000ms exceeded.
  132 |     await page.getByRole('button', { name: '提交 Provider' }).click();
  133 |     await providerCreateResponse;
  134 |     await expect(page.getByRole('status')).toContainText('Provider 创建成功');
  135 |     await expect(page.locator('body')).not.toContainText('sk-live-red-contract-secret');
  136 | 
  137 |     expect(apiHits).toEqual(expect.arrayContaining([
  138 |       'GET /api/v1/prompt-templates',
  139 |       'POST /api/v1/prompt-templates',
  140 |       'GET /api/v1/llm-providers',
  141 |       'POST /api/v1/llm-providers',
  142 |     ]));
  143 |   });
  144 | 
  145 |   test('prompt page renders request_id on validation failure and provider page has empty state', async ({ page }) => {
  146 |     await page.goto(`${webBaseURL}/provider?fixture=empty`);
  147 |     await expect(page.getByTestId('provider-empty')).toContainText('暂无 Provider');
  148 | 
  149 |     await page.goto(`${webBaseURL}/prompt`);
  150 |     await page.getByRole('button', { name: '新建 Prompt' }).click();
  151 |     await page.getByRole('button', { name: '提交 Prompt' }).click();
  152 |     const app = page.locator('main');
  153 |     await expect(app.getByRole('alert')).toContainText('VALIDATION_ERROR');
  154 |     await expect(app.getByRole('alert')).toContainText(/req-/);
  155 |   });
  156 | });
  157 | 
```