# Iteration 1：通用内容项目入口技术设计

## 1. 概述

本次设计在现有 Iteration 0 的 Go API Server 与 Next.js Web Admin 壳层上，新增通用内容项目入口能力：系统大盘、内容类型 / 项目模板、内容项目、Prompt 模板、LLM Provider 管理，以及对应前端页面入口。

整体方案采用现有分层风格：`router -> handler -> module service -> DTO/model`。当前工程尚未接入真实 PostgreSQL，因此 02 阶段只生成可编译接口骨架，04 阶段按测试契约补齐内存实现与可替换 repository 接口；数据库持久化通过 migration 先固化表结构，不在本迭代引入外部运行依赖。

核心约束：

- Core 命名保持内容类型无关，不出现 Novel / Book / Chapter 作为核心资源。
- 所有 HTTP 响应继续使用 `success/data/error/request_id` envelope。
- Provider API Key 只能存储于服务边界内部，所有对外 DTO 只暴露 `api_key_masked`。
- 状态变更必须返回 `operation_log_id`，并在服务层通过 operation log 接口留痕。
- 骨架代码只包含类型、接口、路由和方法签名，不实现业务逻辑。

## 2. Impact Analysis

| 模块 / 文件 | 影响程度 | 说明 |
|---|---|---|
| `apps/api-server/internal/http/api/response.go` | 修改 | 补充 `FORBIDDEN`、`CONFLICT` 等 Iteration 1 需要的错误码常量。 |
| `apps/api-server/internal/http/router.go` | 修改 | 挂载 dashboard、content-types、projects、prompt-templates、llm-providers 路由；保持 `NewRouter(system.Service, logger)` 签名不变，避免影响现有测试和调用方。 |
| `apps/api-server/internal/http/handlers/*` | 新增 | 为新增模块提供 HTTP handler 骨架。 |
| `apps/api-server/internal/modules/dashboard` | 新增 | 系统大盘摘要 DTO 与 service 接口。 |
| `apps/api-server/internal/modules/content` | 新增 | ContentType、ContentProject、ProjectOverview、OperationLog 相关 DTO、模型与 service 接口。 |
| `apps/api-server/internal/modules/prompt` | 新增 | PromptTemplate DTO、模型与 service 接口。 |
| `apps/api-server/internal/modules/llm` | 新增 | LLM Provider DTO、模型与 service 接口，包含脱敏响应契约。 |
| `apps/api-server/migrations/00002_create_content_entry_tables.sql` | 新增 | 固化内容类型、项目、Prompt、Provider 及状态日志关联字段。 |
| `openapi/openapi.yaml` | 修改 | 补充本迭代所有 endpoint、schema、错误响应和 security 描述。 |
| `apps/web-admin/lib/api.ts` | 修改 | 补充前端 API envelope 类型与新增接口 client 函数签名。 |
| `apps/web-admin/app/page.tsx` | 修改 | 首页 / 系统大盘、项目管理、项目详情壳层、项目模板管理入口骨架。 |
| `apps/web-admin/app/prompt/page.tsx` | 新增 | Prompt 模板管理页面骨架。 |
| `apps/web-admin/app/provider/page.tsx` | 新增 | 模型 Provider 管理页面骨架。 |
| `.cube/iterations/feature-1/skeleton-map.yaml` | 新增 | 记录骨架文件与 Development Tasks 的覆盖关系。 |

### 接口兼容性分析

现有 `/api/v1/health`、`/api/v1/system/*` 和 `/openapi.yaml` 不变。新增接口均为新路径，不破坏现有调用方。`router.NewRouter` 函数签名保持不变，现有单元测试无需迁移。

### 数据兼容性分析

本迭代新增表，不修改已有 `operation_log` 表。`content_project` 通过字段表达项目状态，项目暂停动作新增操作日志记录，不迁移历史数据。

## 3. Flow Design

### 3.1 列表加载流程

1. Web Admin 页面加载，调用 `apps/web-admin/lib/api.ts` 中对应 fetch 函数。
2. HTTP 请求进入 `/api/v1` 路由，先通过 Bearer token 占位认证。
3. Handler 解析 query：`page`、`page_size`、`sort`、`order` 以及业务筛选项。
4. Handler 将请求 DTO 传给对应 module service。
5. Service 返回分页列表 DTO。
6. Handler 通过 `api.WriteSuccess` 输出统一 envelope。
7. 校验失败时返回 `VALIDATION_ERROR`；服务层资源不存在或冲突时映射为 `NOT_FOUND` / `CONFLICT`。

### 3.2 新增资源流程

1. 页面提交表单。
2. Handler 解析 JSON body 到 request DTO。
3. Handler / service 校验必填字段、长度、配置结构与唯一性。
4. Service 创建资源并返回 ID 与展示字段。
5. Handler 返回 201 + 统一 envelope。
6. 页面显示成功反馈并刷新关联列表。

### 3.3 获取项目动态表单 Schema 流程

1. 新建项目页面根据用户选择的内容类型调用 `/api/v1/content-types/{id}/project-schema`。
2. Service 根据 id 查找 ContentType。
3. 找到则返回 `project_schema`；不存在则返回 `NOT_FOUND`。
4. 页面根据 schema 渲染动态表单。

### 3.4 暂停项目流程

1. 用户在项目详情壳层点击暂停并填写 `reason` 与可选 `note`。
2. Handler 解析 body 并校验 `reason` 必填。
3. Service 查找项目并校验当前状态是否允许暂停。
4. Service 将项目状态置为 `paused`。
5. Service 通过 operation log writer 写入状态变更日志。
6. Service 返回新状态与 `operation_log_id`。
7. 日志写入失败时，接口不得报告状态变更成功。

### 3.5 Provider API Key 脱敏流程

1. 用户新增 Provider 时提交 `api_key`。
2. Service 接收明文 key，但对外响应只构造 `api_key_masked`。
3. 列表查询同样只返回 `api_key_masked`。
4. Handler、日志和前端类型均不暴露明文 `api_key` 字段。

### 3.6 异常流程处理

- 认证缺失：路由中间件返回 `UNAUTHORIZED`。
- 参数错误：Handler / service 返回 `VALIDATION_ERROR` 与字段级 details。
- 资源不存在：返回 `NOT_FOUND`。
- 唯一键或状态冲突：返回 `CONFLICT`。
- 未预期错误：返回 `INTERNAL_ERROR`，不包含敏感信息。

## 4. Table Design

目标方言：PostgreSQL。

### 4.1 `content_type`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGSERIAL | PRIMARY KEY | 内容类型 ID |
| code | TEXT | NOT NULL UNIQUE | 类型编码 |
| name | TEXT | NOT NULL | 展示名称 |
| project_schema | JSONB | NOT NULL DEFAULT '{}' | 项目动态表单 schema |
| enabled | BOOLEAN | NOT NULL DEFAULT TRUE | 是否启用 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 更新时间 |

索引：`idx_content_type_enabled(enabled)`。

### 4.2 `content_project`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGSERIAL | PRIMARY KEY | 项目 ID |
| name | TEXT | NOT NULL | 项目名称 |
| content_type_id | BIGINT | NOT NULL REFERENCES content_type(id) | 所属内容类型 |
| status | TEXT | NOT NULL | `active` / `paused` / `draft` |
| project_config | JSONB | NOT NULL DEFAULT '{}' | 项目配置 |
| target_platform | TEXT | NOT NULL DEFAULT '' | 目标平台 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 更新时间 |

索引：`idx_content_project_status(status)`、`idx_content_project_content_type(content_type_id)`。

### 4.3 `prompt_template`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGSERIAL | PRIMARY KEY | Prompt 模板 ID |
| code | TEXT | NOT NULL UNIQUE | 模板编码 |
| agent_code | TEXT | NOT NULL DEFAULT '' | Agent 编码 |
| template | TEXT | NOT NULL | Prompt 内容 |
| variables | JSONB | NOT NULL DEFAULT '[]' | 变量列表 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 更新时间 |

索引：`idx_prompt_template_agent_code(agent_code)`。

### 4.4 `llm_provider_config`

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGSERIAL | PRIMARY KEY | Provider ID |
| provider_type | TEXT | NOT NULL | Provider 类型 |
| base_url | TEXT | NOT NULL | OpenAI-compatible endpoint |
| api_key | TEXT | NOT NULL | 明文字段仅限服务内部使用，禁止响应和日志输出 |
| api_key_masked | TEXT | NOT NULL | 脱敏展示值 |
| enabled | BOOLEAN | NOT NULL DEFAULT TRUE | 是否启用 |
| created_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMPTZ | NOT NULL DEFAULT NOW() | 更新时间 |

唯一约束：`UNIQUE(provider_type, base_url)`。

### SQL Contract

- Target dialect：PostgreSQL。
- DDL 必须使用 `CREATE TABLE IF NOT EXISTS` 以支持本地重复初始化。
- JSON 字段使用 JSONB，不拼接用户输入生成 SQL。
- 禁止在 migration 中写入明文 secret 示例数据。
- 禁止 `DROP TABLE` 删除非本 migration 创建的表；Down 只允许回滚本 migration 新增对象。
- 禁止 `TRUNCATE`、危险 DML seed、硬编码 API Key、token、password 或真实凭据。
- 禁止不可逆 destructive DDL；新增约束必须命名明确，便于回滚。
- 查询层后续实现必须使用参数绑定，禁止字符串拼接业务参数。

## 5. API Design

所有接口前缀为 `/api/v1`，均要求 Bearer token，占位认证沿用现有中间件。所有成功响应使用统一 envelope，失败响应使用统一 error envelope。

### 5.0 错误码矩阵

| 错误码 | 适用规则 |
|---|---|
| `VALIDATION_ERROR` | 任意 path、query、header、body 参数不合法；无入参 endpoint 标记为“不适用”。 |
| `UNAUTHORIZED` | 所有 `/api/v1` endpoint 缺失或无效 Bearer token。 |
| `FORBIDDEN` | 所有 `/api/v1` endpoint 认证通过但无权限访问资源或动作。 |
| `NOT_FOUND` | 读取、创建依赖或状态变更涉及的目标资源不存在；纯集合列表 endpoint 标记为“不适用”。 |
| `CONFLICT` | 唯一键冲突、幂等冲突、业务状态不允许当前动作；只读列表和只读摘要 endpoint 标记为“不适用”。 |
| `INTERNAL_ERROR` | 未预期内部错误、日志写入失败、序列化失败或未分类服务错误。 |

每个 endpoint 下列出的错误码是该接口的“必须支持或明确映射”的异常分支；未列出的统一错误码视为该 endpoint 不适用。OpenAPI 必须在 responses 中体现相同矩阵。

### 5.1 Dashboard

#### GET `/dashboard/summary`

- Query：无。
- Response data：`DashboardSummaryResponse`
  - `project_count: integer`
  - `pending_review_count: integer`
  - `pending_publish_count: integer`
  - `failed_task_count: integer`
  - `today_cost: number`
- 错误码：`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`。

### 5.2 Content Types

#### GET `/content-types`

- Query：`page`、`page_size`、`sort`、`order`、`enabled`。
- Response data：`PagedResponse<ContentTypeResponse>`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`。

#### POST `/content-types`

- Body：`CreateContentTypeRequest { code, name, project_schema }`。
- Response data：`CreateContentTypeResponse { content_type_id }`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`CONFLICT`、`INTERNAL_ERROR`。

#### GET `/content-types/{id}/project-schema`

- Path：`id`。
- Response data：`ProjectSchemaResponse { content_type_id, project_schema }`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`。

### 5.3 Projects

#### GET `/projects`

- Query：`page`、`page_size`、`sort`、`order`、`status`、`content_type`。
- Response data：`PagedResponse<ProjectResponse>`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`。

#### POST `/projects`

- Body：`CreateProjectRequest { name, content_type_id, project_config }`。
- Response data：`CreateProjectResponse { project_id, status }`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`INTERNAL_ERROR`。

#### GET `/projects/{id}/overview`

- Path：`id`。
- Response data：`ProjectOverviewResponse { project_id, progress, pending_actions, cost }`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`。

#### POST `/projects/{id}/pause`

- Path：`id`。
- Body：`PauseProjectRequest { reason, note }`。
- Response data：`PauseProjectResponse { project_id, status, operation_log_id }`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`INTERNAL_ERROR`。

### 5.4 Prompt Templates

#### GET `/prompt-templates`

- Query：`page`、`page_size`、`sort`、`order`、`agent_code`。
- Response data：`PagedResponse<PromptTemplateResponse>`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`。

#### POST `/prompt-templates`

- Body：`CreatePromptTemplateRequest { code, template, variables }`。
- Response data：`CreatePromptTemplateResponse { prompt_template_id }`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`CONFLICT`、`INTERNAL_ERROR`。

### 5.5 LLM Providers

#### GET `/llm-providers`

- Query：`page`、`page_size`、`sort`、`order`。
- Response data：`PagedResponse<LLMProviderResponse>`，只包含 `api_key_masked`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`。

#### POST `/llm-providers`

- Body：`CreateLLMProviderRequest { provider_type, base_url, api_key }`。
- Response data：`CreateLLMProviderResponse { provider_id, api_key_masked }`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`CONFLICT`、`INTERNAL_ERROR`。

## 6. Module Design

### 6.1 `dashboard` 模块

职责：聚合系统级摘要数据。

接口：

```go
type Service interface {
    Summary(ctx context.Context) (SummaryResponse, error)
}
```

### 6.2 `content` 模块

职责：管理内容类型、内容项目、项目概览和项目暂停状态变更。

接口：

```go
type Service interface {
    ListContentTypes(ctx context.Context, req ListContentTypesRequest) (PagedContentTypesResponse, error)
    CreateContentType(ctx context.Context, req CreateContentTypeRequest) (CreateContentTypeResponse, error)
    ProjectSchema(ctx context.Context, id string) (ProjectSchemaResponse, error)
    ListProjects(ctx context.Context, req ListProjectsRequest) (PagedProjectsResponse, error)
    CreateProject(ctx context.Context, req CreateProjectRequest) (CreateProjectResponse, error)
    ProjectOverview(ctx context.Context, id string) (ProjectOverviewResponse, error)
    PauseProject(ctx context.Context, id string, req PauseProjectRequest) (PauseProjectResponse, error)
}
```

### 6.3 `prompt` 模块

职责：管理 Prompt 模板列表和新增。

接口：

```go
type Service interface {
    ListTemplates(ctx context.Context, req ListTemplatesRequest) (PagedTemplatesResponse, error)
    CreateTemplate(ctx context.Context, req CreateTemplateRequest) (CreateTemplateResponse, error)
}
```

### 6.4 `llm` 模块

职责：管理 Provider 配置，对外保证 API Key 脱敏。

接口：

```go
type Service interface {
    ListProviders(ctx context.Context, req ListProvidersRequest) (PagedProvidersResponse, error)
    CreateProvider(ctx context.Context, req CreateProviderRequest) (CreateProviderResponse, error)
}
```

### 6.5 `handlers` 模块

职责：解析 HTTP 请求、调用 service、映射错误码、输出统一 envelope。Handler 不持有业务状态，不直接读写数据库。

### 6.6 前端模块

- `apps/web-admin/lib/api.ts`：定义新增 DTO 类型和 fetch 函数签名。
- `apps/web-admin/app/page.tsx`：首页、项目管理、项目详情壳层、项目模板管理入口骨架。
- `apps/web-admin/app/prompt/page.tsx`：Prompt 模板管理页面骨架。
- `apps/web-admin/app/provider/page.tsx`：Provider 管理页面骨架。

## 7. Output Contract

| 产出 | 输入 | 输出 | 正确性规则 | 类型 |
|---|---|---|---|---|
| GET `/dashboard/summary` / `Summary()` | Bearer token | Dashboard summary envelope | 返回项目数、待审稿、待发布、失败任务、今日成本；错误码矩阵：UNAUTHORIZED、FORBIDDEN、INTERNAL_ERROR | Web/API（type id: web-e2e） |
| GET `/content-types` / `ListContentTypes()` | page、page_size、sort、order、enabled | ContentType 分页 envelope | 支持分页、筛选、排序；空集合返回 items=[] 而非错误；错误码矩阵：VALIDATION_ERROR、UNAUTHORIZED、FORBIDDEN、INTERNAL_ERROR | Web/API（type id: web-e2e） |
| POST `/content-types` / `CreateContentType()` | code、name、project_schema | content_type_id envelope | 必填校验、code 唯一；错误码矩阵：VALIDATION_ERROR、UNAUTHORIZED、FORBIDDEN、CONFLICT、INTERNAL_ERROR | Web/API（type id: web-e2e） |
| GET `/content-types/{id}/project-schema` / `ProjectSchema()` | content type id | project_schema envelope | id 非法返回 VALIDATION_ERROR，不存在返回 NOT_FOUND；错误码矩阵：VALIDATION_ERROR、UNAUTHORIZED、FORBIDDEN、NOT_FOUND、INTERNAL_ERROR | Web/API（type id: web-e2e） |
| GET `/projects` / `ListProjects()` | page、page_size、sort、order、status、content_type | Project 分页 envelope | 支持分页、筛选、排序；空集合返回 items=[]；错误码矩阵：VALIDATION_ERROR、UNAUTHORIZED、FORBIDDEN、INTERNAL_ERROR | Web/API（type id: web-e2e） |
| POST `/projects` / `CreateProject()` | name、content_type_id、project_config | project_id、status envelope | content_type_id 必须存在，project_config 符合 schema；错误码矩阵：VALIDATION_ERROR、UNAUTHORIZED、FORBIDDEN、NOT_FOUND、CONFLICT、INTERNAL_ERROR | Web/API（type id: web-e2e） |
| GET `/projects/{id}/overview` / `ProjectOverview()` | project id | 项目进度、待处理、成本 envelope | id 非法返回 VALIDATION_ERROR，不存在返回 NOT_FOUND；错误码矩阵：VALIDATION_ERROR、UNAUTHORIZED、FORBIDDEN、NOT_FOUND、INTERNAL_ERROR | Web/API（type id: web-e2e） |
| POST `/projects/{id}/pause` / `PauseProject()` | project id、reason、note | project_id、status、operation_log_id envelope | reason 必填，状态冲突返回 CONFLICT，日志写入失败不得成功；错误码矩阵：VALIDATION_ERROR、UNAUTHORIZED、FORBIDDEN、NOT_FOUND、CONFLICT、INTERNAL_ERROR | Integration（type id: integration） |
| GET `/prompt-templates` / `ListTemplates()` | page、page_size、sort、order、agent_code | PromptTemplate 分页 envelope | 支持分页、筛选、排序；错误码矩阵：VALIDATION_ERROR、UNAUTHORIZED、FORBIDDEN、INTERNAL_ERROR | Web/API（type id: web-e2e） |
| POST `/prompt-templates` / `CreateTemplate()` | code、template、variables | prompt_template_id envelope | code/template 必填，code 唯一；错误码矩阵：VALIDATION_ERROR、UNAUTHORIZED、FORBIDDEN、CONFLICT、INTERNAL_ERROR | Web/API（type id: web-e2e） |
| GET `/llm-providers` / `ListProviders()` | page、page_size、sort、order | Provider 分页 envelope | 响应只含 api_key_masked，不含明文 api_key；错误码矩阵：VALIDATION_ERROR、UNAUTHORIZED、FORBIDDEN、INTERNAL_ERROR | Web/API + integration（type id: integration） |
| POST `/llm-providers` / `CreateProvider()` | provider_type、base_url、api_key | provider_id、api_key_masked envelope | 响应只含 api_key_masked，唯一冲突返回 CONFLICT；错误码矩阵：VALIDATION_ERROR、UNAUTHORIZED、FORBIDDEN、CONFLICT、INTERNAL_ERROR | Web/API + integration（type id: integration） |
| Migration SQL | migration runner | PostgreSQL tables | DDL 语法可执行，JSONB/索引/约束符合 Table Design，不含禁止 SQL 模式 | SQL/query（type id: sql-query） |
| OpenAPI Contract | API Design | OpenAPI 3.0 YAML | 每个 endpoint 包含 summary、description、tags、operationId、parameters、requestBody、responses、security、examples | Library/API contract（type id: library） |
| Web Admin 首页/项目管理 | 浏览器访问 `/`，页面动作 | 首页、项目列表、项目模板、新建项目、暂停项目、详情入口 | 页面动作映射 fetchDashboardSummary、fetchProjects、createProject、pauseProject、fetchProjectOverview、fetchContentTypes、createContentType、fetchProjectSchema；加载态、空状态、错误态、成功反馈存在 | Integration（type id: integration） |
| Web Admin Prompt/Provider | 浏览器访问 `/prompt`、`/provider`，页面动作 | Prompt/Provider 列表和新增入口 | 页面动作映射 fetchPromptTemplates、createPromptTemplate、fetchLLMProviders、createLLMProvider；Provider 页面不展示明文 api_key | Integration（type id: integration） |

### Testing Standards

- `standards/testing/web-e2e.md`：所有新增 HTTP endpoint 需要经公开 HTTP/API 入口测试，覆盖成功、校验失败和领域失败。
- `standards/testing/integration.md`：项目暂停、Provider 脱敏、前端页面到 API client 的链路需要跨组件测试。
- `standards/testing/sql-query.md`：migration DDL 需要验证目标方言、关键表、索引和禁止明文 secret 示例数据。
- `standards/testing/library.md`：API response 共享契约、OpenAPI 文档契约和前端 API client 类型作为 library/API contract 验证。

项目级 `workflow.yaml` 当前 `project.features` 为空，但本次设计新增 HTTP endpoint，因此本迭代显式声明 Web/API 测试规范。

## 8. Change Log

| 文件 | 类型 | 原因 |
|---|---|---|
| `apps/api-server/internal/http/api/response.go` | 修改 | 补充 Iteration 1 需要的错误码常量。 |
| `apps/api-server/internal/http/router.go` | 修改 | 挂载新增模块路由，保持已有系统路由兼容。 |
| `apps/api-server/internal/http/handlers/dashboard.go` | 新增 | Dashboard HTTP handler 骨架。 |
| `apps/api-server/internal/http/handlers/content.go` | 新增 | ContentType / Project HTTP handler 骨架。 |
| `apps/api-server/internal/http/handlers/prompt.go` | 新增 | PromptTemplate HTTP handler 骨架。 |
| `apps/api-server/internal/http/handlers/llm.go` | 新增 | LLM Provider HTTP handler 骨架。 |
| `apps/api-server/internal/modules/dashboard/dto.go` | 新增 | Dashboard DTO。 |
| `apps/api-server/internal/modules/dashboard/service.go` | 新增 | Dashboard service 接口与骨架构造。 |
| `apps/api-server/internal/modules/content/dto.go` | 新增 | ContentType / Project DTO。 |
| `apps/api-server/internal/modules/content/service.go` | 新增 | Content service 接口与骨架构造。 |
| `apps/api-server/internal/modules/prompt/dto.go` | 新增 | Prompt DTO。 |
| `apps/api-server/internal/modules/prompt/service.go` | 新增 | Prompt service 接口与骨架构造。 |
| `apps/api-server/internal/modules/llm/dto.go` | 新增 | LLM Provider DTO。 |
| `apps/api-server/internal/modules/llm/service.go` | 新增 | LLM Provider service 接口与骨架构造。 |
| `apps/api-server/migrations/00002_create_content_entry_tables.sql` | 新增 | 新增本迭代数据表。 |
| `openapi/openapi.yaml` | 修改 | 补充本迭代 API 契约。 |
| `apps/web-admin/lib/api.ts` | 修改 | 新增前端 DTO 与 API client 函数签名。 |
| `apps/web-admin/app/page.tsx` | 修改 | 首页与项目管理入口骨架。 |
| `apps/web-admin/app/prompt/page.tsx` | 新增 | Prompt 模板管理页面骨架。 |
| `apps/web-admin/app/provider/page.tsx` | 新增 | Provider 管理页面骨架。 |
| `.cube/iterations/feature-1/skeleton-map.yaml` | 新增 | 记录骨架覆盖关系。 |

## 9. Development Tasks

- Task-01：补充 API 错误码与分页契约
  - 所属模块：api
  - 简要描述：扩展统一响应错误码，定义分页 DTO 契约，供新增 handler 复用
  - 涉及接口/方法：api.WriteError(), api.Envelope
  - 输入：HTTP 请求上下文、错误码、字段错误详情
  - 输出：统一错误响应 envelope
  - 产出类型：library
  - 功能类型：共享 API 契约（type id: library）
  - 是否跨组件：否
- Task-02：新增数据库迁移契约
  - 所属模块：migrations
  - 简要描述：创建 content_type、content_project、prompt_template、llm_provider_config 表结构
  - 涉及接口/方法：goose migration
  - 输入：PostgreSQL migration runner
  - 输出：可执行 DDL
  - 产出类型：sql-query
  - 功能类型：SQL DDL 契约（type id: sql-query）
  - 是否跨组件：否
- Task-03：实现系统大盘摘要 API
  - 所属模块：dashboard
  - 简要描述：提供 GET /api/v1/dashboard/summary 的 DTO、service、handler 和路由
  - 涉及接口/方法：Summary(), DashboardHandler.Summary()
  - 输入：HTTP GET 请求
  - 输出：DashboardSummaryResponse envelope
  - 产出类型：web-e2e
  - 功能类型：HTTP API endpoint（type id: web-e2e）
  - 是否跨组件：是（组件链路：Router -> DashboardHandler -> DashboardService -> API Envelope）
- Task-04：实现内容类型管理 API
  - 所属模块：content
  - 简要描述：提供内容类型列表、新增、项目 schema 查询 API
  - 涉及接口/方法：ListContentTypes(), CreateContentType(), ProjectSchema()
  - 输入：分页筛选 query、创建 body、内容类型 id
  - 输出：分页列表、content_type_id、project_schema envelope
  - 产出类型：web-e2e
  - 功能类型：HTTP API endpoint（type id: web-e2e）
  - 是否跨组件：是（组件链路：Router -> ContentHandler -> ContentService -> API Envelope）
- Task-05：实现内容项目管理 API
  - 所属模块：content
  - 简要描述：提供项目列表、新建项目、项目概览、暂停项目 API，并返回 operation_log_id
  - 涉及接口/方法：ListProjects(), CreateProject(), ProjectOverview(), PauseProject()
  - 输入：分页筛选 query、创建 body、项目 id、暂停原因
  - 输出：分页列表、project_id、overview、pause result envelope
  - 产出类型：integration
  - 功能类型：项目状态变更链路（type id: integration）
  - 是否跨组件：是（组件链路：Router -> ContentHandler -> ContentService -> OperationLog -> API Envelope）
- Task-06：实现 Prompt 模板管理 API
  - 所属模块：prompt
  - 简要描述：提供 Prompt 模板列表和新增 API
  - 涉及接口/方法：ListTemplates(), CreateTemplate()
  - 输入：分页筛选 query、创建 body
  - 输出：分页列表、prompt_template_id envelope
  - 产出类型：web-e2e
  - 功能类型：HTTP API endpoint（type id: web-e2e）
  - 是否跨组件：是（组件链路：Router -> PromptHandler -> PromptService -> API Envelope）
- Task-07：实现 LLM Provider 管理 API 与 API Key 脱敏
  - 所属模块：llm
  - 简要描述：提供 Provider 列表和新增 API，确保响应只返回 api_key_masked
  - 涉及接口/方法：ListProviders(), CreateProvider(), MaskAPIKey()
  - 输入：分页 query、provider_type、base_url、api_key
  - 输出：分页列表、provider_id、api_key_masked envelope
  - 产出类型：integration
  - 功能类型：敏感字段脱敏链路（type id: integration）
  - 是否跨组件：是（组件链路：Router -> LLMHandler -> LLMService -> API Envelope）
- Task-08：补充 OpenAPI 契约
  - 所属模块：openapi
  - 简要描述：将本迭代所有接口、schema、错误响应、security 和 examples 写入 OpenAPI
  - 涉及接口/方法：openapi.yaml
  - 输入：API Design
  - 输出：OpenAPI 3.0 YAML
  - 产出类型：library
  - 功能类型：API 文档契约（type id: library）
  - 是否跨组件：否
- Task-09：实现 Web Admin 首页、项目管理与项目模板交互入口
  - 所属模块：web-admin
  - 简要描述：首页展示系统大盘；项目管理支持列表、新建项目、项目详情概览入口、暂停项目动作；项目模板管理支持内容类型列表、新增和动态 schema 查询入口；所有动作显示加载、空、错误和成功反馈
  - 涉及接口/方法：HomePage(), fetchDashboardSummary(), fetchProjects(), createProject(), fetchProjectOverview(), pauseProject(), fetchContentTypes(), createContentType(), fetchProjectSchema()
  - 输入：浏览器访问 `/`，列表筛选参数，新增表单数据，暂停原因
  - 输出：首页、项目管理、项目详情壳层、项目模板管理 UI 与 API client 调用契约
  - 产出类型：integration
  - 功能类型：前端页面到 API client 链路（type id: integration）
  - 是否跨组件：是（组件链路：NextPage -> Form/Action -> API Client -> HTTP API Contract）
- Task-10：实现 Prompt 与 Provider 管理页面交互入口
  - 所属模块：web-admin
  - 简要描述：新增 `/prompt` 与 `/provider` 页面骨架，支持 Prompt 列表、新增、错误反馈；Provider 列表、新增、错误反馈和 API Key 脱敏展示
  - 涉及接口/方法：PromptPage(), ProviderPage(), fetchPromptTemplates(), createPromptTemplate(), fetchLLMProviders(), createLLMProvider()
  - 输入：浏览器访问 `/prompt`、`/provider`，列表筛选参数，新增表单数据
  - 输出：Prompt / Provider 管理 UI 与 API client 调用契约
  - 产出类型：integration
  - 功能类型：前端页面到 API client 链路（type id: integration）
  - 是否跨组件：是（组件链路：NextPage -> Form/Action -> API Client -> HTTP API Contract）
- Task-11：生成骨架映射与阶段追踪文件
  - 所属模块：cube
  - 简要描述：生成 skeleton-map.yaml，确保每个 Development Task 至少映射到一个骨架文件，便于 03 阶段测试编写
  - 涉及接口/方法：skeleton-map.yaml
  - 输入：Change Log 与 Development Tasks
  - 输出：骨架文件到任务的 100% 覆盖映射
  - 产出类型：none
  - 功能类型：阶段追踪文件（type id: none）
  - 是否跨组件：否

## 10. 安全设计

- Provider 明文 `api_key` 只允许出现在 `CreateLLMProviderRequest` 输入 DTO 和服务内部存储模型中，不允许出现在 response DTO、列表 DTO、OpenAPI response schema、前端展示类型和日志字段中。
- `operation_log.metadata` 继续沿用已有 secret/token/password 检查约束。
- 所有新增接口沿用 Bearer token 占位认证；后续真实鉴权可替换中间件，不改变 handler/service 接口。
- 错误响应不回显请求体中的敏感字段。

## 11. 技术选型说明

- HTTP 框架继续使用现有 Chi，不引入新框架。
- 数据访问在本迭代骨架中通过 service/repository 接口隔离；真实 PostgreSQL 实现可在后续任务中替换，不影响 handler 与 DTO 契约。
- 前端继续使用 Next.js App Router，不引入 UI 组件库，以减少当前迭代改动范围。
