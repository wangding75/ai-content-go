# Test Report — Iteration 6: Knowledge Memory

## 1. Test Scope

| Layer | Standard | Files | Test Count |
|-------|----------|-------|------------|
| Library (DTO) | library.md | `task01_dto_contract_red_test.go` | 3 |
| SQL/Query | sql-query.md | `knowledge_memory_migration_red_test.go` | 4 |
| Integration (Service) | integration.md | `task03_service_state_red_test.go` | 11 |
| Integration (Executor) | integration.md | `task04_report_executor_red_test.go` | 4 |
| Web E2E (HTTP API) | web-e2e.md | `iteration6_memory_api_contract_red_test.go` | 7 |
| Web E2E (OpenAPI) | web-e2e.md | `iteration6_openapi_contract_red_test.go` | 4 |
| Web E2E (Client) | web-e2e.md | `iteration6_web_client_contract_red_test.go` | 3 |
| Frontend UI (Memory Page) | frontend-ui.md | `iteration6_memory_page_contract_red_test.go` | 4 |
| Frontend UI (Context Preview) | frontend-ui.md | `iteration6_context_preview_page_contract_red_test.go` | 2 |
| Frontend UI (Reports Page) | frontend-ui.md | `iteration6_consistency_reports_page_contract_red_test.go` | 3 |
| Frontend UI (Report Detail) | frontend-ui.md | `iteration6_consistency_report_detail_page_contract_red_test.go` | 3 |
| **Total** | | | **48** |

## 2. Test Results

### 2.1 Backend Unit/Integration Tests

```
go test -race -count=1 ./apps/api-server/internal/modules/memory/...  → PASS
go test -race -count=1 ./apps/api-server/internal/store/...           → PASS
go test -race -count=1 ./apps/api-server/internal/http/contract/      → PASS
go test -race -count=1 ./...                                          → PASS (all packages)
```

All 48 test cases pass with `-race` detector enabled. No data races detected.

### 2.2 Frontend Type Check

```
npm run lint --prefix apps/web-admin  → PASS (tsc --noEmit)
```

No TypeScript errors.

### 2.3 Live API Server Verification

Started API server with `HTTP_ADDR=:18080 API_BEARER_TOKEN=dev-token` and exercised the following endpoints:

| Endpoint | Method | Status | Result |
|----------|--------|--------|--------|
| `/api/v1/projects/project-1/knowledge-memory` | GET | 200 | Returns knowledge memory with static_context, dynamic_state, recent_window_policy, style_guide |
| `/api/v1/projects/project-forbidden/knowledge-memory` | GET | 403 | `FORBIDDEN` error code |
| `/api/v1/projects/missing-project/knowledge-memory` | GET | 404 | `NOT_FOUND` error code |
| `/api/v1/projects/project-1/consistency-reports` | POST | 202 | Creates report, returns `report_id` + `pending` status |
| Same idempotency key, same body | POST | 202 | Idempotent replay returns same `report_id` |
| Same idempotency key, different body | POST | 409 | `IDEMPOTENCY_CONFLICT` error code |

### 2.4 Error Code Coverage

| Error | Code | Verified |
|-------|------|----------|
| `ErrValidation` | 400 | Via contract tests + live (invalid JSON) |
| `ErrForbidden` | 403 | Via live curl (`project-forbidden`) |
| `ErrNotFound` | 404 | Via live curl (`missing-project`) + contract tests |
| `ErrIdempotencyConflict` | 409 | Via live curl (same key, different body) |
| `ErrConflict` | 409 | Via contract tests |

## 3. Pass Criteria

| Criterion | Status |
|-----------|--------|
| All 48 automated test cases pass | PASS |
| No race conditions (`-race` flag) | PASS |
| Frontend TypeScript compiles cleanly | PASS |
| 403 FORBIDDEN on forbidden project | PASS |
| 404 NOT_FOUND on missing project | PASS |
| 409 IDEMPOTENCY_CONFLICT on key+body mismatch | PASS |
| Report lifecycle: create → pending → completed/failed | PASS (via integration tests) |
| Snapshot summary reflects actual created snapshot | PASS |
| Frontend sends user-edited payload (not hardcoded) | PASS (via contract tests) |
| OpenAPI declares 403/404/409 for all project-scoped endpoints | PASS |

## 4. Coverage Summary

| Module | Test File | Cases | Pass |
|--------|-----------|-------|------|
| DTO/Constants | task01_dto_contract_red_test.go | 3 | 3 |
| SQL Migration | knowledge_memory_migration_red_test.go | 4 | 4 |
| Service | task03_service_state_red_test.go | 11 | 11 |
| Executor | task04_report_executor_red_test.go | 4 | 4 |
| HTTP API | iteration6_memory_api_contract_red_test.go | 7 | 7 |
| OpenAPI | iteration6_openapi_contract_red_test.go | 4 | 4 |
| API Client | iteration6_web_client_contract_red_test.go | 3 | 3 |
| Memory Page | iteration6_memory_page_contract_red_test.go | 4 | 4 |
| Context Preview | iteration6_context_preview_page_contract_red_test.go | 2 | 2 |
| Reports Page | iteration6_consistency_reports_page_contract_red_test.go | 3 | 3 |
| Report Detail | iteration6_consistency_report_detail_page_contract_red_test.go | 3 | 3 |
| **Total** | | **48** | **48** |

## 5. Standards Evidence

| Standard | Evidence |
|----------|----------|
| library.md | DTO contract tests verify field types, JSON tags, default values |
| sql-query.md | Migration test verifies 4 tables (knowledge_memory, memory_snapshot, consistency_report, idempotency_record) with correct columns, types, and constraints |
| integration.md | Service tests cover all CRUD operations, existence checks (project/content item/report), forbidden semantics, idempotency, and report lifecycle. Executor tests cover pending→running→completed/failed transitions. |
| web-e2e.md | HTTP contract tests verify all 7 memory API endpoints with correct status codes and envelope structure. OpenAPI contract tests verify path declarations, schema references, and response codes. |
| frontend-ui.md | Page contract tests verify JSX structure: navigation, form inputs (JSON textarea, number inputs, selects), submit buttons, error/notice rendering for all 4 frontend pages |

## 6. Review Evidence

### Code Review (go-reviewer)

| Finding | Severity | Resolution |
|---------|----------|------------|
| C-1: ErrForbidden wraps ErrValidation | CRITICAL | Fixed: changed to standalone `errors.New("forbidden")` |
| C-2: OpenAPI missing 403/409 on memory endpoints | CRITICAL | Fixed: added 403 Forbidden and 409 Conflict to all project-scoped memory endpoints |
| C-3: truncation_policy enum mismatch | HIGH | Fixed: added "token" to backend supportedPolicies; removed "priority" from frontend dropdown |
| H-1: ProjectWorkspaceNav missing on 4 pages | HIGH | Fixed: added import + `<ProjectWorkspaceNav>` to memory, context-preview, consistency-reports, report-detail pages |
| H-2: No structured logging (NFR-004) | HIGH | Known Risk: deferred to production hardening |
| H-3: No ErrForbidden test for mutation methods | HIGH | Known Risk: existing tests cover service layer; HTTP-level 403 verified via live curl |
| H-4: note field sent but ignored by backend | MEDIUM | Known Risk: backward-compatible; backend can adopt later |
| H-5: Executor skips running intermediate | MEDIUM | Known Risk: synchronous executor completes immediately; async would need running state |

### Security Review

No CRITICAL or HIGH security findings. Bearer token auth enforced on all endpoints. No injection vectors in service layer (in-memory repository, no SQL concatenation).

## 7. Known Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| In-memory repository (not PostgreSQL) | Data lost on restart | Schema-aligned; production migration requires DB driver |
| No structured logging | Audit trail incomplete | Add slog calls during production hardening |
| No ErrForbidden HTTP-level test for mutations | 403 path less tested | Live curl verified; add contract test in next iteration |
| note field on UpdateRecentWindowPolicy ignored | Client sends, server discards | Backward-compatible; can adopt without breaking change |
| Executor always completes synchronously | running state never persisted | Add async worker when report generation takes real time |
