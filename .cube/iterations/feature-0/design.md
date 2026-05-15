# Iteration 0 Design：项目脚手架与基础工程

## 1. 概述

本次设计为 AI Content Factory 建立最小可运行工程底座：Go API Server 提供 `/api/v1` 基础系统接口，PostgreSQL/goose 迁移建立操作日志基线，Next.js 管理台壳层预留系统检查页面，CI 基线验证 Go 构建和测试。

核心原则：

- Go 后端统一采用轻量分层：`cmd/api` → `internal/app` → `internal/http` → `internal/modules/system` → `internal/store`。
- API 响应统一使用 `success/data/error/request_id` 信封结构。
- Core 命名保持内容类型无关，不引入 Novel / Book / Chapter 作为基础资源。
- 设计阶段仅生成接口骨架，业务实现留到 04-development。
- GitHub 复用检索因本地 `gh` 不存在且 GitHub MCP fetch 失败未完成；库行为参考 Context7 的 chi v5 文档和项目需求文档。

## 2. Impact Analysis

| 模块 / 文件范围 | 类型 | 影响程度 | 说明 |
|---|---|---:|---|
| `go.mod` | 新增 | 高 | 初始化 Go module，引入 chi、pgx、validator 等后续实现依赖。 |
| `apps/api-server/cmd/api/main.go` | 新增 | 高 | API Server 入口，负责装配配置、日志、路由与 HTTP Server。 |
| `apps/api-server/internal/app` | 新增 | 高 | 应用装配层，暴露 `NewServer`。 |
| `apps/api-server/internal/config` | 新增 | 中 | 配置结构与加载入口骨架。 |
| `apps/api-server/internal/http` | 新增 | 高 | Router、统一响应、错误码、系统 handler 骨架。 |
| `apps/api-server/internal/modules/system` | 新增 | 高 | 系统健康、信息、配置、数据库和迁移状态服务接口。 |
| `apps/api-server/internal/store` | 新增 | 中 | PostgreSQL health checker 和 migration reader 接口骨架。 |
| `apps/api-server/internal/worker` | 新增 | 中 | 异步任务入队和 job_id 返回契约骨架。 |
| `apps/api-server/migrations` | 新增 | 中 | goose 迁移基线，建立 `operation_log`。 |
| `apps/web-admin` | 新增 | 中 | Next.js 管理台壳层页面与 API client 骨架。 |
| `openapi/openapi.yaml` | 新增 | 中 | 本迭代接口 OpenAPI 占位与契约。 |
| `.github/workflows/ci.yml` | 新增 | 中 | CI 基线。 |

### 对现有接口的兼容性分析

当前仓库无既有业务接口，本次新增 `/api/v1` 下系统接口，不存在向后兼容风险。

### 对现有数据的兼容性分析

当前仓库无既有数据库迁移。本次新增 `operation_log` 表，不修改存量表，无数据兼容风险。

## 3. Flow Design

### 3.1 API 请求流程

1. 客户端请求 `/api/v1/system/*` 或 `/api/v1/health`。
2. Chi Router 通过中间件生成或透传 `request_id`。
3. Handler 调用 `system.Service`。
4. Service 调用配置检查、数据库检查或迁移状态接口。
5. Handler 通过统一 response writer 输出成功或失败信封。

异常流程：

- 请求处理 panic：Recoverer 捕获并返回 `INTERNAL_ERROR`。
- 配置检查缺失：作为业务数据返回，不视为 HTTP 错误。
- 数据库不可连接：返回非 2xx 状态和 `INTERNAL_ERROR` / 依赖错误信息。
- 迁移状态不可读取：返回 `INTERNAL_ERROR`。

### 3.2 前端页面流程

1. 用户打开管理台默认页。
2. 页面并行请求健康、系统信息、配置检查、数据库检查、迁移状态接口。
3. 页面展示加载态，接口成功后展示数据卡片。
4. 任一接口失败时展示错误码、错误信息和 request_id，不阻断其他卡片展示。
5. Swagger / OpenAPI 页面提供 OpenAPI 入口；文档不可用时展示占位错误态。

### 3.3 迁移与操作日志基线

1. goose 执行迁移创建 `operation_log` 表。
2. 后续状态变更操作通过 `operation_log` 记录操作对象、动作、操作者、原因、request_id 和时间。
3. 本迭代仅建立表结构和写入接口骨架，不实现业务状态变更。

### 3.4 异步任务基线

1. 后续异步动作通过 `worker.Queue` 接口入队。
2. `worker.NewMemoryQueue()` 返回本迭代的内存队列骨架，供测试和后续 HTTP 适配层装配使用。
3. 入队成功立即返回 `job_id`，HTTP 层不等待最终执行结果。
4. 入队失败返回统一错误响应，错误码为 `INTERNAL_ERROR`，错误信息明确表示任务无法创建。
5. 本迭代仅定义队列接口、任务 DTO 和内存占位实现骨架，不实现外部 Redis/asynq 运行时。

## 4. Table Design

### 4.1 `operation_log`

```sql
CREATE TABLE operation_log (
    id BIGSERIAL PRIMARY KEY,
    request_id TEXT NOT NULL,
    actor_id TEXT,
    actor_type TEXT NOT NULL DEFAULT 'system',
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    reason TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_operation_log_resource ON operation_log(resource_type, resource_id);
CREATE INDEX idx_operation_log_created_at ON operation_log(created_at DESC);
CREATE INDEX idx_operation_log_request_id ON operation_log(request_id);
```

字段说明：

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| `id` | BIGSERIAL | PK | 日志 ID。 |
| `request_id` | TEXT | NOT NULL | 关联请求 ID。 |
| `actor_id` | TEXT | NULL | 操作者 ID，系统操作可为空。 |
| `actor_type` | TEXT | NOT NULL | 操作者类型，默认 `system`。 |
| `action` | TEXT | NOT NULL | 操作动作。 |
| `resource_type` | TEXT | NOT NULL | 资源类型，Core 命名保持通用。 |
| `resource_id` | TEXT | NULL | 资源 ID。 |
| `reason` | TEXT | NULL | 操作原因或备注。 |
| `metadata` | JSONB | NOT NULL | 扩展上下文。 |
| `created_at` | TIMESTAMPTZ | NOT NULL | 创建时间。 |

SQL Contract：

- 目标方言：PostgreSQL 14+。
- 迁移工具：goose。
- 必须结构：`operation_log` 表、resource 组合索引、created_at 倒序索引、request_id 索引。
- 禁止模式：不得使用 Novel / Book / Chapter 作为 `resource_type` 固定枚举；不得记录凭证或密钥到 `metadata`。
- 典型输出：执行 up 迁移后 `operation_log` 表存在，执行 down 迁移后表删除。

## 5. API Design

公共约定：所有响应使用 API 信封；所有错误响应包含 `code`、`message`、`details`；所有接口支持可选 `X-Request-Id`，响应返回同一 `request_id` 或服务端生成值。

### GET `/api/v1/health`

- Query：无
- Headers：`X-Request-Id` 可选
- 成功响应 data：
  - `status: string`
  - `service: string`
  - `version: string`
  - `timestamp: string`
- 错误码：`INTERNAL_ERROR`

### GET `/api/v1/system/info`

- Query：无
- 成功响应 data：
  - `app_name: string`
  - `environment: string`
  - `build_commit: string`
- 错误码：`INTERNAL_ERROR`

### GET `/api/v1/system/config-check`

- Query：无
- 成功响应 data：
  - `items: []ConfigCheckItem`
  - `summary: ConfigCheckSummary`
- `ConfigCheckItem`：`key`、`required`、`configured`、`status`
- 错误码：`INTERNAL_ERROR`

### GET `/api/v1/system/db-check`

- Query：无
- 成功响应 data：
  - `database: string`
  - `status: string`
  - `latency_ms: number`
- 错误码：`INTERNAL_ERROR`

### GET `/api/v1/system/migration-status`

- Query：无
- 成功响应 data：
  - `applied_migrations: []MigrationInfo`
  - `pending_migrations: []MigrationInfo`
- `MigrationInfo`：`version`、`name`、`applied_at`
- 错误码：`INTERNAL_ERROR`

### GET `/openapi.yaml`

- Query：无
- 成功响应：OpenAPI YAML 文档
- 错误码：`NOT_FOUND`、`INTERNAL_ERROR`
- 责任归属：`internal/http/router.go` 通过静态文件 handler 暴露；OpenAPI 文件不存在时返回 `NOT_FOUND`。

### 内部异步任务契约

- `worker.Queue.Enqueue(ctx, TaskRequest) (TaskReceipt, error)`
- 输入：任务类型、payload、request_id。
- 输出：`job_id`、任务状态和创建时间。
- 错误码映射：入队失败映射为 `QUEUE_ENQUEUE_FAILED`，兜底为 `INTERNAL_ERROR`。
- 本迭代异步任务的规范化返回 ID 为 `job_id`；`run_id` 保留给后续 WorkflowRun 类异步动作。

- `GET /api/v1/system/db-check` 在数据库不可用时返回 `DEPENDENCY_UNAVAILABLE`，其他未预期错误返回 `INTERNAL_ERROR`。
- `GET /api/v1/system/migration-status` 在迁移状态读取失败时返回 `MIGRATION_READ_FAILED`，其他未预期错误返回 `INTERNAL_ERROR`。
- `worker.Queue.Enqueue` 入队失败映射为 `QUEUE_ENQUEUE_FAILED`。
- JSON 编码失败映射为 `SERIALIZATION_FAILED`。

## 6. Module Design

### 6.1 `internal/http/api`

职责：定义统一响应 DTO、错误 DTO、分页 DTO、错误码常量和 JSON writer。

公共接口：

- `WriteSuccess(w http.ResponseWriter, r *http.Request, status int, data any)`
  - 输入：ResponseWriter、Request、HTTP status、业务数据。
  - 输出：统一成功响应。
  - 异常：编码失败时写 `INTERNAL_ERROR`。
- `WriteError(w http.ResponseWriter, r *http.Request, status int, code ErrorCode, message string, details []ErrorDetail)`
  - 输入：错误状态、错误码、信息和详情。
  - 输出：统一失败响应。

### 6.2 `internal/modules/system`

职责：封装系统检查业务能力。

接口：

```go
type Service interface {
    Health(ctx context.Context) (HealthResponse, error)
    Info(ctx context.Context) (InfoResponse, error)
    ConfigCheck(ctx context.Context) (ConfigCheckResponse, error)
    DBCheck(ctx context.Context) (DBCheckResponse, error)
    MigrationStatus(ctx context.Context) (MigrationStatusResponse, error)
}
```

依赖：`config.Provider`、`store.DBChecker`、`store.MigrationReader`。

### 6.3 `internal/http/handlers`

职责：将 HTTP 请求映射到 system service，并使用统一 response writer。

接口：

- `SystemHandler.Health(w http.ResponseWriter, r *http.Request)`
- `SystemHandler.Info(w http.ResponseWriter, r *http.Request)`
- `SystemHandler.ConfigCheck(w http.ResponseWriter, r *http.Request)`
- `SystemHandler.DBCheck(w http.ResponseWriter, r *http.Request)`
- `SystemHandler.MigrationStatus(w http.ResponseWriter, r *http.Request)`

### 6.4 `internal/store`

职责：定义数据库检查与迁移状态读取边界。

接口：

- `DBChecker.Check(ctx context.Context) (DBCheckResult, error)`
- `MigrationReader.Status(ctx context.Context) (MigrationStatusResult, error)`
- `OperationLogger.Log(ctx context.Context, entry OperationLogEntry) error`

### 6.6 `internal/worker`

职责：定义异步任务入队边界和 job_id 返回契约。

接口：

- `worker.NewMemoryQueue() Queue`
  - 输入：无。
  - 输出：内存队列骨架实例，用于本迭代测试和后续装配。
  - 异常：无。
  - 测试要求：03 阶段 Task-07 测试必须通过该生产入口实例化队列，不得在测试文件内自定义绕过生产代码的 `Queue` 实现。
- `Queue.Enqueue(ctx context.Context, req TaskRequest) (TaskReceipt, error)`
  - 输入：任务类型、payload、request_id。
  - 输出：job_id、状态和创建时间。
  - 异常：空任务类型返回 error；入队失败返回 error，由 HTTP 层映射为 `INTERNAL_ERROR`。

### 6.7 `apps/web-admin`

职责：Next.js 管理台壳层。

页面：

- `/`：系统默认页 / 健康检查页。
- `/swagger-openapi`：OpenAPI 入口页。
- `/system/config-check`：系统配置检查页。

### 6.8 `.github/workflows/ci.yml`

职责：执行基础 CI。

检查项：Go build、Go test；前端目录存在 package 配置后执行前端检查。

## 7. Output Contract

workflow.yaml 当前 `project.features` 为空；本迭代仍包含 HTTP API、SQL 迁移、异步任务契约和前端页面，因此声明以下测试标准：

### 7.1 API 与公共方法契约

| 对象 | 输入 | 输出 | 产出类型 | 正确性规则 | 测试规范 |
|---|---|---|---|---|---|
| `cmd/api main` / `app.NewServer` / `http.NewRouter` | 配置、logger、service 依赖 | 可启动 server / router | `integration` | 路由注册完整；启动不执行外部业务；缺依赖时返回 error | `standards/testing/integration.md` |
| `SystemHandler.*` | ResponseWriter、Request | API 信封响应 | `web-e2e` | handler 只负责 HTTP 映射；service error 映射到明确错误码 | `standards/testing/web-e2e.md` |
| `GET /api/v1/health` | HTTP GET、可选 `X-Request-Id` | `HealthResponse` API 信封 | `web-e2e` | 返回 2xx；包含 status、service、version、timestamp；request_id 可追踪 | `standards/testing/web-e2e.md` |
| `GET /api/v1/system/info` | HTTP GET、可选 `X-Request-Id` | `InfoResponse` API 信封 | `web-e2e` | 返回 app_name、environment、build_commit；缺失 build_commit 以空值呈现 | `standards/testing/web-e2e.md` |
| `GET /api/v1/system/config-check` | HTTP GET | `ConfigCheckResponse` API 信封 | `web-e2e` | 配置缺失仍返回成功；items 标明 configured/status；异常才返回错误信封 | `standards/testing/web-e2e.md` |
| `GET /api/v1/system/db-check` | HTTP GET | `DBCheckResponse` API 信封 | `integration` | 成功返回 database/status/latency_ms；数据库不可用返回错误信封 | `standards/testing/integration.md` |
| `GET /api/v1/system/migration-status` | HTTP GET | `MigrationStatusResponse` API 信封 | `integration` | applied/pending 均为列表；无 pending 返回空列表；读取失败返回错误信封 | `standards/testing/integration.md` |
| `GET /openapi.yaml` | HTTP GET | OpenAPI YAML | `web-e2e` | 文件存在返回 YAML；文件缺失返回 `NOT_FOUND` | `standards/testing/web-e2e.md` |
| `api.WriteSuccess` | writer、request、status、data | 成功 API 信封 | `library` | `success=true`、`error=null`、request_id 透传或生成 | `standards/testing/library.md` |
| `api.WriteError` | writer、request、status、code、message、details | 失败 API 信封 | `library` | `success=false`、`data=null`、HTTP 状态非 2xx、错误码不为空 | `standards/testing/library.md` |
| `system.Service.Health` | context | `HealthResponse` | `library` | 不访问外部依赖；时间戳为当前时间 | `standards/testing/library.md` |
| `system.Service.Info` | context | `InfoResponse` | `library` | 读取配置中的 app/environment/build_commit | `standards/testing/library.md` |
| `system.Service.ConfigCheck` | context | `ConfigCheckResponse` | `library` | required 配置逐项检查，不泄露配置值 | `standards/testing/library.md` |
| `system.Service.DBCheck` | context | `DBCheckResponse` 或 error | `integration` | 调用 DBChecker；保留依赖错误供 handler 映射 | `standards/testing/integration.md` |
| `system.Service.MigrationStatus` | context | `MigrationStatusResponse` 或 error | `integration` | 调用 MigrationReader；保持 applied/pending 顺序稳定 | `standards/testing/integration.md` |
| `store.DBChecker.Check` | context | `DBCheckResult` 或 error | `integration` | 使用 context 控制超时；返回 latency_ms | `standards/testing/integration.md` |
| `store.MigrationReader.Status` | context | `MigrationStatusResult` 或 error | `integration` | 返回已应用和待应用迁移列表 | `standards/testing/integration.md` |
| `store.OperationLogger.Log` | context、`OperationLogEntry` | error | `sql-query` | 写入 operation_log；不得写入 secret；request_id/action/resource_type 必填 | `standards/testing/sql-query.md` |
| `worker.NewMemoryQueue` | 无 | `Queue` | `batch-job` | 返回可调用的内存队列骨架；03 阶段测试可直接实例化生产入口 | `standards/testing/batch-job.md` |
| `worker.Queue.Enqueue` | context、`TaskRequest` | `TaskReceipt` 或 error | `batch-job` | 成功立即返回 job_id；不等待任务完成；空 task type 返回 error；失败返回 error | `standards/testing/batch-job.md` |
| `web-admin/lib/api.ts` functions | 浏览器 fetch 请求 | typed result 或 typed error | `web-e2e` | 失败保留 code/message/request_id；页面可展示错误态 | `standards/testing/web-e2e.md` |

### 7.2 产出类别与集成链路

| 产出 | 类型 | type id | 跨组件 | 组件链路 | 测试规范 |
|---|---|---|---|---|---|
| 系统检查 HTTP API | Web/API | `web-e2e` | 是 | Router -> Handler -> SystemService -> Store/Config -> Response | `standards/testing/web-e2e.md`、`standards/testing/integration.md` |
| `operation_log` 迁移 | SQL/query | `sql-query` | 否 | goose migration -> PostgreSQL schema | `standards/testing/sql-query.md` |
| 异步任务入队契约 | Batch/job | `batch-job` | 否 | Queue.Enqueue -> TaskReceipt | `standards/testing/batch-job.md` |
| Next.js 管理台壳层 | Web/UI | `web-e2e` | 是 | Page -> API client -> HTTP API -> UI state | `standards/testing/web-e2e.md` |
| CI 基线 | CLI | `cli` | 否 | workflow command -> go toolchain | `standards/testing/cli.md` |

### 7.3 SQL Contract

- 目标方言：PostgreSQL 14+。
- expected 结构：`operation_log` 表包含主键、request_id、actor、action、resource、reason、metadata、created_at 字段。
- required indexes：`(resource_type, resource_id)`、`created_at DESC`、`request_id`。
- 查询示例：按 request_id 查询操作日志时使用 `WHERE request_id = $1 ORDER BY created_at DESC`；按资源查询时使用 `WHERE resource_type = $1 AND resource_id = $2 ORDER BY created_at DESC`。
- 禁止模式：不得拼接未绑定用户输入；不得在 metadata 写入凭证；不得把 Novel/Book/Chapter 固定为 Core resource_type；不得无过滤条件扫描 operation_log 作为在线接口默认路径。

### 7.4 前端页面到 API 映射

| 页面 | API | 成功态 | 错误态 |
|---|---|---|---|
| `/` | `/api/v1/health`、`/api/v1/system/info`、`/api/v1/system/db-check`、`/api/v1/system/migration-status` | 展示服务、系统、数据库、迁移卡片 | 卡片内展示 code/message/request_id |
| `/swagger-openapi` | `/openapi.yaml` | 展示文档入口或 YAML 链接 | 展示文档不可用说明 |
| `/system/config-check` | `/api/v1/system/config-check` | 展示配置项列表和 summary | 展示 code/message/request_id |

### 7.5 命名约束检查点

- Core 目录、DTO、接口和表名不得使用 Novel / Book / Chapter 作为基础资源名。
- 原型或后续内容包占位文案不得进入 Core API path、Go package 或数据库表名。

## 8. Change Log

| 文件 | 类型 | 原因 |
|---|---|---|
| `go.mod` | 新增 | 初始化 Go module 和后端依赖。 |
| `apps/api-server/cmd/api/main.go` | 新增 | API Server 入口骨架。 |
| `apps/api-server/internal/app/server.go` | 新增 | HTTP server 装配骨架。 |
| `apps/api-server/internal/config/config.go` | 新增 | 配置结构和加载接口骨架。 |
| `apps/api-server/internal/http/api/response.go` | 新增 | 统一响应 DTO 和 writer 骨架。 |
| `apps/api-server/internal/http/router.go` | 新增 | Chi 路由装配骨架。 |
| `apps/api-server/internal/http/handlers/system.go` | 新增 | 系统接口 handler 骨架。 |
| `apps/api-server/internal/modules/system/dto.go` | 新增 | 系统接口 DTO。 |
| `apps/api-server/internal/modules/system/service.go` | 新增 | 系统 service 接口与占位实现骨架。 |
| `apps/api-server/internal/store/store.go` | 新增 | DB checker、migration reader、operation logger 接口。 |
| `apps/api-server/internal/worker/queue.go` | 新增 | 异步任务队列接口和 job_id 返回契约骨架。 |
| `apps/api-server/migrations/00001_create_operation_log.sql` | 新增 | operation_log 迁移基线。 |
| `openapi/openapi.yaml` | 新增 | 本迭代 OpenAPI 契约占位。 |
| `apps/web-admin/package.json` | 新增 | Next.js 管理台包定义骨架。 |
| `apps/web-admin/app/page.tsx` | 新增 | 默认页 / 健康检查页骨架。 |
| `apps/web-admin/app/swagger-openapi/page.tsx` | 新增 | OpenAPI 入口页骨架。 |
| `apps/web-admin/app/system/config-check/page.tsx` | 新增 | 配置检查页骨架。 |
| `apps/web-admin/lib/api.ts` | 新增 | API client DTO 和函数签名骨架。 |
| `.github/workflows/ci.yml` | 新增 | CI 基线。 |
| `.cube/iterations/feature-0/skeleton-map.yaml` | 新增 | 骨架到任务的映射。 |

## 9. Development Tasks

- Task-01：初始化 Go API Server 工程与启动入口
  - 所属模块：api-server/app
  - 简要描述：建立 Go module、API Server 入口、应用装配和基础路由。
  - 涉及接口/方法：`main()`、`app.NewServer()`、`http.NewRouter()`
  - 输入：配置和依赖实例
  - 输出：可编译的 HTTP server 骨架
  - 产出类型：web-e2e
  - 功能类型：Web/API 服务入口（type id: web-e2e）
  - 是否跨组件：是（组件链路：cmd/api -> app -> router）
- Task-02：实现统一 API 响应契约骨架
  - 所属模块：http/api
  - 简要描述：定义成功响应、错误响应、错误码、request_id 处理和 JSON writer。
  - 涉及接口/方法：`WriteSuccess()`、`WriteError()`、`RequestID()`
  - 输入：HTTP request、status、业务数据或错误信息
  - 输出：统一 API 响应信封
  - 产出类型：library
  - 功能类型：公共库 / 响应契约（type id: library）
  - 是否跨组件：否
- Task-03：实现系统健康与信息接口骨架
  - 所属模块：system/http
  - 简要描述：提供 `/api/v1/health` 和 `/api/v1/system/info` 的 handler、DTO 和 service 方法。
  - 涉及接口/方法：`SystemHandler.Health()`、`SystemHandler.Info()`、`Service.Health()`、`Service.Info()`
  - 输入：HTTP GET 请求
  - 输出：健康状态和系统信息响应
  - 产出类型：web-e2e
  - 功能类型：Web/API endpoint（type id: web-e2e）
  - 是否跨组件：是（组件链路：Router -> Handler -> SystemService -> Response）
- Task-04：实现系统配置检查接口骨架
  - 所属模块：system/config
  - 简要描述：提供 `/api/v1/system/config-check` 的 handler、DTO 和 service 方法。
  - 涉及接口/方法：`SystemHandler.ConfigCheck()`、`Service.ConfigCheck()`
  - 输入：HTTP GET 请求和配置 provider
  - 输出：配置项检查结果
  - 产出类型：web-e2e
  - 功能类型：Web/API endpoint（type id: web-e2e）
  - 是否跨组件：是（组件链路：Router -> Handler -> SystemService -> ConfigProvider -> Response）
- Task-05：实现数据库与迁移状态检查接口骨架
  - 所属模块：system/store
  - 简要描述：提供 `/api/v1/system/db-check` 和 `/api/v1/system/migration-status` 的 handler、DTO、service 和 store 接口。
  - 涉及接口/方法：`SystemHandler.DBCheck()`、`SystemHandler.MigrationStatus()`、`DBChecker.Check()`、`MigrationReader.Status()`
  - 输入：HTTP GET 请求和数据库 / 迁移读取接口
  - 输出：数据库连接状态、延迟、迁移状态
  - 产出类型：integration
  - 功能类型：跨组件集成（type id: integration）
  - 是否跨组件：是（组件链路：Router -> Handler -> SystemService -> Store -> Response）
- Task-06：建立 operation_log 迁移与操作日志接口骨架
  - 所属模块：store/migrations
  - 简要描述：创建 `operation_log` 迁移和 `OperationLogger` 接口，为后续状态变更日志提供基线。
  - 涉及接口/方法：`OperationLogger.Log()`
  - 输入：`OperationLogEntry`
  - 输出：操作日志写入结果
  - 产出类型：sql-query
  - 功能类型：SQL migration / query contract（type id: sql-query）
  - 是否跨组件：否
- Task-07：建立异步任务队列契约骨架
  - 所属模块：worker
  - 简要描述：定义 `Queue.Enqueue`、`NewMemoryQueue`、`TaskRequest`、`TaskReceipt`，确保后续异步动作可通过生产入口实例化队列并立即返回 job_id。
  - 涉及接口/方法：`NewMemoryQueue()`、`Queue.Enqueue()`
  - 输入：context 和 `TaskRequest`
  - 输出：`Queue`、`TaskReceipt` 或 error
  - 产出类型：batch-job
  - 功能类型：批处理 / 异步任务契约（type id: batch-job）
  - 是否跨组件：否
- Task-08：建立 OpenAPI 文档入口骨架
  - 所属模块：openapi
  - 简要描述：提供本迭代 API 的 OpenAPI YAML 占位和静态文档路由。
  - 涉及接口/方法：`GET /openapi.yaml`
  - 输入：HTTP GET 请求
  - 输出：OpenAPI YAML 文档
  - 产出类型：web-e2e
  - 功能类型：Web/API endpoint（type id: web-e2e）
  - 是否跨组件：是（组件链路：Router -> StaticFile -> Response）
- Task-09：建立 Next.js 管理台壳层骨架
  - 所属模块：web-admin
  - 简要描述：提供默认页、OpenAPI 入口页、配置检查页和 API client 函数签名。
  - 涉及接口/方法：`fetchHealth()`、`fetchSystemInfo()`、`fetchConfigCheck()`、页面组件
  - 输入：浏览器访问和 API 响应
  - 输出：页面加载态、成功态、错误态、空态骨架
  - 产出类型：web-e2e
  - 功能类型：Web/UI + API client（type id: web-e2e）
  - 是否跨组件：是（组件链路：Page -> API client -> HTTP API -> UI state）
- Task-10：建立 CI 基线
  - 所属模块：ci
  - 简要描述：提供 GitHub Actions 工作流，执行 Go 构建和测试，前端检查作为后续扩展。
  - 涉及接口/方法：`.github/workflows/ci.yml`
  - 输入：代码提交或 CI 触发
  - 输出：构建 / 测试结果
  - 产出类型：cli
  - 功能类型：CLI / CI command（type id: cli）
  - 是否跨组件：否

## 10. 技术选型说明

- HTTP Router：Chi v5。理由：轻量、标准库友好，Context7 文档展示了 RequestID、Recoverer、Timeout 和 Route group 的常规用法。
- 数据库访问：pgx。理由：符合蓝图中 PostgreSQL + pgx 的方向。
- 迁移：goose。理由：PRD 明确要求 goose 或等价工具，蓝图推荐 goose。
- 日志：slog。理由：Go 标准库能力，满足结构化日志基线。
- 前端：Next.js + TypeScript。理由：需求文档明确指定。

## 11. 安全设计

- Bearer Token 鉴权在本迭代作为契约占位，业务强制鉴权留到后续认证迭代。
- 不在 GET 参数中传递敏感信息。
- 不在日志、OpenAPI 示例或页面中输出真实密钥。
- `operation_log.metadata` 禁止记录凭证、Token 或密钥。
