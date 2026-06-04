# 技术设计方案：Iteration 12 — Article Pack 内容类型扩展

## 1. 概述

本设计将 Article（公众号/知乎/SEO 文章）以 Content Pack 插件化方式接入 AI Content Factory Core。设计遵循最小改动原则，复用现有 ContentType、WorkflowTemplate、GenerationRun、ContentItem、MetricTemplate 等基础设施，新增 `article` 模块（类似于现有 `novel` 模块模式），实现 Article Pack 注册、Article 项目生产、Article 指标配置三大功能。

**核心原则**：Core 层不引入 Article/Novel/Book/Chapter 作为核心资源命名（已在 PRD 验收标准中声明）。

## 2. Impact Analysis

| 模块 | 影响程度 | 说明 |
|------|---------|------|
| `internal/modules/content` | 修改（小） | Service 接口新增 Article 扩展配置 DTO 和查询方法 |
| `internal/modules/article` | **新增** | 全新模块，遵循 novel 模块模式：DTO + Service + 接口定义 |
| `internal/http/handlers` | **新增** | 新增 `article_handler.go`，遵循 novel_handler.go 模式 |
| `internal/http/router.go` | 修改 | 新增 Article 路由组（类似 Novel） |
| `internal/modules/generation` | 无影响 | 仅通过 handler 桥接，不修改 generation 核心代码 |
| `internal/modules/metrics` | 无影响 | 通过 ArticleService.RegisterPack 创建 MetricTemplates（运行时），不修改 metrics 核心代码 |
| `internal/engine` | 无影响 | 复现有 Workflow Engine |
| `apps/web-admin` | **新增页面** | 三张管理台页面（FR-012, FR-013, FR-014） |
| `openapi/openapi.yaml` | **修改** | 新增 Article 相关 API 的 OpenAPI 描述 |

### 兼容性分析

- **接口兼容性**：所有新增 API 路径以 `/api/v1/content-packs/article` 和 `/projects/{projectId}/article` 为前缀，与现有路由无冲突
- **数据兼容性**：Article 扩展配置独立存储（DTO 层面），不修改现有 ContentProject 或 ContentType 的数据结构
- **前端兼容性**：新增页面接入现有导航体系，不修改已有页面

## 3. Flow Design

### FR-1: Article Pack 注册流程

```
Client → POST /api/v1/content-packs/article/register
  → ArticleHandler.RegisterPack
    → ArticleService.RegisterPack
      → 创建 seed ContentType(code=article)
      → 创建默认 WorkflowTemplate + WorkflowTemplateVersion
      → 创建默认 MetricTemplates
      → 写入 operation_log
    → 返回 content_pack_id, content_type_id, workflow_version_ids, metric_template_ids
```

### FR-2: Article 项目扩展配置流程

```
GET /projects/{projectId}/article/config
  → ArticleHandler.GetConfig
    → ArticleService.GetConfig
      → 验证项目存在且 ContentType 为 article
      → 返回 ArticleConfig

PATCH /projects/{projectId}/article/config
  → ArticleHandler.UpdateConfig
    → ArticleService.UpdateConfig
      → 验证 + 校验
      → 更新配置
      → 写入 operation_log
      → 返回 version_id
```

### FR-3: Article 生成运行流程

```
POST /projects/{projectId}/article/generation-runs
  → ArticleHandler.CreateGenerationRun
    → 获取 Article 扩展配置
    → 创建 WorkflowRun
    → 创建 GenerationRun（复用 generation 模块）
    → 提交到 Engine 异步执行
    → AgentTask / LLMCallLog 通过 Engine 写入（复用现有 Engine 的逻辑）
    → 返回 generation_run_id, workflow_run_id

GET generation list / detail → 复用现有 generation 查询路径
  → detail 返回中包含 step_runs, agent_tasks, llm_call_logs（通过 generation 模块桥接）
  → 详情优先读取 content_version（当版本化内容存在时）

GET content snapshot → 复用 content-item 查询路径
  → 返回 title, summary, outline, seo_metadata, source_refs, latest_content_version_id

POST /projects/{projectId}/article/generation-runs/{id}/retry
  → 验证运行状态为 failed（否则 CONFLICT）
  → 创建新 WorkflowRun
  → 创建新 GenerationRun（retry_of=原 run_id）
  → 写入 operation_log（记录 retry_reason）
  → 返回新 generation_run_id, workflow_run_id
```

### FR-4: Article 指标配置流程

```
GET /projects/{projectId}/article/metrics
  → ArticleHandler.GetMetricsConfig
    → ArticleService.GetProjectArticleMetrics
      → 返回模板列表 + 项目启用状态

PATCH /projects/{projectId}/article/metrics
  → ArticleHandler.UpdateMetricsConfig
    → ArticleService.UpdateProjectArticleMetrics
      → 更新启用指标码
      → 写入 operation_log
      → 返回 version_id
```

### 异常流程

- 幂等冲突：所有创建类接口支持 Idempotency-Key，输入变化返回 IDEMPOTENCY_CONFLICT
- 项目不存在 / ContentType 不匹配：返回 NOT_FOUND / FORBIDDEN
- 字段校验失败：返回 VALIDATION_ERROR
- 状态冲突：返回 CONFLICT

## 4. Table Design

本次迭代使用 DTO 层内存存储（与现有模块一致，无独立 Article 数据表）。扩展配置存储在 `article.ArticleConfigResponse` 对象中。核心实体复用已有 Infrastructure：

| 实体 | 复用位置 | 说明 |
|------|---------|------|
| ContentType | content.Service | code 字段设为 "article" |
| WorkflowTemplate | workflow.Service | Article 默认工作流模板 |
| WorkflowTemplateVersion | workflow.Service | 资料整理/大纲/正文/质检四步骤 |
| GenerationRun | generation.Service | 复现有生成运行 |
| ContentItem | generation.Service | 复现有内容项 |
| MetricTemplate | metrics.Service | Article 默认指标模板 |

## 5. API Design

### Article Pack 注册

| Method | Path | 描述 | FR |
|--------|------|------|-----|
| POST | `/api/v1/content-packs/article/register` | 注册 Article Pack | FR-001 |
| GET | `/api/v1/content-packs/article/status` | 查看注册状态 | FR-002 |

#### POST /api/v1/content-packs/article/register

**请求**：
```json
{
  "idempotency_key": "string"
}
```

**响应 (201)**：
```json
{
  "content_pack_id": "content-pack-1",
  "content_type_id": "content-type-new",
  "registered_workflow_version_ids": ["wftv-new"],
  "metric_template_ids": ["mt-1"]
}
```

**错误**：VALIDATION_ERROR, IDEMPOTENCY_CONFLICT

#### GET /api/v1/content-packs/article/status

**响应 (200)**：
```json
{
  "registered": true,
  "content_pack_id": "content-pack-1",
  "content_type": {"id": "...", "code": "article", "name": "Article Pack"},
  "default_workflow_template": {"id": "...", "name": "Article Generation", "status": "published"},
  "default_metrics": [{"metric_code": "views", "name": "阅读量", "unit": "次"}]
}
```

**错误**：NOT_FOUND

### Article 项目扩展配置

| Method | Path | 描述 | FR |
|--------|------|------|-----|
| GET | `/projects/{projectId}/article/config` | 获取 Article 扩展配置 | FR-003 |
| PATCH | `/projects/{projectId}/article/config` | 更新 Article 扩展配置 | FR-004 |

#### GET /projects/{projectId}/article/config

**响应 (200)**：
```json
{
  "topic_style": "string",
  "audience_profile": "string",
  "seo_config": {"keywords": ["string"]},
  "source_policy": "string",
  "structure_policy": "string",
  "default_workflow_template_version_id": "string",
  "enabled_metric_codes": ["string"],
  "version": "string"
}
```

**错误**：NOT_FOUND, FORBIDDEN

#### PATCH /projects/{projectId}/article/config

**请求**：
```json
{
  "topic_style": "string",
  "audience_profile": "string",
  "seo_config": {"keywords": ["string"]},
  "source_policy": "string",
  "structure_policy": "string",
  "default_workflow_template_version_id": "string"
}
```

**响应 (200)**：
```json
{
  "version_id": "article-config-v1",
  "operation_log_id": "oplog-1"
}
```

**错误**：VALIDATION_ERROR, CONFLICT

### Article 生成运行

| Method | Path | 描述 | FR |
|--------|------|------|-----|
| POST | `/projects/{projectId}/article/generation-runs` | 创建文章生成 | FR-005 |
| GET | `/projects/{projectId}/article/generation-runs` | 获取生成列表 | FR-006 |
| GET | `/projects/{projectId}/article/generation-runs/{id}` | 获取生成详情 | FR-007 |
| POST | `/projects/{projectId}/article/generation-runs/{id}/retry` | 失败重试 | FR-008 |
| GET | `/projects/{projectId}/article/content-items/{itemId}` | 获取内容快照 | FR-009 |

#### POST /projects/{projectId}/article/generation-runs

路径参数：
- `projectId` (string, required)：项目 ID

**请求**：
```json
{
  "topic": "string",
  "audience": "string",
  "source_refs": ["string"],
  "seo_keywords": ["string"],
  "outline_required": true,
  "target_platform": "string",
  "generation_config": {}
}
```
Idempotency-Key 通过 HTTP Header 传递（非请求体字段）。

**响应 (202)**：
```json
{
  "generation_run_id": "genrun-article-1",
  "workflow_run_id": "wfr-1",
  "status": "pending"
}
```

**错误**：VALIDATION_ERROR, NOT_FOUND, FORBIDDEN, IDEMPOTENCY_CONFLICT

### Article 指标配置

| Method | Path | 描述 | FR |
|--------|------|------|-----|
| GET | `/projects/{projectId}/article/metrics` | 获取指标配置 | FR-010 |
| PATCH | `/projects/{projectId}/article/metrics` | 更新指标配置 | FR-011 |

## 6. Module Design

### 模块划分

```
internal/modules/article/
├── dto.go          — DTO 定义（请求/响应结构体）
├── service.go      — Service 接口 + 内存实现
└── errors.go       — 错误常量

internal/http/handlers/
├── article.go      — ArticleHandler（新增）

internal/modules/content/
└── service.go      — 扩展：新增 ArticleConfig 相关方法
```

### 接口定义

**ArticleService**（遵循 novel.Service 接口设计风格）：
```
- RegisterPack(ctx, request, idempotencyKey) → (RegisterPackResponse, error)
- GetPackStatus(ctx) → (ArticlePackStatusResponse, error)
- GetConfig(ctx, projectID) → (ArticleConfigResponse, error)
- UpdateConfig(ctx, projectID, request, idempotencyKey) → (UpdateArticleConfigResponse, error)
- CreateGenerationRun(ctx, projectID, request, workflowRunID, idempotencyKey) → (generation.CreateGenerationRunResponse, error)
- ListGenerationRuns(ctx, projectID, request) → (PagedArticleGenerationRunResponse, error)
- GetGenerationRun(ctx, id) → (ArticleGenerationRunDetailResponse, error)
- RetryGenerationRun(ctx, id, request, workflowRunID, idempotencyKey) → (generation.RetryGenerationRunResponse, error)
- GetContentSnapshot(ctx, itemID) → (ArticleContentSnapshotResponse, error)
- GetProjectArticleMetrics(ctx, projectID, request) → (PagedProjectArticleMetricsResponse, error)
- UpdateProjectArticleMetrics(ctx, projectID, request, idempotencyKey) → (UpdateProjectArticleMetricsResponse, error)
```

### 依赖关系

```
ArticleHandler → ArticleService
ArticleHandler → workflow.Service (for creating workflow runs)
ArticleHandler → engine.Submitter (for async execution)
ArticleHandler → content.Service (for project validation)
ArticleHandler → metrics.Service (for metric template queries)
```

## 7. Output Contract

### 功能类型标记

| 功能 | type id | 引用规范 |
|------|---------|---------|
| Article Pack 注册 API (POST/GET) | web-e2e | `standards/testing/web-e2e.md` |
| Article 生成运行 API (POST/GET/POST retry) | web-e2e | `standards/testing/web-e2e.md` |
| Article 扩展配置 API (GET/PATCH) | web-e2e | `standards/testing/web-e2e.md` |
| Article 指标配置 API (GET/PATCH) | web-e2e | `standards/testing/web-e2e.md` |
| Article 前端页面（3 张页面） | frontend-ui | `standards/testing/frontend-ui.md` |
| Article 生成运行时 Handler → Service → Engine 组件链路 | integration | `standards/testing/integration.md` |

> 注意：workflow.yaml features 为空列表，以上 type id 基于本次迭代实际内容推断。建议将 `web-api`、`frontend-ui` 补充到 workflow.yaml。

### 各 API 正确性规则

- 每个 API 必须返回标准 Envelope（success/data/error/request_id）
- 创建类 API 返回 201/202 状态码
- 查询类 API 返回 200 状态码
- 列表 API 必须支持 page/page_size/sort/order 分页参数
- 所有创建类 API 支持 Idempotency-Key 头
- PATCH 类 API 变更必须写入 operation_log
- 错误码使用 api-spec.md 定义的规范错误码

### 各 API 错误码汇总

| API | 可能错误码 |
|-----|-----------|
| POST /api/v1/content-packs/article/register | VALIDATION_ERROR, IDEMPOTENCY_CONFLICT |
| GET /api/v1/content-packs/article/status | NOT_FOUND |
| GET /projects/{projectId}/article/config | NOT_FOUND, FORBIDDEN |
| PATCH /projects/{projectId}/article/config | VALIDATION_ERROR, CONFLICT, NOT_FOUND |
| POST /projects/{projectId}/article/generation-runs | VALIDATION_ERROR, NOT_FOUND, FORBIDDEN, IDEMPOTENCY_CONFLICT |
| GET /projects/{projectId}/article/generation-runs | VALIDATION_ERROR, NOT_FOUND, FORBIDDEN |
| GET /projects/{projectId}/article/generation-runs/{id} | VALIDATION_ERROR, NOT_FOUND, FORBIDDEN |
| POST /projects/{projectId}/article/generation-runs/{id}/retry | VALIDATION_ERROR, NOT_FOUND, FORBIDDEN, CONFLICT, IDEMPOTENCY_CONFLICT |
| GET /projects/{projectId}/article/content-items/{itemId} | NOT_FOUND, FORBIDDEN |
| GET /projects/{projectId}/article/metrics | VALIDATION_ERROR, NOT_FOUND, FORBIDDEN |
| PATCH /projects/{projectId}/article/metrics | VALIDATION_ERROR, NOT_FOUND, FORBIDDEN, IDEMPOTENCY_CONFLICT |

## 8. Change Log

| 文件 | 变更类型 | 原因 |
|------|---------|------|
| `internal/modules/content/service.go` | **无影响** | content.Service 已具备 CreateContentType/ListProjects 等能力，Article 模块直接调用即可 |
| `internal/modules/article/dto.go` | **新增** | Article 模块 DTO 定义 |
| `internal/modules/article/errors.go` | **新增** | Article 模块错误常量 |
| `internal/modules/article/service.go` | **新增** | Article 模块 Service 接口和内存实现 |
| `internal/http/handlers/article.go` | **新增** | Article HTTP Handler |
| `internal/http/router.go` | **修改** | 新增 Article 路由组（类似 Novel） |
| `openapi/openapi.yaml` | **修改** | 新增 Article 相关 API 的 OpenAPI 描述 |
| `apps/web-admin/app/article-pack/page.tsx` | **新增** | FR-012 页面 |
| `apps/web-admin/app/projects/[projectId]/article/page.tsx` | **新增** | FR-013 页面 |
| `apps/web-admin/app/projects/[projectId]/article/metrics/page.tsx` | **新增** | FR-014 页面 |
| `apps/web-admin/components/article/` | **新增** | Article 相关前端组件 |
| `apps/web-admin/lib/api.ts` | **修改** | 新增 Article API 客户端函数 |

## 9. Development Tasks

- Task-01：定义 Article DTO、错误常量和 Service 接口
  - 任务类型：contract
  - 所属模块：api-server/article
  - 简要描述：定义 RegisterPack/GetPackStatus/GetConfig/UpdateConfig/GenerationRun/Retry/Metrics 等 DTO 和常量，包含 Service 接口、空实现和构造函数
  - 涉及接口/方法：article.Service、NewService()
  - 输入：各 API request DTO
  - 输出：各 API response DTO 或 error
  - 依赖任务：无
  - 数据操作：无
  - 修改边界：只新增 internal/modules/article/dto.go、errors.go、service.go（包含空方法体）
  - 禁止行为：不得写业务逻辑；不得访问数据库或外部系统
  - 产出类型：integration
  - 功能类型：Article 模块接口契约（type id: integration）
  - 是否跨组件：否

- Task-02：实现 Article Pack 注册业务逻辑
  - 任务类型：business-implementation
  - 所属模块：api-server/article
  - 简要描述：注册 Article ContentType（通过 content.Service）、默认 WorkflowTemplate（通过 workflow.Service）、默认 MetricTemplates（通过 metrics.Service），支持幂等和状态查询。RegisterPack 在运行时创建上述实体，不依赖种子数据。
  - 涉及接口/方法：RegisterPack()、GetPackStatus()
  - 输入：RegisterPackRequest、Idempotency-Key
  - 输出：RegisterPackResponse、ArticlePackStatusResponse 或 error
  - 依赖任务：Task-01（Service 接口）
  - 数据操作：读/写 content.Service 的 contentTypes；读/写 workflow.Service 的 templates/versions；读/写 metrics.Service 的 templates；写 operation_log（模拟）
  - 修改边界：只替换 RegisterPack() 和 GetPackStatus() 的空实现，不删除或重写 service.go
  - 禁止行为：不得使用内存存储替代声明的数据操作
  - 产出类型：integration
  - 功能类型：Article Pack 注册业务实现（type id: integration）
  - 是否跨组件：是（组件链路：ArticleHandler -> ArticleService -> ContentService + WorkflowService + MetricsService）

- Task-03：实现 Article 扩展配置业务逻辑
  - 任务类型：business-implementation
  - 所属模块：api-server/article
  - 简要描述：支持项目级别 Article 扩展配置的获取和更新，写入 operation_log
  - 涉及接口/方法：GetConfig()、UpdateConfig()
  - 输入：projectId、UpdateArticleConfigRequest、Idempotency-Key
  - 输出：ArticleConfigResponse、UpdateArticleConfigResponse 或 error
  - 依赖任务：Task-01（Service 接口、DTO）
  - 数据操作：读/写 article 内部配置存储；写 operation_log
  - 修改边界：只替换 GetConfig() 和 UpdateConfig() 的空实现，允许通过 content.Service 验证项目存在和 ContentType
  - 禁止行为：不得修改 content 模块的 Project 数据结构或添加新方法
  - 产出类型：integration
  - 功能类型：Article 扩展配置业务实现（type id: integration）
  - 是否跨组件：否

- Task-04：实现 Article 生成运行业务逻辑
  - 任务类型：business-implementation
  - 所属模块：api-server/article
  - 简要描述：创建文章生成运行（经 WorkflowRun + GenerationRun 执行），支持列表、详情查询、失败重试（写入 operation_log 记录 retry_reason）、内容快照。详情查询优先返回 versioned content 信息。AgentTask/LLMCallLog 通过 Engine 桥接写入。
  - 涉及接口/方法：CreateGenerationRun()、ListGenerationRuns()、GetGenerationRun()、RetryGenerationRun()、GetContentSnapshot()
  - 输入：CreateArticleGenerationRunRequest、ListGenerationRunsRequest、RetryGenerationRunRequest、Idempotency-Key
  - 输出：generation.CreateGenerationRunResponse / RetryGenerationRunResponse、ArticleGenerationRunDetailResponse、ArticleContentSnapshotResponse
  - 依赖任务：Task-01（Service 接口、DTO）、Task-03（Article 扩展配置）
  - 数据操作：读 article 配置；读 content_version 表；读 generation runs/items；写 generation runs/items；写 operation_log（retry 原因）
  - 修改边界：只替换 CreateGenerationRun/ListGenerationRuns/GetGenerationRun/RetryGenerationRun/GetContentSnapshot 的空实现，允许调用 workflow.Service（创建 WorkflowRun）、generation.Service（创建/查询 GenerationRun 和 ContentItem）、content.Service（项目验证）
  - 禁止行为：不得绕过 WorkflowRun/GenerationRun 直接写 ContentItem
  - 产出类型：integration
  - 功能类型：Article 生成业务实现（type id: integration）
  - 是否跨组件：是（组件链路：ArticleHandler -> ArticleService -> WorkflowService + GenerationService + OperationLog）

- Task-05：实现 Article 指标配置业务逻辑
  - 任务类型：business-implementation
  - 所属模块：api-server/article
  - 简要描述：支持项目 Article 指标模板查询、启用/停用配置，写入 operation_log
  - 涉及接口/方法：GetProjectArticleMetrics()、UpdateProjectArticleMetrics()
  - 输入：projectId、UpdateProjectArticleMetricsRequest、Idempotency-Key
  - 输出：PagedProjectArticleMetricsResponse、UpdateProjectArticleMetricsResponse
  - 依赖任务：Task-01（Service 接口、DTO）
  - 数据操作：读 metrics templates；读/写 article 内部指标配置；写 operation_log
  - 修改边界：只替换指标配置相关的空实现
  - 禁止行为：不得直接写入 MetricRecord；不得修改 metrics.Service 数据
  - 产出类型：integration
  - 功能类型：Article 指标配置业务实现（type id: integration）
  - 是否跨组件：否

- Task-06：实现 Article HTTP Handler 和路由注册
  - 任务类型：api
  - 所属模块：api-server/http
  - 简要描述：实现 ArticleHandler（遵循 NovelHandler 模式），注册所有 Article 路由到 router。Handler 持有 content.Service（项目验证）、workflow.Service（创建 WorkflowRun）、metrics.Service（指标查询）、engine.Submitter（异步执行）等多模块依赖。
  - 涉及接口/方法：ArticleHandler 各方法、NewArticleHandler()、router Route 调用
  - 输入：HTTP 请求
  - 输出：HTTP 响应
  - 依赖任务：Task-02（RegisterPack）、Task-04（生成运行）、Task-05（指标配置）
  - 数据操作：无（代理到 ArticleService）
  - 修改边界：只新增 article_handler.go；只修改 router.go 新增 Article 路由组
  - 禁止行为：不得修改已有 handler 的逻辑；不得移除已有路由
  - 产出类型：web-e2e
  - 功能类型：Article API 端点（type id: web-e2e）
  - 是否跨组件：是（组件链路：HTTP Router -> ArticleHandler -> ArticleService + ContentService + WorkflowService + MetricsService + Engine）

- Task-07：实现 Article 前端管理台页面（FR-012 Article Pack 页）
  - 任务类型：ui
  - 所属模块：web-admin
  - 简要描述：实现 `/content-packs/article` 页面，展示 Article Pack 注册状态、schema 摘要、默认 workflow、默认 metrics，支持注册/重新注册
  - 涉及接口/方法：ArticlePackPage 组件、API 客户端函数
  - 输入：无
  - 输出：渲染页面
  - 依赖任务：Task-06（Handler 和路由就绪）
  - 数据操作：调用 GET /api/v1/content-packs/article/status、POST /api/v1/content-packs/article/register
  - 修改边界：只新增 page.tsx 和相关组件
  - 禁止行为：不得修改已有页面
  - 产出类型：frontend-ui
  - 功能类型：Article Pack 管理页面（type id: frontend-ui）
  - 是否跨组件：是（组件链路：Web Browser -> Next.js Page -> API Handler）

- Task-08：实现 Article 前端管理台页面（FR-013 内容规划与生产页）
  - 任务类型：ui
  - 所属模块：web-admin
  - 简要描述：实现 `/projects/:projectId/article` 页面，展示 Article 扩展配置、生成输入表单、运行列表、生成详情、失败重试入口
  - 涉及接口/方法：ArticleProjectPage 组件、API 客户端函数
  - 输入：无
  - 输出：渲染页面
  - 依赖任务：Task-06（Handler 和路由就绪）
  - 数据操作：调用生成运行相关 API
  - 修改边界：只新增 page.tsx 和相关组件
  - 禁止行为：不得修改已有页面
  - 产出类型：frontend-ui
  - 功能类型：Article 内容规划页面（type id: frontend-ui）
  - 是否跨组件：是（组件链路：Web Browser -> Next.js Page -> API Handler）

- Task-09：实现 Article 前端管理台页面（FR-014 指标配置页）
  - 任务类型：ui
  - 所属模块：web-admin
  - 简要描述：实现 `/projects/:projectId/article/metrics` 页面，展示指标模板、平台差异、项目启用状态，支持启用/停用
  - 涉及接口/方法：ArticleMetricsPage 组件、API 客户端函数
  - 输入：无
  - 输出：渲染页面
  - 依赖任务：Task-06（Handler 和路由就绪）
  - 数据操作：调用指标配置 API
  - 修改边界：只新增 page.tsx 和相关组件
  - 禁止行为：不得修改已有页面
  - 产出类型：frontend-ui
  - 功能类型：Article 指标配置页面（type id: frontend-ui）
  - 是否跨组件：是（组件链路：Web Browser -> Next.js Page -> API Handler）
