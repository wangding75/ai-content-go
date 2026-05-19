# Iteration 6 Design：Knowledge Memory 记忆系统

## 1. 概述

本次设计在内容生成与审稿能力之后补齐项目级 Knowledge Memory 基础能力。整体方案沿用现有分层：`handlers` 负责 HTTP 契约、请求解析和统一响应，`modules/memory` 承载记忆领域 DTO、状态常量、服务接口和可测试骨架，前端通过 `apps/web-admin/lib/api.ts` 调用 `/api/v1` 接口，并在项目工作区新增记忆上下文、一致性报告、上下文预览和报告详情页面。

核心原则和约束：
- Core 命名保持内容类型无关，只使用 project、content_item、memory、snapshot、report 等通用概念。
- DynamicState 允许人工纠偏，但必须通过专门动作提交 `reason`、`changes`、`source_refs`，不得提供自由文本编辑。
- 上下文预览不落库，生成上下文快照才落库并返回 `context_snapshot_id`。
- 一致性检查必须可查询 `pending` / `running` / `completed` / `failed` 状态，验收可使用确定性规则或测试数据生成 completed 报告和结构化 issues。
- RecentContentWindow 仅提供最小策略配置，不做复杂权重、多策略模板、按 Agent 单独配置或自动调参。

## 2. Impact Analysis

| 模块 / 文件 | 影响程度 | 说明 |
|---|---|---|
| `apps/api-server/internal/modules/memory/dto.go` | 新增 | 定义 Knowledge Memory、Snapshot、Consistency Report、请求 / 响应 DTO、状态常量和领域错误。 |
| `apps/api-server/internal/modules/memory/service.go` | 新增 | 定义 Memory Service 接口与可编译骨架实现，承载纠偏、策略配置、上下文装配和报告状态契约。 |
| `apps/api-server/internal/modules/memory/executor.go` | 新增 | 定义一致性报告确定性执行器骨架，负责 pending/running/completed/failed 状态推进与结构化 issues 生成契约。 |
| `apps/api-server/internal/http/handlers/memory.go` | 新增 | 定义 memory HTTP handler、请求解析、统一响应和错误映射骨架。 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册 Iteration 6 记忆与一致性报告 API 路由。 |
| `apps/api-server/migrations/00008_create_knowledge_memory_tables.sql` | 新增 | 新增 `knowledge_memory`、`memory_snapshot`、`consistency_report` 表结构。 |
| `openapi/openapi.yaml` | 修改 | 增加 Knowledge Memory API 路径和 Schema 契约。 |
| `apps/web-admin/lib/api.ts` | 修改 | 增加记忆与一致性报告 TypeScript 类型和 API client 方法。 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 修改 | 增加“记忆上下文”“一致性报告”“上下文预览”导航入口。 |
| `apps/web-admin/app/projects/[projectId]/memory/page.tsx` | 新增 | 项目记忆上下文页面骨架。 |
| `apps/web-admin/app/projects/[projectId]/memory/context-preview/page.tsx` | 新增 | 上下文预览和生成快照页面骨架。 |
| `apps/web-admin/app/projects/[projectId]/consistency-reports/page.tsx` | 新增 | 一致性报告列表页面骨架。 |
| `apps/web-admin/app/projects/[projectId]/consistency-reports/[reportId]/page.tsx` | 新增 | 一致性报告独立详情页面骨架。 |
| `.cube/iterations/feature-6/skeleton-map.yaml` | 新增 | 记录骨架文件与 Development Tasks 的映射。 |

兼容性分析：
- 现有 API 不删除、不改名，仅新增 `/api/v1/projects/{projectId}/knowledge-memory`、`/api/v1/projects/{projectId}/consistency-reports` 等路径，向后兼容。
- `content_item` 和 `content_version` 作为来源引用被读取，不修改既有表结构。
- 前端新增项目工作区路由和导航，不改变已有页面路径。
- 新表均以 `project_id` 隔离，不跨项目污染上下文。

## 3. Flow Design

### 3.1 查看与维护项目记忆

1. 用户从项目工作区导航进入 `/projects/:projectId/memory`。
2. 页面调用 `GET /api/v1/projects/{projectId}/knowledge-memory`。
3. Handler 校验 projectId 并调用 `memory.Service.GetKnowledgeMemory`。
4. Service 返回 StaticContext、DynamicState、RecentContentWindowPolicy、StyleGuide、version、updated_at 和 recent_snapshot_summary。
5. 页面渲染四类上下文、策略、版本、recent_snapshot_summary、分页快照列表和错误态；快照列表必须展示来源、Token 预算、估算 Token、截断策略、创建时间，并支持空态、加载态和错误态。

异常流程：项目不存在返回 `NOT_FOUND`；无权限返回 `FORBIDDEN`；接口失败按统一错误结构展示 `request_id`。

### 3.2 人工修正与纠偏

- StaticContext 修正：用户提交 `static_context` 与 `note`，服务生成新版本并返回 `operation_log_id`。
- StyleGuide 修正：用户提交 `style_guide` 与 `note`，服务生成新版本并返回 `operation_log_id`。
- DynamicState 纠偏：用户提交 `reason`、`changes`、`source_refs`，服务只做纠偏动作，不允许自由覆盖整段文本；成功后生成新的 `memory_snapshot` 并返回 `operation_log_id`。
- RecentContentWindow 最小策略配置：用户提交 `item_count`、`token_limit`、`truncation_policy`，服务更新策略并记录日志。

异常流程：必填字段缺失返回 `VALIDATION_ERROR`；来源引用无效返回 `VALIDATION_ERROR`；版本冲突返回 `CONFLICT`；日志或快照写入失败时操作不伪成功。

### 3.3 上下文预览与快照生成

1. 用户进入 `/projects/:projectId/memory/context-preview`。
2. 点击“预览上下文”时调用 `GET /api/v1/projects/{projectId}/knowledge-memory/context-preview`，该动作不落库。
3. 系统按 `StaticContext -> StyleGuide -> DynamicState -> RecentContentWindow -> 当前任务输入` 顺序装配预览结果。
4. 用户点击“生成上下文快照”时调用 `POST /api/v1/projects/{projectId}/knowledge-memory/assemble-context`。
5. 该接口同步完成快照生成，Service 写入 `memory_snapshot` 并返回 `context_snapshot_id`、`estimated_tokens`、`truncation_policy`。

异常流程：预算不合法返回 `VALIDATION_ERROR`；幂等键重复但请求体不同返回 `IDEMPOTENCY_CONFLICT`；无来源时预览返回空来源和说明。

### 3.4 动态状态更新

1. 内容生成、审稿或用户动作基于内容单元提交 summary、changes、source_version_id。
2. Handler 调用 `UpdateDynamicState`。
3. Service 校验内容单元与来源版本，更新 DynamicState，并生成 `memory_snapshot`。
4. 返回 `memory_snapshot_id` 和 `dynamic_state_version`。

异常流程：内容单元不存在返回 `NOT_FOUND`；来源版本无效返回 `VALIDATION_ERROR`；重复提交通过幂等键返回相同业务结果。

### 3.5 一致性报告

1. 用户在 `/projects/:projectId/consistency-reports` 点击创建一致性检查。
2. `POST /api/v1/projects/{projectId}/consistency-reports` 写入 `consistency_report(status=pending)`，返回 `report_id` 和初始状态。
3. `memory.ReportExecutor` 接收 report_id，状态推进为 `running`，执行确定性规则或测试数据生成。
4. 执行成功时写入结构化 issues、issue_count、severity_summary、completed_at，并将状态置为 `completed`；执行失败时写入 error_code/error_message 并将状态置为 `failed`。
5. 列表页展示状态、创建时间、issue_count、severity_summary。
6. 用户进入 `/projects/:projectId/consistency-reports/:reportId` 查看结构化问题项。

异常流程：检查范围不合法返回 `VALIDATION_ERROR`；报告不存在返回 `NOT_FOUND`；执行失败进入 `failed` 并保留错误原因。

## 4. Table Design

目标方言：PostgreSQL。

### 4.1 `knowledge_memory`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | 记忆记录 ID。 |
| `project_id` | TEXT | NOT NULL, UNIQUE | 项目边界。 |
| `static_context` | JSONB | NOT NULL DEFAULT '{}'::jsonb | 长期稳定资料摘要与规则。 |
| `dynamic_state` | JSONB | NOT NULL DEFAULT '{}'::jsonb | 当前动态状态，不保存完整正文。 |
| `recent_window_policy` | JSONB | NOT NULL DEFAULT '{}'::jsonb | item_count、token_limit、truncation_policy。 |
| `style_guide` | JSONB | NOT NULL DEFAULT '{}'::jsonb | 风格、语气、禁用表达等。 |
| `version` | INTEGER | NOT NULL DEFAULT 1 | 记忆版本。 |
| `operation_log_id` | TEXT | NULL | 最近一次人工修正或纠偏日志。 |
| `created_at` | TIMESTAMPTZ | NOT NULL | 创建时间。 |
| `updated_at` | TIMESTAMPTZ | NOT NULL | 更新时间。 |

索引：`idx_knowledge_memory_project(project_id)`。

### 4.2 `memory_snapshot`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | 快照 ID。 |
| `project_id` | TEXT | NOT NULL | 项目 ID。 |
| `content_item_id` | TEXT | NULL | 关联内容单元。 |
| `source_type` | TEXT | NOT NULL | `assemble_context` / `dynamic_state_update` / `dynamic_state_correction`。 |
| `source_id` | TEXT | NOT NULL DEFAULT '' | 来源业务记录 ID。 |
| `assembled_context` | JSONB | NOT NULL DEFAULT '{}'::jsonb | 装配结果或状态快照。 |
| `source_refs` | JSONB | NOT NULL DEFAULT '[]'::jsonb | 来源引用。 |
| `token_budget` | INTEGER | NOT NULL DEFAULT 0 | Token 预算。 |
| `estimated_tokens` | INTEGER | NOT NULL DEFAULT 0 | 估算 Token。 |
| `truncation_policy` | TEXT | NOT NULL DEFAULT 'time' | 截断策略。 |
| `triggered_by` | TEXT | NOT NULL DEFAULT '' | 触发来源。 |
| `created_at` | TIMESTAMPTZ | NOT NULL | 创建时间。 |

索引：`idx_memory_snapshot_project_created(project_id, created_at DESC)`、`idx_memory_snapshot_content_item(content_item_id)`。

### 4.3 `consistency_report`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | 报告 ID。 |
| `project_id` | TEXT | NOT NULL | 项目 ID。 |
| `range` | JSONB | NOT NULL DEFAULT '{}'::jsonb | 检查范围。 |
| `scope` | TEXT | NOT NULL DEFAULT 'project' | 检查作用域。 |
| `severity_threshold` | TEXT | NOT NULL DEFAULT 'low' | 严重级别阈值。 |
| `status` | TEXT | NOT NULL CHECK | `pending` / `running` / `completed` / `failed`。 |
| `issue_count` | INTEGER | NOT NULL DEFAULT 0 | 问题数量。 |
| `severity_summary` | JSONB | NOT NULL DEFAULT '{}'::jsonb | 严重级别汇总。 |
| `issues` | JSONB | NOT NULL DEFAULT '[]'::jsonb | 结构化 issues。 |
| `source_snapshot_id` | TEXT | NULL | 来源快照。 |
| `error_code` | TEXT | NULL | 失败错误码。 |
| `error_message` | TEXT | NULL | 失败信息。 |
| `created_at` | TIMESTAMPTZ | NOT NULL | 创建时间。 |
| `completed_at` | TIMESTAMPTZ | NULL | 完成时间。 |

索引：`idx_consistency_report_project_created(project_id, created_at DESC)`、`idx_consistency_report_project_status_created(project_id, status, created_at DESC)`、`idx_consistency_report_project_severity(project_id, severity_threshold)`。

### 4.4 `idempotency_record`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | 幂等记录 ID。 |
| `scope` | TEXT | NOT NULL | 资源范围，例如 project 或 content_item。 |
| `endpoint` | TEXT | NOT NULL | 幂等接口标识。 |
| `idempotency_key` | TEXT | NOT NULL | 请求头中的 Idempotency-Key。 |
| `request_hash` | TEXT | NOT NULL | 规范化请求体哈希。 |
| `response_ref_type` | TEXT | NOT NULL | `memory_snapshot` / `consistency_report` 等结果类型。 |
| `response_ref_id` | TEXT | NOT NULL | 首次成功产生的业务记录 ID。 |
| `created_at` | TIMESTAMPTZ | NOT NULL | 创建时间。 |

约束：`UNIQUE(scope, endpoint, idempotency_key)`。相同 scope、endpoint、idempotency_key 且 request_hash 相同的重放返回原 response_ref；request_hash 不同返回 `IDEMPOTENCY_CONFLICT`。

索引：`idx_idempotency_scope_endpoint(scope, endpoint)`。

SQL Contract：本迭代不生成动态 SQL；迁移 DDL 固定，目标方言 PostgreSQL。禁止模式：记忆表无 project_id 隔离、快照缺少来源引用、报告 issues 只保存不可解析长文本、列表查询缺少 project/status/created_at 索引、幂等敏感接口缺少持久化幂等记录。

## 5. API Design

所有响应均使用统一 envelope：`success/data/error/request_id`。所有 Iteration 6 API（包含 GET 查询接口和所有 mutating API）都必须要求 Bearer Token。幂等敏感接口使用 `Idempotency-Key`。

| Method | Path | 用途 | 成功状态 |
|---|---|---|---|
| GET | `/api/v1/projects/{projectId}/knowledge-memory` | 查看项目记忆 | 200 |
| PATCH | `/api/v1/projects/{projectId}/knowledge-memory/static-context` | 修正 StaticContext | 200 |
| PATCH | `/api/v1/projects/{projectId}/knowledge-memory/style-guide` | 修正 StyleGuide | 200 |
| PATCH | `/api/v1/projects/{projectId}/knowledge-memory/dynamic-state-correction` | 人工纠偏 DynamicState | 200 |
| PATCH | `/api/v1/projects/{projectId}/knowledge-memory/recent-window-policy` | 配置 RecentContentWindow 最小策略 | 200 |
| GET | `/api/v1/projects/{projectId}/knowledge-memory/snapshots` | 记忆快照列表 | 200 |
| GET | `/api/v1/projects/{projectId}/knowledge-memory/context-preview` | 预览上下文，不落库 | 200 |
| POST | `/api/v1/projects/{projectId}/knowledge-memory/assemble-context` | 生成上下文快照 | 200 |
| POST | `/api/v1/content-items/{id}/update-dynamic-state` | 更新动态状态 | 200 |
| POST | `/api/v1/projects/{projectId}/consistency-reports` | 创建一致性报告 | 202 |
| GET | `/api/v1/projects/{projectId}/consistency-reports` | 报告列表 | 200 |
| GET | `/api/v1/projects/{projectId}/consistency-reports/{reportId}` | 报告详情 | 200 |

### 5.1 GET `/api/v1/projects/{projectId}/knowledge-memory`

Path：`projectId` 必填。

Response：`KnowledgeMemoryResponse { id, project_id, static_context, dynamic_state, recent_window_policy, style_guide, version, updated_at, recent_snapshot_summary }`。

错误码：`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`。

### 5.2 PATCH StaticContext / StyleGuide

Request：`UpdateStaticContextRequest { static_context, note }`；`UpdateStyleGuideRequest { style_guide, note }`。

Response：`MemoryUpdateResponse { version, operation_log_id }`。

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`INTERNAL_ERROR`。

### 5.3 PATCH DynamicState Correction

Header：`Idempotency-Key` 必填。

Request：`CorrectDynamicStateRequest { reason, changes, source_refs }`。

Response：`DynamicStateCorrectionResponse { memory_snapshot_id, dynamic_state_version, operation_log_id }`。

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`。

### 5.4 PATCH Recent Window Policy

Request：`UpdateRecentWindowPolicyRequest { item_count, token_limit, truncation_policy }`。

Response：`RecentWindowPolicyResponse { item_count, token_limit, truncation_policy, version, operation_log_id }`。

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`INTERNAL_ERROR`。

### 5.5 GET Snapshots / Context Preview

Snapshots Query：`content_item_id` 可选，`page/page_size/sort/order` 可选。

Preview Query：`content_item_id` 可选，`purpose` 必填，`budget` 必填。

Responses：`PagedMemorySnapshotsResponse`；`ContextPreviewResponse { sources, token_budget, estimated_tokens, truncation_policy, preview_text }`。

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`。

### 5.6 POST Assemble Context / Update Dynamic State

Header：`Idempotency-Key` 必填。

Requests：`AssembleContextRequest { purpose, budget, content_item_id }`；`UpdateDynamicStateRequest { summary, changes, source_version_id }`。

Responses：`AssembleContextResponse { context_snapshot_id, estimated_tokens, truncation_policy }`；`UpdateDynamicStateResponse { memory_snapshot_id, dynamic_state_version }`。

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`。

### 5.7 Consistency Reports

Create Header：`Idempotency-Key` 必填。

Create Request：`CreateConsistencyReportRequest { range, scope, severity_threshold }`。

List Query：`status` 可选，`page/page_size/sort/order` 可选。

Detail Response：`ConsistencyReportDetailResponse`，issues 至少包含 `issue_id`、`severity`、`type`、`title`、`description`、`affected_content_items`、`suggestion`。

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`。

## 6. Module Design

### 6.1 Backend `memory` module

职责：
- 管理项目级 Knowledge Memory 四类上下文。
- 管理 DynamicState 人工纠偏和 RecentContentWindow 最小策略。
- 提供上下文预览与快照生成契约。
- 提供一致性报告创建、列表和详情查询契约。
- 返回标准领域错误，供 Handler 映射统一错误码。

服务接口：
- `GetKnowledgeMemory(ctx, projectID) (KnowledgeMemoryResponse, error)`
- `UpdateStaticContext(ctx, projectID, req) (MemoryUpdateResponse, error)`
- `UpdateStyleGuide(ctx, projectID, req) (MemoryUpdateResponse, error)`
- `CorrectDynamicState(ctx, projectID, req, idempotencyKey) (DynamicStateCorrectionResponse, error)`
- `UpdateRecentWindowPolicy(ctx, projectID, req) (RecentWindowPolicyResponse, error)`
- `ListSnapshots(ctx, projectID, req) (PagedMemorySnapshotsResponse, error)`
- `PreviewContext(ctx, projectID, req) (ContextPreviewResponse, error)`
- `AssembleContext(ctx, projectID, req, idempotencyKey) (AssembleContextResponse, error)`
- `UpdateDynamicState(ctx, contentItemID, req, idempotencyKey) (UpdateDynamicStateResponse, error)`
- `CreateConsistencyReport(ctx, projectID, req, idempotencyKey) (CreateConsistencyReportResponse, error)`
- `ListConsistencyReports(ctx, projectID, req) (PagedConsistencyReportsResponse, error)`
- `GetConsistencyReport(ctx, projectID, reportID) (ConsistencyReportDetailResponse, error)`

### 6.2 Consistency Report Executor

职责：
- 接收 `report_id`，将报告从 `pending` 推进到 `running`。
- 使用确定性规则或测试数据生成 completed 报告和结构化 issues。
- 成功时写入 issue_count、severity_summary、issues、completed_at，并置为 `completed`。
- 失败时写入 error_code、error_message，并置为 `failed`。
- 暴露 `RunConsistencyReport(ctx, reportID) (ConsistencyReportDetailResponse, error)` 供 Service 或 worker 调用。

### 6.3 HTTP Handler

职责：
- 从 chi path/query/header/body 读取输入。
- 调用 Memory Service。
- 将领域错误映射为统一 API 错误码。
- 对一致性报告创建返回 202；上下文装配和动态状态更新同步完成并返回 200。

### 6.4 Frontend

- `lib/api.ts` 增加 memory 类型与函数。
- `ProjectWorkspaceNav` 增加 `memory`、`consistency-reports`、`memory/context-preview` 项。
- 新增四个页面，全部使用 `page-shell`、`page-hero`、`card`、`app-nav` 等现有样式模式，展示 loading / empty / error / success。

## 7. Output Contract

| 产物 | 输入 | 输出 | 正确性规则 | 类型 | 测试规范 |
|---|---|---|---|---|---|
| Memory DTO contracts | PRD/API Design | Go DTO、状态常量、错误变量 | recent_snapshot_summary、结构化 issues、状态枚举和错误变量完整 | library | `standards/testing/library.md` |
| Memory HTTP APIs | HTTP request + Bearer token | Envelope JSON | 状态码、schema、错误码、request_id 正确 | web-e2e | `standards/testing/web-e2e.md` |
| Memory service | DTO + context + idempotency key | DTO / error | 纠偏必填、策略校验、快照 ID、operation_log_id、报告状态规则正确 | integration | `standards/testing/integration.md` |
| Memory SQL tables | Migration SQL | PostgreSQL tables | DDL 可执行，约束和索引符合 Table Design | sql-query | `standards/testing/sql-query.md` |
| Web Admin memory pages | 用户导航和 API 响应 | 渲染页面 | 导航可达、刷新不 404、错误态展示 request_id、按钮有反馈 | web-e2e | `standards/testing/web-e2e.md` |

`workflow.yaml` 当前 `project.features` 为 `[]`，但本迭代新增 HTTP endpoint、SQL migration、领域 DTO 契约和前后端跨组件链路，因此按功能本身显式引用 `library`、`web-e2e`、`integration`、`sql-query`。

跨组件链路：
- `HTTP Router -> MemoryHandler -> Memory Service -> API Response`
- `Web Admin Page -> lib/api.ts -> API Server -> Memory Service`
- `Migration DDL -> Store schema contract -> Service response contract`
- `ConsistencyReportsPage -> API Client -> MemoryHandler -> Memory Service -> deterministic executor/worker contract -> consistency_report store -> List/Detail API`

SQL Contract：本迭代不生成动态 SQL；迁移 DDL 固定，目标方言 PostgreSQL。禁止模式见 Table Design。

## 8. Change Log

| 文件 | 变更类型 | 原因 |
|---|---|---|
| `apps/api-server/internal/modules/memory/dto.go` | 新增 | 提供记忆、快照、报告 DTO、状态常量和错误码骨架。 |
| `apps/api-server/internal/modules/memory/service.go` | 新增 | 提供 Memory Service 接口和可编译骨架实现。 |
| `apps/api-server/internal/modules/memory/executor.go` | 新增 | 提供一致性报告执行器骨架，明确 pending/running/completed/failed 状态推进。 |
| `apps/api-server/internal/http/handlers/memory.go` | 新增 | 提供 Knowledge Memory HTTP handler 骨架。 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册 Iteration 6 API 路由。 |
| `apps/api-server/migrations/00008_create_knowledge_memory_tables.sql` | 新增 | 新增记忆、快照、一致性报告表结构。 |
| `openapi/openapi.yaml` | 修改 | 增加 Knowledge Memory API 契约。 |
| `apps/web-admin/lib/api.ts` | 修改 | 增加记忆与一致性报告 API client 契约。 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 修改 | 增加记忆与一致性报告导航入口。 |
| `apps/web-admin/app/projects/[projectId]/memory/page.tsx` | 新增 | 实现记忆上下文页面骨架。 |
| `apps/web-admin/app/projects/[projectId]/memory/context-preview/page.tsx` | 新增 | 实现上下文预览页面骨架。 |
| `apps/web-admin/app/projects/[projectId]/consistency-reports/page.tsx` | 新增 | 实现一致性报告列表页面骨架。 |
| `apps/web-admin/app/projects/[projectId]/consistency-reports/[reportId]/page.tsx` | 新增 | 实现一致性报告详情页面骨架。 |
| `.cube/iterations/feature-6/skeleton-map.yaml` | 新增 | 映射骨架文件和开发任务。 |

## 9. Development Tasks

- Task-01：定义记忆领域 DTO 与状态常量
  - 所属模块：memory
  - 简要描述：定义 knowledge_memory、memory_snapshot、consistency_report、idempotency_record 对应请求响应、状态常量和领域错误。
  - 涉及接口/方法：memory DTO types
  - 输入：记忆、快照、报告、幂等相关 JSON 字段
  - 输出：Go DTO、状态常量、错误变量
  - 产出类型：library
  - 功能类型：领域模型契约（type id: library）
  - 是否跨组件：否
- Task-02：设计并新增记忆数据库迁移
  - 所属模块：store
  - 简要描述：新增 knowledge_memory、memory_snapshot、consistency_report、idempotency_record 表、约束和索引。
  - 涉及接口/方法：migration 00008
  - 输入：PostgreSQL migration
  - 输出：可执行 DDL
  - 产出类型：sql-query
  - 功能类型：固定 DDL（type id: sql-query）
  - 是否跨组件：否
- Task-03：实现 Memory Service 骨架与状态接口
  - 所属模块：memory
  - 简要描述：提供记忆读取、StaticContext 修正、StyleGuide 修正、DynamicState 纠偏、策略配置、快照、预览、同步装配、动态状态更新和报告接口。
  - 涉及接口/方法：memory.Service
  - 输入：context、请求 DTO、项目 ID、内容单元 ID、幂等键
  - 输出：响应 DTO 或领域错误
  - 产出类型：integration
  - 功能类型：服务状态机（type id: integration）
  - 是否跨组件：是（HTTP Handler -> Memory Service -> operation_log/snapshot/report/idempotency contract）
- Task-04：实现一致性报告执行器骨架
  - 所属模块：memory
  - 简要描述：新增 ReportExecutor，负责 report_id 接收、pending->running、completed/failed 状态推进、结构化 issues 生成和失败原因持久化契约。
  - 涉及接口/方法：ReportExecutor.RunConsistencyReport()
  - 输入：context、reportID
  - 输出：ConsistencyReportDetailResponse 或领域错误
  - 产出类型：integration
  - 功能类型：异步报告执行契约（type id: integration）
  - 是否跨组件：是（Service -> Executor -> report store contract）
- Task-05：实现记忆 HTTP API 骨架与路由注册
  - 所属模块：http
  - 简要描述：新增 memory handler 并注册所有 Iteration 6 API 路由。
  - 涉及接口/方法：MemoryHandler methods, NewRouter()
  - 输入：HTTP path/query/header/body
  - 输出：统一 envelope JSON
  - 产出类型：web-e2e
  - 功能类型：REST API（type id: web-e2e）
  - 是否跨组件：是（HTTP Router -> Handler -> Memory Service）
- Task-06：补充 OpenAPI 记忆接口契约
  - 所属模块：openapi
  - 简要描述：为所有记忆与一致性报告 API 增加 path、request、response、错误码和 security；所有 Iteration 6 GET 与 mutating endpoint 都必须声明 Bearer Token security。
  - 涉及接口/方法：openapi.yaml
  - 输入：API Design
  - 输出：OpenAPI 3.0 契约
  - 产出类型：web-e2e
  - 功能类型：API contract（type id: web-e2e）
  - 是否跨组件：否
- Task-07：扩展 Web Admin Memory API client
  - 所属模块：web-admin
  - 简要描述：在 lib/api.ts 增加记忆与一致性报告类型和 API 调用函数。
  - 涉及接口/方法：fetchKnowledgeMemory(), updateStaticContext(), updateStyleGuide(), correctDynamicState(), updateRecentWindowPolicy(), fetchMemorySnapshots(), previewContext(), assembleContext(), updateDynamicState(), createConsistencyReport(), fetchConsistencyReports(), fetchConsistencyReport()
  - 输入：页面参数和表单数据
  - 输出：APIEnvelope 响应
  - 产出类型：web-e2e
  - 功能类型：前端 API client（type id: web-e2e）
  - 是否跨组件：是（Web Page -> API Client -> HTTP API）
- Task-08：实现项目记忆上下文页面与导航入口
  - 所属模块：web-admin
  - 简要描述：新增 `/projects/:projectId/memory`，支持四类上下文展示、StaticContext / StyleGuide 修正、DynamicState 纠偏、策略配置、分页快照列表和错误态。
  - 涉及接口/方法：MemoryPage, ProjectWorkspaceNav, fetchMemorySnapshots()
  - 输入：projectId、表单字段、分页
  - 输出：记忆上下文页面、分页快照列表、加载/空/错误/成功态
  - 产出类型：web-e2e
  - 功能类型：Web UI（type id: web-e2e）
  - 是否跨组件：是（Navigation -> Page -> API Client）
- Task-09：实现上下文预览页面
  - 所属模块：web-admin
  - 简要描述：新增 `/projects/:projectId/memory/context-preview`，支持预览上下文和生成上下文快照两个动作。
  - 涉及接口/方法：ContextPreviewPage
  - 输入：projectId、purpose、budget、content_item_id
  - 输出：上下文来源、Token、截断策略、preview_text、context_snapshot_id
  - 产出类型：web-e2e
  - 功能类型：Web UI + sync API（type id: web-e2e）
  - 是否跨组件：是（Page -> API Client -> HTTP API -> Memory Service）
- Task-10：实现一致性报告列表页面
  - 所属模块：web-admin
  - 简要描述：新增 `/projects/:projectId/consistency-reports`，支持列表、状态筛选、创建报告、严重级别汇总和详情入口。
  - 涉及接口/方法：ConsistencyReportsPage
  - 输入：projectId、status、range、scope、severity_threshold
  - 输出：报告列表页面、report_id/status 反馈
  - 产出类型：web-e2e
  - 功能类型：Web UI + async API（type id: web-e2e）
  - 是否跨组件：是（Page -> API Client -> HTTP API -> Memory Service -> ReportExecutor）
- Task-11：实现一致性报告详情页面
  - 所属模块：web-admin
  - 简要描述：新增 `/projects/:projectId/consistency-reports/:reportId`，展示报告状态、来源快照、结构化 issues、空态、加载态、错误态和失败原因。
  - 涉及接口/方法：ConsistencyReportDetailPage, fetchConsistencyReport()
  - 输入：projectId、reportId
  - 输出：报告详情页面
  - 产出类型：web-e2e
  - 功能类型：Web UI（type id: web-e2e）
  - 是否跨组件：是（Page -> API Client -> HTTP API）

