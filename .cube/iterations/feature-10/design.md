# Iteration 10 Design：Project Portfolio 多项目管理

## Impact Analysis

### Confirmed Approach

采用“新增模块方案”：新增独立 Portfolio 后端模块与前端页面树，只在路由、依赖注入、导航、OpenAPI 等入口处做最小增量接入。

### Rationale

- Portfolio 是系统级跨项目管理与观察视图，边界独立于现有项目工作台、生产、发布、指标、策略建议执行链路。
- 现有代码已经按 `modules/{domain}`、`handlers/{domain}.go`、`router.go` option injection、`apps/web-admin/lib/api.ts` client helper 的方式组织。
- 新增模块可以复用统一 API envelope、分页模型、幂等键、PostgreSQL store 可选注入、memory store fallback、Next.js 页面状态模式。
- 设计契约保留 PRD 要求的 `scope_type`、`owner_id`、`health_policy`、member `role`、snapshot `date_range` / `health_status` / `risk_summary` / `cost_summary` / `strategy_summary` / `source_refs` / `calculated_at` 字段。

### Affected Areas

| Area | Change | Risk | Mitigation |
|---|---|---:|---|
| Backend module | 新增 `apps/api-server/internal/modules/portfolio` | Low | 独立模块，沿用 metrics/strategy 的 Service/Store/DTO/errors 模式 |
| HTTP handlers | 新增 Portfolio handler | Low | 统一使用 `api.WriteSuccess` / `api.WriteError` / `decodeJSON` / `parsePagination` |
| Router | 注册 `/api/v1/portfolios...` 路由，增加 `DELETE` CORS method | Medium | 只增量添加 Portfolio option、handler 和 routes，不改变现有 routes |
| App server | 当 `DatabaseURL` 存在时注入 Postgres PortfolioStore | Low | 沿用 metrics store 注入模式；无 DB 时使用 module 默认 memory store |
| Database | 新增 `project_portfolio`、`portfolio_project`、`portfolio_status_snapshot` | Medium | 使用 additive migration `00012`，不修改现有表 |
| OpenAPI | 增加 Portfolio endpoints/schemas | Low | 与 PRD API 清单保持一一对应 |
| Frontend API client | 增加 Portfolio types 和 request helpers | Low | 复用现有 request/idempotencyHeaders/page envelope 模式 |
| Frontend pages | 新增 `/portfolios` 页面树 | Medium | 独立页面，复用现有 page shell、error、toast、table/card 模式 |
| Global navigation | 增加 `Portfolio 管理` nav item | Low | 只添加一项，match `/portfolios` 子路由 |

### Explicit Non-Impact

- 不调用 `WorkflowRun` 创建或执行接口。
- 不调用 `AgentRuntime`。
- 不创建 `ContentItem`。
- 不执行、确认、忽略或重试策略建议。
- 不复制现有成本、指标、发布、策略建议事实表。
- 不修改 `docs/requirements/`。

## Flow Design

### Portfolio List Flow

```text
Admin opens /portfolios
  -> web-admin calls GET /api/v1/portfolios?page=&page_size=&q=&status=&scope_type=&owner_id=
  -> PortfolioHandler.ListPortfolios parses pagination/search/filter query
  -> portfolio.Service.ListPortfolios validates pagination and filter values
  -> Store.ListPortfolios returns paged portfolio summaries
  -> handler returns API envelope
  -> UI renders cards/table with loading, empty, error states
```

### Create / Edit Portfolio Flow

```text
Admin submits Portfolio form
  -> UI calls POST /api/v1/portfolios or PATCH /api/v1/portfolios/{id}
  -> optional Idempotency-Key is passed for create/update form submissions
  -> handler decodes JSON body and route id
  -> service validates name/status/scope_type/owner_id/health_policy fields
  -> service checks idempotency when key exists
  -> store creates or updates project_portfolio
  -> handler returns created/updated PortfolioDetail
  -> UI refreshes list/detail and shows success status
```

### Portfolio Detail Flow

```text
Admin opens /portfolios/{portfolioId}
  -> UI calls:
     - GET /api/v1/portfolios/{id}
     - GET /api/v1/portfolios/{id}/projects
     - GET /api/v1/portfolios/{id}/strategy-summary
  -> service verifies portfolio exists
  -> store reads portfolio and member summaries
  -> strategy summary is read-only aggregation of existing strategy suggestion state
  -> UI renders overview, member preview, and strategy suggestion summary
```

### Add / Remove Member Project Flow

```text
Admin adds or removes a project
  -> add: POST /api/v1/portfolios/{id}/projects
  -> remove: DELETE /api/v1/portfolios/{id}/projects/{projectId} with { reason, note }
  -> service validates portfolio_id/project_id/role/priority/weight or removal reason
  -> store enforces unique portfolio-project membership and records operation metadata
  -> service records audit fields on membership change
  -> handler returns updated member or removal confirmation with operation_id
  -> UI refreshes member list and summary cards
```

### Priority / Weight Update Flow

```text
Admin opens /portfolios/{portfolioId}/projects
  -> UI lists portfolio members
  -> Admin updates priority, weight, or role
  -> UI calls PATCH /api/v1/portfolios/{id}/projects/{projectId}/priority
  -> service validates priority >= 1 and weight >= 0
  -> store updates only membership priority fields
  -> handler returns updated member
  -> UI updates row and toast
```

### Health Snapshot Recalculation Flow

```text
Admin opens /portfolios/{portfolioId}/health
  -> UI calls GET health-summary, cost-summary, status-snapshots
  -> Admin chooses date_range and force option, then clicks recalculate
  -> UI calls POST /api/v1/portfolios/{id}/status-snapshots/recalculate
     with { date_range, force }
  -> service validates portfolio, date_range, and membership scope
  -> service creates a snapshot reference for asynchronous recalculation
  -> handler returns snapshot_id or job_id immediately
  -> UI polls status-snapshots and refreshes latest summary/history
```

The recalculation endpoint must not perform multi-project aggregation synchronously in the HTTP request path. The 04-development implementation may complete the queued calculation in-process for the MVP, but the API contract remains non-blocking and reference-based.

### Read-Only Source Aggregation

```text
Portfolio Service
  -> reads member project IDs from portfolio_project
  -> aggregates existing domain data by project_id:
     - MetricRecord-derived health signals
     - LLMCallLog-derived model cost signals
     - PublishJob-derived publish status signals
     - StrategySuggestion-derived suggestion counts
  -> records source_refs with source type, source id/query key, and updated_at
  -> returns summaries with source labels, source timestamps, and explainable risk/cost/strategy breakdowns
```

## Table Design

### `project_portfolio`

| Column | Type | Constraint | Notes |
|---|---|---|---|
| id | text | primary key | Portfolio ID generated by service |
| name | text | not null | Display name |
| description | text | not null default '' | Optional description |
| scope_type | text | not null | Portfolio scope, e.g. `manual` |
| owner_id | text | not null default '' | Owner/operator identifier |
| health_policy | jsonb | not null default '{}' | Explainable health scoring policy |
| status | text | not null | `active` / `archived` |
| created_at | timestamptz | not null default now() | Creation time |
| updated_at | timestamptz | not null default now() | Last update time |

Indexes:

- `idx_project_portfolio_status_created_at(status, created_at desc)`
- `idx_project_portfolio_scope_type(scope_type)`
- `idx_project_portfolio_owner_id(owner_id)`
- `idx_project_portfolio_name(name)`

### `portfolio_project`

| Column | Type | Constraint | Notes |
|---|---|---|---|
| portfolio_id | text | not null | References `project_portfolio(id)` |
| project_id | text | not null | Existing content project ID |
| role | text | not null default 'member' | Member role in portfolio |
| priority | integer | not null | Lower number means higher priority |
| weight | numeric(10,2) | not null default 1 | Aggregation weight |
| note | text | not null default '' | Membership note |
| added_by | text | not null default '' | Audit user identifier |
| created_at | timestamptz | not null default now() | Add time |
| updated_at | timestamptz | not null default now() | Last membership update |

Constraints:

- Primary key: `(portfolio_id, project_id)`
- Check: `priority >= 1`
- Check: `weight >= 0`

Indexes:

- `idx_portfolio_project_project_id(project_id)`
- `idx_portfolio_project_role(role)`
- `idx_portfolio_project_priority(portfolio_id, priority asc)`

### `portfolio_status_snapshot`

| Column | Type | Constraint | Notes |
|---|---|---|---|
| id | text | primary key | Snapshot ID generated by service |
| portfolio_id | text | not null | References `project_portfolio(id)` |
| date_range_start | date | not null | Snapshot source window start |
| date_range_end | date | not null | Snapshot source window end |
| health_score | numeric(5,2) | not null | 0-100 aggregate health score |
| health_status | text | not null | `healthy` / `watch` / `critical` / `pending` |
| total_projects | integer | not null | Member count at calculation time |
| active_projects | integer | not null | Active source project count |
| warning_projects | integer | not null | Projects needing attention |
| estimated_monthly_cost | numeric(12,2) | not null | Aggregated estimated monthly cost |
| currency | text | not null default 'CNY' | Cost currency |
| risk_summary | jsonb | not null default '{}' | Explainable risk breakdown |
| cost_summary | jsonb | not null default '{}' | By-model and by-project cost summary |
| strategy_summary | jsonb | not null default '{}' | Strategy suggestion count and top items |
| source_refs | jsonb | not null default '[]' | Source type/id/timestamp references |
| calculation_status | text | not null | `queued` / `running` / `completed` / `failed` |
| calculated_at | timestamptz | null | Completion time |
| created_at | timestamptz | not null default now() | Snapshot request time |

Indexes:

- `idx_portfolio_status_snapshot_portfolio_created(portfolio_id, created_at desc)`
- `idx_portfolio_status_snapshot_status(calculation_status)`
- `idx_portfolio_status_snapshot_date_range(portfolio_id, date_range_start, date_range_end)`

### Migration Rules

- Add new migration `apps/api-server/migrations/00012_create_portfolio_tables.sql`.
- Use goose comments `-- +goose Up` and `-- +goose Down`.
- Use additive DDL only.
- Down migration drops snapshot table, membership table, then portfolio table.

## API Design

All endpoints are under `/api/v1`, require existing bearer auth middleware, return the standard envelope:

```json
{
  "success": true,
  "data": {},
  "error": null,
  "request_id": "req_xxx"
}
```

### Endpoints

| Method | Path | Handler | Request | Response |
|---|---|---|---|---|
| POST | `/portfolios` | `CreatePortfolio` | `CreatePortfolioRequest` | `PortfolioDetailResponse` |
| GET | `/portfolios` | `ListPortfolios` | query pagination + `q` + `status` + `scope_type` + `owner_id` | `PagedPortfoliosResponse` |
| GET | `/portfolios/{portfolioId}` | `GetPortfolio` | path | `PortfolioDetailResponse` |
| PATCH | `/portfolios/{portfolioId}` | `UpdatePortfolio` | `UpdatePortfolioRequest` | `PortfolioDetailResponse` |
| POST | `/portfolios/{portfolioId}/projects` | `AddProject` | `AddPortfolioProjectRequest` | `PortfolioProjectResponse` |
| GET | `/portfolios/{portfolioId}/projects` | `ListProjects` | query pagination + `role` | `PagedPortfolioProjectsResponse` |
| PATCH | `/portfolios/{portfolioId}/projects/{projectId}/priority` | `UpdateProjectPriority` | `UpdatePortfolioProjectPriorityRequest` | `PortfolioProjectResponse` |
| DELETE | `/portfolios/{portfolioId}/projects/{projectId}` | `RemoveProject` | `RemovePortfolioProjectRequest` | `RemovePortfolioProjectResponse` |
| POST | `/portfolios/{portfolioId}/status-snapshots/recalculate` | `RecalculateStatusSnapshot` | `RecalculatePortfolioStatusSnapshotRequest` | `RecalculatePortfolioStatusSnapshotResponse` |
| GET | `/portfolios/{portfolioId}/status-snapshots` | `ListStatusSnapshots` | query pagination + date range | `PagedPortfolioStatusSnapshotsResponse` |
| GET | `/portfolios/{portfolioId}/health-summary` | `GetHealthSummary` | path + date range query | `PortfolioHealthSummaryResponse` |
| GET | `/portfolios/{portfolioId}/cost-summary` | `GetCostSummary` | path + date range query | `PortfolioCostSummaryResponse` |
| GET | `/portfolios/{portfolioId}/strategy-summary` | `GetStrategySummary` | path + date range query | `PortfolioStrategySummaryResponse` |

### Recalculation Contract

Request:

```json
{
  "date_range": {
    "start": "2026-05-01",
    "end": "2026-05-31"
  },
  "force": false
}
```

Response:

```json
{
  "portfolio_id": "pf_xxx",
  "snapshot_id": "pss_xxx",
  "job_id": "pss_xxx",
  "calculation_status": "queued"
}
```

### Remove Project Contract

Request:

```json
{
  "reason": "project no longer belongs to this management scope",
  "note": "moved to another portfolio"
}
```

Response:

```json
{
  "portfolio_id": "pf_xxx",
  "project_id": "proj_xxx",
  "operation_id": "ppo_xxx",
  "removed": true
}
```

### Idempotency

- `POST /portfolios` accepts optional `Idempotency-Key`.
- `PATCH /portfolios/{portfolioId}` accepts optional `Idempotency-Key`.
- `POST /portfolios/{portfolioId}/projects` accepts optional `Idempotency-Key`.
- `PATCH /portfolios/{portfolioId}/projects/{projectId}/priority` accepts optional `Idempotency-Key`.
- `POST /portfolios/{portfolioId}/status-snapshots/recalculate` accepts optional `Idempotency-Key`.
- Idempotency conflict maps to `IDEMPOTENCY_CONFLICT` with HTTP 409.

### Error Mapping

| Domain Error | HTTP | API Error Code |
|---|---:|---|
| `portfolio.ErrValidation` | 400 | `VALIDATION_ERROR` |
| `portfolio.ErrForbidden` | 403 | `FORBIDDEN` |
| `portfolio.ErrNotFound` | 404 | `NOT_FOUND` |
| `portfolio.ErrConflict` | 409 | `CONFLICT` |
| `portfolio.ErrIdempotencyConflict` | 409 | `IDEMPOTENCY_CONFLICT` |
| default | 500 | `INTERNAL_ERROR` |

### OpenAPI Updates

`openapi/openapi.yaml` must add:

- All Portfolio paths listed above.
- Schemas for each request/response DTO.
- Shared pagination usage consistent with existing endpoints.
- Error response references using the existing standard envelope/error schema style.

## Module Design

### Backend Package Layout

```text
apps/api-server/internal/modules/portfolio/
  dto.go
  errors.go
  store.go
  memory_store.go
  postgres_store.go
  service.go
```

### Backend Interfaces

```go
type Service interface {
    CreatePortfolio(ctx context.Context, req CreatePortfolioRequest, idempotencyKey string) (PortfolioDetailResponse, error)
    ListPortfolios(ctx context.Context, req ListPortfoliosRequest) (PagedPortfoliosResponse, error)
    GetPortfolio(ctx context.Context, portfolioID string) (PortfolioDetailResponse, error)
    UpdatePortfolio(ctx context.Context, portfolioID string, req UpdatePortfolioRequest, idempotencyKey string) (PortfolioDetailResponse, error)
    AddProject(ctx context.Context, portfolioID string, req AddPortfolioProjectRequest, idempotencyKey string) (PortfolioProjectResponse, error)
    ListProjects(ctx context.Context, portfolioID string, req ListPortfolioProjectsRequest) (PagedPortfolioProjectsResponse, error)
    UpdateProjectPriority(ctx context.Context, portfolioID string, projectID string, req UpdatePortfolioProjectPriorityRequest, idempotencyKey string) (PortfolioProjectResponse, error)
    RemoveProject(ctx context.Context, portfolioID string, projectID string, req RemovePortfolioProjectRequest) (RemovePortfolioProjectResponse, error)
    RecalculateStatusSnapshot(ctx context.Context, portfolioID string, req RecalculatePortfolioStatusSnapshotRequest, idempotencyKey string) (RecalculatePortfolioStatusSnapshotResponse, error)
    ListStatusSnapshots(ctx context.Context, portfolioID string, req ListPortfolioStatusSnapshotsRequest) (PagedPortfolioStatusSnapshotsResponse, error)
    GetHealthSummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioHealthSummaryResponse, error)
    GetCostSummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioCostSummaryResponse, error)
    GetStrategySummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioStrategySummaryResponse, error)
}
```

```go
type Store interface {
    CreatePortfolio(ctx context.Context, item PortfolioDetailResponse) error
    UpdatePortfolio(ctx context.Context, item PortfolioDetailResponse) error
    GetPortfolio(ctx context.Context, portfolioID string) (*PortfolioDetailResponse, error)
    ListPortfolios(ctx context.Context, req ListPortfoliosRequest) ([]PortfolioListItem, int, error)
    AddProject(ctx context.Context, item PortfolioProjectResponse) error
    UpdateProject(ctx context.Context, item PortfolioProjectResponse) error
    RemoveProject(ctx context.Context, portfolioID string, projectID string, req RemovePortfolioProjectRequest) error
    GetProject(ctx context.Context, portfolioID string, projectID string) (*PortfolioProjectResponse, error)
    ListProjects(ctx context.Context, portfolioID string, req ListPortfolioProjectsRequest) ([]PortfolioProjectResponse, int, error)
    InsertStatusSnapshot(ctx context.Context, item PortfolioStatusSnapshotResponse) error
    ListStatusSnapshots(ctx context.Context, portfolioID string, req ListPortfolioStatusSnapshotsRequest) ([]PortfolioStatusSnapshotResponse, int, error)
    GetLatestStatusSnapshot(ctx context.Context, portfolioID string) (*PortfolioStatusSnapshotResponse, error)
    QueryHealthSummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioHealthSummaryResponse, error)
    QueryCostSummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioCostSummaryResponse, error)
    QueryStrategySummary(ctx context.Context, portfolioID string, req PortfolioSummaryRequest) (PortfolioStrategySummaryResponse, error)
    CheckIdempotency(ctx context.Context, scope string, endpoint string, key string, hash string) (refType string, refID string, conflict bool, err error)
    StoreIdempotency(ctx context.Context, scope string, endpoint string, key string, hash string, refType string, refID string) error
}
```

### Handler Layout

`apps/api-server/internal/http/handlers/portfolio.go` adds:

- `PortfolioHandler` with a `portfolio.Service` field.
- Constructor `NewPortfolioHandler(service portfolio.Service) *PortfolioHandler`.
- One method per endpoint in API Design.
- `writePortfolioError` mapper mirroring metrics/strategy patterns.

### Router Integration

`apps/api-server/internal/http/router.go` adds:

- `portfolio` module import.
- `portfolioService portfolio.Service` in router config.
- `WithPortfolioService(svc portfolio.Service) RouterOption`.
- Default `portfolio.NewService()` handler setup when no service injected.
- Route group under `/api/v1/portfolios`.
- `DELETE` in CORS allowed methods.

### App Server Integration

`apps/api-server/internal/app/server.go` adds:

- `portfolio` module import.
- When `cfg.DatabaseURL != ""`, instantiate `portfolio.NewPostgresStore(db)` and inject `httpserver.WithPortfolioService(portfolio.NewService(portfolioStore))`.
- No DB path uses default memory store via router.

### Frontend Module Layout

```text
apps/web-admin/app/portfolios/
  page.tsx
  [portfolioId]/page.tsx
  [portfolioId]/projects/page.tsx
  [portfolioId]/health/page.tsx
```

### Frontend API Client

`apps/web-admin/lib/api.ts` adds:

- DTO types matching backend JSON responses.
- Functions:
  - `createPortfolio`
  - `fetchPortfolios`
  - `fetchPortfolio`
  - `updatePortfolio`
  - `addPortfolioProject`
  - `fetchPortfolioProjects`
  - `updatePortfolioProjectPriority`
  - `removePortfolioProject`
  - `recalculatePortfolioStatusSnapshot`
  - `fetchPortfolioStatusSnapshots`
  - `fetchPortfolioHealthSummary`
  - `fetchPortfolioCostSummary`
  - `fetchPortfolioStrategySummary`

### Frontend Navigation

`apps/web-admin/app/global-nav.tsx` adds one nav item:

- `href: '/portfolios'`
- `label: 'Portfolio 管理'`
- `match`: `/portfolios` and all subroutes.

## Output Contract

### Backend DTO Contract

Status values:

```text
PortfolioStatusActive = "active"
PortfolioStatusArchived = "archived"
PortfolioScopeManual = "manual"
PortfolioProjectRoleMember = "member"
SnapshotStatusQueued = "queued"
SnapshotStatusRunning = "running"
SnapshotStatusCompleted = "completed"
SnapshotStatusFailed = "failed"
HealthStatusHealthy = "healthy"
HealthStatusWatch = "watch"
HealthStatusCritical = "critical"
HealthStatusPending = "pending"
```

Core response fields:

```json
{
  "id": "pf_xxx",
  "name": "Novel 增长组合",
  "description": "跨项目增长组合",
  "scope_type": "manual",
  "owner_id": "growth-team",
  "health_policy": { "warning_threshold": 60, "critical_threshold": 40 },
  "status": "active",
  "project_count": 3,
  "latest_health_score": 82.5,
  "latest_health_status": "healthy",
  "estimated_monthly_cost": 386,
  "currency": "CNY",
  "created_at": "2026-05-29T10:00:00Z",
  "updated_at": "2026-05-29T10:00:00Z"
}
```

Member response fields:

```json
{
  "portfolio_id": "pf_xxx",
  "project_id": "proj_xxx",
  "project_name": "Novel Demo",
  "content_type": "novel",
  "role": "member",
  "priority": 1,
  "weight": 1,
  "note": "核心增长项目",
  "added_by": "operator",
  "created_at": "2026-05-29T10:00:00Z",
  "updated_at": "2026-05-29T10:00:00Z"
}
```

Snapshot response fields:

```json
{
  "id": "pss_xxx",
  "portfolio_id": "pf_xxx",
  "date_range": { "start": "2026-05-01", "end": "2026-05-31" },
  "health_score": 82.5,
  "health_status": "healthy",
  "total_projects": 3,
  "active_projects": 2,
  "warning_projects": 1,
  "estimated_monthly_cost": 386,
  "currency": "CNY",
  "risk_summary": { "high_risk_projects": 1 },
  "cost_summary": { "by_model": [{ "model": "claude-sonnet", "estimated_cost": 128.5 }] },
  "strategy_summary": { "pending": 2, "confirmed": 1 },
  "source_refs": [{ "source": "metrics", "source_id": "project:proj_xxx", "updated_at": "2026-05-29T10:00:00Z" }],
  "calculation_status": "completed",
  "calculated_at": "2026-05-29T10:05:00Z",
  "created_at": "2026-05-29T10:00:00Z"
}
```

Health summary response fields:

```json
{
  "portfolio_id": "pf_xxx",
  "date_range": { "start": "2026-05-01", "end": "2026-05-31" },
  "health_score": 82.5,
  "health_status": "healthy",
  "total_projects": 3,
  "active_projects": 2,
  "warning_projects": 1,
  "risk_summary": { "high_risk_projects": 1 },
  "latest_snapshot_at": "2026-05-29T10:00:00Z",
  "calculated_at": "2026-05-29T10:05:00Z",
  "source_refs": [
    { "source": "metrics", "source_id": "project:proj_xxx", "updated_at": "2026-05-29T10:00:00Z" }
  ]
}
```

Cost summary response fields:

```json
{
  "portfolio_id": "pf_xxx",
  "date_range": { "start": "2026-05-01", "end": "2026-05-31" },
  "estimated_monthly_cost": 386,
  "currency": "CNY",
  "by_model": [
    { "model": "claude-sonnet", "estimated_cost": 128.5, "currency": "CNY" }
  ],
  "project_costs": [
    { "project_id": "proj_xxx", "project_name": "Novel Demo", "estimated_cost": 128.5, "currency": "CNY" }
  ],
  "calculated_at": "2026-05-29T10:05:00Z",
  "source_refs": [
    { "source": "llm_call_logs", "source_id": "project:proj_xxx", "updated_at": "2026-05-29T10:00:00Z" }
  ]
}
```

Strategy summary response fields:

```json
{
  "portfolio_id": "pf_xxx",
  "date_range": { "start": "2026-05-01", "end": "2026-05-31" },
  "pending": 2,
  "confirmed": 1,
  "ignored": 0,
  "executed": 3,
  "execution_failed": 1,
  "top_suggestions": [
    { "project_id": "proj_xxx", "suggestion_id": "sg_xxx", "type": "optimize", "status": "pending", "title": "优化发布时间" }
  ],
  "calculated_at": "2026-05-29T10:05:00Z",
  "source_refs": [
    { "source": "strategy_suggestions", "source_id": "project:proj_xxx", "updated_at": "2026-05-29T10:00:00Z" }
  ]
}
```

### UI Contract

- `/portfolios` shows list, create form, portfolio status, scope type, owner, project count, latest health, estimated cost.
- `/portfolios/{portfolioId}` shows overview, member preview, health policy, source metadata, and strategy summary.
- `/portfolios/{portfolioId}/projects` shows member list, add project form, role, priority/weight updates, remove action.
- `/portfolios/{portfolioId}/health` shows health summary, cost summary, by-model cost, recalculate form with date range/force, snapshot history.
- All pages include loading, empty, error, and success states.
- Error states show user-readable message and `request_id` when available.

## Change Log

### Create

- `apps/api-server/internal/modules/portfolio/dto.go`
- `apps/api-server/internal/modules/portfolio/errors.go`
- `apps/api-server/internal/modules/portfolio/store.go`
- `apps/api-server/internal/modules/portfolio/memory_store.go`
- `apps/api-server/internal/modules/portfolio/postgres_store.go`
- `apps/api-server/internal/modules/portfolio/service.go`
- `apps/api-server/internal/http/handlers/portfolio.go`
- `apps/api-server/migrations/00012_create_portfolio_tables.sql`
- `apps/web-admin/app/portfolios/page.tsx`
- `apps/web-admin/app/portfolios/[portfolioId]/page.tsx`
- `apps/web-admin/app/portfolios/[portfolioId]/projects/page.tsx`
- `apps/web-admin/app/portfolios/[portfolioId]/health/page.tsx`
- `.cube/iterations/feature-10/skeleton-map.yaml`

### Modify

- `apps/api-server/internal/http/router.go`
- `apps/api-server/internal/app/server.go`
- `openapi/openapi.yaml`
- `apps/web-admin/lib/api.ts`
- `apps/web-admin/app/global-nav.tsx`

### Read Only

- `.cube/iterations/feature-10/prd.md`
- `docs/requirements/iteration-10-portfolio-management.md`
- `docs/requirements/ai-content-factory-clickable-prototype.html`
- `docs/requirements/api-contract-standard.md`
- `docs/requirements/00-product-blueprint.md`

## Development Tasks

- Task-01：Portfolio 后端模块契约
  - 任务类型：contract
  - 所属模块：api-server/portfolio
  - 简要描述：定义 Portfolio DTO、错误常量、Service 接口、Store 接口、memory store skeleton、Postgres store skeleton 和 Service skeleton。
  - 涉及接口/方法：portfolio.Service、portfolio.Store、NewService()、NewMemoryStore()、NewPostgresStore()
  - 输入：Portfolio API request DTO、portfolioID、projectID、idempotencyKey
  - 输出：Portfolio API response DTO 或 error
  - 依赖任务：无
  - 数据操作：无
  - 修改边界：只新增 Portfolio 模块类型、接口、构造函数和最小可编译占位实现
  - 禁止行为：不得实现完整业务逻辑；不得调用 WorkflowRun、AgentRuntime、ContentItem 创建或策略建议执行接口
  - 产出类型：integration
  - 功能类型：Portfolio 后端模块接口契约（type id: integration）
  - 是否跨组件：否

- Task-02：Portfolio 数据库迁移
  - 任务类型：migration
  - 所属模块：api-server/migrations
  - 简要描述：创建 project_portfolio、portfolio_project、portfolio_status_snapshot 三张表的 goose 迁移脚本。
  - 涉及接口/方法：无
  - 输入：无
  - 输出：00012_create_portfolio_tables.sql
  - 依赖任务：Task-01（Portfolio 数据模型）
  - 数据操作：写 project_portfolio 表；写 portfolio_project 表；写 portfolio_status_snapshot 表
  - 修改边界：只新增 00012_create_portfolio_tables.sql
  - 禁止行为：不得修改既有迁移文件；不得修改 docs/requirements/
  - 产出类型：sql-query
  - 功能类型：Portfolio 数据库迁移脚本（type id: sql-query）
  - 是否跨组件：否

- Task-03：Portfolio HTTP handler 与路由契约
  - 任务类型：contract
  - 所属模块：api-server/http
  - 简要描述：新增 PortfolioHandler，注册 Portfolio REST endpoints，补充 DELETE CORS 支持。
  - 涉及接口/方法：NewPortfolioHandler()、CreatePortfolio()、ListPortfolios()、GetPortfolio()、UpdatePortfolio()、AddProject()、ListProjects()、UpdateProjectPriority()、RemoveProject()、RecalculateStatusSnapshot()、ListStatusSnapshots()、GetHealthSummary()、GetCostSummary()、GetStrategySummary()
  - 输入：HTTP path/query/body/header
  - 输出：统一 API envelope
  - 依赖任务：Task-01（Portfolio Service 接口）
  - 数据操作：无
  - 修改边界：只新增 handler 并最小修改 router.go
  - 禁止行为：不得改变既有路由语义；不得跳过 bearer auth；不得同步执行多项目快照聚合
  - 产出类型：integration
  - 功能类型：Portfolio HTTP 接口契约（type id: integration）
  - 是否跨组件：是（组件链路：Router → PortfolioHandler → PortfolioService）

- Task-04：Portfolio app server 注入契约
  - 任务类型：contract
  - 所属模块：api-server/app
  - 简要描述：当 DatabaseURL 存在时创建 Postgres PortfolioStore 并通过 router option 注入。
  - 涉及接口/方法：app.NewServer()、httpserver.WithPortfolioService()
  - 输入：config.Config.DatabaseURL
  - 输出：http.Server
  - 依赖任务：Task-01（NewPostgresStore 与 NewService）
  - 数据操作：连接 PostgreSQL 句柄复用
  - 修改边界：只修改 server.go 的依赖注入
  - 禁止行为：不得新增独立数据库连接生命周期；不得影响 metrics store 注入
  - 产出类型：integration
  - 功能类型：Portfolio 服务依赖注入契约（type id: integration）
  - 是否跨组件：是（组件链路：app.NewServer → router option → PortfolioService）

- Task-05：Portfolio OpenAPI 契约
  - 任务类型：contract
  - 所属模块：openapi
  - 简要描述：在 openapi/openapi.yaml 中声明 Portfolio endpoints 和操作 ID。
  - 涉及接口/方法：createPortfolio、listPortfolios、getPortfolio、updatePortfolio、addPortfolioProject、listPortfolioProjects、updatePortfolioProjectPriority、removePortfolioProject、recalculatePortfolioStatusSnapshot、listPortfolioStatusSnapshots、getPortfolioHealthSummary、getPortfolioCostSummary、getPortfolioStrategySummary
  - 输入：OpenAPI path/request schema
  - 输出：OpenAPI path/response schema
  - 依赖任务：Task-03（HTTP endpoint 列表）
  - 数据操作：无
  - 修改边界：只修改 openapi/openapi.yaml
  - 禁止行为：不得移除或重命名既有 API path
  - 产出类型：integration
  - 功能类型：Portfolio OpenAPI 契约（type id: integration）
  - 是否跨组件：否

- Task-06：Portfolio 前端 API client 契约
  - 任务类型：contract
  - 所属模块：web-admin/api-client
  - 简要描述：在 apps/web-admin/lib/api.ts 中新增 Portfolio 类型和 request helper。
  - 涉及接口/方法：createPortfolio()、fetchPortfolios()、fetchPortfolio()、updatePortfolio()、addPortfolioProject()、fetchPortfolioProjects()、updatePortfolioProjectPriority()、removePortfolioProject()、recalculatePortfolioStatusSnapshot()、fetchPortfolioStatusSnapshots()、fetchPortfolioHealthSummary()、fetchPortfolioCostSummary()、fetchPortfolioStrategySummary()
  - 输入：Portfolio request params
  - 输出：APIEnvelope<Portfolio response>
  - 依赖任务：Task-03（HTTP endpoint contract）、Task-05（OpenAPI contract）
  - 数据操作：无
  - 修改边界：只追加 Portfolio 类型和函数
  - 禁止行为：不得绕过统一 request()；不得硬编码 secret 或 token
  - 产出类型：web-e2e
  - 功能类型：Portfolio 前端 API client 契约（type id: web-e2e）
  - 是否跨组件：是（组件链路：Next.js page → api.ts → HTTP API）

- Task-07：Portfolio 前端页面与导航骨架
  - 任务类型：contract
  - 所属模块：web-admin/portfolio
  - 简要描述：新增 Portfolio 列表、详情、项目优先级、健康成本页面骨架，并在 global-nav 中增加入口。
  - 涉及接口/方法：/portfolios、/portfolios/[portfolioId]、/portfolios/[portfolioId]/projects、/portfolios/[portfolioId]/health、GlobalNav
  - 输入：Portfolio route params 和 API client responses
  - 输出：可渲染页面骨架、loading/empty/error 状态、导航入口
  - 依赖任务：Task-06（前端 API client 契约）
  - 数据操作：UI renders cards/table with loading, empty, error states
  - 修改边界：只新增 Portfolio 页面树并最小修改 global-nav.tsx
  - 禁止行为：不得修改 docs/requirements/；不得在页面中直接拼接敏感认证信息
  - 产出类型：web-e2e
  - 功能类型：Portfolio 前端页面骨架（type id: web-e2e）
  - 是否跨组件：是（组件链路：GlobalNav → Next.js page → api.ts）
