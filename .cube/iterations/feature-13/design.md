# 技术设计方案：Iteration 13 — Social Post Pack 内容类型扩展

## 1. 概述

本次设计将 Social Post Pack 作为新的 Content Pack 接入 AI Content Factory Core，并在项目工作区交付 Social 内容生成、多版本文案管理、标签/封面文案资产生成三条业务链路。整体方案复用现有 Article Pack 的路由形态、API Envelope、Web Admin 页面结构与导航风格；同时由于本次 PRD 明确要求新增持久化表、`content_version` 绑定、`operation_log` 审计以及真实的异步追踪链路，`socialpost` 模块采用与 `metrics` / `portfolio` / `strategy` 一致的 `Service + Store + PostgresStore` 结构，而不是继续沿用 Article 当前的纯内存实现。

核心设计决策如下：

- **最小改动**：沿用 `/api/v1/content-packs/{pack}`、`/api/v1/projects/{projectId}/{pack}` 的现有 URL 体系，前端沿用 `page-shell`、`page-hero`、`card`、`table-card` 等现有页面组织方式。
- **接口归口明确**：HTTP Handler 只负责请求解析、错误映射和 Envelope 输出；`socialpost.Service` 负责业务编排；`socialpost.Store` 负责 Social Post 自有表、`content_version`、`operation_log`、`idempotency_record` 的持久化。
- **真实持久化优先**：`social_post_extension`、`social_post_variant`、`social_post_asset` 三张表必须落库；状态变更必须写 `operation_log`；主选版本必须落 `content_version`。
- **追踪链路复用既有基础设施**：短内容生成、标签生成、封面文案生成均通过既有 `workflow.Service + engine.Submitter` 创建并执行 `workflow_run`，由异步工作流继续生成 `agent_task` / `llm_call_log`，满足 NFR 可追踪性。
- **一致性与并发安全**：同一 `content_item` 任意时刻只能有一个 `selected` 版本；通过数据库局部唯一索引 + 事务内归档旧主选并设置新主选实现。

## 2. Impact Analysis

| 模块 | 影响程度 | 说明 |
|------|---------|------|
| `apps/api-server/internal/modules/socialpost/` | 新增 | 新增 Social Post 领域模块，承载 DTO、错误、Service、Store、MemoryStore、PostgresStore |
| `apps/api-server/internal/http/handlers/social_post.go` | 新增 | 新增 SocialPostHandler，风格对齐 `article.go` |
| `apps/api-server/internal/http/router.go` | 修改 | 注册 Social Post 的 11 个 HTTP 端点，并增加 `WithSocialPostService` 注入选项 |
| `apps/api-server/internal/app/server.go` | 修改 | 当 DB 可用时注入 `socialpost.NewPostgresStore(db)` |
| `apps/api-server/migrations/00014_create_social_post_tables.sql` | 新增 | 新增 Social Post 三张表及唯一索引 |
| `openapi/openapi.yaml` | 修改 | 补充 11 个 Social Post API 的路径、Envelope schema、错误响应 |
| `apps/web-admin/lib/api.ts` | 修改 | 新增 Social Post API 类型与客户端函数 |
| `apps/web-admin/app/social-post-pack/page.tsx` | 新增 | Pack 管理页 |
| `apps/web-admin/app/projects/[projectId]/social-post/page.tsx` | 新增 | 项目 Social 配置与生成页 |
| `apps/web-admin/app/projects/[projectId]/social-post/variants/page.tsx` | 新增 | 候选文案管理页 |
| `apps/web-admin/app/projects/[projectId]/social-post/assets/page.tsx` | 新增 | 资产页 |
| `apps/web-admin/app/global-nav.tsx` | 修改 | 追加 Pack 管理入口 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 修改 | 追加项目内 Social Post 三个子页面入口 |
| 既有 `content` / `workflow` / `engine` 模块 | 复用 | 复用 ContentType / ContentProject / ContentItem / WorkflowRun / 异步执行能力，不新增并行基础设施 |
| 既有 `metrics` 能力 | 复用 | 复用 `metric_template` 表，仅注册 Social 默认指标模板，不新增 metrics 表 |
| 既有 `review` 查询语义 | 复用 | 主版本选择后写入 `content_version`，后续审稿/发布链路继续使用既有 review/publish 能力 |

### 兼容性分析

- **接口兼容性**：全部为新增路径，不修改现有 Article、Generation、Review、Metrics API 的请求/响应结构。
- **数据兼容性**：仅新增 Social Post 自有表和索引，不改已有表结构；通过 `content_item_id`、`generation_run_id`、`workflow_run_id`、`content_version_id` 与既有数据闭环。
- **前端兼容性**：新增页面与导航入口均为追加式改动，不破坏现有导航和页面布局。
- **运行时兼容性**：开发环境允许 `NewMemoryStore()` 以支持骨架编译和最小运行；集成/生产环境在 `DatabaseURL` 存在时必须使用 `PostgresStore`，以满足真实落库与审计要求。

### 方案比较

| 方案 | 描述 | 优点 | 缺点 |
|------|------|------|------|
| 方案 A：照搬 Article Pack 的内存模式 | 仅复制 Article 的 Service/Handler/UI 形态 | 改动最少，骨架最容易搭起 | 无法满足 PRD 要求的三张持久化表、`content_version` 绑定、`operation_log` 审计、并发唯一主选 |
| 方案 B：复用 Article 的路由/UI 形态 + 新增 Store/PostgresStore | API/UI 复用 Article，数据持久化对齐 metrics/portfolio/strategy 模式 | 同时满足“最小改动”和“真实持久化”，03/04/05 可围绕真实链路编写测试与验收 | 需要新增 Store 契约并修改 `router.go` / `server.go` |

**推荐方案：方案 B。** 它是在当前仓库结构下满足 PRD 和阶段检查要求的最小改动方案。

## 3. Flow Design

### 3.1 Social Post Pack 查看

```text
Admin UI → GET /api/v1/content-packs/social-post/status
  → SocialPostHandler.GetPackStatus
    → socialpost.Service.GetPackStatus
      → content.Service 获取 content_type(code=social_post)
      → workflow.Service 获取 code=social_post_generation 的 template / versions
      → metrics.Service 获取 content_type=social_post 的 metric templates
      → 组装 schema / workflows / metrics / current_version
      → 若 Pack 未注册，返回 NOT_FOUND
```

### 3.2 Social Post Pack 注册

```text
Admin UI → POST /api/v1/content-packs/social-post/register
  → SocialPostHandler.RegisterPack
    → socialpost.Service.RegisterPack
      → 校验 schema / workflows / metrics / version
      → store.CheckIdempotency(scope=social-post:pack, endpoint=register)
      → content.Service.CreateContentType(code=social_post)
      → workflow.Service.CreateTemplate(code=social_post_generation)
      → workflow.Service.CreateVersion(...) 并发布/设为 current_version
      → metrics.Service.CreateTemplate(...) × 7 写入默认指标模板
      → store.StoreIdempotency(..., response_ref_type=content_type, response_ref_id=content_type_id)
      → 返回 content_pack_id / content_type_id / registered_version
```

说明：Pack 注册状态不使用独立 `pack_registration` 表，注册结果直接由既有 `content_type`、`workflow_template(_version)`、`metric_template` 组合推导，避免新增冗余表。

### 3.3 项目 Social 配置查询与更新

```text
Admin UI → GET /api/v1/projects/{projectId}/social-post/config
  → SocialPostHandler.GetConfig
    → socialpost.Service.GetConfig
      → content.Service 校验项目存在
      → store.GetExtensionByProjectID(projectId)
      → 若不存在，返回默认配置结构（非错误）

Admin UI → PATCH /api/v1/projects/{projectId}/social-post/config
  → SocialPostHandler.UpdateConfig
    → socialpost.Service.UpdateConfig
      → 校验 target_platforms / default_variant_count / forbidden_terms / policy 字段
      → store.CheckIdempotency(scope=social-post:config:{projectId}, endpoint=patch-config)
      → store.UpsertExtension(...)
      → store.InsertOperationLog(actor, resource, from_state, to_state, reason)
      → store.StoreIdempotency(...)
      → 返回 version_id / operation_log_id
```

### 3.4 短内容生成

```text
Admin UI → POST /api/v1/projects/{projectId}/social-post/generation-runs
  → SocialPostHandler.CreateGenerationRun
    → socialpost.Service.CreateGenerationRun
      → 校验 topic / platform / version_count <= 10 / asset_options
      → content.Service 校验项目和 social_post content_type
      → store.CheckIdempotency(scope=social-post:generation:{projectId}, endpoint=create-run)
      → 创建或关联 ContentItem（复用既有 ContentItem 语义）
      → workflow.Service.CreateRun(template=social_post_generation)
      → engine.Submitter.Submit(workflow_run_id)
      → store.InsertOperationLog(...) 记录触发动作
      → store.StoreIdempotency(...)
      → 返回 generation_run_id / workflow_run_id / status=running
```

异步执行约束：

- 真实执行由既有 workflow engine 继续推进。
- `workflow_run` 下游步骤自动产生 `agent_task` / `llm_call_log`，满足 NFR-001。
- 当 Social 输出解析成功后，业务实现写入 `social_post_variant` 新记录；重新生成只新增历史版本，不覆盖旧记录。
- API 层状态映射规则：底层 `workflow_run.status=succeeded` 对外映射为 `completed`。

### 3.5 生成详情查询

```text
Admin UI → GET /api/v1/projects/{projectId}/social-post/generation-runs/{id}
  → SocialPostHandler.GetGenerationRun
    → socialpost.Service.GetGenerationRun
      → 读取既有 generation_run / workflow_run 信息
      → 读取关联 variants
      → 读取 workflow run error / trace summary
      → 组装 status / content_item_id / variants / error / workflow_run_id
```

### 3.6 候选文案列表与主版本选择

```text
Admin UI → GET /api/v1/projects/{projectId}/social-post/variants
  → SocialPostHandler.ListVariants
    → socialpost.Service.ListVariants
      → 按 project_id + content_item_id + status + platform 分页查询 social_post_variant
      → 返回 items + pagination

Admin UI → POST /api/v1/projects/{projectId}/social-post/variants/{variantId}/select
  → SocialPostHandler.SelectVariant
    → socialpost.Service.SelectVariant
      → 校验 variant 存在且属于 project/content_item
      → 开启事务
        → SELECT ... FOR UPDATE 锁定同 content_item 下候选行
        → 将旧 selected 版本归档为 archived
        → 写入 content_version(source=generation)
        → 更新目标 variant.status=selected / content_version_id / selected_at
        → 写 operation_log
      → 提交事务
      → 返回 selected_variant_id / content_version_id / operation_log_id
```

并发安全规则：

- `social_post_variant` 增加局部唯一索引：同一 `content_item_id` 在 `status='selected'` 时只能存在一行。
- 选择主版本必须在单事务内完成，否则视为实现缺陷。

### 3.7 标签与封面文案生成及资产查询

```text
Admin UI → POST /api/v1/projects/{projectId}/social-post/assets/tags:generate
  → SocialPostHandler.GenerateTags
    → socialpost.Service.GenerateTags
      → 校验 project / content_item / variant / platform / count
      → store.CheckIdempotency(scope=social-post:asset-tags:{projectId}, endpoint=generate-tags)
      → workflow.Service.CreateRun(template=social_post_tags)
      → engine.Submitter.Submit(workflow_run_id)
      → store.InsertOperationLog(...)
      → store.StoreIdempotency(...)
      → 返回 generation_run_id / workflow_run_id / status=running

Admin UI → POST /api/v1/projects/{projectId}/social-post/assets/cover-copy:generate
  → 与 tags 类似，template=social_post_cover_copy

异步执行成功：
  → 将结果写入 social_post_asset
  → tags 结果保存在 result.tags
  → cover_copy 结果保存在 result.items

Admin UI → GET /api/v1/projects/{projectId}/social-post/assets
  → SocialPostHandler.GetAssets
    → socialpost.Service.GetAssets
      → 过滤 project_id / content_item_id / platform / variant_id
      → 返回 tags / cover_copy / asset_suggestions / source_runs
```

### 3.8 异常流程

- Pack 未注册：`GET /content-packs/social-post/status` 返回 `NOT_FOUND`
- 参数校验失败、状态流转非法、`version_count > 10`：返回 `VALIDATION_ERROR`
- 资源不存在（project、content_item、variant、generation_run）：返回 `NOT_FOUND`
- 幂等键命中但请求体 hash 改变：返回 `IDEMPOTENCY_CONFLICT`
- 注册版本冲突、重复注册不同 schema 版本、状态不允许重复选择：返回 `CONFLICT`
- LLM 输出解析失败：返回 `AGENT_OUTPUT_INVALID`，并保留 `workflow_run` / `agent_task` / `llm_call_log` 失败证据
- 任一状态变更写审计失败：作为内部错误返回，避免业务状态已变更但无审计记录

## 4. Table Design

本次新增迁移文件 `00014_create_social_post_tables.sql`，采用与现有迁移一致的 goose `Up/Down` 结构。

### 4.1 social_post_extension

项目级 Social 配置。每个项目最多一条。

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS social_post_extension (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    target_platforms JSONB NOT NULL DEFAULT '[]'::jsonb,
    default_variant_count INTEGER NOT NULL DEFAULT 3 CHECK (default_variant_count > 0 AND default_variant_count <= 10),
    caption_length_policy TEXT NOT NULL DEFAULT 'short',
    hashtag_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
    cover_copy_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
    tone_style TEXT NOT NULL DEFAULT '',
    forbidden_terms JSONB NOT NULL DEFAULT '[]'::jsonb,
    config_version INTEGER NOT NULL DEFAULT 1 CHECK (config_version > 0),
    operation_log_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id)
);

CREATE INDEX IF NOT EXISTS idx_social_post_extension_project ON social_post_extension(project_id);
```

字段约束：

- `project_id` 与既有 `content_project.id` 对齐，服务层按字符串透传。
- 仅保存内容生成策略，不保存平台账号或敏感凭证。
- `config_version` 对应 PRD 的版本字段；API 输出也使用 `config_version`。

### 4.2 social_post_variant

候选文案版本表。

```sql
CREATE TABLE IF NOT EXISTS social_post_variant (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    content_item_id TEXT NOT NULL,
    generation_run_id TEXT NOT NULL,
    workflow_run_id TEXT NOT NULL,
    variant_index INTEGER NOT NULL CHECK (variant_index > 0),
    platform TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    hashtags JSONB NOT NULL DEFAULT '[]'::jsonb,
    cover_copy TEXT NOT NULL DEFAULT '',
    tone_style TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL CHECK (status IN ('generated', 'selected', 'rejected', 'archived')) DEFAULT 'generated',
    content_version_id TEXT NOT NULL DEFAULT '',
    selected_at TIMESTAMPTZ,
    operation_log_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(content_item_id, generation_run_id, variant_index)
);

CREATE INDEX IF NOT EXISTS idx_social_post_variant_project_created ON social_post_variant(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_social_post_variant_content_item_status ON social_post_variant(content_item_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_social_post_variant_platform_status ON social_post_variant(platform, status, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_social_post_variant_selected_unique
    ON social_post_variant(content_item_id)
    WHERE status = 'selected';
```

字段约束：

- `generation_run_id` 对应既有 `generation_run.id`
- `workflow_run_id` 对应既有 `workflow_run.id`
- `content_version_id` 在主选后写入，用于进入既有审稿/发布链路
- 重新生成仅新增新记录，不覆盖历史候选版本

### 4.3 social_post_asset

标签与封面文案资产结果表。

```sql
CREATE TABLE IF NOT EXISTS social_post_asset (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    content_item_id TEXT NOT NULL,
    source_variant_id TEXT NOT NULL DEFAULT '',
    asset_type TEXT NOT NULL CHECK (asset_type IN ('tags', 'cover_copy')),
    platform TEXT NOT NULL,
    generation_run_id TEXT NOT NULL,
    workflow_run_id TEXT NOT NULL,
    result JSONB NOT NULL DEFAULT '{}'::jsonb,
    asset_suggestions JSONB NOT NULL DEFAULT '[]'::jsonb,
    operation_log_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_social_post_asset_project_type_created ON social_post_asset(project_id, asset_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_social_post_asset_content_item_platform ON social_post_asset(content_item_id, platform, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_social_post_asset_variant_type ON social_post_asset(source_variant_id, asset_type, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS social_post_asset;
DROP TABLE IF EXISTS social_post_variant;
DROP TABLE IF EXISTS social_post_extension;
```

字段约束：

- `source_variant_id` 对齐 PRD 的 `source_variant_id`
- `asset_suggestions` 对应 PRD FR-032 的额外建议输出
- `result` 结构：
  - `tags`：`{"tags": ["#618"], "style": "trending"}`
  - `cover_copy`：`{"items": [{"text": "夏日大促", "style": "warm"}]}`

### 4.4 复用既有表与行映射

| 现有表 | 复用方式 |
|--------|---------|
| `content_type` | 注册 `code=social_post` |
| `content_project` | 通过 `project_id` 关联项目 |
| `generation_run` | 承载异步运行主记录 |
| `content_item` | Social 内容主资源 |
| `workflow_template` / `workflow_template_version` / `workflow_run` | 承载生成、标签、封面文案工作流 |
| `metric_template` | 注册默认指标模板 |
| `content_version` | 主版本选择后创建新版本记录 |
| `operation_log` | 配置更新、主选切换、触发动作写审计 |
| `idempotency_record` | 注册、配置更新、生成、资产触发等幂等接口复用 |

### 4.5 Social 默认指标模板映射

本次不新增 `metric_template` 表，仅通过既有 Metrics 能力注册 7 条默认模板。字段映射对齐现有 schema：

| metric_code | metric_name | content_type | platform | unit | value_type | aggregation_method | period | required | enabled |
|-------------|-------------|--------------|----------|------|------------|--------------------|--------|----------|---------|
| `impressions` | 曝光 | `social_post` | `generic` | `count` | `integer` | `sum` | `day` | `false` | `true` |
| `clicks` | 点击 | `social_post` | `generic` | `count` | `integer` | `sum` | `day` | `false` | `true` |
| `likes` | 点赞 | `social_post` | `generic` | `count` | `integer` | `sum` | `day` | `false` | `true` |
| `favorites` | 收藏 | `social_post` | `generic` | `count` | `integer` | `sum` | `day` | `false` | `true` |
| `comments` | 评论 | `social_post` | `generic` | `count` | `integer` | `sum` | `day` | `false` | `true` |
| `shares` | 转发 | `social_post` | `generic` | `count` | `integer` | `sum` | `day` | `false` | `true` |
| `follow_conversion` | 关注转化 | `social_post` | `generic` | `count` | `integer` | `sum` | `day` | `false` | `true` |

## 5. API Design

本次共新增 **11 个 HTTP 端点**，全部遵循 `/api/v1`、统一 Envelope、Bearer Auth、`request_id`、统一错误码约定。

### 5.1 API 列表

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/v1/content-packs/social-post/status` | 查看 Pack 配置 |
| POST | `/api/v1/content-packs/social-post/register` | 注册 Pack |
| GET | `/api/v1/projects/{projectId}/social-post/config` | 查看项目 Social 配置 |
| PATCH | `/api/v1/projects/{projectId}/social-post/config` | 更新项目 Social 配置 |
| POST | `/api/v1/projects/{projectId}/social-post/generation-runs` | 触发短内容生成 |
| GET | `/api/v1/projects/{projectId}/social-post/generation-runs/{id}` | 查询生成详情 |
| GET | `/api/v1/projects/{projectId}/social-post/variants` | 查看候选文案列表 |
| POST | `/api/v1/projects/{projectId}/social-post/variants/{variantId}/select` | 选择主版本 |
| POST | `/api/v1/projects/{projectId}/social-post/assets/tags:generate` | 触发标签生成 |
| POST | `/api/v1/projects/{projectId}/social-post/assets/cover-copy:generate` | 触发封面文案生成 |
| GET | `/api/v1/projects/{projectId}/social-post/assets` | 查看资产结果 |

### 5.2 GET /api/v1/content-packs/social-post/status

**成功响应 data**

```json
{
  "content_pack_id": "cp_social_post",
  "content_type": {
    "id": "13",
    "code": "social_post",
    "name": "Social Post Pack",
    "project_schema": {
      "target_platforms": {"type": "array"},
      "default_variant_count": {"type": "integer", "maximum": 10}
    },
    "enabled": true
  },
  "schema": {
    "content_type_code": "social_post",
    "project_fields": ["target_platforms", "default_variant_count", "tone_style"]
  },
  "workflows": [
    {"template_id": "31", "code": "social_post_generation", "name": "Social Post Generation", "current_version": "45"},
    {"template_id": "32", "code": "social_post_tags", "name": "Social Post Tags", "current_version": "46"},
    {"template_id": "33", "code": "social_post_cover_copy", "name": "Social Post Cover Copy", "current_version": "47"}
  ],
  "metrics": [
    {"metric_code": "impressions", "metric_name": "曝光", "unit": "count", "platform": "generic"}
  ],
  "current_version": "2026.06.social-post.v1"
}
```

**错误码**：`NOT_FOUND`

### 5.3 POST /api/v1/content-packs/social-post/register

**请求体**

```json
{
  "schema": {
    "content_type_code": "social_post",
    "name": "Social Post Pack",
    "project_schema": {
      "target_platforms": {"type": "array"},
      "default_variant_count": {"type": "integer", "maximum": 10}
    }
  },
  "workflows": [
    {"code": "social_post_generation", "name": "Social Post Generation"},
    {"code": "social_post_tags", "name": "Social Post Tags"},
    {"code": "social_post_cover_copy", "name": "Social Post Cover Copy"}
  ],
  "metrics": [
    {"metric_code": "impressions", "metric_name": "曝光", "unit": "count"},
    {"metric_code": "clicks", "metric_name": "点击", "unit": "count"}
  ],
  "version": "2026.06.social-post.v1"
}
```

`Idempotency-Key` 通过 Header 传递。

**成功响应 data**

```json
{
  "content_pack_id": "cp_social_post",
  "content_type_id": "13",
  "registered_version": "2026.06.social-post.v1"
}
```

**错误码**：`VALIDATION_ERROR`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`

### 5.4 GET /api/v1/projects/{projectId}/social-post/config

**成功响应 data**

```json
{
  "target_platforms": ["xiaohongshu", "wechat_channels"],
  "default_variant_count": 3,
  "caption_length_policy": "short",
  "hashtag_policy": {"mode": "auto", "count": 5},
  "cover_copy_policy": {"mode": "auto", "count": 2},
  "tone_style": "professional",
  "forbidden_terms": [],
  "config_version": 1
}
```

配置不存在时返回默认结构，不返回错误。

**错误码**：`NOT_FOUND`

### 5.5 PATCH /api/v1/projects/{projectId}/social-post/config

**请求体**

```json
{
  "target_platforms": ["xiaohongshu"],
  "default_variant_count": 3,
  "caption_length_policy": "short",
  "hashtag_policy": {"mode": "auto", "count": 5},
  "cover_copy_policy": {"mode": "manual", "count": 1},
  "tone_style": "friendly",
  "forbidden_terms": ["绝对化用语"]
}
```

**成功响应 data**

```json
{
  "version_id": "social-post-config-2",
  "operation_log_id": "oplog-social-post-config-2"
}
```

**错误码**：`VALIDATION_ERROR`、`NOT_FOUND`、`IDEMPOTENCY_CONFLICT`

### 5.6 POST /api/v1/projects/{projectId}/social-post/generation-runs

**请求体**

```json
{
  "topic": "618 促销预热",
  "source_content_item_id": "",
  "platform": "xiaohongshu",
  "version_count": 3,
  "tone_style": "friendly",
  "asset_options": {
    "generate_tags": true,
    "generate_cover_copy": false
  }
}
```

**成功响应 data**

```json
{
  "generation_run_id": "genrun-social-1",
  "workflow_run_id": "52",
  "status": "running"
}
```

**错误码**：`VALIDATION_ERROR`、`NOT_FOUND`、`IDEMPOTENCY_CONFLICT`、`AGENT_OUTPUT_INVALID`

### 5.7 GET /api/v1/projects/{projectId}/social-post/generation-runs/{id}

**成功响应 data**

```json
{
  "generation_run_id": "genrun-social-1",
  "workflow_run_id": "52",
  "status": "completed",
  "content_item_id": "content-item-1",
  "trace": {
    "agent_task_ids": ["agent-task-1"],
    "llm_call_log_ids": ["llm-log-1"]
  },
  "variants": [
    {
      "id": "variant-1",
      "variant_index": 1,
      "platform": "xiaohongshu",
      "title": "标题",
      "body": "正文",
      "hashtags": ["#618"],
      "cover_copy": "封面",
      "tone_style": "friendly",
      "status": "generated",
      "created_at": "2026-06-05T10:00:00Z"
    }
  ],
  "error": ""
}
```

**错误码**：`NOT_FOUND`

### 5.8 GET /api/v1/projects/{projectId}/social-post/variants

Query 参数：`content_item_id`、`status`、`platform`、`page`、`page_size`

**成功响应 data**

```json
{
  "items": [
    {
      "id": "variant-1",
      "content_item_id": "content-item-1",
      "variant_index": 1,
      "platform": "xiaohongshu",
      "title": "标题",
      "body": "正文",
      "hashtags": ["#618"],
      "cover_copy": "封面",
      "tone_style": "friendly",
      "status": "generated",
      "content_version_id": "",
      "created_at": "2026-06-05T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "has_next": false
  }
}
```

**错误码**：`NOT_FOUND`

### 5.9 POST /api/v1/projects/{projectId}/social-post/variants/{variantId}/select

**请求体**

```json
{
  "content_item_id": "content-item-1",
  "note": "选择第 1 版作为主版本"
}
```

**成功响应 data**

```json
{
  "selected_variant_id": "variant-1",
  "content_version_id": "version-content-item-1-1",
  "operation_log_id": "oplog-variant-1"
}
```

**错误码**：`VALIDATION_ERROR`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`

### 5.10 POST /api/v1/projects/{projectId}/social-post/assets/tags:generate

**请求体**

```json
{
  "content_item_id": "content-item-1",
  "variant_id": "variant-1",
  "platform": "xiaohongshu",
  "count": 5,
  "style": "trending"
}
```

**成功响应 data**

```json
{
  "generation_run_id": "asset-run-1",
  "workflow_run_id": "66",
  "status": "running"
}
```

**错误码**：`VALIDATION_ERROR`、`NOT_FOUND`、`IDEMPOTENCY_CONFLICT`、`AGENT_OUTPUT_INVALID`

### 5.11 POST /api/v1/projects/{projectId}/social-post/assets/cover-copy:generate

请求体与成功响应结构同上，仅 `asset_type=cover_copy`。

**错误码**：`VALIDATION_ERROR`、`NOT_FOUND`、`IDEMPOTENCY_CONFLICT`、`AGENT_OUTPUT_INVALID`

### 5.12 GET /api/v1/projects/{projectId}/social-post/assets

Query 参数：`content_item_id`、`platform`、`variant_id`

**成功响应 data**

```json
{
  "tags": [
    {
      "id": "asset-1",
      "platform": "xiaohongshu",
      "source_variant_id": "variant-1",
      "generation_run_id": "asset-run-1",
      "result": {"tags": ["#618"]},
      "created_at": "2026-06-05T10:00:00Z"
    }
  ],
  "cover_copy": [
    {
      "id": "asset-2",
      "platform": "xiaohongshu",
      "source_variant_id": "variant-1",
      "generation_run_id": "asset-run-2",
      "result": {"items": [{"text": "夏日大促", "style": "warm"}]},
      "created_at": "2026-06-05T10:05:00Z"
    }
  ],
  "asset_suggestions": ["优先保留品牌词"],
  "source_runs": ["asset-run-1", "asset-run-2"]
}
```

无资产时返回空数组。

**错误码**：`NOT_FOUND`

## 6. Module Design

### 6.1 socialpost 模块结构

```text
apps/api-server/internal/modules/socialpost/
├── dto.go
├── errors.go
├── service.go
├── store.go
├── memory_store.go
└── postgres_store.go
```

### 6.2 核心接口

#### Service

```go
type Service interface {
    GetPackStatus(ctx context.Context) (SocialPostPackStatusResponse, error)
    RegisterPack(ctx context.Context, req RegisterSocialPostPackRequest, idempotencyKey string) (RegisterSocialPostPackResponse, error)

    GetConfig(ctx context.Context, projectID string) (SocialPostConfigResponse, error)
    UpdateConfig(ctx context.Context, projectID string, req UpdateSocialPostConfigRequest, idempotencyKey string) (UpdateSocialPostConfigResponse, error)

    CreateGenerationRun(ctx context.Context, projectID string, req CreateSocialPostGenerationRunRequest, idempotencyKey string) (CreateSocialPostGenerationRunResponse, error)
    GetGenerationRun(ctx context.Context, projectID, generationRunID string) (SocialPostGenerationRunDetailResponse, error)

    ListVariants(ctx context.Context, projectID string, req ListSocialPostVariantsRequest) (PagedSocialPostVariantsResponse, error)
    SelectVariant(ctx context.Context, projectID, variantID string, req SelectSocialPostVariantRequest, idempotencyKey string) (SelectSocialPostVariantResponse, error)

    GenerateTags(ctx context.Context, projectID string, req GenerateSocialPostTagsRequest, idempotencyKey string) (GenerateSocialPostAssetResponse, error)
    GenerateCoverCopy(ctx context.Context, projectID string, req GenerateSocialPostCoverCopyRequest, idempotencyKey string) (GenerateSocialPostAssetResponse, error)
    GetAssets(ctx context.Context, projectID string, req GetSocialPostAssetsRequest) (SocialPostAssetsResponse, error)
}
```

#### Store

```go
type Store interface {
    GetExtensionByProjectID(ctx context.Context, projectID string) (*SocialPostConfigRow, error)
    UpsertExtension(ctx context.Context, row SocialPostConfigRow) error

    InsertVariant(ctx context.Context, row SocialPostVariantRow) error
    ListVariants(ctx context.Context, projectID string, req ListSocialPostVariantsRequest) ([]SocialPostVariantResponse, int, error)
    GetVariantByID(ctx context.Context, variantID string) (*SocialPostVariantRow, error)
    SelectVariantInTx(ctx context.Context, input SelectVariantTxInput) (contentVersionID string, operationLogID string, err error)

    InsertAsset(ctx context.Context, row SocialPostAssetRow) error
    ListAssets(ctx context.Context, projectID string, req GetSocialPostAssetsRequest) ([]SocialPostAssetItem, error)

    InsertOperationLog(ctx context.Context, row OperationLogRow) (string, error)
    CheckIdempotency(ctx context.Context, scope, endpoint, key, hash string) (refType string, refID string, conflict bool, err error)
    StoreIdempotency(ctx context.Context, scope, endpoint, key, hash, refType, refID string) error
}
```

### 6.3 依赖关系

```text
SocialPostHandler
  → socialpost.Service
    → socialpost.Store
    → content.Service
    → workflow.Service
    → metrics.Service
    → engine.Submitter
```

### 6.4 责任边界

- `SocialPostHandler`
  - 读取 path/query/body/header
  - 提取 `Idempotency-Key`
  - 调用 `socialpost.Service`
  - 将领域错误映射为 `api.ErrorValidation` / `api.ErrorNotFound` / `api.ErrorConflict` / `api.ErrorIdempotencyConflict` / `api.ErrorAgentOutputInvalid`
- `socialpost.Service`
  - 校验业务输入
  - 编排 `content` / `workflow` / `metrics` / `engine` / `store`
  - 负责状态映射（例如 `succeeded -> completed`）
- `socialpost.PostgresStore`
  - 负责 `social_post_*` 表、`content_version`、`operation_log`、`idempotency_record` 的 SQL 读写
  - `SelectVariantInTx` 必须使用单事务保障唯一主选
- `socialpost.MemoryStore`
  - 仅用于本地最小运行与骨架编译，不得作为 04/05 阶段真实实现的替代

### 6.5 前端页面设计

- `app/social-post-pack/page.tsx`
  - 对齐 `article-pack/page.tsx`
  - 功能：查看 Pack 状态、显示 schema/workflows/metrics/current_version、执行注册、处理 404 空态
- `app/projects/[projectId]/social-post/page.tsx`
  - 对齐 `projects/[projectId]/article/page.tsx`
  - 功能：配置表单、触发生成、查看 run 详情和 trace
- `app/projects/[projectId]/social-post/variants/page.tsx`
  - 功能：筛选、分页、选择主版本、展示 `content_version_id`
- `app/projects/[projectId]/social-post/assets/page.tsx`
  - 功能：触发 tags / cover_copy 生成，查看资产与 `asset_suggestions`
- `global-nav.tsx`
  - 新增 `/social-post-pack`
- `workspace-nav.tsx`
  - 新增 `social-post`、`social-post/variants`、`social-post/assets`

## 7. Output Contract

### 7.1 功能类型识别

`workflow.yaml` 的 `project.features` 当前为空列表，但本次迭代实际包含 HTTP API、前端页面和跨组件链路，因此按实际内容推断如下：

| 产物 | 业务描述 | type id | 测试规范 |
|------|---------|---------|---------|
| 11 个 Social Post HTTP 端点 | Pack 注册、配置、生成、候选文案、资产接口 | `web-e2e` | `standards/testing/web-e2e.md` |
| 4 个前端页面 | Pack 管理、项目配置/生成、候选文案、资产页 | `frontend-ui` | `standards/testing/frontend-ui.md` |
| Handler → Service → Store / Content / Workflow / Engine 链路 | 跨组件业务链 | `integration` | `standards/testing/integration.md` |

说明：本次没有 SQL/query generator 型产物，迁移文件属于 schema 变更而非 `sql-query` 类型测试触发对象。

### 7.2 类型化测试条目

- `web-e2e`：Social Post HTTP API full roundtrip
- `frontend-ui`：Social Post Admin pages full roundtrip
- `integration`：Social Post generation / selection / asset chain

### 7.3 正确性规则

- **Pack 状态契约**：已注册时必须返回 `schema`、`workflows`、`metrics`、`current_version`；未注册时返回 `NOT_FOUND`
- **Pack 注册幂等**：相同 `Idempotency-Key` + 相同请求体重复注册返回相同 `registered_version`；请求体 hash 变化返回 `IDEMPOTENCY_CONFLICT`
- **配置默认值**：配置不存在时 `GET config` 返回默认结构，不返回 `null` 或 404
- **成本约束**：`version_count` 服务端上限为 10
- **可追踪性**：所有生成动作必须关联 `workflow_run_id`，并能追溯到 `agent_task` / `llm_call_log`
- **唯一主选**：同一 `content_item_id` 任意时刻最多一个 `status=selected`
- **版本绑定**：`SelectVariant` 成功后必须返回非空 `content_version_id`，并同步写入 `social_post_variant.content_version_id`
- **资产关联完整性**：所有资产结果都必须关联 `project_id`、`content_item_id`、`source_variant_id`、`generation_run_id`、`workflow_run_id`
- **审计完整性**：配置更新、主选切换、生成触发均必须返回 `operation_log_id` 或至少在持久层成功写审计后再返回成功
- **错误展示协议**：所有失败响应遵循统一 Envelope，前端必须展示 `error.code`、`error.message`、`request_id`

### 7.4 集成链路

| 链路 | 组件链 | 说明 |
|------|-------|------|
| Pack 注册 | `SocialPostHandler -> socialpost.Service -> content.Service -> workflow.Service -> metrics.Service -> socialpost.Store` | 注册内容类型、工作流版本、默认指标模板并记录幂等 |
| 项目配置更新 | `SocialPostHandler -> socialpost.Service -> socialpost.Store` | 更新配置并写 `operation_log` |
| Social 内容生成 | `SocialPostHandler -> socialpost.Service -> content.Service -> workflow.Service -> engine.Submitter -> socialpost.Store` | 触发异步生成并持久化候选版本 |
| 主版本选择 | `SocialPostHandler -> socialpost.Service -> socialpost.Store` | 事务内归档旧主选、写 `content_version`、更新新主选、写审计 |
| 标签/封面文案生成 | `SocialPostHandler -> socialpost.Service -> workflow.Service -> engine.Submitter -> socialpost.Store` | 触发异步资产生成并查询结果 |

## 8. Change Log

| 文件 | 变更类型 | 变更原因 |
|------|---------|---------|
| `apps/api-server/internal/modules/socialpost/dto.go` | 新增 | 定义 Social Post DTO、列表项、配置对象、资产对象 |
| `apps/api-server/internal/modules/socialpost/errors.go` | 新增 | 定义领域错误常量 |
| `apps/api-server/internal/modules/socialpost/service.go` | 新增 | 定义 Service 接口、构造函数、空实现骨架 |
| `apps/api-server/internal/modules/socialpost/store.go` | 新增 | 定义 Store 接口、事务输入输出、行对象 |
| `apps/api-server/internal/modules/socialpost/memory_store.go` | 新增 | 提供开发/测试默认内存骨架实现 |
| `apps/api-server/internal/modules/socialpost/postgres_store.go` | 新增 | 提供 PostgresStore 骨架和 SQL 方法签名 |
| `apps/api-server/internal/http/handlers/social_post.go` | 新增 | 新增 SocialPostHandler |
| `apps/api-server/internal/http/router.go` | 修改 | 追加 `WithSocialPostService` 与 11 个端点注册 |
| `apps/api-server/internal/app/server.go` | 修改 | 在 DB 可用时注入 SocialPost PostgresStore |
| `apps/api-server/migrations/00014_create_social_post_tables.sql` | 新增 | 创建 Social Post 三张表及唯一索引 |
| `openapi/openapi.yaml` | 修改 | 新增 11 个 Social Post 路径和 schema |
| `apps/web-admin/lib/api.ts` | 修改 | 新增 Social Post API client 与 TS types |
| `apps/web-admin/app/social-post-pack/page.tsx` | 新增 | Pack 管理页骨架 |
| `apps/web-admin/app/projects/[projectId]/social-post/page.tsx` | 新增 | 项目 Social 配置与生成页骨架 |
| `apps/web-admin/app/projects/[projectId]/social-post/variants/page.tsx` | 新增 | 候选文案管理页骨架 |
| `apps/web-admin/app/projects/[projectId]/social-post/assets/page.tsx` | 新增 | 资产页骨架 |
| `apps/web-admin/app/global-nav.tsx` | 修改 | 追加 Social Post Pack 导航入口 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 修改 | 追加项目工作区 Social 导航入口 |
| `.cube/iterations/feature-13/skeleton-map.yaml` | 新增 | 记录骨架文件与 Development Tasks 的覆盖关系 |
| `apps/web-admin/app/external-automation/n8n/page.tsx` | 修改 | 修复既存 TS 类型错误（对齐 API 实际签名），确保 `npx tsc --noEmit` 通过 |

## 9. Development Tasks

- Task-01：定义 Social Post DTO、错误常量和 Service / Store 接口
  - 任务类型：contract
  - 所属模块：api-server/socialpost
  - 简要描述：定义 Pack、配置、生成、候选文案、资产相关 DTO，以及 Service / Store 接口、事务输入输出、构造函数和错误常量，供 handler、OpenAPI 和测试编译引用。
  - 涉及接口/方法：socialpost.Service、socialpost.Store、NewService()、NewMemoryStore()、NewPostgresStore()
  - 输入：各 API request DTO、Store 查询参数、事务输入对象
  - 输出：各 API response DTO、错误常量、接口签名
  - 依赖任务：无
  - 数据操作：无
  - 修改边界：只新增 `dto.go`、`errors.go`、`service.go`、`store.go`、`memory_store.go`、`postgres_store.go` 的骨架定义
  - 禁止行为：不得写业务逻辑；不得访问数据库或外部系统
  - 产出类型：integration
  - 功能类型：Social Post 模块契约定义（type id: integration）
  - 是否跨组件：否

- Task-02：创建 Social Post 持久化表迁移
  - 任务类型：migration
  - 所属模块：api-server/migrations
  - 简要描述：新增 00014 迁移文件，创建 `social_post_extension`、`social_post_variant`、`social_post_asset` 三张表、索引与 selected 唯一约束。
  - 涉及接口/方法：无（DDL 文件）
  - 输入：无
  - 输出：迁移 SQL 文件
  - 依赖任务：无
  - 数据操作：无（DDL 定义）
  - 修改边界：只新增 `00014_create_social_post_tables.sql`
  - 禁止行为：不得修改已有迁移文件；不得重定义 `metric_template`、`content_version`、`idempotency_record`
  - 产出类型：integration
  - 功能类型：Social Post 数据模型定义（type id: integration）
  - 是否跨组件：否

- Task-03：实现 Social Post Pack 注册与状态查询后端链路
  - 任务类型：api
  - 所属模块：api-server/http
  - 简要描述：实现 Pack 状态查询与注册接口，完成 schema/workflows/metrics/current_version 输出、注册参数校验、幂等处理和默认指标模板注册。
  - 涉及接口/方法：GetPackStatus()、RegisterPack()、Service.GetPackStatus()、Service.RegisterPack()
  - 输入：RegisterSocialPostPackRequest、Idempotency-Key
  - 输出：SocialPostPackStatusResponse、RegisterSocialPostPackResponse
  - 依赖任务：Task-01（接口契约）
  - 数据操作：读 `content_type`、`workflow_template`、`workflow_template_version`、`metric_template`；写 `content_type`、`workflow_template`、`workflow_template_version`、`metric_template`；读写 `idempotency_record`
  - 修改边界：只在 `social_post.go`、`service.go`、`router.go`、`server.go`、`lib/api.ts`、`openapi/openapi.yaml` 的 Social Post 相关位置新增或替换空实现
  - 禁止行为：不得在 Handler 中写业务逻辑；不得新增独立 pack registration 表替代既有推导模型
  - 产出类型：web-e2e
  - 功能类型：Pack 注册与查看 HTTP 端点（type id: web-e2e）
  - 是否跨组件：是（组件链路：SocialPostHandler -> socialpost.Service -> content.Service -> workflow.Service -> metrics.Service -> socialpost.Store）

- Task-04：实现项目 Social 配置查询与更新链路
  - 任务类型：api
  - 所属模块：api-server/http
  - 简要描述：实现项目配置查询与更新接口，返回默认配置、维护 `config_version`、写入 `operation_log` 与幂等记录。
  - 涉及接口/方法：GetConfig()、UpdateConfig()、Service.GetConfig()、Service.UpdateConfig()
  - 输入：projectId、UpdateSocialPostConfigRequest、Idempotency-Key
  - 输出：SocialPostConfigResponse、UpdateSocialPostConfigResponse
  - 依赖任务：Task-01、Task-02
  - 数据操作：读 `content_project`、`social_post_extension`；写 `social_post_extension`、`operation_log`；读写 `idempotency_record`
  - 修改边界：只在 Social Post 相关 handler/service/store/OpenAPI/API client 位置新增或替换空实现
  - 禁止行为：不得直接把默认配置硬编码在前端；不得省略审计记录
  - 产出类型：web-e2e
  - 功能类型：项目配置 HTTP 端点（type id: web-e2e）
  - 是否跨组件：是（组件链路：SocialPostHandler -> socialpost.Service -> socialpost.Store）

- Task-05：实现短内容生成触发链路
  - 任务类型：business-implementation
  - 所属模块：api-server/socialpost
  - 简要描述：实现短内容生成触发，校验输入、创建或关联 ContentItem、创建 WorkflowRun、提交异步执行、记录幂等和审计。
  - 涉及接口/方法：CreateGenerationRun()、Service.CreateGenerationRun()
  - 输入：CreateSocialPostGenerationRunRequest、Idempotency-Key
  - 输出：CreateSocialPostGenerationRunResponse
  - 依赖任务：Task-01、Task-02
  - 数据操作：读 `content_project`、`content_type`、`social_post_extension`；写 `content_item`、`generation_run`、`workflow_run`、`operation_log`；读写 `idempotency_record`；调用 `engine.Submitter`
  - 修改边界：只在 Social Post 相关 handler/service/store/router/server/OpenAPI/API client 位置实现；不得重写既有 generation 模块
  - 禁止行为：不得同步阻塞等待生成结果；不得绕开 workflow engine 直接伪造 trace 数据
  - 产出类型：integration
  - 功能类型：短内容生成业务实现（type id: integration）
  - 是否跨组件：是（组件链路：SocialPostHandler -> socialpost.Service -> content.Service -> workflow.Service -> engine.Submitter）

- Task-06：实现生成详情查询与 trace 汇总链路
  - 任务类型：api
  - 所属模块：api-server/socialpost
  - 简要描述：实现生成详情查询，聚合 workflow 状态、候选文案结果和 trace 摘要，对外输出 `running/completed/failed`。
  - 涉及接口/方法：GetGenerationRun()、Service.GetGenerationRun()
  - 输入：projectId、generationRunID
  - 输出：SocialPostGenerationRunDetailResponse
  - 依赖任务：Task-01、Task-05
  - 数据操作：读 `generation_run`、`workflow_run`、`social_post_variant`、`agent_task`、`llm_call_log`
  - 修改边界：只在 Social Post 相关 handler/service/store/OpenAPI/API client 位置实现
  - 禁止行为：不得忽略失败 trace；不得直接暴露内部状态枚举而不做对外映射
  - 产出类型：web-e2e
  - 功能类型：生成详情 HTTP 端点（type id: web-e2e）
  - 是否跨组件：是（组件链路：SocialPostHandler -> socialpost.Service -> socialpost.Store -> workflow/agent/llm read model）

- Task-07：实现候选文案结果持久化与列表查询链路
  - 任务类型：business-implementation
  - 所属模块：api-server/socialpost
  - 简要描述：实现异步生成成功后的 `social_post_variant` 写入、按条件分页查询候选文案列表。
  - 涉及接口/方法：InsertVariant()、ListVariants()、Service.ListVariants()
  - 输入：ListSocialPostVariantsRequest
  - 输出：PagedSocialPostVariantsResponse
  - 依赖任务：Task-01、Task-02、Task-05
  - 数据操作：写 `social_post_variant`；读 `social_post_variant`
  - 修改边界：只在 `service.go`、`store.go`、`postgres_store.go`、`memory_store.go` 与 Social Post handler/OpenAPI/API client 位置实现
  - 禁止行为：不得覆盖历史版本；不得以前端 mock 数据替代真实列表查询
  - 产出类型：integration
  - 功能类型：候选文案查询业务实现（type id: integration）
  - 是否跨组件：是（组件链路：SocialPostHandler -> socialpost.Service -> socialpost.Store）

- Task-08：实现主版本选择事务链路
  - 任务类型：business-implementation
  - 所属模块：api-server/socialpost
  - 简要描述：实现主版本选择的事务逻辑，归档旧主选、创建 `content_version`、更新新主选并写审计。
  - 涉及接口/方法：SelectVariant()、Service.SelectVariant()、Store.SelectVariantInTx()
  - 输入：variantId、SelectSocialPostVariantRequest、Idempotency-Key
  - 输出：SelectSocialPostVariantResponse
  - 依赖任务：Task-01、Task-02、Task-07
  - 数据操作：读 `social_post_variant`；写 `social_post_variant`、`content_version`、`operation_log`；读写 `idempotency_record`
  - 修改边界：只在 Social Post 相关 handler/service/store/OpenAPI/API client 位置实现
  - 禁止行为：不得在事务外更新 selected 状态；不得绕过 `content_version` 创建直接返回空版本 ID
  - 产出类型：integration
  - 功能类型：主版本选择业务实现（type id: integration）
  - 是否跨组件：是（组件链路：SocialPostHandler -> socialpost.Service -> socialpost.Store）

- Task-09：实现标签生成与封面文案生成触发链路
  - 任务类型：business-implementation
  - 所属模块：api-server/socialpost
  - 简要描述：实现 tags / cover_copy 两类资产异步生成触发，创建 WorkflowRun、写幂等与审计，并返回运行标识。
  - 涉及接口/方法：GenerateTags()、GenerateCoverCopy()
  - 输入：GenerateSocialPostTagsRequest、GenerateSocialPostCoverCopyRequest、Idempotency-Key
  - 输出：GenerateSocialPostAssetResponse
  - 依赖任务：Task-01、Task-02、Task-07
  - 数据操作：读 `social_post_variant`、`content_item`；写 `workflow_run`、`generation_run`、`operation_log`；读写 `idempotency_record`；调用 `engine.Submitter`
  - 修改边界：只在 Social Post 相关 handler/service/store/router/server/OpenAPI/API client 位置实现
  - 禁止行为：不得同步生成并阻塞 HTTP；不得省略 traceable workflow run
  - 产出类型：integration
  - 功能类型：资产生成业务实现（type id: integration）
  - 是否跨组件：是（组件链路：SocialPostHandler -> socialpost.Service -> workflow.Service -> engine.Submitter -> socialpost.Store）

- Task-10：实现资产结果持久化与查询链路
  - 任务类型：business-implementation
  - 所属模块：api-server/socialpost
  - 简要描述：实现 tags / cover_copy 结果写入 `social_post_asset`，并按项目、平台、版本查询资产及 `asset_suggestions`。
  - 涉及接口/方法：InsertAsset()、GetAssets()、Service.GetAssets()
  - 输入：GetSocialPostAssetsRequest
  - 输出：SocialPostAssetsResponse
  - 依赖任务：Task-01、Task-02、Task-09
  - 数据操作：写 `social_post_asset`；读 `social_post_asset`
  - 修改边界：只在 Social Post 相关 handler/service/store/OpenAPI/API client 位置实现
  - 禁止行为：不得把资产结果仅保存在组件状态或内存临时变量中
  - 产出类型：integration
  - 功能类型：资产查询业务实现（type id: integration）
  - 是否跨组件：是（组件链路：SocialPostHandler -> socialpost.Service -> socialpost.Store）

- Task-11：更新 OpenAPI Social Post 契约
  - 任务类型：contract
  - 所属模块：openapi
  - 简要描述：为 11 个 Social Post API 增加路径、参数、Envelope schema、trace 字段和错误响应定义。
  - 涉及接口/方法：`openapi.yaml` 路径与 schema
  - 输入：无
  - 输出：OpenAPI 路径定义与组件定义
  - 依赖任务：Task-03、Task-04、Task-05、Task-06、Task-08、Task-09、Task-10
  - 数据操作：无
  - 修改边界：只修改 `openapi/openapi.yaml` 的 Social Post 相关内容
  - 禁止行为：不得删除已有路径；不得改写现有统一 Envelope schema
  - 产出类型：none
  - 功能类型：API 文档契约（type id: none）
  - 是否跨组件：否

- Task-12：实现 Social Post API 客户端与类型定义
  - 任务类型：contract
  - 所属模块：web-admin/api
  - 简要描述：在 `lib/api.ts` 中增加 Social Post TS 类型、API client 函数、幂等 header 透传和错误结构映射。
  - 涉及接口/方法：`fetchSocialPostPackStatus()`、`registerSocialPostPack()`、`fetchSocialPostConfig()`、`updateSocialPostConfig()` 等
  - 输入：前端页面请求参数
  - 输出：TS 类型和 API 调用函数
  - 依赖任务：Task-03、Task-04、Task-05、Task-06、Task-08、Task-09、Task-10
  - 数据操作：调用 Social Post HTTP API
  - 修改边界：只修改 `apps/web-admin/lib/api.ts` 的 Social Post 相关内容
  - 禁止行为：不得删除已有 API 类型；不得引入绕过统一 Envelope 的请求方式
  - 产出类型：frontend-ui
  - 功能类型：前端 API 契约（type id: frontend-ui）
  - 是否跨组件：否

- Task-13：实现 Social Post Pack 管理页面
  - 任务类型：ui
  - 所属模块：web-admin
  - 简要描述：新增 Pack 管理页，支持加载状态、404 空态、错误态、成功态和注册按钮交互。
  - 涉及接口/方法：`app/social-post-pack/page.tsx`
  - 输入：无
  - 输出：Pack 状态展示页
  - 依赖任务：Task-03、Task-12、Task-17
  - 数据操作：调用 `GET/POST /api/v1/content-packs/social-post/*`
  - 修改边界：只新增 `app/social-post-pack/page.tsx`；只在 `global-nav.tsx` 追加入口
  - 禁止行为：不得用静态文案替代真实 API 加载；不得删除现有导航项
  - 产出类型：frontend-ui
  - 功能类型：Pack 管理前端页面（type id: frontend-ui）
  - 是否跨组件：否

- Task-14：实现项目 Social 配置与生成页面
  - 任务类型：ui
  - 所属模块：web-admin
  - 简要描述：新增项目级 Social 页面，包含配置表单、生成表单、运行状态和 trace 展示区域。
  - 涉及接口/方法：`app/projects/[projectId]/social-post/page.tsx`
  - 输入：projectId
  - 输出：配置与生成一体页
  - 依赖任务：Task-04、Task-05、Task-06、Task-12、Task-17
  - 数据操作：调用配置查询/更新、生成创建、生成详情查询 API
  - 修改边界：只新增 `app/projects/[projectId]/social-post/page.tsx`；只在 `workspace-nav.tsx` 追加入口
  - 禁止行为：不得省略加载态、空态、错误态；不得把 run 结果写死在页面
  - 产出类型：frontend-ui
  - 功能类型：项目 Social 配置与生成页（type id: frontend-ui）
  - 是否跨组件：否

- Task-15：实现候选文案管理页面
  - 任务类型：ui
  - 所属模块：web-admin
  - 简要描述：新增候选文案分页页，支持按 `content_item` / `status` / `platform` 过滤并触发主版本选择。
  - 涉及接口/方法：`app/projects/[projectId]/social-post/variants/page.tsx`
  - 输入：projectId、筛选条件
  - 输出：候选文案表格页
  - 依赖任务：Task-07、Task-08、Task-12、Task-17
  - 数据操作：调用 `GET variants`、`POST select`
  - 修改边界：只新增 `app/projects/[projectId]/social-post/variants/page.tsx`；只在 `workspace-nav.tsx` 追加入口
  - 禁止行为：不得用 mock 数据填充主表；不得省略分页结构与主选结果反馈
  - 产出类型：frontend-ui
  - 功能类型：候选文案前端页面（type id: frontend-ui）
  - 是否跨组件：否

- Task-16：实现资产管理页面
  - 任务类型：ui
  - 所属模块：web-admin
  - 简要描述：新增资产页，支持查看 tags / cover_copy 结果、查看 `asset_suggestions` 并触发两类资产生成。
  - 涉及接口/方法：`app/projects/[projectId]/social-post/assets/page.tsx`
  - 输入：projectId、content_item_id、variant_id、platform
  - 输出：资产列表与触发表单
  - 依赖任务：Task-09、Task-10、Task-12、Task-17
  - 数据操作：调用 `POST tags:generate`、`POST cover-copy:generate`、`GET assets`
  - 修改边界：只新增 `app/projects/[projectId]/social-post/assets/page.tsx`；只在 `workspace-nav.tsx` 追加入口
  - 禁止行为：不得把结果仅保存在组件常量中；不得跳过真实 API 查询层
  - 产出类型：frontend-ui
  - 功能类型：资产管理前端页面（type id: frontend-ui）
  - 是否跨组件：否

- Task-17：接入全局与项目导航入口
  - 任务类型：ui
  - 所属模块：web-admin/navigation
  - 简要描述：在全局导航中新增 Social Post Pack 入口，在项目工作区导航中新增 Social Post、Variants、Assets 入口，确保四个页面可直接访问。
  - 涉及接口/方法：`GlobalNav()`、`ProjectWorkspaceNav()`
  - 输入：projectId、当前 pathname
  - 输出：导航链接
  - 依赖任务：无
  - 数据操作：无
  - 修改边界：只修改 `app/global-nav.tsx`、`app/projects/[projectId]/workspace-nav.tsx`
  - 禁止行为：不得移除现有导航项；不得更改现有匹配逻辑语义
  - 产出类型：frontend-ui
  - 功能类型：前端导航接入（type id: frontend-ui）
  - 是否跨组件：否
