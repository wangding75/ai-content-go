# Iteration 3 Design：Novel Pack 新书规划流程

## 1. 概述

本次设计在现有 Go API Server、in-memory Service、Chi Router、Workflow Engine、AgentTask、LLMCallLog 和 Next.js Web Admin 基础上，新增 Novel Pack 新书规划能力。整体采用最小改动方案：Core 继续保持内容类型无关，Novel 专属能力集中在 `novel` 模块、`/projects/{projectId}/novel/*` API、前端项目工作区页面和迁移契约中。

核心思路：

- 后端新增 `apps/api-server/internal/modules/novel`，承载 planning_run、planning_snapshot、topic candidate、worldview、character、arc 的 DTO 与 in-memory Service。
- HTTP 层新增 `NovelHandler`，路由挂在 `/api/v1/projects/{projectId}/novel/*`。
- 启动规划时由 Novel Service 校验项目与 Novel Pack 边界，并通过既有 `workflow.Service.CreateRun` 创建 `WorkflowRun`，再由 `engine.Submitter` 异步提交。
- 规划详情聚合 Novel Pack 业务记录与既有 WorkflowRun / StepRun / AgentTask / LLMCallLog 可追踪摘要。
- 前端新增项目工作区规划、候选选题确认、世界观、人物、大纲页面，复用 `lib/api.ts`、统一 envelope 错误结构、全局导航和现有 CSS 管理台样式。
- 数据库通过新增 PostgreSQL migration 固化表结构契约；当前开发仍沿用现有 in-memory 实现。

关键约束：

- Core 层不新增 `novel_*`、Book、Chapter 作为核心资源名。
- 所有接口使用 `/api/v1`、Bearer Auth、统一 `success/data/error/request_id` envelope。
- 启动规划和确认选题强制支持 `Idempotency-Key`。
- 页面验收以 `docs/requirements/ai-content-factory-clickable-prototype.html` 为视觉和交互基准。

## 2. Impact Analysis

| 模块/文件 | 影响程度 | 影响说明 |
|---|---|---|
| `apps/api-server/internal/modules/novel/dto.go` | 新增 | Novel Pack 规划、候选选题、世界观、人物、大纲 DTO。 |
| `apps/api-server/internal/modules/novel/service.go` | 新增 | Novel Pack in-memory Service、状态流转、幂等占位、业务查询接口。 |
| `apps/api-server/internal/http/handlers/novel.go` | 新增 | 暴露本迭代 Novel Pack HTTP API，统一错误映射。 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册 `/projects/{id}/novel/*` 路由，并注入 workflow service 与 engine submitter。 |
| `apps/api-server/migrations/00005_create_novel_planning_tables.sql` | 新增 | PostgreSQL DDL 契约，定义 Novel Pack 扩展表和索引。 |
| `openapi/openapi.yaml` | 修改 | 增加本迭代 API paths、schemas、examples、错误响应引用。 |
| `apps/web-admin/lib/api.ts` | 修改 | 新增 Novel Planning 类型与 API client 函数。 |
| `apps/web-admin/app/page.tsx` | 修改 | 在项目详情/工作区视图中增加当前项目的内容规划入口，不硬编码项目 ID。 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 新增 | 项目工作区局部导航，动态使用当前 `projectId` 并支持规划、世界观、人物、大纲高亮。 |
| `apps/web-admin/app/global-nav.tsx` | 修改 | 仅保留系统级入口，不硬编码 Novel 项目路由。 |
| `apps/web-admin/app/projects/[projectId]/planning/page.tsx` | 新增 | 内容规划页，包含启动规划、运行列表、候选选题弹窗。 |
| `apps/web-admin/app/projects/[projectId]/planning/topics/page.tsx` | 新增 | 候选选题确认直达/刷新恢复入口。 |
| `apps/web-admin/app/projects/[projectId]/novel/worldview/page.tsx` | 新增 | 世界观查看与编辑页。 |
| `apps/web-admin/app/projects/[projectId]/novel/characters/page.tsx` | 新增 | 人物管理页。 |
| `apps/web-admin/app/projects/[projectId]/novel/arcs/page.tsx` | 新增 | 大纲管理页。 |
| `.cube/iterations/feature-3/skeleton-map.yaml` | 新增 | 记录骨架文件与 Development Tasks 对应关系。 |

接口兼容性：现有 API 路由不删除、不改语义；只新增 Novel Pack API。`NewRouter(systemService, logger)` 签名保持不变，新增服务仍在 router 内部构造。

数据兼容性：新增 migration 只创建新表和索引，不修改已有表；当前实现为 in-memory，无存量数据迁移风险。

## 3. Flow Design

### 3.1 启动新书规划

1. 用户从 `/projects/{projectId}/planning` 填写 `genre`、`audience`、`count`、`template_version_id`、`input_override`。
2. 前端调用 `POST /api/v1/projects/{projectId}/novel/planning-runs`，携带 `Idempotency-Key`。
3. `NovelHandler.CreatePlanningRun` 解码请求并读取 path `projectId` 与幂等键。
4. `novel.Service.CreatePlanningRun` 校验：项目 ID 非空、模板版本 ID 非空、count > 0、Novel Pack 内容类型边界、幂等键重复语义。
5. Handler 通过既有 `workflow.Service.CreateRun` 创建 WorkflowRun，input 中包含 Novel 规划输入与 `content_pack=novel`。
6. Novel Service 记录 `planning_run`，关联 `workflow_run_id`，初始化 `planning_snapshot` 与候选选题占位数据。
7. Handler 调用 `engine.Submitter.Submit(workflow_run_id)`，立即返回 202 与 `{planning_run_id, workflow_run_id, status}`。
8. 后台 Workflow Engine 继续负责 StepRun、AgentTask、LLMCallLog 追踪。

异常流程：

- 请求体非法或缺少幂等键：`VALIDATION_ERROR`。
- 项目不存在或模板版本不存在：`NOT_FOUND`。
- 项目不是 Novel Pack：`CONFLICT`。
- 相同幂等键请求体不一致：`IDEMPOTENCY_CONFLICT`。
- WorkflowRun 创建失败：`WORKFLOW_RUN_FAILED`。

### 3.2 查看规划运行

1. 内容规划页调用 `GET /api/v1/projects/{projectId}/novel/planning-runs` 查询分页历史。
2. 详情调用 `GET /api/v1/projects/{projectId}/novel/planning-runs/{runId}`。
3. Novel Service 返回业务运行、候选选题、StepRun 摘要、AgentTask 摘要、LLMCallLog 摘要。
4. 页面渲染运行状态、失败原因、候选选题和可继续操作入口。

异常流程：运行不存在返回 `NOT_FOUND`；运行不属于项目返回 `FORBIDDEN`；分页参数非法返回 `VALIDATION_ERROR`。

### 3.3 候选选题确认

1. 用户在内容规划页打开候选选题确认弹窗；`/projects/{projectId}/planning/topics` 作为直达入口加载同一确认体验。
2. 用户输入 `note` 并提交。
3. 前端调用 `POST /api/v1/projects/{projectId}/novel/topics/{topicId}/confirm`，携带 `Idempotency-Key`。
4. Novel Service 校验候选属于项目且 status 为 `candidate`。
5. 成功后状态变为 `confirmed`，生成 `confirmed_topic_id` 与 `operation_log_id`，返回前后状态。

异常流程：候选不存在返回 `NOT_FOUND`；候选已确认或状态非法返回 `CONFLICT`；幂等键冲突返回 `IDEMPOTENCY_CONFLICT`。

### 3.4 世界观、人物、大纲维护

- 世界观：GET 查询当前版本；PATCH 保存新版本并返回 `version_id`、`operation_log_id`。
- 人物：GET 分页查询；POST 新增人物并返回 `character_id`、`operation_log_id`。
- 大纲：GET 分页查询弧线大纲，结果关联 planning_run / planning_snapshot。

异常流程统一映射为 `VALIDATION_ERROR`、`NOT_FOUND`、`FORBIDDEN`、`CONFLICT`、`INTERNAL_ERROR`。

### 3.5 前端页面链路

```
GlobalNav / 项目工作区入口
  -> ProjectPlanningPage
     -> lib/api.ts Novel Planning client
        -> API Server NovelHandler
           -> NovelService + WorkflowService + EngineSubmitter
```

页面必须展示 loading、empty、error、success，并在 error 中显示 `error.code`、`error.message`、`request_id`。

## 4. Table Design

目标方言：PostgreSQL。当前开发实现为 in-memory，DDL 用于后续持久化契约和测试。

### content_asset

本迭代不修改 Core 表定义，只要求 Novel 规划产物可通过 `content_asset` 或 Novel 扩展表关联 `content_project`。

### planning_run

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | text | primary key | 业务规划运行 ID |
| project_id | text | not null | ContentProject ID |
| workflow_run_id | text | not null unique | 关联 WorkflowRun |
| template_version_id | text | not null | 已发布 WorkflowTemplateVersion |
| status | text | not null | pending / running / success / failed / cancelled |
| genre | text | not null | 类型 |
| audience | text | not null | 目标读者 |
| candidate_count | integer | not null check > 0 | 候选数量 |
| input_override | jsonb | not null default '{}' | 输入覆盖 |
| idempotency_key | text | null | 启动规划幂等键 |
| created_at | timestamptz | not null | 创建时间 |
| updated_at | timestamptz | not null | 更新时间 |

索引：`idx_planning_run_project_id`、`idx_planning_run_workflow_run_id`、`idx_planning_run_status`、`idx_planning_run_idempotency_key`。

### planning_snapshot

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | text | primary key | 快照 ID |
| planning_run_id | text | not null | 规划运行 ID |
| project_id | text | not null | 项目 ID |
| snapshot_type | text | not null | topics / worldview / characters / arcs |
| payload | jsonb | not null default '{}' | 快照内容 |
| created_at | timestamptz | not null | 创建时间 |

索引：`idx_planning_snapshot_run_id`、`idx_planning_snapshot_project_id`。

### novel_topic_candidate

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | text | primary key | candidate_id |
| project_id | text | not null | 项目 ID |
| planning_run_id | text | not null | 规划运行 ID |
| snapshot_id | text | not null | 快照 ID |
| title | text | not null | 选题标题 |
| logline | text | not null | 一句话卖点 |
| status | text | not null | candidate / confirmed / rejected |
| score | numeric(5,2) | not null | 评分 |
| reason | text | not null | 推荐理由 |
| confirmed_topic_id | text | null | 确认后主题 ID |
| created_at | timestamptz | not null | 创建时间 |
| updated_at | timestamptz | not null | 更新时间 |

索引：`idx_novel_topic_candidate_project_id`、`idx_novel_topic_candidate_run_id`、`idx_novel_topic_candidate_status`。

### novel_worldview

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | text | primary key | 世界观版本 ID |
| project_id | text | not null | 项目 ID |
| planning_run_id | text | null | 来源规划运行 |
| snapshot_id | text | null | 来源快照 |
| worldview | jsonb | not null default '{}' | 世界观内容 |
| forbidden_rules | jsonb | not null default '[]' | 禁止项 |
| version | integer | not null | 版本号 |
| created_at | timestamptz | not null | 创建时间 |
| updated_at | timestamptz | not null | 更新时间 |

索引：`idx_novel_worldview_project_id`。

### novel_character

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | text | primary key | character_id |
| project_id | text | not null | 项目 ID |
| planning_run_id | text | null | 来源规划运行 |
| snapshot_id | text | null | 来源快照 |
| name | text | not null | 人物名 |
| role | text | not null | 角色类型 |
| profile | jsonb | not null default '{}' | 人物设定 |
| created_at | timestamptz | not null | 创建时间 |
| updated_at | timestamptz | not null | 更新时间 |

索引：`idx_novel_character_project_id`、`idx_novel_character_role`。

### novel_arc

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | text | primary key | arc_id |
| project_id | text | not null | 项目 ID |
| planning_run_id | text | null | 来源规划运行 |
| snapshot_id | text | null | 来源快照 |
| title | text | not null | 大纲标题 |
| summary | text | not null | 摘要 |
| order_index | integer | not null | 排序 |
| created_at | timestamptz | not null | 创建时间 |
| updated_at | timestamptz | not null | 更新时间 |

索引：`idx_novel_arc_project_id`、`idx_novel_arc_order`。

### operation_log

复用既有 `operation_log` 契约。确认选题、编辑世界观、新增人物必须返回 `operation_log_id`。

## 5. API Design

所有 API 使用统一 envelope、Bearer Auth、`X-Request-Id`，失败响应返回统一错误结构。

### POST /api/v1/projects/{projectId}/novel/planning-runs

- Header：`Idempotency-Key` 必填。
- Path：`projectId`。
- Body：`{genre: string, audience: string, count: number, template_version_id: string, input_override?: object}`。
- 202：`{planning_run_id, workflow_run_id, status}`。
- Errors：`VALIDATION_ERROR(400)`, `UNAUTHORIZED(401)`, `FORBIDDEN(403)`, `NOT_FOUND(404)`, `CONFLICT(409)`, `IDEMPOTENCY_CONFLICT(409)`, `WORKFLOW_RUN_FAILED(422)`, `INTERNAL_ERROR(500)`。

### GET /api/v1/projects/{projectId}/novel/planning-runs

- Query：`page`、`page_size`、`sort`、`order`、`status`。
- 200：`{items: [PlanningRunResponse], pagination}`。
- Errors：`VALIDATION_ERROR(400)`, `UNAUTHORIZED(401)`, `FORBIDDEN(403)`, `NOT_FOUND(404)`, `INTERNAL_ERROR(500)`。

### GET /api/v1/projects/{projectId}/novel/planning-runs/{runId}

- 200：`PlanningRunDetailResponse`，包含候选选题、`workflow_run_id`、step_runs 摘要、agent_tasks 摘要、llm_call_logs 摘要。
- Errors：`UNAUTHORIZED(401)`, `FORBIDDEN(403)`, `NOT_FOUND(404)`, `INTERNAL_ERROR(500)`。

### POST /api/v1/projects/{projectId}/novel/topics/{topicId}/confirm

- Header：`Idempotency-Key` 必填。
- Body：`{note: string}`。
- 200：`{confirmed_topic_id, previous_status, current_status, operation_log_id}`。
- Errors：`VALIDATION_ERROR(400)`, `UNAUTHORIZED(401)`, `FORBIDDEN(403)`, `NOT_FOUND(404)`, `CONFLICT(409)`, `IDEMPOTENCY_CONFLICT(409)`, `INTERNAL_ERROR(500)`。

### GET /api/v1/projects/{projectId}/novel/worldview

- 200：`WorldviewResponse`，包含 `project_id`、`worldview`、`forbidden_rules`、`version_id`、`version`。
- Errors：`UNAUTHORIZED(401)`, `FORBIDDEN(403)`, `NOT_FOUND(404)`, `INTERNAL_ERROR(500)`。

### PATCH /api/v1/projects/{projectId}/novel/worldview

- Body：`{worldview: object, forbidden_rules: array, note: string}`。
- 200：`{version_id, operation_log_id}`。
- Errors：`VALIDATION_ERROR(400)`, `UNAUTHORIZED(401)`, `FORBIDDEN(403)`, `NOT_FOUND(404)`, `INTERNAL_ERROR(500)`。

### GET /api/v1/projects/{projectId}/novel/characters

- Query：`page`、`page_size`、`sort`、`order`、`role`。
- 200：`{items: [CharacterResponse], pagination}`。
- Errors：`VALIDATION_ERROR(400)`, `UNAUTHORIZED(401)`, `FORBIDDEN(403)`, `NOT_FOUND(404)`, `INTERNAL_ERROR(500)`。

### POST /api/v1/projects/{projectId}/novel/characters

- Body：`{name: string, role: string, profile: object, note: string}`。
- 201：`{character_id, operation_log_id}`。
- Errors：`VALIDATION_ERROR(400)`, `UNAUTHORIZED(401)`, `FORBIDDEN(403)`, `NOT_FOUND(404)`, `CONFLICT(409)`, `INTERNAL_ERROR(500)`。

### GET /api/v1/projects/{projectId}/novel/arcs

- Query：`page`、`page_size`、`sort`、`order`。
- 200：`{items: [ArcResponse], pagination}`。
- Errors：`VALIDATION_ERROR(400)`, `UNAUTHORIZED(401)`, `FORBIDDEN(403)`, `NOT_FOUND(404)`, `INTERNAL_ERROR(500)`。

## 6. Module Design

### novel module

目录：

```text
apps/api-server/internal/modules/novel/
  dto.go
  service.go
```

职责：

- 管理 Novel Pack 规划运行、候选选题、规划快照、世界观、人物和大纲的业务 DTO 与 in-memory 状态。
- 负责 Novel Pack 边界校验、状态流转、幂等键冲突检测和 operation_log_id 生成。
- 不直接依赖 HTTP；由 handler 组合 workflow service 与 engine submitter。

接口：

```go
type Service interface {
    CreatePlanningRun(ctx context.Context, projectID string, req CreatePlanningRunRequest, workflowRunID string, idempotencyKey string) (CreatePlanningRunResponse, error)
    ListPlanningRuns(ctx context.Context, projectID string, req ListPlanningRunsRequest) (PagedPlanningRunsResponse, error)
    GetPlanningRun(ctx context.Context, projectID, runID string) (PlanningRunDetailResponse, error)
    ConfirmTopic(ctx context.Context, projectID, topicID string, req ConfirmTopicRequest, idempotencyKey string) (ConfirmTopicResponse, error)
    GetWorldview(ctx context.Context, projectID string) (WorldviewResponse, error)
    UpdateWorldview(ctx context.Context, projectID string, req UpdateWorldviewRequest) (UpdateWorldviewResponse, error)
    ListCharacters(ctx context.Context, projectID string, req ListCharactersRequest) (PagedCharactersResponse, error)
    CreateCharacter(ctx context.Context, projectID string, req CreateCharacterRequest) (CreateCharacterResponse, error)
    ListArcs(ctx context.Context, projectID string, req ListArcsRequest) (PagedArcsResponse, error)
}
```

错误变量：`ErrValidation`, `ErrNotFound`, `ErrForbidden`, `ErrConflict`, `ErrIdempotencyConflict`, `ErrWorkflowRunFailed`。

### http handlers

新增 `NovelHandler`：

- `CreatePlanningRun()`：组合 Novel Service、Workflow Service、Engine Submitter。
- `ListPlanningRuns()`、`GetPlanningRun()`。
- `ConfirmTopic()`。
- `GetWorldview()`、`UpdateWorldview()`。
- `ListCharacters()`、`CreateCharacter()`。
- `ListArcs()`。

错误映射：

- `novel.ErrValidation` → 400 `VALIDATION_ERROR`
- `novel.ErrForbidden` → 403 `FORBIDDEN`
- `novel.ErrNotFound` → 404 `NOT_FOUND`
- `novel.ErrConflict` → 409 `CONFLICT`
- `novel.ErrIdempotencyConflict` → 409 `IDEMPOTENCY_CONFLICT`
- `novel.ErrWorkflowRunFailed` → 422 `WORKFLOW_RUN_FAILED`

### router integration

在现有 `/api/v1` scope 中注册：

```text
POST  /projects/{id}/novel/planning-runs
GET   /projects/{id}/novel/planning-runs
GET   /projects/{id}/novel/planning-runs/{runId}
POST  /projects/{id}/novel/topics/{topicId}/confirm
GET   /projects/{id}/novel/worldview
PATCH /projects/{id}/novel/worldview
GET   /projects/{id}/novel/characters
POST  /projects/{id}/novel/characters
GET   /projects/{id}/novel/arcs
```

### web-admin

- `lib/api.ts` 新增 Novel Planning 类型与函数。
- 项目详情/工作区视图新增“内容规划”入口，链接使用当前项目的动态 `projectId`。
- `projects/[projectId]/workspace-nav.tsx` 提供项目工作区局部导航，覆盖内容规划、候选选题、世界观、人物、大纲，并支持当前路由高亮。
- `global-nav.tsx` 不硬编码具体项目的 Novel 路由，仅保留系统级导航。
- 新增页面全部使用 `page-shell`、`page-hero`、`card`、`table-card`、`badge`、`dialog-backdrop` 等现有 CSS。
- 页面错误态统一使用 `PageError`，展示 code/message/request_id。

## 7. Output Contract

| 产出 | 输入 | 输出 | 产出类型 | 功能类型 | 是否跨组件 | 测试规范 |
|---|---|---|---|---|---|---|
| Novel Planning HTTP API | HTTP path/query/header/body | 统一 envelope、HTTP status、业务 DTO、错误码 | web-e2e | Novel Planning API（type id: web-e2e） | 是（Router -> NovelHandler -> NovelService -> WorkflowService -> EngineSubmitter） | `standards/testing/web-e2e.md`, `standards/testing/integration.md` |
| Novel Service | projectID、DTO、workflowRunID、幂等键 | 规划运行、候选选题、世界观、人物、大纲 DTO | integration | Novel Pack service 状态机（type id: integration） | 是（NovelService -> WorkflowRun 关联数据） | `standards/testing/integration.md` |
| PostgreSQL migration DDL | Table Design | `00005_create_novel_planning_tables.sql` | sql-query | DDL contract（type id: sql-query） | 否 | `standards/testing/sql-query.md` |
| OpenAPI contract | API Design | paths、schemas、examples、security | integration | API 文档契约（type id: integration） | 否 | `standards/testing/integration.md` |
| Web Admin Novel pages | 用户导航、表单、按钮、API envelope | 样式化页面、弹窗、Toast、错误态 | web-e2e | Web Admin 页面（type id: web-e2e） | 是（GlobalNav -> Page -> lib/api -> API Server） | `standards/testing/web-e2e.md`, `standards/testing/integration.md` |

### SQL Contract

目标方言：PostgreSQL。

固定 DDL 文件：`apps/api-server/migrations/00005_create_novel_planning_tables.sql`。

必须包含：

- `planning_run`
- `planning_snapshot`
- `novel_topic_candidate`
- `novel_worldview`
- `novel_character`
- `novel_arc`

关键结构：

- `planning_run.workflow_run_id` 必须唯一，保证业务运行可追踪到 WorkflowRun。
- `planning_snapshot.planning_run_id` 外键关联 `planning_run.id`，`planning_snapshot.project_id` 与 run 的项目保持一致。
- `novel_topic_candidate.planning_run_id` 外键关联 `planning_run.id`，`novel_topic_candidate.snapshot_id` 外键关联 `planning_snapshot.id`。
- `novel_worldview`、`novel_character`、`novel_arc` 的 `planning_run_id` 和 `snapshot_id` 可空但若存在必须关联对应规划运行和快照。
- 所有 `novel_*` 表必须包含 `project_id`，保证 Novel Pack 扩展资源只挂在项目下。
- 候选选题必须包含 `status`、`score`、`reason`、`snapshot_id`，同一项目同一候选 ID 唯一。
- 人工确认不写入 Book / Chapter 表；确认状态通过 `novel_topic_candidate.status` 和 `operation_log` 追踪。

禁止模式：

- 不创建 Core 表 `book`、`chapter`、`novel`。
- 不把 `novel_*` 字段加入 Core 通用表。
- 不在 DDL 中存储明文模型凭据、Token 或 Secret。

典型输入输出：

- 输入：创建规划运行 `{project_id=project-1, template_version_id=wftv-1, genre=fantasy, audience=young-adult, count=3}`。
- 输出：`planning_run` 一行、`planning_snapshot` 一行、若干 `novel_topic_candidate` 行，且 `planning_run.workflow_run_id` 可关联 WorkflowRun。

## 8. Change Log

| 文件 | 类型 | 原因 |
|---|---|---|
| `.cube/iterations/feature-3/design.md` | 新增 | 本阶段设计文档。 |
| `.cube/iterations/feature-3/skeleton-map.yaml` | 新增 | 骨架文件与 Development Tasks 映射。 |
| `apps/api-server/internal/modules/novel/dto.go` | 新增 | Novel Pack DTO 骨架。 |
| `apps/api-server/internal/modules/novel/service.go` | 新增 | Novel Pack Service interface 与骨架实现。 |
| `apps/api-server/internal/http/handlers/novel.go` | 新增 | Novel Pack HTTP handler 骨架。 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册 Novel Pack API 路由。 |
| `apps/api-server/migrations/00005_create_novel_planning_tables.sql` | 新增 | Novel Pack 数据表契约。 |
| `openapi/openapi.yaml` | 修改 | 追加 Novel Pack API 契约说明。 |
| `apps/web-admin/lib/api.ts` | 修改 | Novel Planning API client 类型与函数。 |
| `apps/web-admin/app/page.tsx` | 修改 | 项目详情/工作区增加当前项目内容规划入口。 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 新增 | 项目工作区局部导航骨架。 |
| `apps/web-admin/app/global-nav.tsx` | 修改 | 保持系统级导航，不硬编码项目级 Novel 路由。 |
| `apps/web-admin/app/projects/[projectId]/planning/page.tsx` | 新增 | 内容规划页骨架。 |
| `apps/web-admin/app/projects/[projectId]/planning/topics/page.tsx` | 新增 | 候选选题直达页骨架。 |
| `apps/web-admin/app/projects/[projectId]/novel/worldview/page.tsx` | 新增 | 世界观页骨架。 |
| `apps/web-admin/app/projects/[projectId]/novel/characters/page.tsx` | 新增 | 人物管理页骨架。 |
| `apps/web-admin/app/projects/[projectId]/novel/arcs/page.tsx` | 新增 | 大纲管理页骨架。 |

## 9. Development Tasks

- Task-01：定义 Novel Pack DTO 与 Service 契约
  - 所属模块：novel
  - 简要描述：定义规划运行、候选选题、世界观、人物、大纲 DTO、分页响应、错误变量和 Service interface。
  - 涉及接口/方法：novel.Service, CreatePlanningRun(), ListPlanningRuns(), GetPlanningRun(), ConfirmTopic(), GetWorldview(), UpdateWorldview(), ListCharacters(), CreateCharacter(), ListArcs()
  - 输入：projectID、分页参数、规划请求、确认请求、世界观请求、人物请求。
  - 输出：Novel Pack DTO 与错误。
  - 产出类型：library
  - 功能类型：Novel Pack 服务契约（type id: library）
  - 是否跨组件：否
- Task-02：实现 Novel Planning Service 状态与幂等规则
  - 所属模块：novel
  - 简要描述：实现 in-memory planning_run、planning_snapshot、候选选题、世界观、人物、大纲、状态流转、operation_log_id 和幂等冲突规则。
  - 涉及接口/方法：novelService.CreatePlanningRun(), ConfirmTopic(), UpdateWorldview(), CreateCharacter()
  - 输入：Service 请求 DTO、workflow_run_id、Idempotency-Key。
  - 输出：业务响应、状态变更、operation_log_id。
  - 产出类型：integration
  - 功能类型：Novel Pack 状态机（type id: integration）
  - 是否跨组件：是（组件链路：NovelHandler -> NovelService -> WorkflowService）
- Task-03：暴露 Novel Planning HTTP API
  - 所属模块：http/novel
  - 简要描述：实现 NovelHandler、路由注册、请求解析、Idempotency-Key 校验、统一 envelope 和错误映射。
  - 涉及接口/方法：NovelHandler.CreatePlanningRun(), ListPlanningRuns(), GetPlanningRun(), ConfirmTopic(), GetWorldview(), UpdateWorldview(), ListCharacters(), CreateCharacter(), ListArcs()
  - 输入：HTTP path/query/header/body。
  - 输出：HTTP status、统一 API envelope、错误码。
  - 产出类型：web-e2e
  - 功能类型：Novel Planning HTTP API（type id: web-e2e）
  - 是否跨组件：是（组件链路：HTTP Router -> NovelHandler -> NovelService -> WorkflowService -> EngineSubmitter）
- Task-04：补齐 Novel Planning 数据迁移与 OpenAPI 契约
  - 所属模块：contract
  - 简要描述：新增 PostgreSQL DDL contract 和 OpenAPI paths/schemas/examples/security，覆盖本迭代全部接口和表结构。
  - 涉及接口/方法：00005_create_novel_planning_tables.sql, openapi.yaml
  - 输入：Table Design 与 API Design。
  - 输出：DDL、OpenAPI contract。
  - 产出类型：sql-query
  - 功能类型：DDL 与 API contract（type id: sql-query）
  - 是否跨组件：否
- Task-05：扩展 Web Admin API client 与项目工作区导航入口
  - 所属模块：web-admin
  - 简要描述：在 `lib/api.ts` 增加 Novel Planning 类型与函数，在项目详情/工作区视图中加入当前项目的内容规划入口，并新增项目工作区局部导航支持规划、世界观、人物、大纲高亮。
  - 涉及接口/方法：fetchPlanningRuns(), createPlanningRun(), fetchPlanningRun(), confirmTopic(), fetchWorldview(), updateWorldview(), fetchCharacters(), createCharacter(), fetchArcs(), ProjectWorkspaceNav
  - 输入：当前 projectId、页面表单、分页、筛选和状态动作参数。
  - 输出：typed APIEnvelope、动态项目级导航链接。
  - 产出类型：library
  - 功能类型：前端 API client 与项目工作区导航（type id: library）
  - 是否跨组件：否
- Task-06：实现内容规划页与候选确认弹窗
  - 所属模块：web-admin
  - 简要描述：实现 `/projects/[projectId]/planning` 和 `/projects/[projectId]/planning/topics`，覆盖启动规划、运行列表、详情摘要、候选选题弹窗、确认动作、loading/empty/error/success。
  - 涉及接口/方法：PlanningPage, TopicsPage, createPlanningRun(), fetchPlanningRuns(), fetchPlanningRun(), confirmTopic()
  - 输入：用户导航、规划表单、候选确认 note。
  - 输出：样式化页面、弹窗、Toast、错误态。
  - 产出类型：web-e2e
  - 功能类型：内容规划页面（type id: web-e2e）
  - 是否跨组件：是（组件链路：GlobalNav -> PlanningPage -> lib/api -> API Server）
- Task-07：实现世界观编辑页
  - 所属模块：web-admin
  - 简要描述：实现 `/projects/[projectId]/novel/worldview`，支持查看、编辑、保存、空态、错误态和成功反馈。
  - 涉及接口/方法：WorldviewPage, fetchWorldview(), updateWorldview()
  - 输入：worldview、forbidden_rules、note。
  - 输出：版本信息、operation_log_id、页面反馈。
  - 产出类型：web-e2e
  - 功能类型：世界观编辑页面（type id: web-e2e）
  - 是否跨组件：是（组件链路：GlobalNav -> WorldviewPage -> lib/api -> API Server）
- Task-08：实现人物管理页
  - 所属模块：web-admin
  - 简要描述：实现 `/projects/[projectId]/novel/characters`，支持分页、角色筛选、新增人物、loading/empty/error/success。
  - 涉及接口/方法：CharactersPage, fetchCharacters(), createCharacter()
  - 输入：page、page_size、role、name、profile、note。
  - 输出：人物列表、character_id、operation_log_id。
  - 产出类型：web-e2e
  - 功能类型：人物管理页面（type id: web-e2e）
  - 是否跨组件：是（组件链路：GlobalNav -> CharactersPage -> lib/api -> API Server）
- Task-09：实现大纲管理页
  - 所属模块：web-admin
  - 简要描述：实现 `/projects/[projectId]/novel/arcs`，支持分页、排序、规划来源展示、loading/empty/error/success。
  - 涉及接口/方法：ArcsPage, fetchArcs()
  - 输入：page、page_size、sort、order。
  - 输出：弧线大纲列表与来源信息。
  - 产出类型：web-e2e
  - 功能类型：大纲管理页面（type id: web-e2e）
  - 是否跨组件：是（组件链路：ProjectWorkspaceNav -> ArcsPage -> lib/api -> API Server）
- Task-10：补齐 Novel Planning 自动化测试覆盖
  - 所属模块：test-contract
  - 简要描述：为 DTO 校验、幂等冲突、候选确认状态流转、统一错误响应、分页查询、OpenAPI contract、DDL contract、前端成功态和失败态编写测试。
  - 涉及接口/方法：novel service tests, novel handler contract tests, OpenAPI contract tests, migration contract tests, Playwright e2e
  - 输入：设计中的 API、Table Design、Output Contract 和页面交互。
  - 输出：可在 03 阶段先失败、04 阶段实现后通过的测试契约。
  - 产出类型：integration
  - 功能类型：Novel Planning 测试契约（type id: integration）
  - 是否跨组件：是（组件链路：tests -> HTTP Router -> NovelHandler -> NovelService -> Web Admin）
