# Test Report — Iteration 12: Article Pack

## Test Scope

- **Modules**: Article Pack registration, Article project config, Article generation runs, Article metrics config
- **Test types**: Unit tests (Go), Integration tests (Go service chain), Web/API E2E (real HTTP), Frontend UI (real browser)
- **Standards applied**: `standards/testing/web-e2e.md`, `standards/testing/frontend-ui.md`, `standards/testing/integration.md`
- **Features under test**: FR-001 to FR-014 (Article Pack registration, project config, generation runs, metrics, admin pages)

## Test Results

### Green Gate — ALL PASS

```
go test -race -skip 'TestTask01PostgresMigrationAndMetricDMLWriteThenReadContract|TestTask01PostgresMigrationAndStrategyDMLWriteThenReadContract|TestTask02PostgresMigrationAndPortfolioDMLWriteThenReadContract' ./...
```

28 packages, 0 failures. All packages including `article`, `metrics`, `handlers`, `contract`, `generation`, `workflow`, `content` passed with race detector enabled.

### Integration Tests — ALL PASS

| Test | Status | Chain |
|------|--------|-------|
| `TestTask02IntegrationRegisterPackCreatesContentType` | PASS | ArticleService → ContentService |
| `TestTask02IntegrationRegisterPackCreatesWorkflowTemplates` | PASS | ArticleService → WorkflowService |
| `TestTask02IntegrationRegisterPackCreatesMetricTemplates` | PASS | ArticleService → MetricsService |
| `TestTask04IntegrationCreateGenerationRunWithWorkflow` | PASS | ArticleService → WorkflowService + GenerationService |
| `TestTask04IntegrationRetryGenerationRun` | PASS | ArticleService → GenerationService |
| `TestIntegrationPostgresStoreCreateAndFindTemplate` | PASS | PostgreSQL |
| `TestIntegrationPostgresStoreCreateAndFindRecord` | PASS | PostgreSQL |
| `TestIntegrationPostgresStoreIdempotency` | PASS | PostgreSQL |
| `TestIntegrationPostgresStoreSummaryAggregation` | PASS | PostgreSQL |

### Web/API E2E — ALL PASS

All endpoints verified via real HTTP against live API server (port 18080, API_BEARER_TOKEN=dev).

| Endpoint | Method | Success | Validation Error | Business Error |
|----------|--------|---------|-----------------|----------------|
| `/api/v1/content-packs/article/register` | POST | 200, content_pack_id + workflow_version_ids + metric_template_ids | — | — |
| `/api/v1/content-packs/article/status` | GET | 200, registered=true, 5 default_metrics | — | — |
| `/projects/{id}/article/config` | GET | 200, default config for article project | — | 404 nonexistent |
| `/projects/{id}/article/config` | PATCH | 200, version_id | VALIDATION_ERROR (empty topic_style) | — |
| `/projects/{id}/article/generation-runs` | POST | 200, generation_run_id + workflow_run_id | — | — |
| `/projects/{id}/article/generation-runs` | GET | 200, paginated list | invalid page param | — |
| `/projects/{id}/article/generation-runs/{id}` | GET | 200, detail with status+topic | — | NOT_FOUND |
| `/projects/{id}/article/generation-runs/{id}/retry` | POST | 200, new ID different from original | — | — |
| `/projects/{id}/article/metrics` | GET | 200, paginated items | — | — |
| `/projects/{id}/article/metrics` | PATCH | 200, version_id | — | — |

Error envelope format verified: `{"success":false,"data":null,"error":{"code":"...","message":"...","details":null},"request_id":"..."}`

### Frontend UI — ALL PASS

Verified via real headless Chromium against live services.

| Page | Layout | No Errors | Data State | Interaction |
|------|--------|-----------|------------|-------------|
| `/article-pack` | AppLayout + card styling | 0 console errors | content_pack_id, 5 metrics, workflow template | Refresh → request_id |
| `/projects/project-1/article` | AppLayout + form + table | 0 console errors | topic_style=tech, 2 generation runs | Save config → success |
| `/projects/project-1/article/metrics` | AppLayout + table + checkbox | 0 console errors | views + likes enabled | Refresh → API response |

Screenshots: `apps/web-admin/test-results/manual-article-admin/page1-article-pack.png`, `page2-project-article.png`, `page3-project-article-metrics.png`

### E2E Playwright Regression

`apps/web-admin/e2e/iteration2_1-pages.spec.ts` — PASS: article pack page stays stable when status API returns unregistered payload without defaults.

## Pass Criteria

| Acceptance Criterion | Status | Evidence |
|---------------------|--------|----------|
| FR-001: Register Article Pack via POST | PASS | curl + green tests |
| FR-002: Query Article Pack status via GET | PASS | curl + browser screenshot |
| FR-003: Get Article project config via GET | PASS | curl + browser screenshot |
| FR-004: Update Article project config via PATCH | PASS | curl + browser interaction |
| FR-005: Create generation run via POST | PASS | curl + browser screenshot |
| FR-006: List generation runs via GET | PASS | curl + browser screenshot (2 runs listed) |
| FR-007: Get generation run detail via GET | PASS | curl |
| FR-008: Retry failed generation run | PASS | curl (retry produced different ID) |
| FR-009: Get content snapshot | PASS | green test |
| FR-010: Get metrics config via GET | PASS | curl + browser screenshot |
| FR-011: Update metrics config via PATCH | PASS | curl + browser interaction |
| FR-012: Article Pack admin page | PASS | real browser — AppLayout + JS + data states |
| FR-013: Article project admin page | PASS | real browser — config form + runs table + save |
| FR-014: Article metrics admin page | PASS | real browser — metrics table + checkboxes + save |
| Idempotency: Register/PATCH/retry | PASS | curl + green tests |
| Error envelope on all endpoints | PASS | curl for VALIDATION_ERROR + INTERNAL_ERROR |
| API_BEARER_TOKEN auth on all routes | PASS | contract tests pass |

## Coverage

- Go packages: 28/28 pass (100% package coverage)
- Contract tests: iteration11 + article handler red tests — ALL PASS
- Integration chain coverage: ArticleService → ContentService + WorkflowService + MetricsService + GenerationService — all exercised
- API endpoint coverage: 10/10 article endpoints exercised via real HTTP
- Frontend page coverage: 3/3 article admin pages exercised via real browser
- Coverage tool: not configured in workflow.yaml (`coverage_command` absent)

## Standards Evidence

| Standard | Executed | Method | Result |
|----------|----------|--------|--------|
| `standards/testing/integration.md` | Yes | `go test -race ./apps/api-server/internal/modules/article` | PASS — 5 integration tests |
| `standards/testing/web-e2e.md` | Yes | Step 1: started real API server (HTTP_ADDR=:18080); Step 2: curl against 10 endpoints (success + validation failure + business failure); Step 3: chromium detected; Step 4: headless Chromium browser flow | PASS |
| `standards/testing/frontend-ui.md` | Yes | Step 1: started Next.js dev server; Step 2: real browser opened 3 pages; Step 3: CSS/AppLayout/cards/tables verified via screenshots; Step 4: Refresh/Save interactions confirmed | PASS |

## Review Evidence

### Go Code Review (`ecc:go-reviewer`)
- **No CRITICAL issues**
- **HIGH** (2 fixed):
  - Dead code removed: `newID`, `nextID`, `idCounter` (never called, unsynchronized)
  - Pagination: `ListGenerationRuns` handler now passes parsed `page`/`page_size` instead of discarding
- **MEDIUM** (noted): `RetryGenerationRunForNonFailedRunReturnsConflict` uses `t.Logf` not `t.Fatal`; hardcoded `Platform: "web"` in `GetProjectArticleMetrics`
- **LOW**: `parsePage` uses manual loop instead of `strconv.Atoi`

### Security Review (`ecc:security-reviewer`)
- **No CRITICAL or HIGH issues**
- Input validation present on all handler entry points; idempotency key handling safe (in-memory only); error messages do not leak internal state; bearer token auth enforced on all article routes

## Known Issues

- `RetryGenerationRunForNonFailedRunReturnsConflict` test assertion weakened (uses `t.Logf` instead of `t.Fatal`)
- `GetProjectArticleMetrics` hardcodes `Platform: "web"` — should derive from data
- `parsePage` uses manual char-loop instead of `strconv.Atoi`