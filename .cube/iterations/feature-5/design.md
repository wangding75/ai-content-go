# Iteration 5 Design：审稿与质量控制

## 1. 概述

本次设计在现有内容生成闭环之后补齐通用审稿与质量控制能力。整体方案沿用当前 Go API Server 的分层方式：`handlers` 负责 HTTP 契约与统一响应，`modules/review` 承载审稿领域 DTO、状态与服务接口，前端通过 `apps/web-admin/lib/api.ts` 调用 `/api/v1` 接口并在项目工作区新增审稿导航与页面。

核心约束：
- Core 层使用 `content_review`、`content_version`、`review_report` 等通用概念，不引入 Book / Chapter / Novel 作为核心资源。
- 审稿状态变更必须写入 `operation_log`，失败时不允许静默成功。
- AI 质检生成和打回重生成都按异步任务处理，HTTP 接口只返回 `run_id` / `job_id` / `workflow_run_id` 等可追踪 ID。
- 编辑后通过的编辑字段由 Content Pack / content type 决定，Core 只保存通用 JSONB patch 与版本快照。

## 2. Impact Analysis

| 模块 / 文件 | 影响程度 | 说明 |
|---|---|---|
| `apps/api-server/internal/modules/review/dto.go` | 新增 | 定义审稿、版本、质检报告、请求 / 响应 DTO、状态常量和错误。 |
| `apps/api-server/internal/modules/review/service.go` | 新增 | 定义 Review Service 接口与骨架实现，承载状态流转、质检触发、版本创建契约。 |
| `apps/api-server/internal/http/handlers/review.go` | 新增 | 定义 HTTP Handler、请求解析、统一响应和错误映射骨架。 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册 Iteration 5 审稿相关路由。 |
| `apps/api-server/migrations/00007_create_content_review_tables.sql` | 新增 | 新增 `content_review`、`content_version`、`review_report` 表。 |
| `openapi/openapi.yaml` | 修改 | 增加审稿 API 路径、Schema 与响应契约。 |
| `apps/web-admin/lib/api.ts` | 修改 | 增加审稿相关 TypeScript 类型和 API client 方法。 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 修改 | 增加项目工作区“审稿中心”导航入口。 |
| `apps/web-admin/app/content-items/[itemId]/page.tsx` | 修改 | 修复 Next.js 15 动态路由 params 类型，解除前端构建阻塞。 |
| `apps/web-admin/app/generation-runs/[runId]/page.tsx` | 修改 | 修复 Next.js 15 动态路由 params 类型，解除前端构建阻塞。 |
| `apps/web-admin/app/projects/[projectId]/reviews/page.tsx` | 新增 | 项目审稿中心页面。 |
| `apps/web-admin/app/content-reviews/[reviewId]/page.tsx` | 新增 | 审稿详情页面。 |
| `apps/web-admin/app/content-reviews/[reviewId]/ai-report/page.tsx` | 新增 | AI 质检报告页面。 |
| `apps/web-admin/app/content-reviews/[reviewId]/edit-approve/page.tsx` | 新增 | 编辑后通过页面。 |
| `.cube/iterations/feature-5/skeleton-map.yaml` | 新增 | 记录骨架文件与 Development Tasks 的映射。 |

兼容性分析：
- 现有 API 不删除、不改名，仅新增 `/api/v1/content-reviews`、`/api/v1/content-items/{id}/reviews`、`/api/v1/content-items/{id}/versions` 等路径，向后兼容。
- `content_item` 表只通过外键被引用，不在本阶段修改原表结构。
- 前端新增路由和导航，不改变已有页面路径。

## 3. Flow Design

### 3.1 创建审稿

1. 用户在审稿中心或内容单元入口选择内容单元并提交审稿类型。
2. `ReviewHandler.CreateReview` 读取 `contentItemID`、请求体和 `Idempotency-Key`。
3. `review.Service.CreateReview` 校验内容单元状态、审稿类型和重复创建规则。
4. 服务创建 `content_review` 记录，状态为 `pending`，并返回 `review_id`。
5. 页面展示成功反馈并跳转或刷新列表。

异常流程：内容单元不存在返回 `NOT_FOUND`；状态不允许或重复创建返回 `CONFLICT`；缺少审稿类型或幂等键返回 `VALIDATION_ERROR`。

### 3.2 查看审稿与版本

1. 列表页调用 `ListReviews(project_id, status, page, sort)`。
2. 详情页调用 `GetReview(review_id)`，返回正文、报告摘要和版本入口所需数据。
3. 版本历史调用 `ListContentVersions(content_item_id)`，返回按版本号倒序的版本列表。

异常流程：资源不存在返回 `NOT_FOUND`；非法分页 / 状态返回 `VALIDATION_ERROR`；无数据返回空 `items`。

### 3.3 AI 质检报告

1. 用户在 AI 质检报告页点击“生成质检报告”。
2. Handler 校验 `Idempotency-Key`，调用 Workflow Engine 或任务提交接口创建异步任务。
3. `review.Service.TriggerAIReport` 记录报告状态为 `generating`，返回 `report_id`、`job_id` / `workflow_run_id`。
4. 页面展示生成中状态，用户可刷新或重新进入报告页查看结果。

异常流程：重复触发且请求体不同返回 `IDEMPOTENCY_CONFLICT`；审稿状态不允许质检返回 `CONFLICT`；任务创建失败返回 `WORKFLOW_RUN_FAILED` 或 `LLM_PROVIDER_ERROR`。

### 3.4 通过、打回、编辑后通过

- 通过：`ApproveReview` 校验状态为可审核状态，更新 `content_review.status=approved`，写入 `operation_log`，返回 `operation_log_id`。
- 仅打回：`RejectReview` 校验原因，更新 `content_review.status=rejected`，写入 `operation_log`，不创建重生成任务。
- 打回并重生成：`RejectReview` 在状态更新和日志写入成功后创建异步重生成任务，并返回 `regeneration_run_id` 或 `job_id`。
- 编辑后通过：`ApproveWithEdit` 校验 Content Pack 提供的 editable fields，创建 `content_version`，更新审稿为 `approved_with_edit`，写入 `operation_log`。

异常流程：状态流转非法返回 `CONFLICT`；日志写入失败时整个状态变更失败；重生成任务创建失败时返回失败响应，不让用户误认为已触发重生成。

## 4. Table Design

目标方言：PostgreSQL。

### 4.1 `content_review`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | 审稿记录 ID。 |
| `project_id` | TEXT | NOT NULL | 项目 ID。 |
| `content_item_id` | TEXT | NOT NULL, references `content_item(id)` | 被审内容单元。 |
| `review_type` | TEXT | NOT NULL | `manual` / `ai` / `combined`。 |
| `status` | TEXT | NOT NULL CHECK | `pending` / `in_review` / `approved` / `rejected` / `approved_with_edit`。 |
| `current_version_id` | TEXT | NULL | 当前关联内容版本。 |
| `report_id` | TEXT | NULL | 当前质检报告。 |
| `note` | TEXT | NOT NULL DEFAULT '' | 通过备注。 |
| `reject_reason` | TEXT | NOT NULL DEFAULT '' | 打回原因。 |
| `regenerate_instruction` | TEXT | NOT NULL DEFAULT '' | 重生成说明。 |
| `operation_log_id` | TEXT | NULL | 最近一次状态操作日志。 |
| `created_at` | TIMESTAMPTZ | NOT NULL | 创建时间。 |
| `updated_at` | TIMESTAMPTZ | NOT NULL | 更新时间。 |

索引：`idx_content_review_project_status(project_id, status)`、`idx_content_review_content_item(content_item_id)`、`idx_content_review_updated_at(updated_at DESC)`。

### 4.2 `content_version`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | 内容版本 ID。 |
| `content_item_id` | TEXT | NOT NULL, references `content_item(id)` | 内容单元。 |
| `project_id` | TEXT | NOT NULL | 项目 ID。 |
| `version_no` | INTEGER | NOT NULL CHECK > 0 | 版本号。 |
| `source` | TEXT | NOT NULL | `generation` / `edit_approve` / `regeneration`。 |
| `title` | TEXT | NOT NULL DEFAULT '' | 标题快照。 |
| `body` | TEXT | NOT NULL DEFAULT '' | 正文快照。 |
| `editable_fields` | JSONB | NOT NULL DEFAULT '{}'::jsonb | Content Pack 字段快照。 |
| `summary` | TEXT | NOT NULL DEFAULT '' | 可读摘要。 |
| `created_at` | TIMESTAMPTZ | NOT NULL | 创建时间。 |

约束：`UNIQUE(content_item_id, version_no)`。索引：`idx_content_version_item_version(content_item_id, version_no DESC)`。

### 4.3 `review_report`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | TEXT | PRIMARY KEY | 报告 ID。 |
| `review_id` | TEXT | NOT NULL, references `content_review(id)` | 审稿记录。 |
| `content_item_id` | TEXT | NOT NULL | 内容单元。 |
| `status` | TEXT | NOT NULL CHECK | `pending` / `generating` / `succeeded` / `failed`。 |
| `quality_score` | INTEGER | NULL CHECK 0-100 | 质量分。 |
| `risk_level` | TEXT | NOT NULL DEFAULT 'unknown' | `low` / `medium` / `high` / `unknown`。 |
| `issues` | JSONB | NOT NULL DEFAULT '[]'::jsonb | 问题项。 |
| `suggestions` | JSONB | NOT NULL DEFAULT '[]'::jsonb | 建议项。 |
| `job_id` | TEXT | NULL | 异步任务 ID。 |
| `workflow_run_id` | TEXT | NULL | 工作流运行 ID。 |
| `error_code` | TEXT | NULL | 失败错误码。 |
| `error_message` | TEXT | NULL | 失败信息。 |
| `created_at` | TIMESTAMPTZ | NOT NULL | 创建时间。 |
| `updated_at` | TIMESTAMPTZ | NOT NULL | 更新时间。 |

索引：`idx_review_report_review_id(review_id)`、`idx_review_report_status(status)`。

## 5. API Design

所有响应均使用统一 envelope：`success/data/error/request_id`。

| Method | Path | 用途 | 成功状态 |
|---|---|---|---|
| GET | `/api/v1/content-reviews` | 审稿列表 | 200 |
| POST | `/api/v1/content-items/{id}/reviews` | 创建审稿 | 201 |
| GET | `/api/v1/content-reviews/{id}` | 审稿详情 | 200 |
| POST | `/api/v1/content-reviews/{id}/ai-report` | 触发 AI 质检生成 | 202 |
| GET | `/api/v1/content-reviews/{id}/ai-report` | 查看 AI 质检报告 | 200 |
| POST | `/api/v1/content-reviews/{id}/approve` | 通过审稿 | 200 |
| POST | `/api/v1/content-reviews/{id}/reject` | 仅打回或打回并重生成 | 202 |
| POST | `/api/v1/content-reviews/{id}/approve-with-edit` | 编辑后通过 | 200 |
| GET | `/api/v1/content-items/{id}/versions` | 版本历史 | 200 |

### 5.1 GET `/api/v1/content-reviews`

Query：`project_id` 必填，`status` 可选，`page/page_size/sort/order` 可选。

响应：`PagedContentReviewsResponse`，包含 `items` 和 `pagination`。

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`。

### 5.2 POST `/api/v1/content-items/{id}/reviews`

Header：`Idempotency-Key` 必填。

Request：`CreateReviewRequest { review_type }`。

Response：`CreateReviewResponse { review_id, status }`。

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`。

### 5.3 GET `/api/v1/content-reviews/{id}`

Response：`ContentReviewDetailResponse`，包含正文、metadata、extension、报告摘要、版本摘要。

错误码：`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`。

### 5.4 POST `/api/v1/content-reviews/{id}/ai-report`

Header：`Idempotency-Key` 必填。

Request：`TriggerAIReportRequest { report_type, config }`。

Response：`TriggerAIReportResponse { report_id, job_id, workflow_run_id, status }`。

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`WORKFLOW_RUN_FAILED`、`LLM_PROVIDER_ERROR`、`INTERNAL_ERROR`。

### 5.5 POST `/api/v1/content-reviews/{id}/reject`

Request：`RejectReviewRequest { reason, regenerate_instruction, trigger_regeneration }`。

Response：`RejectReviewResponse { review_id, status, operation_log_id, regeneration_run_id, job_id }`。

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`WORKFLOW_RUN_FAILED`、`INTERNAL_ERROR`。

### 5.6 POST `/api/v1/content-reviews/{id}/approve-with-edit`

Request：`ApproveWithEditRequest { editable_fields, note }`。

Response：`ApproveWithEditResponse { review_id, status, content_version_id, operation_log_id }`。

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`INTERNAL_ERROR`。

## 6. Module Design

### 6.1 Backend `review` module

职责：
- 管理审稿记录生命周期。
- 管理内容版本历史。
- 管理 AI 质检报告查询与异步触发契约。
- 校验审稿状态流转并返回标准领域错误。

服务接口：
- `ListReviews(ctx, req) (PagedContentReviewsResponse, error)`
- `CreateReview(ctx, contentItemID, req, idempotencyKey) (CreateReviewResponse, error)`
- `GetReview(ctx, id) (ContentReviewDetailResponse, error)`
- `TriggerAIReport(ctx, id, req, workflowRunID, idempotencyKey) (TriggerAIReportResponse, error)`
- `GetAIReport(ctx, id) (ReviewReportResponse, error)`
- `ApproveReview(ctx, id, req) (ApproveReviewResponse, error)`
- `RejectReview(ctx, id, req, regenerationRunID) (RejectReviewResponse, error)`
- `ApproveWithEdit(ctx, id, req) (ApproveWithEditResponse, error)`
- `ListContentVersions(ctx, contentItemID, req) (PagedContentVersionsResponse, error)`

### 6.2 HTTP Handler

职责：
- 从 chi path/query/header/body 读取输入。
- 调用 Review Service。
- 将领域错误映射为统一 API 错误码。
- 对异步触发接口返回 202。

### 6.3 Frontend

- `lib/api.ts` 增加 review 类型与函数。
- `ProjectWorkspaceNav` 增加 `reviews` 项。
- 四个页面全部使用 `page-shell`、`page-hero`、`card`、`app-nav` 等现有样式模式，展示 loading / empty / error / success。

## 7. Output Contract

| 产物 | 输入 | 输出 | 正确性规则 | 类型 | 测试规范 |
|---|---|---|---|---|---|
| Review HTTP APIs | HTTP request + Bearer token | Envelope JSON | 状态码、schema、错误码、request_id 正确 | web-e2e | `standards/testing/web-e2e.md` |
| Review service | DTO + context | DTO / error | 状态流转、幂等、操作日志 ID、异步 ID 规则正确 | integration | `standards/testing/integration.md` |
| Review SQL tables | Migration SQL | PostgreSQL tables | DDL 可执行，约束和索引符合 Table Design | sql-query | `standards/testing/sql-query.md` |
| Web Admin pages | 用户导航和 API 响应 | 渲染页面 | 导航可达、刷新不 404、错误态展示 request_id | web-e2e | `standards/testing/web-e2e.md` |

`workflow.yaml` 当前 `project.features` 为 `[]`，但本迭代新增 HTTP endpoint 和 SQL migration，因此按功能本身显式引用 `web-e2e`、`integration`、`sql-query`。

跨组件链路：
- `HTTP Handler -> Review Service -> Workflow Engine/Queue Contract -> API Response`
- `Web Admin Page -> lib/api.ts -> API Server -> Review Service`

SQL Contract：本迭代不生成动态 SQL；迁移 DDL 固定，目标方言 PostgreSQL。禁止模式：无外键引用的孤立 review/version/report 表、无状态 CHECK 约束、列表查询缺少 project/status 索引。

## 8. Change Log

| 文件 | 变更类型 | 原因 |
|---|---|---|
| `apps/api-server/internal/modules/review/dto.go` | 新增 | 提供审稿 DTO、状态、错误码和响应模型骨架。 |
| `apps/api-server/internal/modules/review/service.go` | 新增 | 提供 Review Service 接口和可编译骨架实现。 |
| `apps/api-server/internal/http/handlers/review.go` | 新增 | 提供审稿 HTTP handler 骨架。 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册 Iteration 5 API 路由。 |
| `apps/api-server/migrations/00007_create_content_review_tables.sql` | 新增 | 新增审稿、版本、报告表结构。 |
| `openapi/openapi.yaml` | 修改 | 增加审稿 API 契约。 |
| `apps/web-admin/lib/api.ts` | 修改 | 增加审稿 API client 契约。 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 修改 | 增加审稿中心导航入口。 |
| `apps/web-admin/app/content-items/[itemId]/page.tsx` | 修改 | 修复 Next.js 15 动态路由 params 类型，解除前端构建阻塞。 |
| `apps/web-admin/app/generation-runs/[runId]/page.tsx` | 修改 | 修复 Next.js 15 动态路由 params 类型，解除前端构建阻塞。 |
| `apps/web-admin/app/projects/[projectId]/reviews/page.tsx` | 新增 | 实现审稿中心页面骨架。 |
| `apps/web-admin/app/content-reviews/[reviewId]/page.tsx` | 新增 | 实现审稿详情页面骨架。 |
| `apps/web-admin/app/content-reviews/[reviewId]/ai-report/page.tsx` | 新增 | 实现 AI 质检报告页面骨架。 |
| `apps/web-admin/app/content-reviews/[reviewId]/edit-approve/page.tsx` | 新增 | 实现编辑后通过页面骨架。 |
| `.cube/iterations/feature-5/skeleton-map.yaml` | 新增 | 映射骨架文件和开发任务。 |

## 9. Development Tasks

- Task-01：定义审稿领域 DTO 与状态常量
  - 所属模块：review
  - 简要描述：定义 content_review、content_version、review_report 对应请求响应、状态常量和领域错误。
  - 涉及接口/方法：review DTO types
  - 输入：审稿、版本、报告相关 JSON 字段
  - 输出：Go DTO、状态常量、错误变量
  - 产出类型：library
  - 功能类型：领域模型契约（type id: library）
  - 是否跨组件：否
- Task-02：设计并新增审稿数据库迁移
  - 所属模块：store
  - 简要描述：新增 content_review、content_version、review_report 表、约束和索引。
  - 涉及接口/方法：migration 00007
  - 输入：PostgreSQL migration
  - 输出：可执行 DDL
  - 产出类型：sql-query
  - 功能类型：固定 DDL（type id: sql-query）
  - 是否跨组件：否
- Task-03：实现 Review Service 骨架与状态流转接口
  - 所属模块：review
  - 简要描述：提供列表、创建、详情、质检触发、报告查询、通过、打回、编辑后通过、版本列表接口。
  - 涉及接口/方法：review.Service
  - 输入：context、请求 DTO、幂等键、异步运行 ID
  - 输出：响应 DTO 或领域错误
  - 产出类型：integration
  - 功能类型：服务状态机（type id: integration）
  - 是否跨组件：是（HTTP Handler -> Review Service -> operation_log/Workflow contract）
- Task-04：实现审稿 HTTP API 骨架与路由注册
  - 所属模块：http
  - 简要描述：新增 review handler 并注册所有 Iteration 5 API 路由。
  - 涉及接口/方法：ReviewHandler methods, NewRouter()
  - 输入：HTTP path/query/header/body
  - 输出：统一 envelope JSON
  - 产出类型：web-e2e
  - 功能类型：REST API（type id: web-e2e）
  - 是否跨组件：是（HTTP Router -> Handler -> Review Service）
- Task-05：补充 OpenAPI 审稿接口契约
  - 所属模块：openapi
  - 简要描述：为所有审稿 API 增加 path、request、response、错误码和 security。
  - 涉及接口/方法：openapi.yaml
  - 输入：API Design
  - 输出：OpenAPI 3.0 契约
  - 产出类型：web-e2e
  - 功能类型：API contract（type id: web-e2e）
  - 是否跨组件：否
- Task-06：扩展 Web Admin API client
  - 所属模块：web-admin
  - 简要描述：在 lib/api.ts 增加审稿类型和 API 调用函数。
  - 涉及接口/方法：fetchContentReviews(), createContentReview(), fetchContentReview(), triggerAIReport(), approveReview(), rejectReview(), approveWithEdit(), fetchContentVersions()
  - 输入：页面参数和表单数据
  - 输出：APIEnvelope 响应
  - 产出类型：web-e2e
  - 功能类型：前端 API client（type id: web-e2e）
  - 是否跨组件：是（Web Page -> API Client -> HTTP API）
- Task-07：实现项目审稿中心页面与导航入口
  - 所属模块：web-admin
  - 简要描述：新增 `/projects/:projectId/reviews`，支持列表、筛选、创建审稿入口和错误态。
  - 涉及接口/方法：ReviewsPage, ProjectWorkspaceNav
  - 输入：projectId、status、分页
  - 输出：审稿中心页面
  - 产出类型：web-e2e
  - 功能类型：Web UI（type id: web-e2e）
  - 是否跨组件：是（Navigation -> Page -> API Client）
- Task-08：实现审稿详情页面
  - 所属模块：web-admin
  - 简要描述：新增 `/content-reviews/:reviewId`，展示正文、报告摘要、版本入口和状态操作。
  - 涉及接口/方法：ContentReviewDetailPage
  - 输入：reviewId
  - 输出：审稿详情页面
  - 产出类型：web-e2e
  - 功能类型：Web UI（type id: web-e2e）
  - 是否跨组件：是（Page -> API Client -> HTTP API）
- Task-09：实现 AI 质检报告页面与异步触发入口
  - 所属模块：web-admin
  - 简要描述：新增 `/content-reviews/:reviewId/ai-report`，展示报告并支持触发异步生成。
  - 涉及接口/方法：AIReportPage
  - 输入：reviewId、report_type、config
  - 输出：报告页面、job_id/workflow_run_id 反馈
  - 产出类型：web-e2e
  - 功能类型：Web UI + async API（type id: web-e2e）
  - 是否跨组件：是（Page -> API Client -> HTTP API -> Workflow contract）
- Task-10：实现编辑后通过页面
  - 所属模块：web-admin
  - 简要描述：新增 `/content-reviews/:reviewId/edit-approve`，按 content type editable_fields 提交编辑并通过。
  - 涉及接口/方法：EditApprovePage
  - 输入：reviewId、editable_fields、note
  - 输出：content_version_id、operation_log_id、成功或错误态
  - 产出类型：web-e2e
  - 功能类型：Web UI（type id: web-e2e）
  - 是否跨组件：是（Page -> API Client -> HTTP API -> Review Service）
