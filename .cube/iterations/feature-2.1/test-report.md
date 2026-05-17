# Iteration 2.1 Test Report

## Corrected 05-testing Conclusion

The previous 05-testing conclusion was incomplete: backend/API/basic route contracts were passing, but frontend visual and JS interaction acceptance had not been explicitly validated. This report corrects that gap by separating Contract, Functional, and Visual/UI evidence.

## Test Scope

- Backend library, handler, contract, integration, store, engine, and worker tests under `./...`.
- Web Admin TypeScript contract via `npm --prefix apps/web-admin run lint`.
- Iteration 1 Web Admin regression E2E: `e2e/iteration1-ui.spec.ts`.
- Iteration 2 navigation regression E2E: `e2e/iteration2-navigation.spec.ts`.
- Iteration 2.1 visual and functional Web Admin E2E: `e2e/iteration2_1-pages.spec.ts`.
- Source-contract regressions for Web Admin home/provider pages.

## Test Results

| Category | Check | Result | Evidence |
|---|---|---:|---|
| Contract PASS | `go test -race ./...` | PASS | All Go packages passed; `internal/http/contract` passed with race detector. |
| Contract PASS | Focused Web Admin contract tests | PASS | Home empty-fixture and prompt/provider source contracts passed. |
| Contract PASS | `npm --prefix apps/web-admin run lint` | PASS | `tsc --noEmit` completed successfully. |
| Functional PASS | `WEB_BASE_URL=http://127.0.0.1:3005 npm --prefix apps/web-admin run test:ui -- e2e/iteration1-ui.spec.ts` | PASS | 4/4 tests passed. |
| Functional PASS | `WEB_BASE_URL=http://127.0.0.1:3005 npm --prefix apps/web-admin run test:ui -- e2e/iteration2-navigation.spec.ts` | PASS | 5/5 tests passed. |
| Functional PASS / Visual UI PASS | `WEB_BASE_URL=http://127.0.0.1:3005 npm --prefix apps/web-admin run test:ui -- e2e/iteration2_1-pages.spec.ts` | PASS | 3/3 tests passed; CSS and JS interaction acceptance covered. |

## Frontend Visual/UI PASS Evidence

- Global CSS is imported from `app/layout.tsx` through `app/globals.css`.
- Unified layout styling is applied to navigation, main content, page hero, cards, tables, forms, buttons, alerts, status messages, badges, dialogs, pagination, and metric cards.
- Playwright asserts non-default CSS behavior:
  - `styled-page-shell` uses CSS grid layout.
  - `Iteration 2 navigation` uses flex layout.
  - `.page-hero` and buttons have non-default border radius.
  - `.card` has explicit white surface background.
- Visual smoke screenshots are captured for:
  - `test-results/iteration2_1-schedules-visual.png`
  - `test-results/iteration2_1-n8n-visual.png`
  - `test-results/iteration2_1-cost-visual.png`

## Frontend Functional PASS Evidence

- `/workflow/schedules`
  - Styled AppLayout renders.
  - Status filter works.
  - New schedule dialog opens and submits.
  - Schedule test-run produces result feedback.
  - Enable/disable status operation is clickable and produces result feedback.
  - Schedule detail panel opens.
  - Pagination controls are visible.
- `/external-automation/n8n`
  - Styled provider and binding forms render.
  - Provider submission produces result feedback.
  - Binding submission produces result feedback.
  - Plaintext token is not rendered after submission.
  - Token input uses password masking.
- `/llm/cost-summary`
  - Styled metric cards render.
  - Refresh action produces result feedback.
  - Model filter renders.
  - Detail jump opens a model detail panel.

## Pass Criteria

- Backend tests pass under race detector.
- Web Admin TypeScript compilation passes.
- Required locked Playwright regression suites pass against the real backend and Web Admin dev server.
- Iteration 2.1 frontend visual acceptance confirms pages are not browser-default bare HTML.
- Iteration 2.1 frontend JS acceptance confirms navigation, filtering, pagination, form submission, dialog/detail display, status operation, and result feedback.
- Provider API keys and external provider tokens remain masked in frontend flows.

## Coverage

- Schedule lifecycle and production-plan baseline contracts.
- Runtime trigger execution and workflow integration.
- LLM cost-summary API and Web Admin route.
- External automation provider/binding API and Web Admin route.
- Historical Web Admin page rendering contracts.
- Iteration 1 dashboard, projects, detail shell, content templates, prompt, and provider roundtrip flows.
- Iteration 2 route navigation and direct refresh behavior.
- Iteration 2.1 CSS and JS acceptance for schedules, n8n automation, and LLM cost summary pages.

## Standards Evidence

- `library`: Go module/service tests passed in full `go test -race ./...`.
- `integration`: HTTP handler, contract, store, and workflow integration tests passed in full Go suite.
- `web-e2e`: Iteration 1, Iteration 2 navigation, and Iteration 2.1 visual/functional Playwright suites passed.
- `sql-query`: Migration/store contract tests passed in full Go suite.

## Review Evidence

- TypeScript reviewer checked the Web Admin contract fixes and identified unsafe automatic seed-project mutation probes.
- Automatic page-load pause probe was removed.
- TypeScript reviewer checked the UI acceptance work and identified plaintext token input exposure.
- n8n token input was changed to `type="password"` and the visual/functional suite was rerun successfully.
- Backend project creation generates unique in-memory project IDs to keep repeat E2E runs stable.
