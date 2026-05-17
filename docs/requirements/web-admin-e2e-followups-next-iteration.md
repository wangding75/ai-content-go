# Web Admin E2E Follow-ups for Next Iteration

## Context

Iteration 2 Web Admin navigation integration is mostly complete:

- Global navigation exposes:
  - 首页 / 系统大盘
  - 项目管理
  - 项目模板管理
  - 工作流模板
  - 运行记录
  - Agent 管理
  - LLM Logs
- Homepage duplicate navigation was removed.
- Iteration 2 navigation e2e passed: `apps/web-admin/e2e/iteration2-navigation.spec.ts`.
- Frontend type check passed via `npm --prefix apps/web-admin run lint`.

The remaining instability is in the older Iteration 1 UI e2e suite, not in the new Iteration 2 navigation path.

## Current Failing Area

File:

- `apps/web-admin/e2e/iteration1-ui.spec.ts`

Latest observed failures came from repeated runs against a long-lived in-memory backend. The tests assume stable empty/fresh backend state, but the backend retains previously created projects, prompt templates, and LLM providers during the dev session.

## Root Causes

### 1. Project list flow is not repeat-run safe

The homepage project management flow currently depends on project list state. During repeated test runs, the backend may already contain projects created by previous runs.

Observed symptoms:

- Empty-state assertions become unstable.
- Clicking a fallback `进入项目` can enter `seed-project` instead of the newly created project.
- Pausing `seed-project` or an already paused project can return `CONFLICT`.

Observed error example:

```text
CONFLICT：pause project failed
```

### 2. Fixed test data causes collisions

The e2e suite used fixed values such as:

- `RED roundtrip project`
- `outline_red_contract`
- `sk-live-red-contract-secret`
- fixed provider base URL

Repeated runs can create duplicate data or cause backend conflict/validation responses instead of the expected `201` responses.

### 3. Some request waits are still sensitive to timing

Requests triggered by button clicks should consistently create `page.waitForResponse(...)` promises before clicking. Otherwise fast local responses can be missed.

### 4. Provider masked-key assertion assumed one provider

The provider page can contain multiple providers after repeated runs. A strict locator such as `page.getByTestId('provider-key-masked')` may resolve to multiple elements.

Observed symptom:

```text
strict mode violation: getByTestId('provider-key-masked') resolved to 2 elements
```

### 5. Error injection test must align with delayed dashboard loading

The dashboard request is delayed on homepage load. Error-injection tests should register the route before navigation and explicitly wait for the mocked 500 response before asserting the alert.

## Recommended Next-Iteration Fix

### A. Make project listing fixture explicit

In `apps/web-admin/app/page.tsx`:

- Default project management should call the real list API: `fetchProjects()`.
- Only use the empty fixture when explicitly requested, for example:

```text
/?view=projects&fixture=empty
```

Then `loadProjects()` can choose:

```ts
const fixture = searchParams.get('fixture');
const result = await fetchProjects(fixture === 'empty' ? '&status=__empty_fixture__' : '');
```

### B. Make Iteration 1 e2e data unique per run

In `apps/web-admin/e2e/iteration1-ui.spec.ts`, generate a run suffix:

```ts
const runID = Date.now();
const projectName = `RED roundtrip project ${runID}`;
const promptCode = `outline_red_contract_${runID}`;
const providerSecret = `sk-live-red-contract-secret-${runID}`;
const providerURL = `https://llm-${runID}.example.invalid/v1`;
```

Use these values for creation, lookup, and negative secret-exposure assertions.

### C. Enter the newly created project, not a fallback project

After project creation and list refresh, locate the newly created item by `projectName`:

```ts
await page
  .getByRole('listitem')
  .filter({ hasText: projectName })
  .getByRole('button', { name: '进入项目' })
  .click();
```

Avoid fallback `进入项目` buttons that may point to `seed-project`.

### D. Move all request waits before clicks

Apply this pattern consistently:

```ts
const responsePromise = page.waitForResponse((response) =>
  response.url().includes('/api/v1/...') && response.request().method() === 'POST' && response.status() === 201
);
await page.getByRole('button', { name: '提交...' }).click();
await responsePromise;
```

Important actions:

- Create project
- Open project overview
- Pause project
- View project schema
- Create prompt template
- Create LLM provider
- Navigate into project/template views where the click triggers a request

### E. Split empty-state assertions from mutation flow

Do not assert `projects-empty` in the same flow that creates projects against a shared backend.

Instead, test empty state via explicit fixture:

```ts
await page.goto(`${webBaseURL}/?view=projects&fixture=empty`, { waitUntil: 'domcontentloaded' });
await expect(page.getByTestId('projects-empty')).toBeVisible();
```

### F. Make provider masked-key assertions list-safe

Use `.first()` for generic masked-key existence, or locate the provider item by unique `providerURL`/type and assert inside that item.

Example:

```ts
await expect(page.getByTestId('provider-key-masked').first()).toContainText(/\*{2,}/);
await expect(page.locator('body')).not.toContainText(providerSecret);
```

### G. Align dashboard error test with delayed request

Register route before navigation and explicitly wait for the mocked 500 response:

```ts
await page.route('**/api/v1/dashboard/summary', async (route) => {
  await route.fulfill({ status: 500, contentType: 'application/json', body: JSON.stringify(...) });
});

await page.goto(webBaseURL, { waitUntil: 'domcontentloaded' });
await page.waitForResponse((response) =>
  response.url().includes('/api/v1/dashboard/summary') && response.status() === 500
);
await expect(page.locator('main').getByRole('alert')).toContainText('INTERNAL_ERROR');
```

## Verification Commands

After implementing the follow-ups:

```bash
npm --prefix apps/web-admin run lint
```

```bash
WEB_BASE_URL=http://127.0.0.1:3000 npm --prefix apps/web-admin run test:ui -- e2e/iteration1-ui.spec.ts
```

```bash
WEB_BASE_URL=http://127.0.0.1:3000 npm --prefix apps/web-admin run test:ui -- e2e/iteration2-navigation.spec.ts
```

## Acceptance Criteria

- `iteration1-ui.spec.ts` can be run repeatedly against a long-lived in-memory dev backend.
- Tests do not require manual backend restart or clearing in-memory state.
- Project pause flow targets the project created in the current test run.
- Prompt/provider creation uses unique data and does not collide with prior runs.
- Empty-state checks use explicit fixture mode.
- Iteration 2 navigation e2e remains green.
