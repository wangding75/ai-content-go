# Iteration 2.1 Design：接口契约与调度基线补齐

## 1. 概述

本次设计在现有 Go API Server、in-memory workflow/agent/llm 服务和 Next.js Web Admin 基础上做最小扩展，补齐调度基线、外部自动化、成本汇总、历史页面真实渲染和 E2E 稳定性。核心约束是：本迭代只做手动触发和试跑闭环，不实现真实 cron 自动触发；ProductionPlan 的每日生成数量默认 5 且可配置；n8n 只做外围自动化；Core 命名不引入 Novel / Book / Chapter。

整体方案：

- 后端复用现有 `/api/v1`、统一 envelope、`Idempotency-Key`、operation_log_id 返回风格和 in-memory service 模式。
- schedule 模块从 contract-only 扩展为可列出、创建、启停、试跑和查看触发记录的运行时基线。
- workflow run 继续由现有 workflow service + engine.Submitter 承载，schedule 试跑通过已有 CreateRun 异步链路生成 `workflow_run_id`。
- llm 模块扩展成本汇总接口，external automation 新增轻量 in-memory provider/binding/call-log 服务。
- Web Admin 复用 `lib/api.ts` envelope 解析、错误展示和敏感信息脱敏，新增调度、外部自动化和成本汇总页面，并把历史页面统一成 AppLayout + 卡片 / 表格 / 表单视觉。
- E2E 修复允许同步调整页面 fixture 行为，目标是长生命周期内存后端重复运行稳定。

## 2. Impact Analysis

| 模块 | 影响程度 | 影响说明 |
|---|---|---|
| `apps/api-server/internal/modules/schedule` | 修改 | 扩展 DTO、Service 接口和 in-memory 存储，支持 schedule / production plan / trigger log。 |
| `apps/api-server/internal/http/handlers/schedule.go` | 修改 | 将 panic 骨架替换为列表、创建、启停、试跑、触发记录 handler。 |
| `apps/api-server/internal/http/router.go` | 修改 | 增加 workflow-schedules 的 GET、enable、disable、test-run、triggers 路由；增加 external automation 和 cost summary 路由。 |
| `apps/api-server/internal/modules/llm` | 修改 | 增加成本汇总 DTO 与 Service 方法，基于 callLogs 聚合。 |
| `apps/api-server/internal/http/handlers/llmlog.go` | 修改 | 增加 summary handler。 |
| `apps/api-server/internal/modules/external` | 新增 | 管理外部自动化 Provider、Binding、CallLog，Token 只返回脱敏值。 |
| `apps/api-server/internal/http/handlers/external.go` | 新增 | 暴露外部自动化配置 API。 |
| `apps/api-server/migrations/00004_create_iteration_2_1_tables.sql` | 新增 | 记录目标 PostgreSQL 表结构契约，当前实现仍使用 in-memory。 |
| `openapi/openapi.yaml` | 修改 | 补齐本迭代 API path、schema、security、错误响应。 |
| `apps/web-admin/lib/api.ts` | 修改 | 增加 schedule、external automation、cost summary 类型与 API 函数；调整 projects fixture 参数。 |
| `apps/web-admin/app/global-nav.tsx` | 修改 | 增加生产计划、外部自动化、成本汇总、系统页入口，保持 `aria-current`。 |
| `apps/web-admin/app/page.tsx` | 修改 | 项目列表默认真实 API，显式 `fixture=empty` 才走空态；补强 AppLayout 视觉与项目流稳定性。 |
| `apps/web-admin/app/workflow/schedules/page.tsx` | 新增 | 调度管理页面。 |
| `apps/web-admin/app/external-automation/n8n/page.tsx` | 新增 | 外部自动化 / n8n 页面。 |
| `apps/web-admin/app/llm/cost-summary/page.tsx` | 新增 | 成本汇总页面。 |
| `apps/web-admin/app/swagger-openapi/page.tsx` 与现有 Iteration 2 页面 | 修改 | 补齐样式化渲染、错误态、空态和导航一致性。 |
| `apps/web-admin/e2e/iteration1-ui.spec.ts` | 修改 | 使用唯一测试数据、显式空态 fixture、请求等待前置和列表安全断言。 |
| `apps/web-admin/e2e/iteration2-navigation.spec.ts` | 修改 | 增加本迭代新增导航页面覆盖，并确保原有页面仍 green。 |

兼容性分析：

- 现有 API 路径不删除、不改语义；仅新增路径和字段。
- `GET /api/v1/workflow-runs` 保持兼容，新增来源筛选可选。
- LLM Provider 仍只返回 `api_key_masked`，不暴露明文。
- 页面 fixture 行为改为显式 opt-in，不影响真实项目列表默认行为。
- 数据迁移为新增表，不修改已有表；in-memory 实现不依赖真实数据库。

## 3. Flow Design

### 3.1 创建调度计划

1. Web Admin 调度页面提交 project_id、template_version_id、cron_expression、daily_content_count。
2. `ScheduleHandler.CreateSchedule` 解码请求并读取 `Idempotency-Key`。
3. `schedule.Service.CreateSchedule` 校验必填字段、daily_content_count 大于 0、cron_expression 非空。
4. Service 创建 WorkflowSchedule 与 ProductionPlan in-memory 记录，默认 disabled，返回 `schedule_id`、`next_run_at`、`daily_content_count`。
5. Handler 使用统一 envelope 返回 201；校验失败返回 `VALIDATION_ERROR`，幂等冲突返回 `IDEMPOTENCY_CONFLICT`。

### 3.2 启用 / 停用调度计划

1. 用户在调度页面点击启用或停用并填写 note / reason。
2. Handler 调用 `EnableSchedule` 或 `DisableSchedule`。
3. Service 校验状态流转，已启用再次启用或已停用再次停用返回 `CONFLICT`。
4. 成功后更新 enabled、next_run_at，并返回 previous_enabled、current_enabled、operation_log_id。
5. 页面展示 Toast 并刷新列表。

### 3.3 手动试跑

1. 用户点击试跑。
2. Handler 调用 `schedule.Service.TestRunSchedule` 校验 schedule 存在并生成 trigger log。
3. Handler 使用返回的 project_id/template_version_id/input_override 调用现有 `workflow.Service.CreateRun`，source 标记为 `schedule_manual`。
4. 如果 engine.Submitter 可用则提交异步执行。
5. Handler 返回 202、`workflow_run_id` 和初始状态，并把 trigger log 关联到 run_id。
6. 页面提示可去运行记录查看。

### 3.4 成本汇总

1. 成本汇总页面按 project_id、date_from、date_to、model 请求 `GET /api/v1/llm-call-logs/summary`。
2. LLM service 从现有 callLogs 聚合 calls、tokens、cost、by_model。
3. 日期范围格式不合法返回 `VALIDATION_ERROR`；无数据返回 calls=0、by_model=[]。

### 3.5 外部自动化 / n8n

1. 用户配置 Provider 和 Binding。
2. external service 保存 provider_type、base_url、api_key_masked、binding trigger_event、webhook_url。
3. Token/API Key 明文只用于请求输入，不进入响应和页面。
4. 外部调用失败统一映射为 `EXTERNAL_AUTOMATION_ERROR`，并保留 call log 状态。

### 3.6 E2E 稳定性

1. `iteration1-ui.spec.ts` 每次运行生成唯一 runID。
2. 项目列表真实加载；空态单独通过 `fixture=empty` 访问。
3. 点击触发网络请求前先创建 `waitForResponse` promise。
4. 项目流定位包含本次 projectName 的 listitem，不点击 fallback 进入项目按钮。
5. Provider 脱敏断言在匹配的 provider 项内或 `.first()` 上执行，同时断言 body 不包含本次 secret。

## 4. Table Design

目标方言：PostgreSQL。当前开发实现使用 in-memory store，迁移文件用于契约和后续持久化。

### workflow_schedule

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | text | primary key | 调度 ID |
| project_id | text | not null | ContentProject ID |
| template_version_id | text | not null | WorkflowTemplateVersion ID |
| cron_expression | text | not null | cron 表达式，当前仅保存不自动触发 |
| enabled | boolean | not null default false | 是否启用 |
| next_run_at | timestamptz | null | 下次计划时间展示字段 |
| created_at | timestamptz | not null | 创建时间 |
| updated_at | timestamptz | not null | 更新时间 |

索引：`idx_workflow_schedule_project_id`、`idx_workflow_schedule_enabled`。

### production_plan

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | text | primary key | 生产计划 ID |
| schedule_id | text | not null | workflow_schedule.id |
| project_id | text | not null | ContentProject ID |
| daily_content_count | integer | not null check > 0 | 每日生成数量，默认 5 |
| input_template | jsonb | not null default '{}' | 生成输入模板 |
| created_at | timestamptz | not null | 创建时间 |
| updated_at | timestamptz | not null | 更新时间 |

索引：`idx_production_plan_schedule_id`。

### schedule_trigger_log

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | text | primary key | 触发日志 ID |
| schedule_id | text | not null | workflow_schedule.id |
| trigger_type | text | not null | manual_test / manual_run |
| workflow_run_id | text | null | 关联 WorkflowRun |
| status | text | not null | queued / running / succeeded / failed |
| error | text | null | 错误摘要 |
| triggered_at | timestamptz | not null | 触发时间 |

索引：`idx_schedule_trigger_log_schedule_id`、`idx_schedule_trigger_log_run_id`。

### external_workflow_provider

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | text | primary key | Provider ID |
| provider_type | text | not null | n8n 等 |
| base_url | text | not null | 外部服务 URL |
| token_masked | text | not null | 脱敏 Token |
| enabled | boolean | not null default true | 是否启用 |
| created_at | timestamptz | not null | 创建时间 |

唯一约束：`provider_type + base_url`。

### external_workflow_binding

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | text | primary key | Binding ID |
| provider_id | text | not null | Provider ID |
| trigger_event | text | not null | 触发事件 |
| webhook_url | text | not null | Webhook URL |
| enabled | boolean | not null default true | 是否启用 |
| created_at | timestamptz | not null | 创建时间 |

索引：`idx_external_binding_provider_id`、`idx_external_binding_event`。

### external_workflow_call_log

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | text | primary key | 调用日志 ID |
| provider_id | text | not null | Provider ID |
| binding_id | text | null | Binding ID |
| status | text | not null | succeeded / failed |
| error | text | null | 错误信息 |
| request_id | text | not null | 请求 ID |
| created_at | timestamptz | not null | 创建时间 |

## 5. API Design

所有 API 使用 `/api/v1`、Bearer Auth、统一 envelope、统一错误结构和 `request_id`。

| Method | Path | 成功响应 | 错误码 |
|---|---|---|---|
| GET | `/workflow-schedules` | `PagedWorkflowSchedulesResponse` | `VALIDATION_ERROR`, `INTERNAL_ERROR` |
| POST | `/workflow-schedules` | 201 `CreateScheduleResponse` | `VALIDATION_ERROR`, `IDEMPOTENCY_CONFLICT`, `INTERNAL_ERROR` |
| POST | `/workflow-schedules/{id}/enable` | `ToggleScheduleResponse` | `NOT_FOUND`, `CONFLICT`, `VALIDATION_ERROR`, `INTERNAL_ERROR` |
| POST | `/workflow-schedules/{id}/disable` | `ToggleScheduleResponse` | `NOT_FOUND`, `CONFLICT`, `VALIDATION_ERROR`, `INTERNAL_ERROR` |
| POST | `/workflow-schedules/{id}/test-run` | 202 `TestRunScheduleResponse` | `NOT_FOUND`, `VALIDATION_ERROR`, `WORKFLOW_RUN_FAILED`, `INTERNAL_ERROR` |
| GET | `/workflow-schedules/{id}/triggers` | `PagedScheduleTriggersResponse` | `NOT_FOUND`, `VALIDATION_ERROR`, `INTERNAL_ERROR` |
| GET | `/llm-call-logs/summary` | `LLMCostSummaryResponse` | `VALIDATION_ERROR`, `INTERNAL_ERROR` |
| POST | `/external-automation/providers` | 201 `CreateExternalProviderResponse` | `VALIDATION_ERROR`, `CONFLICT`, `INTERNAL_ERROR` |
| GET | `/external-automation/providers` | `PagedExternalProvidersResponse` | `VALIDATION_ERROR`, `INTERNAL_ERROR` |
| POST | `/external-automation/bindings` | 201 `CreateExternalBindingResponse` | `VALIDATION_ERROR`, `NOT_FOUND`, `CONFLICT`, `INTERNAL_ERROR` |
| GET | `/external-automation/bindings` | `PagedExternalBindingsResponse` | `VALIDATION_ERROR`, `INTERNAL_ERROR` |

### 关键 DTO

- `CreateScheduleRequest`：`project_id`, `template_version_id`, `cron_expression`, `daily_content_count`, `input_template`。
- `WorkflowScheduleResponse`：`id`, `project_id`, `template_version_id`, `cron_expression`, `enabled`, `next_run_at`, `daily_content_count`, `created_at`, `updated_at`。
- `ToggleScheduleRequest`：`reason`, `note`。
- `ToggleScheduleResponse`：`previous_enabled`, `current_enabled`, `next_run_at`, `operation_log_id`。
- `TestRunScheduleRequest`：`input_override`。
- `TestRunScheduleResponse`：`workflow_run_id`, `status`, `trigger_log_id`。
- `LLMCostSummaryResponse`：`calls`, `input_tokens`, `output_tokens`, `tokens`, `cost`, `currency`, `by_model`。
- `CreateExternalProviderRequest`：`provider_type`, `base_url`, `token`。
- `ExternalProviderResponse`：`id`, `provider_type`, `base_url`, `token_masked`, `enabled`。
- `CreateExternalBindingRequest`：`trigger_event`, `provider_id`, `webhook_url`。

## 6. Module Design

### schedule module

职责：管理 WorkflowSchedule、ProductionPlan、ScheduleTriggerLog，提供手动试跑所需的 run 输入。

接口：

```go
type Service interface {
    ListSchedules(ctx context.Context, req ListSchedulesRequest) (PagedSchedulesResponse, error)
    CreateSchedule(ctx context.Context, req CreateScheduleRequest, idempotencyKey string) (CreateScheduleResponse, error)
    EnableSchedule(ctx context.Context, id string, req ToggleScheduleRequest, idempotencyKey string) (ToggleScheduleResponse, error)
    DisableSchedule(ctx context.Context, id string, req ToggleScheduleRequest, idempotencyKey string) (ToggleScheduleResponse, error)
    PrepareTestRun(ctx context.Context, id string, req TestRunScheduleRequest) (PreparedScheduleRun, error)
    CompleteTrigger(ctx context.Context, triggerLogID, workflowRunID, status string) error
    ListTriggers(ctx context.Context, scheduleID string, req ListTriggersRequest) (PagedScheduleTriggersResponse, error)
}
```

### workflow integration

`ScheduleHandler.TestRun` 组合 `schedule.Service` 与既有 `workflow.Service`：先 `PrepareTestRun`，再 `CreateRun`，最后 `CompleteTrigger`。不在 schedule service 内直接依赖 engine，避免循环依赖。

### llm module

新增 `SummaryCallLogs(ctx, req)`，基于现有 `callLogs` 聚合。聚合只读，不改变调用日志。

### external module

新增 `external.Service`，职责为 Provider / Binding 的 in-memory CRUD 和 token masking。错误命名与现有模块一致：`ErrValidation`, `ErrNotFound`, `ErrConflict`, `ErrExternalAutomation`。

### web-admin

新增页面全部使用现有 `APIEnvelope`、`pageErrorFromEnvelope` 和 `redactSensitive`。历史页面改造不新增业务范围，只统一视觉与稳定测试入口。

## 7. Output Contract

| 产出 | 类型 | 正确性规则 | 测试规范 |
|---|---|---|---|
| Schedule API handler 链路 | web-e2e / integration | 成功、校验失败、冲突、not found、request_id、Idempotency-Key 均按 contract 返回 | `standards/testing/web-e2e.md`, `standards/testing/integration.md` |
| Schedule service | library | 默认 daily_content_count=5；非法数量失败；启停状态流转合法；trigger log 关联 run id | `standards/testing/library.md` |
| PostgreSQL migration SQL | sql-query | DDL 包含目标表、索引、约束；不存明文 token/password；字段与 Table Design 一致 | `standards/testing/sql-query.md` |
| LLM cost summary | integration | calls/tokens/cost/by_model 从 call logs 聚合；空数据为 0；日期非法失败 | `standards/testing/integration.md` |
| External automation API | web-e2e / integration | Provider token 脱敏；Binding 校验 provider；错误映射完整 | `standards/testing/web-e2e.md`, `standards/testing/integration.md` |
| Web Admin pages | web-e2e | 导航进入、刷新不 404、loading/empty/error/success、主要按钮反馈 | `standards/testing/web-e2e.md` |
| Iteration 1 E2E stability | web-e2e | 长生命周期后端重复运行不冲突；请求等待前置；fixture 显式 | `standards/testing/web-e2e.md` |

### SQL Contract

目标方言：PostgreSQL。

- 固定 DDL 文件：`apps/api-server/migrations/00004_create_iteration_2_1_tables.sql`。
- 必须包含：`workflow_schedule`、`production_plan`、`schedule_trigger_log`、`external_workflow_provider`、`external_workflow_binding`、`external_workflow_call_log`。
- 必须包含索引：schedule project/enabled、production plan schedule、trigger schedule/run、external binding provider/event。
- 禁止模式：任何 `token`、`secret`、`password` 明文字段进入外部调用日志 metadata；Provider 表只允许 `token_masked`，不允许 `token` 明文字段。
- 典型输入输出：创建 schedule 输入 `daily_content_count=5`，输出 schedule + plan；试跑输出 trigger log + workflow run id。

## 8. Change Log

| 文件 | 类型 | 原因 |
|---|---|---|
| `.cube/iterations/feature-2.1/design.md` | 新增 | 本阶段设计文档。 |
| `.cube/iterations/feature-2.1/skeleton-map.yaml` | 新增 | 记录骨架覆盖 Development Tasks。 |
| `apps/api-server/internal/modules/schedule/dto.go` | 修改 | 扩展 schedule / production plan / trigger DTO。 |
| `apps/api-server/internal/modules/schedule/service.go` | 修改 | 扩展 Service 接口和 in-memory 方法签名。 |
| `apps/api-server/internal/http/handlers/schedule.go` | 修改 | 增加 schedule HTTP handler 方法签名。 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册本迭代新增 API 路由。 |
| `apps/api-server/internal/modules/llm/dto.go` | 修改 | 增加成本汇总 DTO。 |
| `apps/api-server/internal/modules/llm/service.go` | 修改 | 增加成本汇总 Service 方法签名。 |
| `apps/api-server/internal/http/handlers/llmlog.go` | 修改 | 增加成本汇总 handler。 |
| `apps/api-server/internal/modules/external/dto.go` | 新增 | 外部自动化 DTO。 |
| `apps/api-server/internal/modules/external/service.go` | 新增 | 外部自动化 Service 骨架。 |
| `apps/api-server/internal/http/handlers/external.go` | 新增 | 外部自动化 handler 骨架。 |
| `apps/api-server/migrations/00004_create_iteration_2_1_tables.sql` | 新增 | 数据表契约。 |
| `openapi/openapi.yaml` | 修改 | API contract 更新。 |
| `apps/web-admin/lib/api.ts` | 修改 | 前端 API 类型与函数。 |
| `apps/web-admin/app/global-nav.tsx` | 修改 | 导航新增入口。 |
| `apps/web-admin/app/page.tsx` | 修改 | 历史页面渲染与 E2E fixture 修复。 |
| `apps/web-admin/app/workflow/schedules/page.tsx` | 新增 | 调度管理页面。 |
| `apps/web-admin/app/external-automation/n8n/page.tsx` | 新增 | n8n 页面。 |
| `apps/web-admin/app/llm/cost-summary/page.tsx` | 新增 | 成本汇总页面。 |
| `apps/web-admin/app/swagger-openapi/page.tsx` | 修改 | OpenAPI 入口渲染补齐。 |
| `apps/web-admin/app/workflow/templates/page.tsx` | 修改 | 历史页面渲染补齐。 |
| `apps/web-admin/app/workflow/runs/page.tsx` | 修改 | 历史页面渲染补齐。 |
| `apps/web-admin/app/agent/tasks/page.tsx` | 修改 | 历史页面渲染补齐。 |
| `apps/web-admin/app/llm/logs/page.tsx` | 修改 | 历史页面渲染补齐。 |
| `apps/web-admin/e2e/iteration1-ui.spec.ts` | 修改 | E2E 重复运行稳定性。 |
| `apps/web-admin/e2e/iteration2-navigation.spec.ts` | 修改 | 新增页面导航覆盖。 |
| `apps/api-server/internal/http/schedule_test.go` | 新增 | Schedule API TDD 测试。 |
| `apps/api-server/internal/modules/schedule/service_test.go` | 新增 | Schedule service TDD 测试。 |
| `apps/api-server/internal/modules/llm/summary_test.go` | 新增 | 成本汇总 TDD 测试。 |
| `apps/api-server/internal/modules/external/service_test.go` | 新增 | 外部自动化 TDD 测试。 |
| `apps/api-server/internal/store/iteration_2_1_migration_test.go` | 新增 | DDL contract 测试。 |
| `apps/web-admin/e2e/iteration2_1-pages.spec.ts` | 新增 | 本迭代 Web Admin E2E。 |

## 9. Development Tasks

- Task-01：扩展 Schedule DTO 与 Service 接口骨架
  - 所属模块：schedule
  - 简要描述：定义 WorkflowSchedule、ProductionPlan、TriggerLog、创建、启停、试跑和触发记录相关 DTO 与 Service 方法签名。
  - 涉及接口/方法：schedule.Service, CreateSchedule(), ListSchedules(), EnableSchedule(), DisableSchedule(), PrepareTestRun(), CompleteTrigger(), ListTriggers()
  - 输入：调度查询、创建、启停、试跑请求。
  - 输出：分页调度、创建响应、启停响应、PreparedScheduleRun、触发记录。
  - 产出类型：library
  - 功能类型：Schedule 服务契约（type id: library）
  - 是否跨组件：否
- Task-02：实现 Schedule Service 手动触发运行时
  - 所属模块：schedule
  - 简要描述：实现 in-memory schedule、production plan、trigger log、默认每日数量 5、可配置每日数量、状态流转和幂等。
  - 涉及接口/方法：scheduleService.CreateSchedule(), EnableSchedule(), DisableSchedule(), PrepareTestRun(), CompleteTrigger()
  - 输入：CreateScheduleRequest、ToggleScheduleRequest、TestRunScheduleRequest。
  - 输出：CreateScheduleResponse、ToggleScheduleResponse、PreparedScheduleRun。
  - 产出类型：integration
  - 功能类型：Schedule service 状态机与触发准备（type id: integration）
  - 是否跨组件：是（组件链路：ScheduleHandler -> ScheduleService -> WorkflowService）
- Task-03：暴露 WorkflowSchedule HTTP API
  - 所属模块：http/schedule
  - 简要描述：实现 schedule handler、路由、错误映射和统一 envelope，试跑接口返回 workflow_run_id。
  - 涉及接口/方法：ScheduleHandler.ListSchedules(), CreateSchedule(), EnableSchedule(), DisableSchedule(), TestRun(), ListTriggers()
  - 输入：HTTP query/path/body/header。
  - 输出：统一 APIEnvelope 和 HTTP status。
  - 产出类型：web-e2e
  - 功能类型：Schedule HTTP API（type id: web-e2e）
  - 是否跨组件：是（组件链路：HTTP Router -> ScheduleHandler -> ScheduleService -> WorkflowService -> EngineSubmitter）
- Task-04：补齐 LLM 成本汇总接口
  - 所属模块：llm
  - 简要描述：基于 LLMCallLog 聚合 calls、tokens、cost、by_model，并暴露 summary HTTP API。
  - 涉及接口/方法：LLMService.SummaryCallLogs(), LLMLogHandler.Summary()
  - 输入：project_id、date_from、date_to、provider、model。
  - 输出：LLMCostSummaryResponse。
  - 产出类型：integration
  - 功能类型：LLM 日志聚合（type id: integration）
  - 是否跨组件：是（组件链路：HTTP Router -> LLMLogHandler -> LLMService）
- Task-05：新增 External Automation 后端 API
  - 所属模块：external
  - 简要描述：新增 Provider、Binding 的 DTO、Service、Handler 和路由，Provider Token 只返回脱敏值。
  - 涉及接口/方法：external.Service, ExternalHandler.CreateProvider(), ListProviders(), CreateBinding(), ListBindings()
  - 输入：provider_type、base_url、token、trigger_event、webhook_url。
  - 输出：Provider / Binding 创建响应和分页列表。
  - 产出类型：web-e2e
  - 功能类型：External Automation API（type id: web-e2e）
  - 是否跨组件：是（组件链路：HTTP Router -> ExternalHandler -> ExternalService）
- Task-06：补齐 OpenAPI 与数据库迁移契约
  - 所属模块：contract
  - 简要描述：新增 Iteration 2.1 API OpenAPI 描述和 PostgreSQL DDL contract，覆盖所有新增表、字段、索引和禁止明文 token 约束。
  - 涉及接口/方法：openapi.yaml, migration SQL
  - 输入：设计中的 API 和 Table Design。
  - 输出：OpenAPI paths/schemas 和 00004 migration。
  - 产出类型：sql-query
  - 功能类型：DDL 与 API contract（type id: sql-query）
  - 是否跨组件：否
- Task-07：扩展 Web Admin API client 与导航入口
  - 所属模块：web-admin
  - 简要描述：在 `lib/api.ts` 增加 schedule、external automation、cost summary 类型与函数，并在全局导航加入新增页面。
  - 涉及接口/方法：fetchWorkflowSchedules(), createWorkflowSchedule(), testRunSchedule(), fetchLLMCostSummary(), createExternalProvider()
  - 输入：页面筛选、表单和动作参数。
  - 输出：APIEnvelope typed response。
  - 产出类型：library
  - 功能类型：前端 API client 与导航（type id: library）
  - 是否跨组件：否
- Task-08：实现 Iteration 2.1 新增 Web Admin 页面
  - 所属模块：web-admin
  - 简要描述：实现生产计划 / 调度管理、外部自动化 / n8n、成本汇总页面，覆盖 loading、empty、error、success 和主要按钮反馈。
  - 涉及接口/方法：workflow/schedules page, external-automation/n8n page, llm/cost-summary page
  - 输入：用户导航、筛选、表单提交、启停、试跑。
  - 输出：样式化页面、Toast/Alert、导航高亮。
  - 产出类型：web-e2e
  - 功能类型：Web Admin 页面（type id: web-e2e）
  - 是否跨组件：是（组件链路：GlobalNav -> Page -> lib/api -> API Server）
- Task-09：补齐历史页面真实渲染
  - 所属模块：web-admin
  - 简要描述：将 Iteration 0 / 1 / 2 核心页面统一为管理台视觉，补齐卡片、表格、表单、按钮、状态标签、错误态 request_id 和刷新可访问性。
  - 涉及接口/方法：HomePage, Swagger page, workflow templates/runs pages, agent tasks pages, llm logs page
  - 输入：用户访问和交互。
  - 输出：不再出现裸 HTML 的页面渲染。
  - 产出类型：web-e2e
  - 功能类型：历史页面渲染追溯（type id: web-e2e）
  - 是否跨组件：是（组件链路：GlobalNav -> Page -> lib/api -> API Server）
- Task-10：修复 Iteration 1 E2E 重复运行稳定性
  - 所属模块：web-admin-e2e
  - 简要描述：测试使用唯一数据、显式空态 fixture、请求等待前置、项目定位当前运行对象、Provider 脱敏断言列表安全。
  - 涉及接口/方法：iteration1-ui.spec.ts, HomePage fixture behavior
  - 输入：长生命周期内存后端和重复 Playwright 运行。
  - 输出：稳定通过的 iteration1-ui.spec.ts。
  - 产出类型：web-e2e
  - 功能类型：E2E 稳定性（type id: web-e2e）
  - 是否跨组件：是（组件链路：Playwright -> Web Admin -> API Server）
