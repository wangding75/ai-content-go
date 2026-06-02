# Iteration 11 技术设计：平台适配器与浏览器插件

## 1. 概述

本次设计采用用户确认的 Extend modules 方案：平台 Adapter 配置、插件客户端认证和插件发布任务协作扩展既有 `publish` 模块；平台采集日志与人工确认扩展 `metrics` 模块；n8n 外围回调扩展 `external` 模块；管理台继续复用 `AppLayout`、`global-nav`、`lib/api.ts` 和现有 CSS class。

核心约束：

- 新增能力复用现有 `/api/v1` envelope、错误码、`request_id` 和分页约定。
- Adapter 只保存 `credential_ref`，创建/更新时校验引用可用性，不保存平台凭证明文。
- 插件 `api_key` 只在注册/轮换成功响应中返回一次，服务端只保存 hash 和 masked 值。
- 插件发布协作必须绑定既有 `publish_job`，不得新增第二套发布主状态机。
- 插件领取任务必须具备原子锁契约，同一任务并发领取最多一个成功。
- 平台采集结果默认只保存为 `platform_collect_log`，人工确认后才写入 `metric_record`。
- n8n 回调只处理外围事件，必须拒绝创建 WorkflowRun、推进 Agent、绕过发布状态机或修改内容正文的 payload。

## 2. Impact Analysis

| 模块 | 影响程度 | 说明 |
|------|----------|------|
| `apps/api-server/internal/modules/publish` | 修改 | 新增 Adapter 配置、插件客户端、插件认证、插件任务领取/回填 DTO、Service 方法和内存状态。 |
| `apps/api-server/internal/http/handlers/publish.go` | 修改 | 新增 Adapter 管理、插件客户端、插件认证、插件发布任务 handler。 |
| `apps/api-server/internal/modules/metrics` | 修改 | 新增平台采集日志 DTO、Service 方法和 store 接口。 |
| `apps/api-server/internal/modules/metrics/store.go` | 修改 | 增加 collect log 与确认写入所需 store 方法。 |
| `apps/api-server/internal/modules/metrics/memory_store.go` | 修改 | 实现 collect log 内存存储、幂等和确认状态更新。 |
| `apps/api-server/internal/modules/metrics/postgres_store.go` | 修改 | 增加 collect log PostgreSQL 读写骨架/契约常量。 |
| `apps/api-server/internal/http/handlers/metrics.go` | 修改 | 新增平台采集日志提交、列表、详情、确认 handler。 |
| `apps/api-server/internal/modules/external` | 修改 | 新增外围回调 DTO、边界校验、回调日志列表和幂等处理。 |
| `apps/api-server/internal/http/handlers/external.go` | 修改 | 新增回调接收、回调日志列表和测试回调 handler。 |
| `apps/api-server/internal/http/router.go` | 修改 | 将 `/api/v1` 拆成管理端 bearer、插件 api_key、插件 bearer、外部回调 auth 等不同路由组。 |
| `apps/api-server/migrations/00013_create_platform_adapter_extension_tables.sql` | 新增 | 新建 Adapter、插件客户端、插件 token、采集日志、回调日志表，并扩展 publish 表。 |
| `apps/api-server/internal/store/platform_adapter_sql_contract_test.go` | 新增 | SQL 契约测试，约束 DDL、索引、状态枚举、原子锁 SQL 和敏感字段禁止模式。 |
| `apps/api-server/internal/http/contract/iteration11_platform_adapter_contract_red_test.go` | 新增 | Web/API 契约测试。 |
| `openapi/openapi.yaml` | 修改 | 增加 Iteration 11 API paths、schemas、认证和错误响应。 |
| `apps/web-admin/lib/api.ts` | 修改 | 新增 Adapter、插件客户端、采集日志、回调日志 API client 类型和函数。 |
| `apps/web-admin/app/global-nav.tsx` | 修改 | 新增平台 Adapter、插件客户端、采集日志导航入口，保留 n8n 入口。 |
| `apps/web-admin/app/platform-adapters/page.tsx` | 新增 | 平台 Adapter 管理页面。 |
| `apps/web-admin/app/plugin-clients/page.tsx` | 新增 | 插件客户端管理页面。 |
| `apps/web-admin/app/platform-collect-logs/page.tsx` | 新增 | 采集日志列表、详情和人工确认页面。 |
| `apps/web-admin/app/external-automation/n8n/page.tsx` | 修改 | 增加外围边界说明、回调日志列表和测试回调反馈。 |
| `apps/web-admin/e2e/iteration11-platform-adapter-extension.spec.ts` | 新增 | 前端 UI 与前后端链路 E2E 测试。 |

### 兼容性分析

- 现有 API：新增 endpoint，不删除既有 paths；`publish_job` 原有人工复制/发布能力继续可用。
- 现有数据：新增表和 nullable/default 字段；修改 `publish_log.event_type` 约束时必须保留旧事件值。
- 认证：管理端继续使用现有 admin bearer；插件认证、插件任务和外部回调使用独立 middleware，不依赖 admin bearer。
- 前端：新增路由和导航项，已有页面 URL 不变。

## 3. Flow Design

### 3.1 Adapter 配置管理流程

1. 管理员在 `/platform-adapters` 提交平台编码、发布模式、目标类型、字段映射、填充规则、采集规则、`credential_ref` 和启用状态。
2. `PublishHandler.CreatePlatformAdapter` 解码请求并调用 `publish.Service.CreatePlatformAdapter()`。
3. Service 校验平台编码、发布模式、JSON 规则结构、敏感字段禁止模式，并校验 `credential_ref` 指向可用的外部自动化 Provider/Binding 或凭证引用；不可用凭证返回 `VALIDATION_ERROR`，无权限访问凭证返回 `FORBIDDEN`。
4. Service 写入 `platform_adapter_config`，版本从 1 开始，写 `operation_log`。
5. 列表和详情从 `platform_adapter_config` 读取，按平台、发布模式、启用状态分页筛选。
6. 编辑/启停校验 `expected_version`。当 `enabled=false` 时，Service 通过 `publish_job.adapter_config_id` 查询该 Adapter 关联任务；对旧数据再用 `publish_target.platform + publish_target.target_type` 回退匹配。若仍有 `queued`、`copied`、有效锁定或不可迁移任务，返回 `CONFLICT`，不更新 Adapter。
7. 通过校验后更新配置版本，写 `platform_adapter_revision` 与 `operation_log`。

异常流程：平台编码缺失、规则 JSON 不合法、凭证引用不可用返回 `VALIDATION_ERROR`；凭证无权访问返回 `FORBIDDEN`；版本冲突或停用时仍有不可处理任务返回 `CONFLICT`；不存在返回 `NOT_FOUND`。

### 3.2 插件客户端注册、密钥轮换与短期认证流程

1. 管理员在 `/plugin-clients` 注册 Chrome 插件客户端。
2. `publish.Service.RegisterPluginClient()` 校验名称、版本和 scope，生成 `api_key_once`，只保存 `api_key_hash` 与 `api_key_masked`。
3. 页面只在创建或轮换成功弹窗中展示 `api_key_once`；刷新或关闭后不可再次查看。
4. 插件调用 `POST /api/v1/plugin-auth/token`，该 endpoint 不使用 admin bearer，只接收 `api_key` 和 `client_version`。
5. Service 根据 hash 查找启用客户端，校验版本和 scope，认证失败写审计，认证成功签发短期 `access_token` 并保存 `token_hash`、scope 和过期时间。
6. 禁用客户端后，新 token 签发失败；插件任务接口每次都校验 token hash、过期时间、客户端状态和 scope。

异常流程：api_key 无效、客户端禁用、版本不兼容或 scope 不足返回 `UNAUTHORIZED` 或 `FORBIDDEN`，错误信息不得泄露密钥校验细节。

### 3.3 插件发布任务协作流程

1. publish_job 在创建时即确定 Adapter：优先由请求中的 `adapter_config_id` 指定；未指定时根据 `publish_target.platform + publish_target.target_type` 精确匹配唯一启用 Adapter 并写入 `publish_job.adapter_config_id`、`adapter_version`。无法匹配或多重匹配时，创建任务返回 `VALIDATION_ERROR`，不得留到插件领取时猜测。
2. 插件使用 access_token 拉取可处理 publish_job，Service 按 token scope、Adapter 启用状态、项目、平台、目标类型和状态过滤；只返回已有 `adapter_config_id` 且 Adapter 仍启用的任务。
3. 插件领取任务时，Service 使用原子锁获取规则：输入 `job_id`、`client_id`、`lock_ttl_seconds`、当前时间和 token scope；谓词为任务状态可领取、Adapter 启用、客户端授权、没有有效锁或锁已过期；实现必须使用单条条件 `UPDATE ... WHERE ... AND (locked_until IS NULL OR locked_until < now()) RETURNING ...` 或事务内 `SELECT ... FOR UPDATE`，禁止事务外 read-then-write。
4. 原子领取成功时写入 `plugin_lock_id`、`plugin_client_id`、`locked_until`、`adapter_config_id`、`adapter_version`，返回填充载荷、payload_hash、content_version_id、adapter_id 和 adapter_version；并发领取同一任务时只能一个成功，其他返回 `CONFLICT`。
5. 插件填充页面后提交 filled 事件，Service 同时校验 `lock_id`、`plugin_client_id`、payload_hash、任务状态和锁未过期，只记录填充事件并将任务推进到 `copied`，不视为发布完成。
6. 用户在平台人工确认发布后，插件提交 published 结果；Service 校验幂等键、lock_id、plugin_client_id、payload_hash、external_url 和状态流转，更新 publish_job 为 `published`，写 publish_log 和 operation_log。
7. 插件提交 failed 结果时，Service 校验失败原因和幂等键，更新 publish_job 为 `failed`，记录 retryable 和平台错误摘要。
8. 锁过期后允许其他同 scope 插件重新领取；filled/published/failed 必须同时校验 lock_id 和 plugin_client_id。

异常流程：锁冲突、锁无效或过期返回 `CONFLICT`；Adapter 禁用或 scope 不足返回 `FORBIDDEN`；幂等请求体不一致返回 `IDEMPOTENCY_CONFLICT`。

### 3.4 平台采集日志与指标确认流程

1. 插件或外围自动化提交平台采集日志，包含项目、平台、目标账号、`publish_job_id`、external_url、raw_payload、parsed_metrics 和 collected_at；`source_type=external_callback` 时还必须提供 `binding_id` 或 `X-External-Binding-Id`。由于既有 `metric_record.publish_job_id` 为必填，任何采集日志必须绑定有效 `publish_job_id`；缺失或无效 `publish_job_id` 直接返回 `VALIDATION_ERROR` 且不持久化。
2. `POST /api/v1/platform-collect-logs` 接受两种认证：插件 bearer token（source_type=`extension`）或外部自动化 callback token/signature（source_type=`external_callback`）。外部自动化提交必须用 `binding_id`/`X-External-Binding-Id` 定位 binding；token 模式校验该 binding 的 `callback_token_hash`，signature 模式用该 binding 的 `signing_secret_ref` 验签。
3. `metrics.Service.SubmitPlatformCollectLog()` 校验认证来源与 source_type 匹配、binding 身份、边界字段、parsed_metrics 结构；Service 必须从 `publish_job -> publish_target` 派生 `content_item_id`、`content_version_id`、`target_id`、`content_type` 和 platform 上下文，并校验请求显式字段不得冲突；只有具备有效 publish_job 上下文的解析失败或指标校验失败才持久化 error_summary。
4. 管理台按项目、平台、状态筛选采集日志，详情展示 raw_payload、parsed_metrics、关联对象和错误摘要。
5. 用户人工确认后，`ConfirmPlatformCollectLogMetrics()` 校验日志状态、指标模板匹配、确认记录字段完整性和重复写入，复用既有 MetricRecord 写入语义，记录 source_type=`extension` 或 `external_callback`、source_ref=`platform_collect_log:{id}`。
6. 确认成功后采集日志状态变为 `confirmed`，返回 metric_record_ids 和 operation_log_id。

异常流程：认证缺失返回 `UNAUTHORIZED`；来源越权返回 `FORBIDDEN`；模板不匹配返回 `VALIDATION_ERROR`；重复确认返回 `CONFLICT`；幂等冲突返回 `IDEMPOTENCY_CONFLICT`；日志不存在返回 `NOT_FOUND`。

### 3.5 n8n 外围回调边界流程

1. n8n 或外部自动化 Provider 调用 `POST /api/v1/external-automation/callbacks`，使用 binding token 或签名认证，`binding_id` 和 stable event ID 不能替代认证。
2. `external.Service.ReceiveCallback()` 校验 binding、认证、事件类型、payload schema 和幂等键。
3. 事件类型只允许 `notification.sent`、`external_sync.completed`、`alert.raised`、`platform_collect.submitted`、`lightweight_collect.completed`。
4. 如果 payload 包含创建 WorkflowRun、推进 Agent、绕过发布状态机或直接修改内容正文的意图，Service 拒绝并记录 `boundary_violation=true`。
5. 成功或拒绝都写 `external_callback_log`；成功返回 accepted=true 和 callback_log_id，边界拒绝返回 `FORBIDDEN` 与 callback_log_id。
6. `/external-automation/n8n` 页面通过回调日志列表展示回调处理结果；“测试回调”使用同一真实 callback endpoint，发送 `stable_event_id=test-{binding_id}` 和测试 payload，页面展示 callback_log_id 或错误。

异常流程：未知事件类型或 schema 错误返回 `VALIDATION_ERROR`；binding 不存在返回 `NOT_FOUND`；认证失败返回 `UNAUTHORIZED`；幂等冲突返回 `IDEMPOTENCY_CONFLICT`；越界事件返回 `FORBIDDEN`。

## 4. Table Design

### 4.1 `platform_adapter_config`

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | TEXT | PRIMARY KEY | Adapter ID |
| platform | TEXT | NOT NULL | 平台编码 |
| display_name | TEXT | NOT NULL | 展示名称 |
| publish_mode | TEXT | NOT NULL CHECK (publish_mode IN ('manual_plugin', 'external_callback', 'manual_only')) | 发布模式 |
| target_type | TEXT | NOT NULL | 目标类型 |
| field_mapping | JSONB | NOT NULL DEFAULT '{}'::jsonb | 字段映射 |
| fill_rules | JSONB | NOT NULL DEFAULT '{}'::jsonb | 插件填充规则 |
| collect_rules | JSONB | NOT NULL DEFAULT '{}'::jsonb | 采集规则 |
| credential_ref | TEXT | NOT NULL DEFAULT '' | 凭证引用，不保存明文 |
| enabled | BOOLEAN | NOT NULL DEFAULT TRUE | 启用状态 |
| version | INTEGER | NOT NULL DEFAULT 1 CHECK (version > 0) | 配置版本 |
| operation_log_id | TEXT | | 最近操作日志 |
| created_at / updated_at | TIMESTAMPTZ | NOT NULL | 时间戳 |

索引：`UNIQUE(platform, target_type)`、`idx_platform_adapter_enabled`、`idx_platform_adapter_mode`。

### 4.2 `platform_adapter_revision`

字段：`id TEXT PRIMARY KEY`、`adapter_id TEXT NOT NULL`、`version INTEGER NOT NULL`、`change_reason TEXT NOT NULL`、`snapshot JSONB NOT NULL`、`operation_log_id TEXT`、`created_at TIMESTAMPTZ NOT NULL`。索引：`idx_platform_adapter_revision_adapter_version`。

### 4.3 `plugin_client`

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | TEXT | PRIMARY KEY | 客户端 ID |
| name | TEXT | NOT NULL UNIQUE | 客户端名称 |
| client_type | TEXT | NOT NULL CHECK (client_type IN ('chrome_extension')) | 客户端类型 |
| version | TEXT | NOT NULL | 客户端版本 |
| scopes | JSONB | NOT NULL DEFAULT '[]'::jsonb | 权限范围 |
| status | TEXT | NOT NULL CHECK (status IN ('enabled', 'disabled')) | 状态 |
| api_key_hash | TEXT | NOT NULL UNIQUE | 密钥哈希 |
| api_key_masked | TEXT | NOT NULL | 脱敏展示 |
| last_active_at | TIMESTAMPTZ | | 最后活跃 |
| operation_log_id | TEXT | | 最近操作日志 |
| created_at / updated_at | TIMESTAMPTZ | NOT NULL | 时间戳 |

索引：`idx_plugin_client_status`、`idx_plugin_client_last_active`。禁止 `api_key`、`api_key_plain`、`password`、`secret_plain` 明文字段。

### 4.4 `plugin_access_token`

字段：`id TEXT PRIMARY KEY`、`client_id TEXT NOT NULL`、`token_hash TEXT NOT NULL UNIQUE`、`scopes JSONB NOT NULL`、`expires_at TIMESTAMPTZ NOT NULL`、`revoked_at TIMESTAMPTZ`、`created_at TIMESTAMPTZ NOT NULL`。索引：`idx_plugin_access_token_client_expires`、`idx_plugin_access_token_hash`。

### 4.5 `publish_target` / `publish_job` 增补字段与事件

`publish_target` 新增字段：`target_type TEXT NOT NULL DEFAULT 'default'`，用于 Adapter 精确映射；`CreatePublishTargetRequest` / `UpdatePublishTargetRequest` 增加 `target_type`，目标响应暴露 `target_type`。

`publish_job` 新增字段：`plugin_lock_id TEXT`、`plugin_client_id TEXT`、`locked_until TIMESTAMPTZ`、`adapter_config_id TEXT`、`adapter_version INTEGER`、`filled_at TIMESTAMPTZ`、`platform_error_summary TEXT NOT NULL DEFAULT ''`。`CreatePublishJobRequest` 增加可选 `adapter_config_id`；未传时 `CreateJob` 必须根据目标的 `platform + target_type` 解析唯一启用 Adapter 并写入 `adapter_config_id` 和 `adapter_version`；历史 job 允许为空，插件列表/锁定只处理可映射到启用 Adapter 的 job。

新增索引：`idx_publish_target_platform_type`、`idx_publish_job_plugin_lock`、`idx_publish_job_adapter_status`、`idx_publish_job_locked_until`。

`publish_log.event_type` 约束必须保留旧值并增补 `plugin_locked`、`plugin_filled`、`plugin_published`、`plugin_failed`。迁移必须通过 drop/recreate CHECK 或等效安全方式修改约束，不得丢失旧数据。

### 4.6 `platform_collect_log`

字段：`id TEXT PRIMARY KEY`、`project_id TEXT NOT NULL`、`platform TEXT NOT NULL`、`target_account TEXT NOT NULL DEFAULT ''`、`publish_job_id TEXT NOT NULL`、`content_item_id TEXT NOT NULL`、`content_version_id TEXT NOT NULL`、`target_id TEXT NOT NULL`、`content_type TEXT NOT NULL`、`external_url TEXT NOT NULL DEFAULT ''`、`source_type TEXT NOT NULL CHECK (source_type IN ('extension','external_callback'))`、`raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb`、`parsed_metrics JSONB NOT NULL DEFAULT '[]'::jsonb`、`status TEXT NOT NULL CHECK (status IN ('received','parse_failed','ready','confirmed','rejected'))`、`error_summary TEXT NOT NULL DEFAULT ''`、`collected_at TIMESTAMPTZ NOT NULL`、`operation_log_id TEXT`、`created_at TIMESTAMPTZ NOT NULL`、`updated_at TIMESTAMPTZ NOT NULL`。

索引：`idx_platform_collect_log_project_status`、`idx_platform_collect_log_platform_collected`、`idx_platform_collect_log_publish_job`。

### 4.7 `external_callback_log`

`external_workflow_binding` 新增字段：`callback_auth_type TEXT NOT NULL DEFAULT 'token' CHECK (callback_auth_type IN ('token','signature'))`、`callback_token_hash TEXT NOT NULL DEFAULT ''`、`signing_secret_ref TEXT NOT NULL DEFAULT ''`、`callback_token_masked TEXT NOT NULL DEFAULT ''`。token 模式必须保存 hash/masked，不保存明文；signature 模式必须保存 secret 引用，不保存 secret 明文。

回调凭证生命周期：新增 `POST /api/v1/external-automation/bindings/{bindingId}/rotate-callback-token`，管理端 bearer required，请求 `reason*`，响应 `binding_id`、`callback_token_once`、`callback_token_masked`、`operation_log_id`；旧 token hash 立即失效。新增 `PATCH /api/v1/external-automation/bindings/{bindingId}/callback-auth`，管理端 bearer required，请求 `callback_auth_type*`、`signing_secret_ref`、`change_reason*`，用于切换 token/signature 模式和校验 secret 引用。OpenAPI、前端 API client、n8n 页面和契约测试必须覆盖 token 一次性展示、hash/masked 持久化和 invalid callback auth。

字段：`id TEXT PRIMARY KEY`、`provider_id TEXT`、`binding_id TEXT NOT NULL`、`event_type TEXT NOT NULL`、`idempotency_key TEXT NOT NULL`、`request_hash TEXT NOT NULL`、`payload JSONB NOT NULL`、`accepted BOOLEAN NOT NULL`、`rejected_reason TEXT NOT NULL DEFAULT ''`、`boundary_violation BOOLEAN NOT NULL DEFAULT FALSE`、`created_at TIMESTAMPTZ NOT NULL`。

索引：`uq_external_callback_idempotency(binding_id, idempotency_key)`、`idx_external_callback_binding_created`、`idx_external_callback_event_accepted`。

## 5. API Design

所有响应使用 `api.WriteSuccess` / `api.WriteError` envelope。

### 5.0 认证矩阵与路由分组

| Endpoint group | Auth mode | Router 设计 | 失败错误 |
|----------------|-----------|-------------|----------|
| Adapter/client 管理 API | admin bearer | `/api/v1` 管理组使用现有 `bearerAuth` | `UNAUTHORIZED`、`FORBIDDEN` |
| `POST /api/v1/plugin-auth/token` | api_key only | 在 admin bearer middleware 外注册，只执行 api_key 解码和 Service 校验 | `UNAUTHORIZED`、`FORBIDDEN` |
| `/api/v1/plugin/publish-jobs...` | plugin bearer token | 插件路由组使用 plugin token middleware，校验 token_hash、过期、client status、scope | `UNAUTHORIZED`、`FORBIDDEN` |
| `POST /api/v1/platform-collect-logs` | plugin bearer 或 external callback auth | 根据 source_type 选择插件 token 或外部签名/binding token middleware | `UNAUTHORIZED`、`FORBIDDEN` |
| n8n callback endpoint | provider/binding token or signature | 外部回调路由组使用 callback auth，不使用 admin bearer | `UNAUTHORIZED`、`FORBIDDEN` |
| 管理台 collect/callback list/detail/confirm | admin bearer | `/api/v1` 管理组 | `UNAUTHORIZED`、`FORBIDDEN` |

### 5.1 Adapter 管理 API

- `POST /api/v1/platform-adapters`
  - 请求：`platform*`、`display_name*`、`publish_mode*`、`target_type*`、`field_mapping*`、`fill_rules*`、`collect_rules*`、`credential_ref`、`enabled`
  - 响应：`adapter_id:string`、`version:int`、`operation_log_id:string`
  - 正确性规则：校验 publish_mode 枚举、JSON 规则结构、credential_ref 可用性、敏感字段禁止模式；重复 platform/target_type 返回冲突。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`CONFLICT`、`INTERNAL_ERROR`
- `GET /api/v1/platform-adapters`
  - 查询：`platform`、`publish_mode`、`enabled`、`page`、`page_size`、`sort`、`order`
  - 响应：`items:PlatformAdapterResponse[]`、`pagination`
  - 正确性规则：sort 仅允许 `platform`、`updated_at`、`version`；分页遵循现有 content pagination。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`
- `GET /api/v1/platform-adapters/{adapterId}`
  - 响应：`PlatformAdapterDetailResponse`，包含配置、规则摘要、版本和更新时间。
  - 错误码：`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`
- `PATCH /api/v1/platform-adapters/{adapterId}`
  - 请求：可编辑字段、`expected_version*`、`change_reason*`
  - 响应：`adapter_id:string`、`version:int`、`operation_log_id:string`
  - 正确性规则：expected_version 必须匹配；credential_ref 更新时必须重新校验；停用前必须查询关联 publish_job，存在 queued/copied/有效锁定任务时返回 `CONFLICT`。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`INTERNAL_ERROR`

### 5.2 插件客户端与认证 API

- `POST /api/v1/plugin-clients`
  - 请求：`name*`、`client_type*`、`version*`、`scopes*`
  - 响应：`client_id:string`、`api_key_once:string`、`api_key_masked:string`
  - 正确性规则：api_key_once 只在本响应返回；scopes 只能包含 `publish:read`、`publish:write`、`collect:write`。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`CONFLICT`、`INTERNAL_ERROR`
- `GET /api/v1/plugin-clients`
  - 查询：`status`、`client_type`、`page`、`page_size`、`sort`、`order`
  - 响应：`items:PluginClientResponse[]`、`pagination`，只包含 `api_key_masked`。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`
- `PATCH /api/v1/plugin-clients/{clientId}`
  - 请求：`status`、`scopes`、`change_reason*`
  - 响应：`client_id:string`、`status:string`、`operation_log_id:string`
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`INTERNAL_ERROR`
- `POST /api/v1/plugin-clients/{clientId}/rotate-key`
  - 请求：`reason*`
  - 响应：`client_id:string`、`api_key_once:string`、`api_key_masked:string`、`operation_log_id:string`
  - 正确性规则：旧 api_key_hash 立即失效；禁用客户端不能轮换。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`INTERNAL_ERROR`
- `POST /api/v1/plugin-auth/token`
  - 请求：`api_key*`、`client_version*`
  - 响应：`access_token:string`、`expires_at:string`、`scopes:string[]`
  - 正确性规则：不要求 admin bearer；认证失败不泄露 hash、是否存在、版本细节。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`

### 5.3 插件发布任务 API

- `GET /api/v1/plugin/publish-jobs`
  - 查询：`project_id`、`platform`、`status`、`page`、`page_size`
  - 响应：`items:PluginPublishJobResponse[]`、`pagination`
  - 正确性规则：只返回 token scope 内且 Adapter 启用的任务；无任务返回空列表。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`
- `POST /api/v1/plugin/publish-jobs/{jobId}/lock`
  - 请求：`lock_ttl_seconds*`
  - 响应：`lock_id:string`、`locked_until:string`、`payload:object`、`payload_hash:string`、`content_version_id:string`、`adapter_config_id:string`、`adapter_version:int`
  - 正确性规则：使用原子锁；并发领取最多一个成功；Adapter 禁用或任务状态不可领取返回错误。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`INTERNAL_ERROR`
- `POST /api/v1/plugin/publish-jobs/{jobId}/filled`
  - 请求：`lock_id*`、`payload_hash*`、`note`
  - 响应：`event_id:string`、`current_status:string`
  - 正确性规则：校验 lock_id、plugin_client_id、payload_hash、未过期锁；不直接标记 published。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`INTERNAL_ERROR`
- `POST /api/v1/plugin/publish-jobs/{jobId}/published`
  - 请求：`lock_id*`、`external_url*`、`published_at*`、`payload_hash*`、`note`，Header `Idempotency-Key*`
  - 响应：`publish_job_id:string`、`current_status:string`、`operation_log_id:string`
  - 正确性规则：重复相同 idempotency key 返回相同结果；请求体不同返回冲突；外部 URL 必须合法。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`
- `POST /api/v1/plugin/publish-jobs/{jobId}/failed`
  - 请求：`lock_id*`、`reason*`、`retryable`、`platform_error_summary`，Header `Idempotency-Key*`
  - 响应：`publish_job_id:string`、`current_status:string`、`operation_log_id:string`
  - 正确性规则：reason 必填；重复提交不重复写 publish_log。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`

### 5.4 平台采集日志 API

- `POST /api/v1/platform-collect-logs`
  - 请求：`project_id*`、`platform*`、`target_account`、`publish_job_id*`、`binding_id`（external_callback 必填，可由 `X-External-Binding-Id` header 提供）、`external_url`、`source_type*`、`raw_payload*`、`parsed_metrics`、`collected_at*`
  - 响应：`collect_log_id:string`、`status:string`
  - 正确性规则：source_type 与认证方式匹配；external_callback 必须用 binding_id 定位 callback auth；publish_job_id 必须有效并派生 content/version/target/content_type；无有效 publish_job_id 返回 VALIDATION_ERROR 且不持久化；失败采集仅在 publish_job 有效时持久化 error_summary。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`CONFLICT`、`INTERNAL_ERROR`
- `GET /api/v1/platform-collect-logs`
  - 查询：`project_id`、`platform`、`status`、`page`、`page_size`、`sort`、`order`
  - 响应：`items:PlatformCollectLogResponse[]`、`pagination`
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`
- `GET /api/v1/platform-collect-logs/{collectLogId}`
  - 响应：`PlatformCollectLogDetailResponse`，包含 raw_payload、parsed_metrics、关联对象和错误摘要。
  - 错误码：`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`
- `POST /api/v1/platform-collect-logs/{collectLogId}/confirm-metrics`
  - 请求：`records*`、`note`，Header `Idempotency-Key*`；每条 record 包含 `metric_template_id*`、`metric_code*`、`metric_date*`、`period*`、`raw_value*`、`normalized_value`、`unit`。
  - 响应：`metric_record_ids:string[]`、`operation_log_id:string`
  - 正确性规则：只允许 ready 日志确认；每条 record 必须匹配 metric_template；最终写入 metric_record 的 `project_id`、`content_item_id`、`content_version_id`、`publish_job_id`、`target_id`、`content_type`、`platform`、`external_url` 必须来自 collect log 的 publish_job 派生上下文，不允许确认请求覆盖；source_ref 必须为 collect log；重复确认返回冲突或相同幂等结果。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`

### 5.5 n8n 外围回调 API

- `POST /api/v1/external-automation/callbacks`
  - 请求：`binding_id*`、`event_type*`、`payload*`、`stable_event_id*`
  - 响应：`accepted:boolean`、`callback_log_id:string`
  - 正确性规则：stable_event_id 作为幂等键；binding token/signature 必须通过 `external_workflow_binding.callback_token_hash` 或 `signing_secret_ref` 校验；越界 payload 被记录并拒绝。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`
- `GET /api/v1/external-automation/callback-logs`
  - 查询：`provider_id`、`binding_id`、`event_type`、`accepted`、`page`、`page_size`、`sort`、`order`
  - 响应：`items:ExternalCallbackLogResponse[]`、`pagination`
  - 正确性规则：sort 仅允许 `created_at`、`event_type`、`accepted`；管理端 bearer required。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`
- `POST /api/v1/external-automation/bindings/{bindingId}/rotate-callback-token`
  - 请求：`reason*`
  - 响应：`binding_id:string`、`callback_token_once:string`、`callback_token_masked:string`、`operation_log_id:string`
  - 正确性规则：callback_token_once 只在本响应返回；服务端只保存 hash/masked；旧 token 立即失效。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`INTERNAL_ERROR`
- `PATCH /api/v1/external-automation/bindings/{bindingId}/callback-auth`
  - 请求：`callback_auth_type*`、`signing_secret_ref`、`change_reason*`
  - 响应：`binding_id:string`、`callback_auth_type:string`、`operation_log_id:string`
  - 正确性规则：signature 模式必须校验 signing_secret_ref 可用；token 模式必须已有 callback_token_hash 或先 rotate token。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`INTERNAL_ERROR`
- `POST /api/v1/external-automation/callbacks/test`
  - 请求：`binding_id*`、`event_type*`、`payload*`
  - 响应：`accepted:boolean`、`callback_log_id:string`
  - 正确性规则：管理端 bearer required；内部复用真实 ReceiveCallback 流程，stable event ID 为 `test-{binding_id}-{request_id}`。
  - 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`

## 6. Module Design

### 6.1 Publish 模块扩展

新增 DTO、常量和 Service 方法，继续由 `publish.NewService()` 返回同一 Service：

```go
CreatePlatformAdapter(ctx context.Context, req CreatePlatformAdapterRequest, idempotencyKey string) (CreatePlatformAdapterResponse, error)
ListPlatformAdapters(ctx context.Context, req ListPlatformAdaptersRequest) (PagedPlatformAdaptersResponse, error)
GetPlatformAdapter(ctx context.Context, adapterID string) (PlatformAdapterDetailResponse, error)
UpdatePlatformAdapter(ctx context.Context, adapterID string, req UpdatePlatformAdapterRequest, idempotencyKey string) (UpdatePlatformAdapterResponse, error)
RegisterPluginClient(ctx context.Context, req RegisterPluginClientRequest, idempotencyKey string) (RegisterPluginClientResponse, error)
ListPluginClients(ctx context.Context, req ListPluginClientsRequest) (PagedPluginClientsResponse, error)
UpdatePluginClient(ctx context.Context, clientID string, req UpdatePluginClientRequest, idempotencyKey string) (UpdatePluginClientResponse, error)
RotatePluginClientKey(ctx context.Context, clientID string, req RotatePluginClientKeyRequest, idempotencyKey string) (RotatePluginClientKeyResponse, error)
AuthenticatePlugin(ctx context.Context, req PluginAuthRequest) (PluginAuthTokenResponse, error)
ListPluginPublishJobs(ctx context.Context, req ListPluginPublishJobsRequest, token string) (PagedPluginPublishJobsResponse, error)
LockPluginPublishJob(ctx context.Context, jobID string, req LockPluginPublishJobRequest, token string) (PluginPublishJobLockResponse, error)
MarkPluginPublishJobFilled(ctx context.Context, jobID string, req MarkPluginPublishJobFilledRequest, token string) (PluginPublishJobFilledResponse, error)
MarkPluginPublishJobPublished(ctx context.Context, jobID string, req MarkPluginPublishJobPublishedRequest, token string, idempotencyKey string) (PluginPublishJobPublishedResponse, error)
MarkPluginPublishJobFailed(ctx context.Context, jobID string, req MarkPluginPublishJobFailedRequest, token string, idempotencyKey string) (PluginPublishJobFailedResponse, error)
```

关键 DTO 字段：

| DTO | 字段 |
|-----|------|
| `PlatformAdapterResponse` | `id`、`platform`、`display_name`、`publish_mode`、`target_type`、`enabled`、`version`、`updated_at` |
| `PlatformAdapterDetailResponse` | `PlatformAdapterResponse` 字段 + `field_mapping`、`fill_rules`、`collect_rules`、`credential_ref`、`rule_summary` |
| `PluginClientResponse` | `id`、`name`、`client_type`、`version`、`scopes`、`status`、`api_key_masked`、`last_active_at` |
| `PluginPublishJobResponse` | `id`、`project_id`、`platform`、`target_display`、`status`、`payload_hash`、`locked_until`、`adapter_config_id`、`adapter_version` |
| `PluginPublishJobLockResponse` | `lock_id`、`locked_until`、`payload`、`payload_hash`、`content_version_id`、`adapter_config_id`、`adapter_version` |

正确性规则：Adapter credential_ref 必须可用；插件 key 永不明文存储；插件 token 校验客户端状态；插件锁必须原子；published/failed 必须幂等。

### 6.2 Metrics 模块扩展

新增平台采集日志 DTO 与 Service 方法：

```go
SubmitPlatformCollectLog(ctx context.Context, req SubmitPlatformCollectLogRequest, auth PlatformCollectLogAuth, idempotencyKey string) (SubmitPlatformCollectLogResponse, error)
ListPlatformCollectLogs(ctx context.Context, req ListPlatformCollectLogsRequest) (PagedPlatformCollectLogsResponse, error)
GetPlatformCollectLog(ctx context.Context, collectLogID string) (PlatformCollectLogDetailResponse, error)
ConfirmPlatformCollectLogMetrics(ctx context.Context, collectLogID string, req ConfirmPlatformCollectLogMetricsRequest, idempotencyKey string) (ConfirmPlatformCollectLogMetricsResponse, error)
```

关键 DTO 字段：`SubmitPlatformCollectLogRequest` 包含 `project_id`、`platform`、`publish_job_id`、`binding_id`、`source_type`、`raw_payload`、`parsed_metrics`、`collected_at`；`PlatformCollectLogAuth` 包含 `source_type`、`plugin_token`、`binding_id`、`callback_auth_header`、`signature_header`。`PlatformCollectLogResponse` 包含 `id`、`project_id`、`platform`、`status`、`publish_job_id`、`content_item_id`、`external_url`、`error_summary`、`collected_at`；`PlatformCollectLogDetailResponse` 增加 `raw_payload`、`parsed_metrics` 和关联对象摘要。Store 方法包括 insert/list/get/update collect log、confirm collect log、check/store idempotency、find metric template、insert metric record。

### 6.3 External 模块扩展

新增回调 DTO 与 Service 方法：

```go
RotateCallbackToken(ctx context.Context, bindingID string, req RotateCallbackTokenRequest, idempotencyKey string) (RotateCallbackTokenResponse, error)
UpdateCallbackAuth(ctx context.Context, bindingID string, req UpdateCallbackAuthRequest, idempotencyKey string) (UpdateCallbackAuthResponse, error)
ReceiveCallback(ctx context.Context, req ExternalCallbackRequest, auth ExternalCallbackAuth, idempotencyKey string) (ExternalCallbackResponse, error)
ListCallbackLogs(ctx context.Context, req ListCallbackLogsRequest) (PagedExternalCallbackLogsResponse, error)
TestCallback(ctx context.Context, req TestExternalCallbackRequest) (ExternalCallbackResponse, error)
```

`ExternalCallbackLogResponse` 包含 `id`、`provider_id`、`binding_id`、`event_type`、`accepted`、`rejected_reason`、`boundary_violation`、`created_at`。边界校验只记录外围事件，不直接调用 Workflow/Agent/Publish 状态推进。

### 6.4 HTTP 层

- `PublishHandler` 增加 Adapter、插件客户端、插件认证、插件任务 handler。
- `MetricsHandler` 增加采集日志提交、列表、详情、确认 handler。
- `ExternalHandler` 增加回调接收、回调日志列表和测试回调 handler。
- `router.go` 拆分 route groups：管理端 bearer 组、plugin-auth 无 admin bearer 组、plugin bearer 组、external callback auth 组。

### 6.5 前端管理台

- `lib/api.ts` 增加类型和函数，继续使用 `request<T>` 和 `pathSegment()`。
- `global-nav.tsx`：Task-08 添加 `/platform-adapters` 和 `/plugin-clients`；Task-09 添加 `/platform-collect-logs`，保留 `/external-automation/n8n`。
- 新页面必须参考 `docs/requirements/ai-content-factory-clickable-prototype.html`：Adapter 页面复用平台 Adapter 原型区域的配置卡片/表格/规则摘要；n8n 页面复用外部自动化原型区域的 Provider/Binding/回调状态布局；插件客户端和采集日志页面沿用同一管理台卡片、表格、筛选、弹窗和状态标签风格。
- 新页面使用 `page-shell`、`page-hero`、`card`、`badge`、`form-grid`、`action-row` 样式，必须具备加载态、空态、错误态、成功态和 request_id 展示。

## 7. Output Contract

`workflow.yaml` 当前 `project.features` 为空，但本次迭代实际新增 HTTP API、前端页面、跨组件链路和 SQL DDL/查询契约，因此触发以下类型化测试：

| 类型 | type id | 适用范围 | 测试规范 |
|------|---------|----------|----------|
| REST/OpenAPI/handler contract | web-e2e | 新增 HTTP endpoint、OpenAPI、handler -> service -> envelope | `standards/testing/web-e2e.md` |
| Frontend route/page state/rendering | frontend-ui | 新增或修改管理台页面、导航、CSS、交互和 API client roundtrip | `standards/testing/frontend-ui.md` |
| Cross-service workflow/state transitions | integration | 插件锁定/回填、采集确认写指标、n8n 边界日志链路 | `standards/testing/integration.md` |
| PostgreSQL DDL/query/migration contract | sql-query | 00013 migration、状态约束、索引、原子锁 SQL、敏感字段禁止模式 | `standards/testing/sql-query.md` |

### 7.1 API 与公共方法契约表

| Endpoint / Method | 输入 | 输出 | Auth | type id | 正确性规则 | 副作用 |
|-------------------|------|------|------|---------|------------|--------|
| `CreatePlatformAdapter` / `POST /platform-adapters` | Adapter config + Idempotency-Key | adapter_id/version/operation_log_id | admin bearer | web-e2e | 校验 credential_ref、规则 JSON、唯一性 | 写 adapter、revision、operation_log |
| `UpdatePlatformAdapter` / `PATCH /platform-adapters/{id}` | editable fields、expected_version、change_reason | adapter_id/version/operation_log_id | admin bearer | web-e2e | 版本匹配；停用前无 queued/copied/locked jobs | 更新 adapter、写 revision/log |
| `RegisterPluginClient` / `POST /plugin-clients` | name/client_type/version/scopes | client_id/api_key_once/api_key_masked | admin bearer | web-e2e | api_key_once 只返回一次，只存 hash | 写 plugin_client、operation_log |
| `AuthenticatePlugin` / `POST /plugin-auth/token` | api_key/client_version | access_token/expires_at/scopes | api_key | web-e2e | 不要求 admin bearer；禁用/版本不兼容失败 | 写 token、审计失败 |
| `LockPluginPublishJob` / `POST /plugin/publish-jobs/{id}/lock` | token/job_id/ttl | lock_id/payload/payload_hash/adapter_version | plugin bearer | integration | 原子锁；并发最多一个成功 | 更新 publish_job、写 publish_log |
| `MarkPluginPublishJobPublished` | lock_id/url/payload_hash/idempotency | status/operation_log_id | plugin bearer | integration | 校验 lock+client+payload；幂等 | 更新 publish_job、写 publish_log/log |
| `SubmitPlatformCollectLog` | collect payload + required publish_job_id + external binding_id/header | collect_log_id/status | plugin/external auth | web-e2e | source_type 与认证匹配；external binding_id 定位 callback auth；publish_job 派生 metric_record 上下文后才进入 ready；无效 publish_job 不持久化 | 写 collect_log |
| `ConfirmPlatformCollectLogMetrics` | records(metric_template_id/metric_code/date/period/raw_value)/note/idempotency | metric_record_ids/operation_log_id | admin bearer | integration | ready 状态、模板匹配、metric_record 必需字段完整、source_ref 绑定日志 | 更新 collect_log、写 metric_record/log |
| `RotateCallbackToken` / `POST /external-automation/bindings/{id}/rotate-callback-token` | reason | callback_token_once/masked/operation_log_id | admin bearer | web-e2e | token 只返回一次，只存 hash/masked，旧 token 失效 | 更新 binding auth、写 operation_log |
| `UpdateCallbackAuth` / `PATCH /external-automation/bindings/{id}/callback-auth` | auth_type/signing_secret_ref/change_reason | binding_id/auth_type/operation_log_id | admin bearer | web-e2e | signature secret ref 校验；token 模式必须已有 token hash | 更新 binding auth、写 operation_log |
| `ReceiveCallback` | binding/event/payload/stable_event_id + token/signature | accepted/callback_log_id | callback auth | web-e2e | 校验 callback_token_hash 或 signing_secret_ref、schema、边界、幂等 | 写 callback_log |
| `ListCallbackLogs` / `GET /callback-logs` | filters/pagination | items/pagination | admin bearer | web-e2e | 只读日志，分页排序白名单 | 无 |
| Frontend pages | 用户点击/表单/筛选 | styled UI state | dev token/admin API | frontend-ui | loading/empty/error/success、真实 API roundtrip、CSS applied | 调用 API |

### 7.2 SQL Contract

- 目标方言：PostgreSQL。
- 迁移文件：`apps/api-server/migrations/00013_create_platform_adapter_extension_tables.sql`。
- 必须包含：`platform_adapter_config`、`platform_adapter_revision`、`plugin_client`、`plugin_access_token`、`platform_collect_log`、`external_callback_log`。
- 必须修改：`publish_target` 新增 `target_type`；`publish_job` 新增插件锁和 Adapter 字段；`publish_log.event_type` 支持插件事件且保留旧事件；`external_workflow_binding` 新增 callback auth hash/ref 字段。
- 关键约束：所有状态/枚举列必须同时具备 `NOT NULL` 与 `CHECK (... IN (...))`；`plugin_client.api_key_hash` 和 `plugin_access_token.token_hash` 必须 UNIQUE；`external_callback_log` 必须有 `(binding_id, idempotency_key)` 唯一约束；`external_workflow_binding.callback_token_hash`/`signing_secret_ref` 不得存明文 token/secret；`platform_collect_log` 必须包含写入 metric_record 所需的 publish_job、content version、target 和 content type 上下文字段。SQL contract test 必须拒绝 nullable 状态/认证枚举列。
- Sort 白名单：Adapter list 允许 `platform`,`updated_at`,`version`；plugin client list 允许 `name`,`status`,`last_active_at`,`updated_at`；collect log list 允许 `collected_at`,`status`,`platform`；callback log list 允许 `created_at`,`event_type`,`accepted`。
- 原子锁 SQL 契约：必须包含条件 `UPDATE publish_job SET plugin_lock_id = ..., plugin_client_id = ..., locked_until = ... WHERE id = $1 AND status IN ('queued','copied','failed') AND (locked_until IS NULL OR locked_until < $now) RETURNING ...` 或 `SELECT ... FOR UPDATE` 事务等效结构；禁止事务外 read-then-write。
- 禁止模式：字符串拼接生成 SQL；动态 ORDER BY 未白名单；明文 `api_key` / `password` / `secret` 持久化列；无 CHECK 或缺少 NOT NULL 的状态/枚举字段。
- 典型输入输出：注册插件客户端只在 API 响应出现 `api_key_once`，迁移和 DTO 只保留 hash/masked；采集日志确认后生成 `metric_record_ids`，source_ref=`platform_collect_log:{id}`。

## 8. Change Log

| 文件 | 类型 | 原因 |
|------|------|------|
| `apps/api-server/migrations/00013_create_platform_adapter_extension_tables.sql` | 新增 | 定义 Adapter、插件客户端、插件 token、插件锁、采集日志和回调日志数据结构。 |
| `apps/api-server/internal/store/platform_adapter_sql_contract_test.go` | 新增 | 约束 SQL 契约、原子锁和敏感字段禁止模式。 |
| `apps/api-server/internal/modules/publish/dto.go` | 修改 | 增加 Adapter、插件客户端、插件认证和插件任务 DTO/常量。 |
| `apps/api-server/internal/modules/publish/service.go` | 修改 | 扩展 Service 接口、内存状态和骨架方法。 |
| `apps/api-server/internal/modules/publish/errors.go` | 修改 | 复用/补充插件认证、凭证校验和锁相关错误映射。 |
| `apps/api-server/internal/http/handlers/publish.go` | 修改 | 增加 Adapter、插件客户端、插件认证和插件任务 HTTP handler。 |
| `apps/api-server/internal/modules/metrics/dto.go` | 修改 | 增加平台采集日志 DTO。 |
| `apps/api-server/internal/modules/metrics/service.go` | 修改 | 扩展 Service 接口和采集日志确认骨架方法。 |
| `apps/api-server/internal/modules/metrics/store.go` | 修改 | 增加 collect log 与确认写入 store 接口方法。 |
| `apps/api-server/internal/modules/metrics/memory_store.go` | 修改 | 实现 collect log 内存存储和幂等状态骨架。 |
| `apps/api-server/internal/modules/metrics/postgres_store.go` | 修改 | 增加 collect log PostgreSQL 查询/写入骨架或契约常量。 |
| `apps/api-server/internal/http/handlers/metrics.go` | 修改 | 增加平台采集日志 HTTP handler。 |
| `apps/api-server/internal/modules/external/dto.go` | 修改 | 增加外部回调请求、响应、测试回调和日志 DTO。 |
| `apps/api-server/internal/modules/external/service.go` | 修改 | 扩展 Service 接口、边界校验和回调日志方法。 |
| `apps/api-server/internal/http/handlers/external.go` | 修改 | 增加回调接收、回调日志列表和测试回调 handler。 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册新增 API routes 和不同认证路由组。 |
| `openapi/openapi.yaml` | 修改 | 增加新增 API 契约。 |
| `apps/web-admin/lib/api.ts` | 修改 | 增加前端 API client 类型和函数。 |
| `apps/web-admin/app/global-nav.tsx` | 修改 | 增加新增管理页面导航入口。 |
| `apps/web-admin/app/platform-adapters/page.tsx` | 新增 | Adapter 管理 UI。 |
| `apps/web-admin/app/plugin-clients/page.tsx` | 新增 | 插件客户端管理 UI。 |
| `apps/web-admin/app/platform-collect-logs/page.tsx` | 新增 | 平台采集日志 UI。 |
| `apps/web-admin/app/external-automation/n8n/page.tsx` | 修改 | 补充外围回调边界、回调日志和测试回调。 |
| `apps/web-admin/e2e/iteration11-platform-adapter-extension.spec.ts` | 新增 | 前端 UI 和全链路测试。 |
| `apps/api-server/internal/http/contract/iteration11_platform_adapter_contract_red_test.go` | 新增 | Web/API 契约测试。 |

## 9. Development Tasks

- Task-01：定义平台适配器与插件协作数据库契约
  - 任务类型：migration
  - 所属模块：api-server/store
  - 简要描述：创建 Adapter、插件客户端、插件 token、采集日志、外部回调日志表，扩展 publish_target 目标类型、publish_job 插件锁/Adapter 字段、publish_log 插件事件和 external binding 回调认证字段。
  - 涉及接口/方法：迁移 SQL、SQL 契约测试
  - 输入：Iteration 11 DDL 契约、状态枚举、索引、原子锁 SQL 规则、callback auth hash/ref 规则、metric_record 上下文字段规则
  - 输出：PostgreSQL migration 和 SQL contract test
  - 依赖任务：无
  - 数据操作：写 platform_adapter_config、platform_adapter_revision、plugin_client、plugin_access_token、platform_collect_log、external_callback_log 表结构；修改 publish_target；修改 publish_job；修改 publish_log 事件约束；修改 external_workflow_binding callback auth 字段
  - 修改边界：只新增 00013 migration 和 SQL 契约测试；只通过 ALTER TABLE 扩展既有 publish 表；只安全调整 publish_log event_type CHECK
  - 禁止行为：不得删除既有表；不得保存 api_key、password、secret 明文字段；不得重写旧迁移；不得使用未白名单动态 ORDER BY
  - 正确性规则：DDL 包含所有 CHECK/UNIQUE/索引；原子锁 SQL 契约可被字符串契约测试识别；外部回调认证具备 hash/ref 存储；采集日志具备 metric_record 写入上下文；敏感明文字段测试必须失败保护
  - 产出类型：sql-query
  - 功能类型：数据库迁移与敏感字段安全契约（type id: sql-query）
  - 是否跨组件：否
- Task-02：实现平台 Adapter 配置管理业务行为
  - 任务类型：business-implementation
  - 所属模块：api-server/publish
  - 简要描述：支持 Adapter 新增、列表筛选、详情、编辑、启停、credential_ref 可用性校验、停用前 publish_job 冲突检查、版本冲突和操作日志返回。
  - 涉及接口/方法：CreatePlatformAdapter()、ListPlatformAdapters()、GetPlatformAdapter()、UpdatePlatformAdapter()
  - 输入：Adapter 配置请求、查询条件、expected_version、change_reason、Idempotency-Key
  - 输出：Adapter 响应、分页响应、operation_log_id 或错误
  - 依赖任务：Task-01（platform_adapter_config 表契约）
  - 数据操作：读写 platform_adapter_config；写 platform_adapter_revision；读外部 credential/provider/binding 引用；停用前按 publish_job.adapter_config_id 与 publish_target 平台/目标类型回退匹配读取 publish_job；写 operation_log
  - 修改边界：只新增 publish DTO/Service/Handler 中 Adapter 相关类型和方法；只在 router.go 注册 Adapter routes；只以向后兼容方式为 publish target/job 增加 Adapter 映射字段
  - 禁止行为：不得保存平台凭证明文；不得新增独立 platform 模块；不得破坏既有 publish target API 语义；不得在仍有 queued/copied/locked 任务时停用 Adapter
  - 正确性规则：credential_ref 不可用返回 VALIDATION_ERROR/FORBIDDEN；expected_version 不匹配返回 CONFLICT；停用冲突返回 CONFLICT
  - 产出类型：web-e2e
  - 功能类型：平台 Adapter 管理 API（type id: web-e2e）
  - 是否跨组件：是（组件链路：Router -> PublishHandler -> PublishService -> platform_adapter_config/publish_job）
- Task-03：实现插件客户端密钥与短期认证业务行为
  - 任务类型：business-implementation
  - 所属模块：api-server/publish
  - 简要描述：支持插件客户端注册、列表、启停、scope 调整、密钥轮换和 api_key 换短期 access_token。
  - 涉及接口/方法：RegisterPluginClient()、ListPluginClients()、UpdatePluginClient()、RotatePluginClientKey()、AuthenticatePlugin()
  - 输入：客户端注册请求、状态变更请求、轮换请求、api_key、client_version
  - 输出：client_id、api_key_once、api_key_masked、access_token、expires_at、operation_log_id 或错误
  - 依赖任务：Task-01（plugin_client 和 plugin_access_token 表契约）
  - 数据操作：读写 plugin_client；写 plugin_access_token；写 operation_log；更新 last_active_at；写认证失败审计
  - 修改边界：只新增 publish DTO/Service/Handler 中插件客户端和认证相关类型和方法；只在 router.go 注册插件客户端和认证 routes
  - 禁止行为：不得保存 api_key 明文；不得泄露密钥校验细节；不得让禁用客户端认证成功；plugin-auth 不得强制 admin bearer
  - 正确性规则：api_key_once 只返回一次；旧 key 轮换后立即失效；token 过期/禁用客户端任务接口失败
  - 产出类型：web-e2e
  - 功能类型：插件客户端认证 API（type id: web-e2e）
  - 是否跨组件：是（组件链路：Router -> PublishHandler -> PublishService -> plugin_client/plugin_access_token）
- Task-04：实现插件发布任务锁定与回填业务行为
  - 任务类型：business-implementation
  - 所属模块：api-server/publish
  - 简要描述：确保 publish target/job 在创建期写入目标类型和 Adapter 映射，插件按 scope 拉取 publish_job、原子领取锁定任务、获取填充载荷、回填 filled/published/failed 并保持幂等。
  - 涉及接口/方法：ListPluginPublishJobs()、LockPluginPublishJob()、MarkPluginPublishJobFilled()、MarkPluginPublishJobPublished()、MarkPluginPublishJobFailed()
  - 输入：CreatePublishTarget target_type、CreatePublishJob adapter_config_id、插件 access_token、任务筛选条件、lock_ttl_seconds、lock_id、payload_hash、发布或失败回填请求、Idempotency-Key
  - 输出：目标 target_type、任务 adapter_config_id/adapter_version、任务列表、lock_id、填充载荷、事件 ID、状态变更结果、operation_log_id 或错误
  - 依赖任务：Task-01（publish_job 插件锁字段）、Task-02（Adapter 可用性）、Task-03（插件 token 与 scope）
  - 数据操作：读 plugin_access_token 和 plugin_client；读 platform_adapter_config；以原子条件更新 publish_job 锁字段；读写 publish_job 状态；写 publish_log；写 operation_log；读写幂等记录
  - 修改边界：只扩展 publish Service、Handler、DTO 和 router 中插件任务相关方法；只追加 publish_job 状态辅助逻辑
  - 禁止行为：不得新增第二套发布主状态机；不得绕过 publish_job；不得使用事务外 read-then-write 领取锁；不得在 lock/client/payload 无效时推进状态；不得重复处理相同幂等键
  - 正确性规则：新建 target 暴露 target_type；新建 job 持久化 adapter_config_id/adapter_version；并发领取同一 job 最多一个成功；filled 不等于 published；published/failed 校验 lock_id 与 plugin_client_id；幂等请求体冲突返回 IDEMPOTENCY_CONFLICT
  - 产出类型：integration
  - 功能类型：插件发布任务协作链路（type id: integration）
  - 是否跨组件：是（组件链路：Plugin API -> PublishHandler -> PublishService -> publish_job/publish_log）
- Task-05：实现平台采集日志与人工确认指标业务行为
  - 任务类型：business-implementation
  - 所属模块：api-server/metrics
  - 简要描述：接收平台采集日志、支持列表和详情、人工确认后写入 MetricRecord 并保留来源。
  - 涉及接口/方法：SubmitPlatformCollectLog()、ListPlatformCollectLogs()、GetPlatformCollectLog()、ConfirmPlatformCollectLogMetrics()、Store collect log methods
  - 输入：采集日志请求、external callback binding_id/header、筛选条件、collect_log_id、确认指标记录、Idempotency-Key
  - 输出：collect_log_id、解析状态、采集日志详情、metric_record_ids、operation_log_id 或错误
  - 依赖任务：Task-01（platform_collect_log 表契约）
  - 数据操作：插入/列表/读取/更新 platform_collect_log；读 metric_template；写 metric_record；写 operation_log；读写幂等记录；确认时更新 collect_log 状态
  - 修改边界：只扩展 metrics DTO/Service/Handler/router；只修改 store.go、memory_store.go、postgres_store.go 中 collect log 所需接口与实现；只复用既有 MetricRecord 写入语义
  - 禁止行为：不得默认自动写入 metric_record；不得丢弃失败采集摘要；不得重复确认污染指标；不得绕过 MetricRecord uniqueness/source 规则；不得整文件重写 store 文件
  - 正确性规则：source_type 与认证方式匹配；external_callback 必须有 binding_id/header 定位 binding；缺失或无效 publish_job_id 返回 VALIDATION_ERROR 且不持久化；ready 状态必须绑定有效 publish_job；确认请求不得覆盖 publish_job 派生上下文；metric_template 不匹配返回 VALIDATION_ERROR；source_ref 必须绑定 collect log
  - 产出类型：integration
  - 功能类型：采集日志确认写入指标链路（type id: integration）
  - 是否跨组件：是（组件链路：MetricsHandler -> MetricsService -> platform_collect_log -> metric_record）
- Task-06：实现 n8n 外围回调边界业务行为
  - 任务类型：business-implementation
  - 所属模块：api-server/external
  - 简要描述：接收 n8n 外围回调，支持 binding callback token 轮换与 signature secret 引用配置，校验 binding、认证、事件类型、schema、幂等键和越界 payload，记录回调日志并提供回调日志列表和测试回调。
  - 涉及接口/方法：RotateCallbackToken()、UpdateCallbackAuth()、ReceiveCallback()、ListCallbackLogs()、TestCallback()
  - 输入：binding_id、callback auth 类型、callback token 轮换原因、signing_secret_ref、event_type、payload、stable_event_id、Idempotency-Key、回调日志筛选条件
  - 输出：callback_token_once/masked、accepted、callback_log_id、回调日志列表或错误
  - 依赖任务：Task-01（external_callback_log 表契约）
  - 数据操作：读写 external binding callback auth 字段；读 external provider；读写 external_callback_log；读写幂等记录
  - 修改边界：只扩展 external DTO/Service/Handler 和 router 中回调相关方法；只记录外围事件
  - 禁止行为：不得创建 WorkflowRun；不得推进 Agent 编排；不得修改内容正文；不得直接改 publish_job 主状态；不得把 stable_event_id 当认证
  - 正确性规则：callback token 只返回一次且只存 hash/masked；signing_secret_ref 必须可校验；binding token/signature 必须有效；越界 payload 返回 FORBIDDEN 且记录 boundary_violation；callback list 支持分页筛选
  - 产出类型：web-e2e
  - 功能类型：n8n 外围回调 API（type id: web-e2e）
  - 是否跨组件：是（组件链路：ExternalHandler -> ExternalService -> external_callback_log）
- Task-07：实现 Iteration 11 OpenAPI 与前端 API Client 契约
  - 任务类型：contract
  - 所属模块：openapi/web-admin
  - 简要描述：为新增 API 补充 OpenAPI paths/schemas/security，并在 `lib/api.ts` 增加类型和请求函数。
  - 涉及接口/方法：openapi paths、PlatformAdapter API client、PluginClient API client、PluginPublishJob API client、PlatformCollectLog API client、ExternalCallback API client、ExternalCallbackAuth API client
  - 输入：API Design 中的请求、响应、认证和错误契约
  - 输出：OpenAPI schema 和 TypeScript API client 函数
  - 依赖任务：Task-02、Task-03、Task-04、Task-05、Task-06 的接口契约
  - 数据操作：无
  - 修改边界：只修改 openapi/openapi.yaml 和 apps/web-admin/lib/api.ts 中 Iteration 11 相关定义
  - 禁止行为：不得改变既有 API client 函数签名；不得删除旧 OpenAPI paths；不得使用 any 表示新增公共响应；不得遗漏 callback-log 和 test callback API
  - 正确性规则：每个 endpoint 有 request/response/error/security 定义；platform collect external_callback 的 binding_id/header 完整建模；callback token rotate/update auth API 完整建模；TypeScript 公共响应类型显式命名
  - 产出类型：web-e2e
  - 功能类型：API 与前端 client 契约（type id: web-e2e）
  - 是否跨组件：是（组件链路：OpenAPI -> lib/api -> Go API）
- Task-08：实现管理台平台 Adapter 与插件客户端页面交互
  - 任务类型：ui
  - 所属模块：web-admin
  - 简要描述：实现 `/platform-adapters` 和 `/plugin-clients` 页面，支持列表、表单、详情/状态操作、密钥一次性展示、加载态、空态、错误态和成功反馈。
  - 涉及接口/方法：PlatformAdaptersPage、PluginClientsPage、global-nav
  - 输入：用户筛选、表单字段、状态操作、密钥轮换操作
  - 输出：页面渲染、Toast/Alert、一次性 api_key 弹窗、导航高亮
  - 依赖任务：Task-07（前端 API client）
  - 数据操作：调用 Adapter 和 plugin client HTTP API
  - 修改边界：只新增两个页面；只为 `/platform-adapters`、`/plugin-clients` 追加 global-nav 项；只使用现有 CSS class 和 API client
  - 禁止行为：不得裸 HTML 无样式；不得刷新 404；不得在页面状态中持久化 api_key_once 超出弹窗生命周期；不得偏离 prototype 管理台视觉
  - 正确性规则：页面基于 clickable prototype 管理台卡片/表格/表单风格；loading/empty/error/success 全覆盖；错误态展示 code/message/request_id
  - 产出类型：frontend-ui
  - 功能类型：管理台 Adapter 与插件客户端页面（type id: frontend-ui）
  - 是否跨组件：是（组件链路：Next Page -> lib/api -> Go API）
- Task-09：实现管理台采集日志与 n8n 回调页面交互
  - 任务类型：ui
  - 所属模块：web-admin
  - 简要描述：实现 `/platform-collect-logs` 页面并增强 `/external-automation/n8n`，支持采集日志筛选、详情、人工确认、回调凭证轮换/签名配置、回调边界说明、测试回调和回调日志反馈。
  - 涉及接口/方法：PlatformCollectLogsPage、ExternalAutomationN8NPage、global-nav
  - 输入：采集日志筛选、详情点击、确认指标记录、callback token 轮换、signature secret 配置、测试回调、回调日志筛选
  - 输出：列表、详情、错误摘要、确认结果、callback_token_once、回调日志、callback_log_id 和边界提示
  - 依赖任务：Task-07（前端 API client）
  - 数据操作：调用 platform_collect_log 和 external callback/callback-log HTTP API
  - 修改边界：只新增采集日志页面；只修改 n8n 页面 Iteration 11 区域；只为 `/platform-collect-logs` 追加 global-nav 项并保留既有 n8n 入口
  - 禁止行为：不得默认自动确认指标；不得隐藏失败采集记录；不得让按钮点击无反馈；不得偏离 prototype 管理台视觉
  - 正确性规则：页面基于 clickable prototype 的外部自动化和管理台风格；采集失败可筛选；确认失败展示 request_id；测试回调展示 callback_log_id 或错误
  - 产出类型：frontend-ui
  - 功能类型：采集日志与 n8n 外围回调页面（type id: frontend-ui）
  - 是否跨组件：是（组件链路：Next Page -> lib/api -> Go API）
