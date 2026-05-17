# Iteration 2 Test Report

## Test Scope

- Backend workflow, agent, LLM log, schedule, engine, HTTP handler, migration, and contract tests.
- Web-admin TypeScript contract/type coverage for workflow templates, workflow runs, agent tasks, and LLM call logs.
- Stage 05 acceptance review for API error envelope display, action feedback, sensitive payload redaction, and scope drift.

## Test Results

- `npm --prefix apps/web-admin run lint`: PASS. See `.cube/iterations/feature-2/web-admin-typecheck.log`.
- Browser UI integration with Playwright Chromium: PASS, 12/12 checks passed against `http://127.0.0.1:3000` and `http://127.0.0.1:18080`.
- `npm --prefix apps/web-admin run test:ui -- e2e/iteration2-navigation.spec.ts`: PASS, 5/5 checks cover homepage navigation into `/workflow/templates`, `/workflow/runs`, `/agent/tasks`, and `/llm/logs`, direct refresh, active-route highlighting, and same-origin `/api/v1/*` proxy requests.
- `go test ./apps/api-server/internal/http/contract`: PASS.
- `go test -race ./...`: FAIL in `TestCancelWorkflowRunEndpointReturns200WithCancelledStatus`; the run can be completed by the asynchronous engine before the cancel request, returning `409 CONFLICT`. See `.cube/iterations/feature-2/test-output.log`.

## Pass Criteria

- Web-admin TypeScript compilation passes with `tsc --noEmit`.
- Browser UI integration passes for the main workflow template/version/run, agent task, LLM log, and structured error-display flows.
- Homepage/global navigation exposes Workflow Templates, Workflow Runs, Agent Tasks, and LLM Logs with active-route highlighting and direct refresh support.
- Go race suite must pass after stabilizing the cancel-run handler test or test router setup.
- Workflow/agent/LLM frontend pages render API failures with message, code, and request_id.
- Main workflow actions expose loading/disabled states and success feedback.
- Workflow and agent detail payloads are redacted before browser display.
- No unrelated deletion remains for `docs/requirements/iteration-1-content-project-entry.md`.

## Coverage

- Service and handler tests cover workflow templates, versions, runs, agent tasks, LLM call logs, schedules, and engine execution.
- Contract tests cover frontend API types and required page behavior for workflow template/run, agent task, and LLM log screens.
- Integration/race coverage verifies the asynchronous WorkflowEngine execution path.

## Standards Evidence

- Web/API evidence follows `standards/testing/web-e2e.md` through executable frontend contract coverage, web-admin type checking, and real Playwright browser flows covering create/open/publish/trigger/detail/list/error paths plus homepage navigation into all Iteration 2 pages.
- Local browser integration now uses a Next.js same-origin `/api/v1/*` proxy so Windows Chrome does not need to connect directly to the WSL backend port.
- Integration evidence includes engine integration tests; current full race-suite status is blocked by the cancel-run timing race noted above.

## Review Evidence

- Code review: PASS after fixes; prior blockers for API error code/message/request_id rendering, action feedback, and unrelated doc deletion were addressed.
- TypeScript review: PASS after fixes; detail-page load failures now display structured API errors.
- Security review: PASS for Stage 05 after fixes; `NEXT_PUBLIC_API_TOKEN` was removed, local dev bearer header is restricted to exact localhost/127.0.0.1/[::1] HTTP origins, and payload redaction covers sensitive keys plus common secret-like string patterns.
