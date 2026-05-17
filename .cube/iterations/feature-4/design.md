# Iteration 4 Design：内容单元生成闭环

## 1. 概述

本设计将 Iteration 4 的内容生成闭环落到现有 Go API Server、Workflow Engine、Novel Pack 和 Next.js 管理台结构中。整体方案采用最小增量：新增 `generation` 后端模块承载 `generation_run`、`content_item` 与 `novel_chapter_extension` 的业务契约，复用现有 `workflow.Service` 创建 WorkflowRun，复用现有 Engine 产生 StepRun、AgentTask、LLMCallLog，并在前端复用项目工作区导航、API envelope、错误态与页面样式。

核心约束：

- 所有生成触发必须关联 WorkflowRun，不允许绕过 Workflow Engine。
- `content_item.status` 状态集合固定为 `planned`、`generating`、`generated`、`generation_failed`、`pending_review`；生成成功最终进入 `pending_review`。
- 外部真实 LLM 调用不是 P0 阻塞验收；但 AgentTask / LLMCallLog 追踪链路必须成立。
- Core 资源命名保持内容类型无关，Novel 字段只进入 `novel_chapter_extension`。
- 前端 4 个页面均为 P0，必须接入 API、导航、状态反馈与 e2e。

## 2. Impact Analysis

| 模块 / 区域 | 影响程度 | 说明 |
|---|---|---|
| `apps/api-server/internal/modules/generation` | 新增 | 新增生成运行、内容单元、Novel 扩展 DTO、服务接口与内存骨架。 |
| `apps/api-server/internal/http/handlers/generation.go` | 新增 | 新增 Iteration 4 HTTP handler，遵循现有 envelope 和错误码映射。 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册生成运行与 ContentItem 路由，复用现有 `workflow.Service` 与 `engine.Submitter`。 |
| `apps/api-server/migrations/00006_create_content_generation_tables.sql` | 新增 | 新增 `generation_run`、`content_item`、`novel_chapter_extension` 表设计。 |
| `openapi/openapi.yaml` | 修改 | 增加本迭代 7 个 API path 与 schema。 |
| `apps/web-admin/lib/api.ts` | 修改 | 增加 GenerationRun、ContentItem 类型与 API client 函数。 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 修改 | 增加内容生产和 ContentItem 导航入口。 |
| `apps/web-admin/app/projects/[projectId]/production/page.tsx` | 新增 | 项目工作区内容生产页。 |
| `apps/web-admin/app/generation-runs/[runId]/page.tsx` | 新增 | 生成运行详情页。 |
| `apps/web-admin/app/generation-runs/[runId]/retry/page.tsx` | 新增 | 可刷新独立失败重试路由，复用重试交互。 |
| `apps/web-admin/app/projects/[projectId]/content-items/page.tsx` | 新增 | 项目 ContentItem 列表页。 |
| `apps/web-admin/app/content-items/[itemId]/page.tsx` | 新增 | ContentItem 详情页，支撑列表查看详情。 |
| `apps/web-admin/e2e/iteration4-content-generation-loop.spec.ts` | 新增 | 覆盖导航、主要按钮、成功态、失败态。 |
| `.cube/iterations/feature-4/skeleton-map.yaml` | 新增 | 声明 Development Tasks 到骨架文件的映射，支撑 02-design skeleton 覆盖检查。 |

兼容性分析：

- 现有 API 不改变请求或响应结构；只新增 path，向后兼容。
- 现有数据库迁移不修改；新增表通过独立迁移追加，兼容存量数据。
- 现有 WorkflowRun 状态为 `pending/running/success/failed/...`，generation 模块对外使用 `pending/running/succeeded/failed/retrying`，通过服务层映射，不改变 workflow 模块契约。
- 现有前端导航只增加入口，不移除已有入口。

## 3. Flow Design

### 3.1 手动生成流程

1. 用户在 `/projects/:projectId/production` 填写 confirmed_topic、worldview、arc、target_count、start_sequence_no、generation_config。
2. 前端携带 `Idempotency-Key` 调用 `POST /api/v1/projects/:projectId/generation-runs`。
3. GenerationHandler 校验幂等 Header 和请求体格式。
4. GenerationHandler 调用 GenerationService 的预检查能力，校验项目、规划资产、可用 workflow template version、可用 Provider 抽象与幂等 payload。
5. GenerationHandler 调用 `workflow.CreateRun()` 创建 WorkflowRun，Input 包含生成配置与规划资产引用。
6. GenerationHandler 将 `workflow_run_id` 传入 GenerationService；GenerationService 创建 GenerationRun，状态为 `pending`，预创建或关联 planned/generating ContentItem 草稿。
7. Handler 调用 `engine.Submit(workflow_run_id)` 异步执行；HTTP 立即返回 `generation_run_id`、`workflow_run_id`、`status`。
8. Engine 处理 WorkflowRun，产生 StepRun、AgentTask、LLMCallLog。
9. GenerationService 根据 WorkflowRun 结果进行状态对账与持久化，更新 GenerationRun 和 ContentItem；成功时 ContentItem 最终状态为 `pending_review`。

职责边界：WorkflowRun 只能由 Handler 通过 `workflow.Service` 创建；GenerationService 不创建 WorkflowRun，只负责预检查、业务记录、幂等响应、状态对账、ContentItem 持久化和 operation_log 关联。若 WorkflowRun 创建成功但 GenerationRun 创建失败，Handler 必须返回 `WORKFLOW_RUN_FAILED` 并记录 operation_log，避免静默孤儿运行。

异常流程：

- 幂等键缺失或过长：`VALIDATION_ERROR`。
- 同一幂等键请求体不同：`IDEMPOTENCY_CONFLICT`。
- 规划资产缺失：`CONFLICT`。
- WorkflowRun 创建失败：`WORKFLOW_RUN_FAILED`。
- Provider / stub 失败：`LLM_PROVIDER_ERROR`，运行进入失败状态。

### 3.2 批量生成流程

批量生成入口与手动生成一致，但输入为 `range`、`batch_size`、`generation_config`。Handler 按批次为可受理单元创建 WorkflowRun，并将 `workflow_run_id` 列表交给 GenerationService 创建业务记录；GenerationService 返回已受理 ID 列表、`accepted_count`、`rejected_count` 和每个拒绝单元的 `sequence_no`、`code`、`message`。HTTP 不等待最终正文生成；若全部单元均不可受理则返回 `CONFLICT` 或 `VALIDATION_ERROR`，部分可受理则返回 `202 Accepted` 并携带 rejected 明细。

### 3.3 失败重试流程

1. 用户从详情页弹窗 / 抽屉或 `/generation-runs/:runId/retry` 输入 reason 和 input_override。
2. 前端调用 `POST /api/v1/generation-runs/:id/retry`，携带 `Idempotency-Key`。
3. GenerationService 校验原 GenerationRun 存在且状态为 `failed`，并完成幂等 payload 校验。
4. GenerationHandler 调用 `workflow.CreateRun()` 新建 WorkflowRun。
5. GenerationHandler 将 `workflow_run_id` 传入 GenerationService；GenerationService 新建 GenerationRun，设置 `retry_of_generation_run_id` 指向原失败运行，不覆盖原记录。
6. Handler 提交新 WorkflowRun，返回 `new_generation_run_id`、`workflow_run_id`、`operation_log_id`。

### 3.4 查询与页面流程

- 内容生产页加载 generation run 列表，并支持状态筛选、分页、排序。
- 生成运行详情页调用详情接口，不依赖列表缓存，展示 WorkflowRun、StepRun 摘要、AgentTask、LLMCallLog、输出 ContentItem、错误信息。
- ContentItem 列表页调用列表接口并支持状态筛选、分页、排序。
- ContentItem 详情页展示正文、扩展字段、版本和来源 GenerationRun。

## 4. Table Design

### 4.1 `generation_run`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | TEXT | PK | 生成运行 ID。 |
| `project_id` | TEXT | NOT NULL | 项目 ID。 |
| `workflow_run_id` | TEXT | NOT NULL UNIQUE | 关联 WorkflowRun。 |
| `template_version_id` | TEXT | NOT NULL | 使用的工作流模板版本。 |
| `status` | TEXT | NOT NULL CHECK | `pending/running/succeeded/failed/retrying`。 |
| `trigger_type` | TEXT | NOT NULL CHECK | `manual/batch/retry`。 |
| `confirmed_topic_id` | TEXT | NULL | 已确认选题。 |
| `worldview_version_id` | TEXT | NULL | 世界观版本。 |
| `arc_id` | TEXT | NULL | 大纲弧线。 |
| `target_count` | INTEGER | NOT NULL CHECK > 0 | 目标生成数量。 |
| `start_sequence_no` | INTEGER | NOT NULL CHECK > 0 | 起始序号。 |
| `generation_config` | JSONB | NOT NULL DEFAULT `{}` | 生成配置。 |
| `retry_of_generation_run_id` | TEXT | NULL FK | 原失败运行。 |
| `idempotency_key` | TEXT | NULL | 幂等键。 |
| `error_code` | TEXT | NULL | 失败错误码。 |
| `error_message` | TEXT | NULL | 失败信息。 |
| `created_at` | TIMESTAMPTZ | NOT NULL | 创建时间。 |
| `updated_at` | TIMESTAMPTZ | NOT NULL | 更新时间。 |

索引：

- `idx_generation_run_project_id(project_id)`
- `idx_generation_run_workflow_run_id(workflow_run_id)`
- `idx_generation_run_status(status)`
- `idx_generation_run_retry_of(retry_of_generation_run_id)`
- `idx_generation_run_project_idempotency_key(project_id, idempotency_key) WHERE idempotency_key IS NOT NULL`

### 4.2 `content_item`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | TEXT | PK | 内容单元 ID。 |
| `project_id` | TEXT | NOT NULL | 项目 ID。 |
| `generation_run_id` | TEXT | NOT NULL FK | 来源生成运行。 |
| `content_type_code` | TEXT | NOT NULL | 内容类型，例如 `novel`。 |
| `status` | TEXT | NOT NULL CHECK | `planned/generating/generated/generation_failed/pending_review`。 |
| `sequence_no` | INTEGER | NOT NULL CHECK > 0 | 项目内序号。 |
| `title` | TEXT | NOT NULL | 内容标题。 |
| `body` | TEXT | NOT NULL DEFAULT '' | 正文。 |
| `version` | INTEGER | NOT NULL DEFAULT 1 | 内容版本。 |
| `metadata` | JSONB | NOT NULL DEFAULT `{}` | 通用元数据。 |
| `error_code` | TEXT | NULL | 生成失败错误码。 |
| `error_message` | TEXT | NULL | 生成失败信息。 |
| `created_at` | TIMESTAMPTZ | NOT NULL | 创建时间。 |
| `updated_at` | TIMESTAMPTZ | NOT NULL | 更新时间。 |

索引：

- `idx_content_item_project_id(project_id)`
- `idx_content_item_generation_run_id(generation_run_id)`
- `idx_content_item_status(status)`
- `idx_content_item_project_sequence(project_id, sequence_no)`
- `uq_content_item_project_type_sequence_version(project_id, content_type_code, sequence_no, version)`

重试与版本语义：失败重试不覆盖原 ContentItem；同一项目、内容类型、序号允许通过递增 `version` 保留多次生成记录，最新成功版本进入 `pending_review`，失败版本保留 `generation_failed` 与错误字段。

### 4.3 `novel_chapter_extension`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `content_item_id` | TEXT | PK FK | 对应 ContentItem。 |
| `project_id` | TEXT | NOT NULL | 项目 ID。 |
| `confirmed_topic_id` | TEXT | NOT NULL | 已确认选题。 |
| `worldview_version_id` | TEXT | NOT NULL | 世界观版本。 |
| `arc_id` | TEXT | NOT NULL | 大纲弧线。 |
| `chapter_no` | INTEGER | NOT NULL CHECK > 0 | Novel Pack 章节序号。 |
| `script` | JSONB | NOT NULL DEFAULT `{}` | 脚本/结构化草稿。 |
| `created_at` | TIMESTAMPTZ | NOT NULL | 创建时间。 |
| `updated_at` | TIMESTAMPTZ | NOT NULL | 更新时间。 |

索引：

- `idx_novel_chapter_extension_project_id(project_id)`
- `idx_novel_chapter_extension_arc_id(arc_id)`

## 5. API Design

所有响应遵循 `.cube/config/api-spec.md` 的 `success/data/error/request_id` envelope。列表响应使用统一 `items/pagination` 结构。

### 5.1 `POST /api/v1/projects/:projectId/generation-runs`

请求 Header：`Idempotency-Key` 必填。

请求体：

```json
{
  "confirmed_topic_id": "topic-1",
  "worldview_version_id": "worldview-v1",
  "arc_id": "arc-1",
  "target_count": 1,
  "start_sequence_no": 1,
  "template_version_id": "wftv-generation",
  "generation_config": {}
}
```

成功响应：`202 Accepted`

```json
{
  "generation_run_id": "genrun-1",
  "workflow_run_id": "wfr-1",
  "status": "pending"
}
```

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`WORKFLOW_RUN_FAILED`、`LLM_PROVIDER_ERROR`、`AGENT_OUTPUT_INVALID`、`INTERNAL_ERROR`。

### 5.2 `POST /api/v1/projects/:projectId/generation-runs/batch`

请求 Header：`Idempotency-Key` 必填。

请求体：

```json
{
  "range": { "start_sequence_no": 1, "end_sequence_no": 5 },
  "batch_size": 5,
  "template_version_id": "wftv-generation",
  "generation_config": {}
}
```

成功响应：`202 Accepted`

```json
{
  "generation_run_ids": ["genrun-1"],
  "workflow_run_ids": ["wfr-1"],
  "accepted_count": 1,
  "rejected_count": 1,
  "rejected_items": [
    { "sequence_no": 2, "code": "CONFLICT", "message": "planning prerequisite is missing" }
  ]
}
```

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`WORKFLOW_RUN_FAILED`、`AGENT_OUTPUT_INVALID`、`INTERNAL_ERROR`。

### 5.3 `GET /api/v1/projects/:projectId/generation-runs`

Query：`status`、`page`、`page_size`、`sort`、`order`。

成功响应：`200 OK`，data 为 `items[]` + `pagination`。

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`。

### 5.4 `GET /api/v1/generation-runs/:id`

成功响应包含 GenerationRun 基础字段、`workflow_run_id`、`step_runs[]`、`agent_tasks[]`、`llm_call_logs[]`、`content_items[]`、`error`。

错误码：`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`。

### 5.5 `POST /api/v1/generation-runs/:id/retry`

请求 Header：`Idempotency-Key` 必填。

请求体：

```json
{
  "reason": "修复失败后重试",
  "input_override": {}
}
```

成功响应：`202 Accepted`

```json
{
  "new_generation_run_id": "genrun-2",
  "workflow_run_id": "wfr-2",
  "operation_log_id": "oplog-1"
}
```

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`WORKFLOW_RUN_FAILED`、`AGENT_OUTPUT_INVALID`、`INTERNAL_ERROR`。

### 5.6 `GET /api/v1/projects/:projectId/content-items`

Query：`status`、`page`、`page_size`、`sort`、`order`。

成功响应：ContentItem 列表和分页信息。

错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`。

### 5.7 `GET /api/v1/content-items/:id`

成功响应包含正文、扩展字段、版本、来源 `generation_run_id`。

错误码：`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`。

## 6. Module Design

### 6.1 `generation` 模块

职责：

- 定义 GenerationRun、ContentItem、NovelChapterExtension 的 DTO 和状态常量。
- 校验前置规划资产，确保 confirmed_topic、worldview、characters、arc / outline 已可用。
- 提供手动生成、批量生成、列表、详情、失败重试、ContentItem 查询接口。
- 管理幂等缓存；同 key 同 payload 返回缓存结果，不同 payload 返回 `ErrIdempotencyConflict`。
- 维护状态流转和 operation_log ID。

服务接口：

```go
type Service interface {
    ValidateGenerationRun(ctx context.Context, projectID string, req CreateGenerationRunRequest, idempotencyKey string) error
    CreateGenerationRun(ctx context.Context, projectID string, req CreateGenerationRunRequest, workflowRunID string, idempotencyKey string) (CreateGenerationRunResponse, error)
    CreateBatchGenerationRuns(ctx context.Context, projectID string, req CreateBatchGenerationRunsRequest, workflowRunIDs []string, idempotencyKey string) (CreateBatchGenerationRunsResponse, error)
    ListGenerationRuns(ctx context.Context, projectID string, req ListGenerationRunsRequest) (PagedGenerationRunsResponse, error)
    GetGenerationRun(ctx context.Context, id string) (GenerationRunDetailResponse, error)
    ValidateRetryGenerationRun(ctx context.Context, id string, req RetryGenerationRunRequest, idempotencyKey string) error
    RetryGenerationRun(ctx context.Context, id string, req RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (RetryGenerationRunResponse, error)
    ReconcileWorkflowResult(ctx context.Context, workflowRunID string) error
    ListContentItems(ctx context.Context, projectID string, req ListContentItemsRequest) (PagedContentItemsResponse, error)
    GetContentItem(ctx context.Context, id string) (ContentItemDetailResponse, error)
}
```

### 6.2 HTTP Handler

`GenerationHandler` 负责：

- 解析 path/query/body/header。
- 调用 `generation.Service` 执行预检查和幂等 payload 校验。
- 调用 `workflow.Service` 创建或重试 WorkflowRun。
- 调用 `generation.Service` 创建业务运行记录。
- 调用 `engine.Submitter` 异步提交 workflow run。
- 将模块错误映射到统一 API 错误码。

### 6.3 Frontend

- `apps/web-admin/lib/api.ts` 新增类型和 API 函数。
- 项目工作区导航新增 `内容生产` 和 `内容单元`。
- 页面复用现有 `page-shell`、`page-hero`、`card`、`badge`、`dialog-panel` 样式。
- 每页实现加载态、空态、错误态、成功 Toast / status。

## 7. Output Contract

项目 `workflow.yaml` 的 `project.features` 当前为空；本迭代仍因实际交付 API、SQL migration、异步生成、前端 e2e 和跨组件链路而声明以下测试规范：

| 产物 | 类型 | 正确性规则 | 测试规范 |
|---|---|---|---|
| Generation API | `web-e2e` | HTTP status、envelope、业务字段、错误码、Header 幂等一致 | `standards/testing/web-e2e.md` |
| Migration SQL | `sql-query` | PostgreSQL DDL 字段、约束、索引与 Core 命名边界正确 | `standards/testing/sql-query.md` |
| Async generation | `batch-job` | 触发后立即返回，后台 workflow 可重试且幂等 | `standards/testing/batch-job.md` |
| Feature chain | `integration` | Handler -> GenerationService -> WorkflowService -> Engine -> Agent/LLM -> API response | `standards/testing/integration.md` |
| Frontend pages | `web-e2e` | 导航进入、按钮点击、成功态、失败态、request_id 展示 | `standards/testing/web-e2e.md` |

SQL Contract：

- 目标方言：PostgreSQL。
- 本迭代不生成动态 SQL；迁移 DDL 是固定 SQL 产物。
- 必须包含 `generation_run`、`content_item`、`novel_chapter_extension` 三张表。
- 必须包含状态 CHECK 约束、幂等唯一索引和项目内内容序号/版本唯一约束。
- 迁移文件必须遵循仓库现有 migration 约定，包含可执行 Up DDL；如仓库引入 Down/rollback 约定，本迁移需补齐对应 rollback 段。
- 禁止在 Core 表/API 字段中使用 `book`、`chapter` 作为核心资源名；`chapter_no` 只允许出现在 `novel_chapter_extension`。
- 典型验证：读取迁移文件，断言表名、状态枚举、索引名、FK 关系、`pending_review` 状态存在。

## 8. Change Log

| 文件 | 类型 | 原因 |
|---|---|---|
| `apps/api-server/internal/modules/generation/dto.go` | 新增 | 定义生成运行与内容单元 DTO、状态和请求响应。 |
| `apps/api-server/internal/modules/generation/service.go` | 新增 | 定义服务接口、错误、幂等与状态骨架。 |
| `apps/api-server/internal/http/handlers/generation.go` | 新增 | 暴露 Iteration 4 API handler。 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册 generation routes。 |
| `apps/api-server/migrations/00006_create_content_generation_tables.sql` | 新增 | 新增本迭代表结构。 |
| `openapi/openapi.yaml` | 修改 | 增加 API path 与 schema。 |
| `apps/web-admin/lib/api.ts` | 修改 | 增加前端 API client。 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 修改 | 增加内容生产与内容单元导航入口。 |
| `apps/web-admin/app/projects/[projectId]/production/page.tsx` | 新增 | 内容生产页。 |
| `apps/web-admin/app/generation-runs/[runId]/page.tsx` | 新增 | 生成运行详情页。 |
| `apps/web-admin/app/generation-runs/[runId]/retry/page.tsx` | 新增 | 失败重试独立路由页。 |
| `apps/web-admin/app/projects/[projectId]/content-items/page.tsx` | 新增 | ContentItem 列表页。 |
| `apps/web-admin/app/content-items/[itemId]/page.tsx` | 新增 | ContentItem 详情页。 |
| `apps/web-admin/e2e/iteration4-content-generation-loop.spec.ts` | 新增 | e2e 覆盖。 |
| `.cube/iterations/feature-4/skeleton-map.yaml` | 新增 | 记录 Development Tasks 与骨架文件映射，保证设计阶段检查可追踪。 |

## 9. Development Tasks

- Task-01：定义生成运行与内容单元 DTO 契约
  - 所属模块：generation
  - 简要描述：定义 GenerationRun、ContentItem、NovelChapterExtension、请求响应 DTO、状态常量和错误。
  - 涉及接口/方法：generation DTO structs
  - 输入：PRD 中 API 请求与状态集合
  - 输出：Go DTO、状态常量、错误变量
  - 产出类型：library
  - 功能类型：后端公共 DTO 契约（type id: library）
  - 是否跨组件：否
- Task-02：实现生成服务接口与状态规则骨架
  - 所属模块：generation
  - 简要描述：实现服务接口、幂等、前置依赖校验、生成运行状态、ContentItem 状态和重试关联。
  - 涉及接口/方法：CreateGenerationRun(), CreateBatchGenerationRuns(), RetryGenerationRun()
  - 输入：projectID、生成请求、workflowRunID、idempotencyKey
  - 输出：GenerationRun 响应、ContentItem 响应、业务错误
  - 产出类型：integration
  - 功能类型：跨组件业务服务（type id: integration）
  - 是否跨组件：是（组件链路：GenerationHandler -> GenerationService -> WorkflowService -> Engine）
- Task-03：暴露生成运行和 ContentItem HTTP API
  - 所属模块：http/handlers
  - 简要描述：新增 GenerationHandler 并在 router 注册 7 个 API。
  - 涉及接口/方法：NewGenerationHandler(), CreateGenerationRun(), CreateBatchGenerationRuns(), ListGenerationRuns(), GetGenerationRun(), RetryGenerationRun(), ListContentItems(), GetContentItem()
  - 输入：HTTP path/query/header/body
  - 输出：统一 API envelope、状态码、错误码
  - 产出类型：web-e2e
  - 功能类型：HTTP API（type id: web-e2e）
  - 是否跨组件：是（组件链路：Router -> Handler -> GenerationService -> WorkflowService）
- Task-04：补充数据库迁移和 OpenAPI 契约
  - 所属模块：migrations/openapi
  - 简要描述：新增 PostgreSQL DDL 并补充 OpenAPI path/schema。
  - 涉及接口/方法：00006 migration, openapi paths
  - 输入：Table Design 与 API Design
  - 输出：迁移 SQL、OpenAPI schema
  - 产出类型：sql-query
  - 功能类型：固定 SQL DDL 与 API 契约（type id: sql-query）
  - 是否跨组件：否
- Task-05：补充前端 API client 契约
  - 所属模块：web-admin/lib
  - 简要描述：在 API client 中定义 GenerationRun、ContentItem 类型与请求函数。
  - 涉及接口/方法：createGenerationRun(), createBatchGenerationRuns(), fetchGenerationRuns(), fetchGenerationRun(), retryGenerationRun(), fetchContentItems(), fetchContentItem()
  - 输入：页面参数和表单数据
  - 输出：APIEnvelope 类型响应
  - 产出类型：library
  - 功能类型：前端 API client（type id: library）
  - 是否跨组件：否
- Task-06：实现内容生产页与项目导航入口
  - 所属模块：web-admin/app
  - 简要描述：新增 `/projects/:projectId/production` 页面并在项目工作区导航接入。
  - 涉及接口/方法：ProductionPage
  - 输入：projectId、生成表单、筛选分页
  - 输出：生成运行列表、手动生成、批量生成、Toast、错误态
  - 产出类型：web-e2e
  - 功能类型：前端页面交互（type id: web-e2e）
  - 是否跨组件：是（组件链路：Next Page -> API Client -> Generation API）
- Task-07：实现生成运行详情和失败重试路由
  - 所属模块：web-admin/app
  - 简要描述：新增 `/generation-runs/:runId` 和 `/generation-runs/:runId/retry`，支持详情展示与重试。
  - 涉及接口/方法：GenerationRunDetailPage, GenerationRetryPage
  - 输入：runId、reason、input_override
  - 输出：追踪详情、新运行 ID、operation_log_id、错误态
  - 产出类型：web-e2e
  - 功能类型：前端页面交互（type id: web-e2e）
  - 是否跨组件：是（组件链路：Next Page -> API Client -> Generation API -> WorkflowService）
- Task-08：实现 ContentItem 列表和详情页面
  - 所属模块：web-admin/app
  - 简要描述：新增 `/projects/:projectId/content-items` 和 `/content-items/:itemId` 页面。
  - 涉及接口/方法：ContentItemsPage, ContentItemDetailPage
  - 输入：projectId、itemId、筛选分页
  - 输出：内容单元列表、详情、正文、扩展字段、错误态
  - 产出类型：web-e2e
  - 功能类型：前端页面交互（type id: web-e2e）
  - 是否跨组件：是（组件链路：Next Page -> API Client -> ContentItem API）
- Task-09：覆盖 Iteration 4 后端契约与异步联调路径
  - 所属模块：api-server
  - 简要描述：覆盖 DTO 校验、幂等冲突、规划资产缺失、异步 202 受理、状态对账、operation_log、统一错误码和 OpenAPI 契约一致性。
  - 涉及接口/方法：GenerationHandler, GenerationService, WorkflowService, Engine Submitter, ReconcileWorkflowResult()
  - 输入：HTTP 请求、幂等 Header、WorkflowRun 结果、AgentTask/LLMCallLog 输出
  - 输出：集成测试证据、API 契约验证、状态持久化断言
  - 产出类型：integration
  - 功能类型：后端联调验收（type id: integration）
  - 是否跨组件：是（组件链路：HTTP -> GenerationService -> WorkflowService -> Engine -> Agent/LLM -> GenerationService reconciliation -> GET APIs）
- Task-10：覆盖 Iteration 4 前端 e2e 与页面联调路径
  - 所属模块：web-admin/e2e
  - 简要描述：覆盖导航进入、主要按钮、成功渲染、失败态 request_id、列表到详情、重试独立路由刷新。
  - 涉及接口/方法：Playwright iteration4 spec
  - 输入：浏览器导航、按钮点击、mock/API 响应
  - 输出：e2e 测试证据
  - 产出类型：web-e2e
  - 功能类型：端到端验收（type id: web-e2e）
  - 是否跨组件：是（组件链路：Browser -> Next Page -> API Client -> Go API）
- Task-11：维护设计骨架映射
  - 所属模块：cube
  - 简要描述：维护 `.cube/iterations/feature-4/skeleton-map.yaml`，确保 Development Tasks 100% 映射到非空骨架文件。
  - 涉及接口/方法：skeleton-map.yaml
  - 输入：Development Tasks 与模块 source_dir
  - 输出：skeleton-map 覆盖检查通过
  - 产出类型：library
  - 功能类型：流程契约（type id: library）
  - 是否跨组件：否
