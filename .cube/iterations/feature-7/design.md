# Iteration 7 技术设计：发布队列与手动发布回填

## 1. 概述

本次设计在现有分层架构中新增 `publish` 业务模块，围绕项目级发布目标、发布任务队列、复制发布载荷和手动发布回填构建闭环。后端沿用现有 `handler -> module service -> DTO -> api.Envelope` 模式，前端沿用 `apps/web-admin/lib/api.ts` API 客户端和项目工作区导航模式，OpenAPI 继续维护单一 `openapi/openapi.yaml`。

核心约束：

- 只支持人工发布闭环，不做平台自动发布和凭证保存。
- 发布目标配置只保存非敏感展示配置，包含 token、cookie、password、secret、credential、api key 等键名或密钥引用时返回 `VALIDATION_ERROR`。
- 发布任务绑定 `content_version_id` 和 `payload_hash`，复制预览不改变状态，点击复制动作才写日志并允许 `queued -> copied`。
- 所有状态变更遵守 `queued/copied/published/failed/canceled` 状态机，其中 `canceled` 只作为枚举预留，不开放接口。
- 所有创建和状态变更接口支持 `Idempotency-Key`，同键同请求返回同结果，同键不同请求返回 `IDEMPOTENCY_CONFLICT`。
- 状态变更必须同时写 `operation_log` 与 `publish_log`；复制动作至少写 `publish_log`。

## 2. Impact Analysis

| 模块/文件 | 影响程度 | 说明 |
| --- | --- | --- |
| `apps/api-server/internal/modules/publish` | 新增 | 新增发布目标、发布任务、复制载荷、状态动作 DTO 与 Service 接口/骨架。 |
| `apps/api-server/internal/http/handlers/publish.go` | 新增 | 新增 HTTP Handler，负责参数解析、统一响应和 publish 错误映射。 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册发布目标、发布队列、详情、复制、回填和重新入队路由。 |
| `apps/api-server/migrations/00009_create_publish_tables.sql` | 新增 | 新增 `publish_target`、`publish_job`、`publish_log` 表。 |
| `openapi/openapi.yaml` | 修改 | 增加本迭代所有接口 path、schema、响应和 examples。 |
| `apps/web-admin/lib/api.ts` | 修改 | 增加 publish 类型与 API client 函数。 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 修改 | 增加发布队列导航项。 |
| `apps/web-admin/app/projects/[projectId]/publish-jobs/page.tsx` | 新增 | 项目发布队列页面骨架。 |
| `apps/web-admin/app/publish-jobs/[jobId]/page.tsx` | 新增 | 发布任务详情页面骨架。 |
| `apps/web-admin/app/publish-jobs/[jobId]/copy/page.tsx` | 新增 | 复制发布内容页面骨架。 |
| `apps/web-admin/app/publish-jobs/[jobId]/backfill/page.tsx` | 新增 | 手动发布回填页面骨架。 |

### 兼容性分析

- API：新增 `/api/v1` 路由，不修改既有接口请求/响应，向后兼容。
- 数据：新增表与索引，不修改既有表；通过 `content_item`、`content_version`、`operation_log` 的文本 ID 建立业务关联，避免破坏现有迁移。
- 前端：项目工作区导航新增一项，不改变既有路径；新增页面可直接刷新访问。
- 安全：沿用现有 Bearer token 中间件；发布目标 config 增加敏感键校验。

## 3. Flow Design

### 3.1 发布目标维护

1. 前端在创建发布任务弹窗中调用 `GET /projects/{projectId}/publish-targets` 获取目标列表。
2. 用户新建或编辑目标时提交平台、账号名称、展示名称、启用状态、展示配置。
3. Handler 校验请求体和 `Idempotency-Key`；Service 检查必填字段和敏感配置键。
4. Service 创建或更新 `publish_target`，写 `operation_log`，返回目标 ID 与配置摘要。
5. 若 config 包含敏感键或空必填字段，返回 `VALIDATION_ERROR`；同键不同请求返回 `IDEMPOTENCY_CONFLICT`。

### 3.2 发布任务创建与队列查看

1. 前端从已审稿通过内容版本选择 `content_item_id/content_version_id/target_id` 创建发布任务。
2. Service 校验内容版本属于对应内容项，且内容审稿状态为 `approved` 或 `approved_with_edit`。
3. Service 生成 `payload_hash`，创建 `publish_job`，状态为 `queued`，写 `operation_log` 与 `publish_log(job_created)`。
4. 列表接口按 `project_id`、`target_id`、`status`、`scheduled_at`、分页和排序查询，返回内容摘要、目标摘要和可执行操作。

### 3.3 复制发布内容

1. 打开复制页调用 `GET /publish-jobs/{id}/copy-payload`。
2. Service 读取 `publish_job` 绑定的 `content_version_id`，构造 title/body/format/platform/target_id/content_version_id/payload_hash。
3. 预览接口不写日志、不改变状态。
4. 用户点击复制按钮后调用 `POST /publish-jobs/{id}/copy`，提交复制范围 `title/body/full`。
5. `queued` 状态任务变为 `copied`；若已是 `copied`，同幂等请求返回原结果，不重复写日志；非法状态返回 `CONFLICT`。

### 3.4 手动发布回填与失败处理

1. `POST /publish-jobs/{id}/mark-published` 仅允许 `copied -> published`；`external_url` 可空，但为空时 `note` 或 `reason` 必填。
2. `POST /publish-jobs/{id}/mark-failed` 允许 `queued/copied -> failed`；`reason` 必填，记录 `retryable` 与最近失败时间。
3. `POST /publish-jobs/{id}/requeue` 允许 `failed -> queued`，以及 `copied -> queued`；`published` 返回 `CONFLICT`。
4. 每次状态变更都写 `operation_log` 与 `publish_log`，日志失败时状态变更失败。

### 3.5 幂等存储与重放

所有写接口复用现有 `idempotency_record` 表契约：`scope` 使用 `publish:{project_id}` 或 `publish:{job_id}`，`endpoint` 使用规范化动作名，`idempotency_key` 来自请求头，`request_hash` 为规范化请求体哈希。为保证可重放响应不被后续状态变化污染，`response_ref_type/response_ref_id` 必须指向可恢复原始响应的稳定记录，而不是只指向当前可变实体。

每个写接口的重放映射：

| endpoint | scope | response_ref_type | response_ref_id | 重放规则 |
| --- | --- | --- | --- | --- |
| `create_publish_target` | `publish:{project_id}` | `publish_target_operation` | `operation_log_id` | 通过 operation_log metadata 恢复 `target_id/operation_log_id`。 |
| `update_publish_target` | `publish:{target_id}` | `publish_target_operation` | `operation_log_id` | 通过 operation_log metadata 恢复更新响应，不读取目标当前状态。 |
| `create_publish_job` | `publish:{project_id}` | `publish_log` | `publish_log_id` | 通过 `publish_log.payload_snapshot` 恢复 `publish_job_id/status/payload_hash/operation_log_id`。 |
| `copy_publish_payload` | `publish:{job_id}` | `publish_log` | `publish_log_id` | 通过 `publish_log.payload_snapshot` 恢复 `previous_status/current_status/copied_at/publish_log_id`。 |
| `mark_published` | `publish:{job_id}` | `publish_log` | `publish_log_id` | 通过 `publish_log.payload_snapshot` 恢复 `external_url/published_at/operation_log_id/publish_log_id`。 |
| `mark_failed` | `publish:{job_id}` | `publish_log` | `publish_log_id` | 通过 `publish_log.payload_snapshot` 恢复 `failed_at/operation_log_id/publish_log_id`。 |
| `requeue_publish_job` | `publish:{job_id}` | `publish_log` | `publish_log_id` | 通过 `publish_log.payload_snapshot` 恢复 `retry_count/scheduled_at/operation_log_id/publish_log_id`。 |

处理规则：

1. 收到写请求后先按 `scope + endpoint + idempotency_key` 查找记录。
2. 未命中时执行校验、状态变更和日志写入，将完整响应快照写入 `operation_log.metadata` 或 `publish_log.payload_snapshot`，成功后保存请求哈希与响应引用。
3. 命中且 `request_hash` 一致时读取响应快照并返回同一业务结果，不重复写 `operation_log` 或 `publish_log`，也不读取可变实体当前状态拼装响应。
4. 命中但 `request_hash` 不一致时返回 `IDEMPOTENCY_CONFLICT`。

### 3.6 异常流程

- 目标、任务或内容版本不存在：`NOT_FOUND`。
- 请求体字段缺失或 URL 非法：`VALIDATION_ERROR`。
- 状态机不允许动作：`CONFLICT`。
- 幂等键复用但请求体不同：`IDEMPOTENCY_CONFLICT`。
- 日志写入失败或内部持久化失败：`INTERNAL_ERROR`。

## 4. Table Design

### 4.1 `publish_target`

```sql
CREATE TABLE IF NOT EXISTS publish_target (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    platform TEXT NOT NULL,
    account_name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    config_summary TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    operation_log_id TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE(project_id, platform, account_name, display_name)
);

CREATE INDEX IF NOT EXISTS idx_publish_target_project_enabled ON publish_target(project_id, enabled);
```

字段说明：

- `config`：仅保存非敏感展示配置；实现阶段必须拒绝敏感键。
- `config_summary`：前端列表展示摘要，避免直接展开复杂配置。
- `enabled`：创建发布任务时只能选择启用目标。

### 4.2 `publish_job`

```sql
CREATE TABLE IF NOT EXISTS publish_job (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    content_item_id TEXT NOT NULL REFERENCES content_item(id),
    content_version_id TEXT NOT NULL REFERENCES content_version(id),
    target_id TEXT NOT NULL REFERENCES publish_target(id),
    status TEXT NOT NULL CHECK (status IN ('queued', 'copied', 'published', 'failed', 'canceled')),
    payload_hash TEXT NOT NULL,
    scheduled_at TIMESTAMPTZ,
    copied_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    external_url TEXT NOT NULL DEFAULT '',
    last_error TEXT NOT NULL DEFAULT '',
    failed_at TIMESTAMPTZ,
    retryable BOOLEAN NOT NULL DEFAULT FALSE,
    retry_count INTEGER NOT NULL DEFAULT 0 CHECK (retry_count >= 0),
    operation_log_id TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_publish_job_project_status_created ON publish_job(project_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_publish_job_target_status ON publish_job(target_id, status);
CREATE INDEX IF NOT EXISTS idx_publish_job_scheduled ON publish_job(project_id, scheduled_at);
CREATE UNIQUE INDEX IF NOT EXISTS uq_publish_job_version_target_active ON publish_job(content_version_id, target_id) WHERE status IN ('queued', 'copied', 'failed');
```

### 4.3 `publish_log`

```sql
CREATE TABLE IF NOT EXISTS publish_log (
    id TEXT PRIMARY KEY,
    publish_job_id TEXT NOT NULL REFERENCES publish_job(id),
    event_type TEXT NOT NULL CHECK (event_type IN ('job_created', 'payload_copied', 'marked_published', 'marked_failed', 'requeued')),
    from_status TEXT NOT NULL DEFAULT '',
    to_status TEXT NOT NULL DEFAULT '',
    actor_id TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    payload_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    operation_log_id TEXT,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_publish_log_job_created ON publish_log(publish_job_id, created_at DESC);
```

### 4.4 幂等记录

本迭代不新增幂等表，复用 `00008_create_knowledge_memory_tables.sql` 中的 `idempotency_record(scope, endpoint, idempotency_key, request_hash, response_ref_type, response_ref_id)`。发布模块实现阶段必须通过 Service 层写入和读取该表契约，并把写接口的完整响应快照存入 `operation_log.metadata` 或 `publish_log.payload_snapshot`；重放时按 3.5 的映射读取快照，不得用当前实体状态重新组装响应。当前骨架阶段仅在设计、DTO 与测试契约中固化行为。

### 4.5 SQL Contract

- 目标方言：PostgreSQL。
- 关键列表查询模板：

```sql
SELECT j.*, t.platform, t.account_name, t.display_name
FROM publish_job j
JOIN publish_target t ON t.id = j.target_id
WHERE j.project_id = $1
  AND ($2::text IS NULL OR j.target_id = $2)
  AND ($3::text IS NULL OR j.status = $3)
  AND ($4::timestamptz IS NULL OR j.scheduled_at >= $4)
ORDER BY /* whitelisted sort: created_at|scheduled_at|status, order: asc|desc */ j.created_at DESC
LIMIT $5 OFFSET $6;
```

- 禁止模式：禁止无谓跨项目查询；禁止拼接未校验排序字段；禁止将 URL、token、cookie 等敏感值写入 `publish_target.config`。
- 参数规则：排序字段只能从 `created_at/scheduled_at/status` 白名单映射；排序方向只能为 `asc/desc`；分页参数沿用 `content.PaginationRequest`。
- 项目边界规则：创建和查询时必须校验 `publish_job.project_id`、`publish_target.project_id`、`content_version.project_id` 与 `content_item.project_id` 一致；数据库外键保证存在性，Service 层保证跨表项目一致性。

## 5. API Design

所有接口均在 `/api/v1` 下，使用 Bearer token 与统一 Envelope。每个接口都必须声明并测试通用错误响应：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`；只读接口不使用 `IDEMPOTENCY_CONFLICT`，列表接口通常不使用 `NOT_FOUND`，但 OpenAPI 必须在每个 operation 的 responses 中显式列出适用错误。

### 发布目标

#### `GET /projects/{projectId}/publish-targets`

- 查询参数：`enabled`、`page`、`page_size`、`sort`、`order`。
- 200：`PagedPublishTargetsResponse`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`。

#### `POST /projects/{projectId}/publish-targets`

- Header：`Idempotency-Key` 必填。
- 请求：`CreatePublishTargetRequest{platform, account_name, display_name, config, enabled}`。
- 201：`CreatePublishTargetResponse{target_id, operation_log_id}`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`。

#### `PATCH /publish-targets/{id}`

- Header：`Idempotency-Key` 必填。
- 请求：`UpdatePublishTargetRequest{platform, account_name, display_name, config, enabled, note}`。
- 200：`UpdatePublishTargetResponse{target_id, operation_log_id}`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`。

### 发布任务

#### `POST /projects/{projectId}/publish-jobs`

- Header：`Idempotency-Key` 必填。
- 请求：`CreatePublishJobRequest{content_item_id, content_version_id, target_id, scheduled_at}`。
- 201：`CreatePublishJobResponse{publish_job_id, status, payload_hash, operation_log_id}`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`。

#### `GET /projects/{projectId}/publish-jobs`

- 查询参数：`target_id`、`status`、`scheduled_from`、`page`、`page_size`、`sort`、`order`。
- 200：`PagedPublishJobsResponse`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`。

#### `GET /publish-jobs/{id}`

- 200：`PublishJobDetailResponse`，包含 job、target、content_version、logs。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`。

#### `GET /publish-jobs/{id}/copy-payload`

- 200：`PublishCopyPayloadResponse{publish_job_id,title,body,format,platform,target_id,content_version_id,payload_hash}`。
- 不写日志，不改变状态。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`INTERNAL_ERROR`。

#### `POST /publish-jobs/{id}/copy`

- Header：`Idempotency-Key` 必填。
- 请求：`CopyPublishPayloadRequest{copy_scope, note}`，`copy_scope` 为 `title/body/full`。
- 200：`CopyPublishPayloadResponse{publish_job_id, previous_status, current_status, copied_at, publish_log_id}`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`。

#### `POST /publish-jobs/{id}/mark-published`

- Header：`Idempotency-Key` 必填。
- 请求：`MarkPublishedRequest{external_url, published_at, reason, note}`。
- 200：`MarkPublishedResponse{publish_job_id, previous_status, current_status, external_url, published_at, operation_log_id, publish_log_id}`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`。

#### `POST /publish-jobs/{id}/mark-failed`

- Header：`Idempotency-Key` 必填。
- 请求：`MarkFailedRequest{reason, retryable, note}`。
- 200：`MarkFailedResponse{publish_job_id, previous_status, current_status, failed_at, operation_log_id, publish_log_id}`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`。

#### `POST /publish-jobs/{id}/requeue`

- Header：`Idempotency-Key` 必填。
- 请求：`RequeuePublishJobRequest{reason, scheduled_at, note}`。
- 200：`RequeuePublishJobResponse{publish_job_id, previous_status, current_status, retry_count, operation_log_id, publish_log_id}`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`。

## 6. Module Design

### 6.1 后端 publish 模块

包路径：`apps/api-server/internal/modules/publish`

职责：

- 定义发布目标、发布任务、复制载荷、发布日志 DTO。
- 定义状态常量与事件常量。
- 提供 `Service` 接口和 `NewService()` 构造函数。
- 在实现阶段封装状态机、敏感配置校验、轻量平台格式化、payload_hash 计算、幂等记录、日志写入和内容审稿通过校验。

核心接口：

```go
type Service interface {
    ListTargets(ctx context.Context, projectID string, req ListPublishTargetsRequest) (PagedPublishTargetsResponse, error)
    CreateTarget(ctx context.Context, projectID string, req CreatePublishTargetRequest, idempotencyKey string) (CreatePublishTargetResponse, error)
    UpdateTarget(ctx context.Context, id string, req UpdatePublishTargetRequest, idempotencyKey string) (UpdatePublishTargetResponse, error)
    CreateJob(ctx context.Context, projectID string, req CreatePublishJobRequest, idempotencyKey string) (CreatePublishJobResponse, error)
    ListJobs(ctx context.Context, projectID string, req ListPublishJobsRequest) (PagedPublishJobsResponse, error)
    GetJob(ctx context.Context, id string) (PublishJobDetailResponse, error)
    GetCopyPayload(ctx context.Context, id string) (PublishCopyPayloadResponse, error)
    CopyPayload(ctx context.Context, id string, req CopyPublishPayloadRequest, idempotencyKey string) (CopyPublishPayloadResponse, error)
    MarkPublished(ctx context.Context, id string, req MarkPublishedRequest, idempotencyKey string) (MarkPublishedResponse, error)
    MarkFailed(ctx context.Context, id string, req MarkFailedRequest, idempotencyKey string) (MarkFailedResponse, error)
    Requeue(ctx context.Context, id string, req RequeuePublishJobRequest, idempotencyKey string) (RequeuePublishJobResponse, error)
}
```

错误：`ErrValidation`、`ErrNotFound`、`ErrForbidden`、`ErrConflict`、`ErrIdempotencyConflict`、`ErrInternal`。

### 6.2 HTTP Handler

文件：`apps/api-server/internal/http/handlers/publish.go`

职责：

- 从 path/query/header/body 解析参数。
- 调用 `publish.Service`。
- 将 publish 错误映射为统一 `api.ErrorCode`。
- 对创建动作返回 201，对状态动作返回 200，对只读接口返回 200。

### 6.3 Router 集成

在 `NewRouter` 内初始化：

```go
publishHandler := handlers.NewPublishHandler(publish.NewService(), logger)
```

新增路由：

```go
r.Get("/projects/{projectId}/publish-targets", publishHandler.ListTargets)
r.Post("/projects/{projectId}/publish-targets", publishHandler.CreateTarget)
r.Patch("/publish-targets/{id}", publishHandler.UpdateTarget)
r.Post("/projects/{projectId}/publish-jobs", publishHandler.CreateJob)
r.Get("/projects/{projectId}/publish-jobs", publishHandler.ListJobs)
r.Get("/publish-jobs/{id}", publishHandler.GetJob)
r.Get("/publish-jobs/{id}/copy-payload", publishHandler.GetCopyPayload)
r.Post("/publish-jobs/{id}/copy", publishHandler.CopyPayload)
r.Post("/publish-jobs/{id}/mark-published", publishHandler.MarkPublished)
r.Post("/publish-jobs/{id}/mark-failed", publishHandler.MarkFailed)
r.Post("/publish-jobs/{id}/requeue", publishHandler.Requeue)
```

### 6.4 前端模块

- `apps/web-admin/lib/api.ts`：增加 publish 类型、列表/详情/动作函数，所有动作函数接收 `idempotencyKey`。
- `ProjectWorkspaceNav`：增加 `publish-jobs` 导航项，label 为 `发布队列`。
- 发布队列页：负责目标筛选、状态筛选、分页、创建任务弹窗/内嵌表单、错误态展示。
- 详情页：负责从详情接口刷新最新状态，展示日志摘要和操作入口。
- 复制页：负责预览载荷和点击复制动作；只有点击按钮调用 `copy` 接口。
- 回填页：负责标记已发布、标记失败、重新入队三类表单。

## 7. Output Contract

### 7.1 API 与方法契约

响应字段硬约束：发布目标列表必须包含 `platform/account_name/display_name/enabled/config_summary`；队列列表必须包含内容摘要、目标摘要、`status/scheduled_at/copied_at/published_at/last_error/retry_count/actions`；详情必须包含 `content_version_id/payload_hash/external_url/logs(actor_id,event_type,from_status,to_status,reason,note,created_at)`；错误态必须展示 `error.code/error.message/request_id`。

| 产物 | 输入 | 输出 | 正确性规则 | 产出类型 |
| --- | --- | --- | --- | --- |
| `ListTargets` / `GET /projects/{projectId}/publish-targets` | project_id、enabled、分页 | 目标分页列表 | 只返回项目内目标；config 只暴露摘要和非敏感配置 | web-e2e |
| `CreateTarget` / `POST /projects/{projectId}/publish-targets` | target 字段、Idempotency-Key | target_id、operation_log_id | 必填校验、敏感 config 拒绝、幂等冲突检测 | web-e2e |
| `UpdateTarget` / `PATCH /publish-targets/{id}` | target 字段、Idempotency-Key | target_id、operation_log_id | 不允许跨项目更新；敏感 config 拒绝 | web-e2e |
| `CreateJob` / `POST /projects/{projectId}/publish-jobs` | content_item_id、content_version_id、target_id、scheduled_at | publish_job_id、queued、payload_hash | 只允许审稿通过版本；同版本同目标活跃任务冲突 | web-e2e, integration |
| `ListJobs` / `GET /projects/{projectId}/publish-jobs` | filters、分页、排序 | 任务分页列表 | SQL Contract 查询；分页字段正确；状态/目标筛选生效 | web-e2e, sql-query |
| `GetJob` / `GET /publish-jobs/{id}` | job id | 详情、版本、目标、日志 | 必须从详情接口取最新状态；不存在返回 NOT_FOUND | web-e2e, integration |
| `GetCopyPayload` / `GET /publish-jobs/{id}/copy-payload` | job id | 复制载荷 | 不写日志、不改状态；payload_hash 与绑定版本一致 | web-e2e, integration |
| `CopyPayload` / `POST /publish-jobs/{id}/copy` | copy_scope、Idempotency-Key | copied 状态、publish_log_id | queued 可变 copied；重复同键不重复日志 | web-e2e, integration |
| `MarkPublished` / `POST /publish-jobs/{id}/mark-published` | external_url、published_at、reason/note、Idempotency-Key | published 状态、日志 ID | external_url 空时必须有原因；只允许 copied | web-e2e, integration |
| `MarkFailed` / `POST /publish-jobs/{id}/mark-failed` | reason、retryable、note、Idempotency-Key | failed 状态、日志 ID | reason 必填；允许 queued/copied | web-e2e, integration |
| `Requeue` / `POST /publish-jobs/{id}/requeue` | reason、scheduled_at、Idempotency-Key | queued 状态、retry_count | failed/copied 可重入队；published 冲突 | web-e2e, integration |
| Web publish pages | 路由参数、API 响应 | 页面状态与交互反馈 | 导航可达；刷新不 404；错误态展示 code/message/request_id | web-e2e |

### 7.2 testing standard 对应

`workflow.yaml` 当前 `project.features` 为空；本次根据 API Design 和 SQL Contract 显式声明类型化测试；03-test-cases 阶段必须创建 `iteration7_publish_contract_test.go`、`publish_sql_contract_test.go` 与 `iteration7-publish-queue.spec.ts`：

- `web-e2e`：所有 HTTP endpoint 与前端页面链路，引用 `standards/testing/web-e2e.md`。
- `integration`：Handler -> Service -> 状态机 -> 日志结果链路，引用 `standards/testing/integration.md`。
- `sql-query`：发布队列列表和迁移索引契约，引用 `standards/testing/sql-query.md` 与 `standards/sql-guidelines.md`。

跨组件链路：

- `PublishQueue Page -> lib/api.ts -> PublishHandler -> PublishService`
- `PublishHandler -> PublishService -> PublishJob state machine -> publish_log/operation_log`
- `PublishService -> content_version/content_review/content_item contract -> publish_job payload_hash`

## 8. Change Log

| 文件 | 类型 | 原因 |
| --- | --- | --- |
| `.cube/iterations/feature-7/design.md` | 新增 | 记录本迭代技术设计。 |
| `.cube/iterations/feature-7/skeleton-map.yaml` | 新增 | 记录骨架文件与 Development Tasks 对应关系。 |
| `apps/api-server/migrations/00009_create_publish_tables.sql` | 新增 | 增加发布目标、发布任务、发布日志表。 |
| `apps/api-server/internal/modules/publish/dto.go` | 新增 | 定义发布模块 DTO 与状态常量。 |
| `apps/api-server/internal/modules/publish/errors.go` | 新增 | 定义发布模块错误 sentinel。 |
| `apps/api-server/internal/modules/publish/service.go` | 新增 | 定义 Service 接口、构造函数与骨架方法。 |
| `apps/api-server/internal/http/handlers/publish.go` | 新增 | 定义发布 API Handler 骨架与错误映射。 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册发布模块路由。 |
| `openapi/openapi.yaml` | 修改 | 增加发布接口契约。 |
| `apps/web-admin/lib/api.ts` | 修改 | 增加发布 API 类型与调用函数。 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 修改 | 增加发布队列导航入口。 |
| `apps/web-admin/app/projects/[projectId]/publish-jobs/page.tsx` | 新增 | 发布队列页面骨架。 |
| `apps/web-admin/app/publish-jobs/[jobId]/page.tsx` | 新增 | 发布详情页面骨架。 |
| `apps/web-admin/app/publish-jobs/[jobId]/copy/page.tsx` | 新增 | 复制页面骨架。 |
| `apps/web-admin/app/publish-jobs/[jobId]/backfill/page.tsx` | 新增 | 回填页面骨架。 |

## 9. Development Tasks

- Task-01：创建发布数据表与 SQL 查询契约
  - 所属模块：api-server/migrations
  - 简要描述：新增 publish_target、publish_job、publish_log 表、约束、索引与队列列表 SQL Contract。
  - 涉及接口/方法：00009_create_publish_tables.sql
  - 输入：PostgreSQL migration
  - 输出：可迁移的发布表结构
  - 产出类型：sql-query
  - 功能类型：发布队列表结构与查询契约（type id: sql-query）
  - 是否跨组件：否
- Task-02：定义发布模块 DTO、状态常量与服务接口
  - 所属模块：api-server/internal/modules/publish
  - 简要描述：定义发布目标、发布任务、复制载荷、日志、动作请求响应和 Service 接口。
  - 涉及接口/方法：publish.Service、NewService()
  - 输入：发布目标、发布任务和动作请求结构
  - 输出：可被 Handler 和测试引用的公共 DTO 与接口
  - 产出类型：library
  - 功能类型：后端模块公共接口（type id: library）
  - 是否跨组件：否
- Task-03：实现发布服务业务规则与幂等契约
  - 所属模块：api-server/internal/modules/publish
  - 简要描述：实现状态机、敏感配置校验、内容审稿通过校验、payload_hash、轻量平台格式化、幂等重放/冲突和 operation_log/publish_log 写入契约。
  - 涉及接口/方法：CreateTarget()、UpdateTarget()、CreateJob()、CopyPayload()、MarkPublished()、MarkFailed()、Requeue()
  - 输入：Service 请求 DTO、Idempotency-Key、内容版本和发布目标状态
  - 输出：稳定的发布状态、日志 ID、payload_hash、幂等重放结果
  - 产出类型：integration
  - 功能类型：发布服务状态机与幂等链路（type id: integration）
  - 是否跨组件：是（组件链路：PublishService -> content_review/content_version/content_item contract -> idempotency_record -> operation_log/publish_log）
- Task-04：实现发布目标 API 骨架
  - 所属模块：api-server/internal/http
  - 简要描述：新增发布目标列表、创建和编辑 Handler 与路由骨架。
  - 涉及接口/方法：ListTargets()、CreateTarget()、UpdateTarget()
  - 输入：HTTP path/query/body/header
  - 输出：统一 Envelope 响应或错误结构
  - 产出类型：web-e2e
  - 功能类型：Web/API 发布目标维护（type id: web-e2e）
  - 是否跨组件：是（组件链路：HTTP Router -> PublishHandler -> PublishService）
- Task-05：实现发布任务创建与队列 API 骨架
  - 所属模块：api-server/internal/http
  - 简要描述：新增创建发布任务、发布队列列表和发布详情 Handler 与路由骨架。
  - 涉及接口/方法：CreateJob()、ListJobs()、GetJob()
  - 输入：project_id、target/status/scheduled filters、content_item_id、content_version_id、target_id
  - 输出：发布任务 ID、队列分页、详情响应
  - 产出类型：integration
  - 功能类型：发布任务 HTTP 到服务链路（type id: integration）
  - 是否跨组件：是（组件链路：HTTP Router -> PublishHandler -> PublishService -> Content/Review contract）
- Task-06：实现复制发布载荷 API 骨架
  - 所属模块：api-server/internal/http
  - 简要描述：新增复制预览和点击复制动作 Handler 与路由骨架。
  - 涉及接口/方法：GetCopyPayload()、CopyPayload()
  - 输入：publish_job_id、copy_scope、Idempotency-Key
  - 输出：复制载荷或 copied 状态动作响应
  - 产出类型：integration
  - 功能类型：复制动作状态链路（type id: integration）
  - 是否跨组件：是（组件链路：HTTP Router -> PublishHandler -> PublishService -> PublishJob state machine -> publish_log）
- Task-07：实现发布回填、失败与重新入队 API 骨架
  - 所属模块：api-server/internal/http
  - 简要描述：新增标记已发布、标记失败和重新入队 Handler 与路由骨架。
  - 涉及接口/方法：MarkPublished()、MarkFailed()、Requeue()
  - 输入：external_url、published_at、reason、retryable、scheduled_at、Idempotency-Key
  - 输出：状态变更响应、operation_log_id、publish_log_id
  - 产出类型：integration
  - 功能类型：发布状态机回填链路（type id: integration）
  - 是否跨组件：是（组件链路：HTTP Router -> PublishHandler -> PublishService -> PublishJob state machine -> operation_log/publish_log）
- Task-08：补充 OpenAPI 发布接口契约
  - 所属模块：openapi
  - 简要描述：为本迭代所有发布接口增加 path、summary、description、tags、operationId、parameters、schema、requestBody、responses、security 和 examples。
  - 涉及接口/方法：openapi/openapi.yaml
  - 输入：API Design
  - 输出：OpenAPI 3.0 发布接口描述
  - 产出类型：web-e2e
  - 功能类型：Web/API 契约文档（type id: web-e2e）
  - 是否跨组件：否
- Task-09：实现前端发布 API client 骨架
  - 所属模块：web-admin/lib
  - 简要描述：增加发布目标、发布任务、复制和回填 API 类型与函数。
  - 涉及接口/方法：fetchPublishTargets()、createPublishTarget()、updatePublishTarget()、fetchPublishJobs()、createPublishJob()、fetchPublishJob()、fetchPublishCopyPayload()、copyPublishPayload()、markPublishJobPublished()、markPublishJobFailed()、requeuePublishJob()
  - 输入：页面参数和表单数据
  - 输出：APIEnvelope 包装的发布响应
  - 产出类型：library
  - 功能类型：前端 API client（type id: library）
  - 是否跨组件：否
- Task-10：实现发布队列导航与列表页面骨架
  - 所属模块：web-admin/app
  - 简要描述：新增项目工作区发布队列导航入口和发布队列页面骨架。
  - 涉及接口/方法：ProjectWorkspaceNav、PublishJobsPage
  - 输入：projectId、筛选和分页状态
  - 输出：发布队列页面、空态/加载态/错误态位置
  - 产出类型：web-e2e
  - 功能类型：前端 Web 页面（type id: web-e2e）
  - 是否跨组件：是（组件链路：PublishQueue Page -> lib/api.ts -> PublishHandler -> PublishService）
- Task-11：实现发布详情、复制与回填页面骨架
  - 所属模块：web-admin/app
  - 简要描述：新增发布详情、复制发布内容、手动发布回填页面骨架与操作入口。
  - 涉及接口/方法：PublishJobDetailPage、PublishJobCopyPage、PublishJobBackfillPage
  - 输入：jobId、复制/回填表单数据
  - 输出：详情、复制和回填页面结构
  - 产出类型：web-e2e
  - 功能类型：前端 Web 页面（type id: web-e2e）
  - 是否跨组件：是（组件链路：Publish pages -> lib/api.ts -> PublishHandler -> PublishService）
