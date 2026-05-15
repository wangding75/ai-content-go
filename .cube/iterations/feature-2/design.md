# Design: Iteration 2 — Workflow Engine 与多 Agent 架构

> 基于 PRD：`.cube/iterations/feature-2/prd.md`  
> 撰写日期：2026-05-15

---

## 1. 概述

本次设计在 Iteration 1 的单模块 Go HTTP API 基础上，新增 `workflow`、`agent`、`schedule` 三个业务模块，扩展 `llm` 模块增加 LLMCallLog，并增加 `engine` 执行层。整体遵循现有分层模式（Service interface + in-memory 实现 + HTTP handler），沿用 Iteration 1 的代码风格和目录结构。

**核心设计原则：**
- 所有新模块均遵循 `modules/{module}/dto.go + service.go` 模式
- `NewRouter` 签名不变；新服务和 engine 在内部创建，保持现有测试不变
- 幂等支持通过 service 内部 `idempotentOps map[string]any` 实现
- Operation log ID 由 service 内部计数器生成，与 content.service 保持一致
- WorkflowEngine 在 `NewRouter` 内创建并以 `context.Background()` 启动 goroutine
- 新增 SQL migration 文件，遵循 goose 格式，定义 iteration 2 的 DB schema

---

## 2. Impact Analysis

| 模块/文件 | 变更类型 | 影响范围 |
|-----------|---------|---------|
| `internal/http/api/response.go` | 修改 | 新增 5 个错误码常量；无破坏性改动 |
| `internal/http/router.go` | 修改 | 新增 18 条路由；`NewRouter` 签名不变 |
| `internal/modules/llm/dto.go` | 修改 | 新增 LLMCallLog DTO；不影响现有 DTO |
| `internal/modules/llm/service.go` | 修改 | 扩展 Service 接口增加 3 个 LLMCallLog 方法；现有方法不变 |
| `internal/modules/workflow/dto.go` | 新增 | 全新文件 |
| `internal/modules/workflow/service.go` | 新增 | 全新文件 |
| `internal/modules/agent/dto.go` | 新增 | 全新文件 |
| `internal/modules/agent/service.go` | 新增 | 全新文件 |
| `internal/modules/schedule/dto.go` | 新增 | 全新文件（contract only） |
| `internal/modules/schedule/service.go` | 新增 | 全新文件（contract only） |
| `internal/engine/engine.go` | 新增 | 全新文件；WorkflowEngine |
| `internal/http/handlers/workflow.go` | 新增 | 全新文件 |
| `internal/http/handlers/agent.go` | 新增 | 全新文件 |
| `internal/http/handlers/llmlog.go` | 新增 | 全新文件 |
| `internal/http/handlers/schedule.go` | 新增 | 全新文件 |
| `migrations/00003_create_workflow_tables.sql` | 新增 | 全新迁移文件 |
| `apps/web-admin/lib/api.ts` | 修改 | 扩展 TypeScript API 类型 |
| `apps/web-admin/app/workflow/templates/page.tsx` | 新增 | 工作流模板管理页 |
| `apps/web-admin/app/workflow/templates/[id]/page.tsx` | 新增 | 工作流模板详情页 |
| `apps/web-admin/app/workflow/runs/page.tsx` | 新增 | 工作流运行记录页 |
| `apps/web-admin/app/workflow/runs/[id]/page.tsx` | 新增 | 工作流运行详情页 |
| `apps/web-admin/app/agent/tasks/page.tsx` | 新增 | AgentTask 列表页 |
| `apps/web-admin/app/agent/tasks/[id]/page.tsx` | 新增 | AgentTask 详情页 |
| `apps/web-admin/app/llm/logs/page.tsx` | 新增 | LLM 调用日志页 |
| `internal/app/server.go` | 无影响 | 新服务在 router.go 内部创建 |
| content、prompt、system、dashboard 模块 | 无影响 | 不涉及 |

**接口兼容性：** 现有 15 条路由均不变动；`NewRouter(systemService, logger)` 签名不变。

**数据兼容性：** 当前为 in-memory 实现，无现存数据受影响。

---

## 3. Flow Design

### 3.1 WorkflowTemplate 发布流程

```
用户
  → POST /api/v1/workflow-templates             (创建模板, status=draft)
  → POST /api/v1/workflow-templates/:id/versions (创建版本+步骤, status=draft)
  → POST /api/v1/workflow-template-versions/:id/publish
      service: 校验 version.status == "draft" → 置为 "published" → 生成 operation_log_id
      返回 {previous_status, current_status, operation_log_id}
```

### 3.2 WorkflowRun 触发与执行流程

```
HTTP 请求线程:
  POST /api/v1/workflow-runs
    → service.CreateRun() → 写入 in-memory, status=pending
    → engine.Submit(runID)
    → 立即返回 {workflow_run_id, status="pending"}

WorkflowEngine goroutine (异步):
  run → status=running
  foreach step in template_version.steps (按 order_index 顺序):
    stepRun = pending → running
    if step_type == "agent":
      agentTask = pending → running
      (mock LLM call) → 生成 LLMCallLog
      agentTask = success/failed
    if step_type == "human_review":
      stepRun = running (等待外部驱动, 本迭代不自动推进)
      break
    if step_type == "condition":
      简单条件判断, 跳过或继续
    if step_type == "system_task":
      立即标记 success (mock)
    stepRun = success/failed/skipped
  run = success / failed (视最后步骤结果)
```

### 3.3 WorkflowRun 取消

```
POST /api/v1/workflow-runs/:id/cancel
  → service.CancelRun() 校验 run.status ∈ {pending, running}
  → 置 run.status = cancelled
  → 生成 operation_log_id
  → engine 下次执行步骤前检测 cancelled 状态自动停止
  → 返回 {previous_status, current_status, operation_log_id}
```

### 3.4 WorkflowRun 重试

```
POST /api/v1/workflow-runs/:id/retry
  → service.RetryRun() 校验 original_run.status == failed
  → 创建新 WorkflowRun, parent_run_id = original_run.id, source=retry
  → engine.Submit(new_run_id)
  → 返回 {new_workflow_run_id, status="pending"}
```

### 3.5 AgentTask 只能内部创建

```
AgentTask 由 WorkflowEngine.runStep() 内部调用 agent.Service.CreateTask()
HTTP 层没有 POST /agent-tasks 端点
```

---

## 4. Table Design

> 以下为逻辑数据模型，当前迭代使用 in-memory 实现；字段命名与规划 DB schema 一致。

### operation_log（已存在，DB schema 见 migrations/00001_create_operation_log.sql）

> In-memory 实现中，operation_log_id 由 service 内部计数器生成为字符串 `"oplog-{n}"`（与 content.service 保持一致）。DB schema 使用 `BIGSERIAL` 主键，两者不冲突：in-memory 返回字符串 ID 仅作占位响应字段，不写入 DB。

| 字段 | 类型（DB） | 说明 |
|------|-----------|------|
| id | BIGSERIAL | in-memory 占位为 "oplog-{n}" 字符串；DB 为自增整数 |
| request_id | TEXT NOT NULL | 关联 HTTP request_id |
| actor_id | TEXT | 操作人 ID（可空） |
| actor_type | TEXT | 默认 "system" |
| action | TEXT | "publish" / "cancel" / "retry" |
| resource_type | TEXT | "workflow_template_version" / "workflow_run" |
| resource_id | TEXT | 关联资源 ID |
| reason | TEXT | 操作原因 |
| metadata | JSONB | 额外信息（可选） |
| created_at | TIMESTAMP | 生成时间 |

### workflow_template

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | "wft-{n}" |
| code | string | 唯一 |
| name | string | |
| content_type | string | 内容类型 |
| category | string | 分类 |
| description | string | 描述 |
| status | string | draft / active / archived |
| created_at | time.Time | |

### workflow_template_version

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | "wftv-{n}" |
| template_id | string | FK → workflow_template |
| version | int | 自增，同模板内唯一 |
| input_schema | map[string]any | |
| output_schema | map[string]any | |
| status | string | draft / published / deprecated |
| created_at | time.Time | |

### workflow_step_template

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | "wfst-{n}" |
| template_version_id | string | FK → workflow_template_version |
| step_code | string | 步骤标识 |
| step_type | string | agent / human_review / condition / system_task |
| agent_code | string | 当 step_type=agent 时填写 |
| order_index | int | 执行顺序，从 1 开始 |
| input_mapping | map[string]any | 输入映射 |
| output_mapping | map[string]any | 输出映射 |

### workflow_run

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | "wfr-{n}" |
| project_id | string | 关联内容项目 |
| template_version_id | string | FK → workflow_template_version |
| status | string | pending / running / success / failed / cancelled |
| input | map[string]any | 输入参数 |
| output | map[string]any | 输出结果 |
| error | string | 错误信息 |
| source | string | manual / retry |
| parent_run_id | string | 重试时关联原 run |
| idempotency_key | string | 幂等键 |
| created_at | time.Time | |
| updated_at | time.Time | |

### workflow_step_run

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | "wfsr-{n}" |
| workflow_run_id | string | FK → workflow_run |
| step_template_id | string | FK → workflow_step_template |
| status | string | pending / running / success / failed / skipped / cancelled |
| input | map[string]any | |
| output | map[string]any | |
| error | string | |
| started_at | *time.Time | |
| finished_at | *time.Time | |

### agent_task

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | "at-{n}" |
| workflow_run_id | string | FK → workflow_run |
| step_run_id | string | FK → workflow_step_run |
| agent_code | string | |
| prompt_template_id | string | 关联 prompt（可空） |
| status | string | pending / running / success / failed / cancelled |
| input | map[string]any | |
| output | map[string]any | |
| error | string | |
| started_at | *time.Time | |
| finished_at | *time.Time | |

### llm_call_log

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | "llmlog-{n}" |
| workflow_run_id | string | FK → workflow_run |
| step_run_id | string | FK → workflow_step_run |
| agent_task_id | string | FK → agent_task |
| provider | string | LLM Provider 名称 |
| model | string | 模型名称 |
| input_tokens | int | |
| output_tokens | int | |
| cost | float64 | |
| currency | string | USD |
| latency_ms | int | |
| status | string | success / failed |
| error | string | |
| request_id | string | 关联 HTTP request_id |
| created_at | time.Time | |

### workflow_schedule（contract only）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 占位 |
| template_version_id | string | |
| project_id | string | |
| schedule_type | string | |
| status | string | not_implemented |

---

## 5. API Design

> 所有接口遵守 `api-spec.md` 的 `success/data/error/request_id` 响应格式。

### WorkflowTemplate

**GET /api/v1/workflow-templates**
- Query: `page`、`page_size`、`sort`、`order`、`content_type`、`category`、`status`
- 200: `{items: [WorkflowTemplateResponse], pagination: PaginationResponse}`
- Errors: `VALIDATION_ERROR(400)`

**POST /api/v1/workflow-templates**
- Body: `{code, name, content_type, category, description}`
- 201: `{workflow_template_id, status}`
- Errors: `VALIDATION_ERROR(400)`, `CONFLICT(409)`

**GET /api/v1/workflow-templates/:id**
- 200: `WorkflowTemplateResponse`
- Errors: `NOT_FOUND(404)`

**GET /api/v1/workflow-templates/:id/versions**
- Query: `page`、`page_size`
- 200: `{items: [WorkflowTemplateVersionResponse], pagination: PaginationResponse}`
- Errors: `NOT_FOUND(404)`

**POST /api/v1/workflow-templates/:id/versions**
- Body: `{input_schema, output_schema, steps: [{step_code, step_type, agent_code, order_index, input_mapping, output_mapping}]}`
- 201: `{template_version_id, step_count, status}`
- Errors: `VALIDATION_ERROR(400)`, `NOT_FOUND(404)`

**GET /api/v1/workflow-template-versions/:id**
- 200: `{...WorkflowTemplateVersionResponse, steps: [WorkflowStepTemplateResponse]}`
- Errors: `NOT_FOUND(404)`

**POST /api/v1/workflow-template-versions/:id/publish**
- Header: `Idempotency-Key` (optional)
- Body: `{note}`
- 200: `{previous_status, current_status, operation_log_id}`
- Errors: `VALIDATION_ERROR(400)`, `NOT_FOUND(404)`, `CONFLICT(409)`, `IDEMPOTENCY_CONFLICT(409)`

### WorkflowRun

**GET /api/v1/workflow-runs**
- Query: `project_id`、`template_version_id`、`status`、`page`、`page_size`、`sort`、`order`
- 200: `{items: [WorkflowRunResponse], pagination: PaginationResponse}`
- Errors: `VALIDATION_ERROR(400)`

**POST /api/v1/workflow-runs**
- Header: `Idempotency-Key` (optional)
- Body: `{project_id, template_version_id, input}`
- 202: `{workflow_run_id, status: "pending"}`
- Errors: `VALIDATION_ERROR(400)`, `NOT_FOUND(404)`, `CONFLICT(409)`, `IDEMPOTENCY_CONFLICT(409)`

**GET /api/v1/workflow-runs/:id**
- 200: `WorkflowRunDetailResponse`（含 input、output、error、step_count、agent_task_count）
- Errors: `NOT_FOUND(404)`

**GET /api/v1/workflow-runs/:id/steps**
- 200: `{items: [WorkflowStepRunResponse]}`
- Errors: `NOT_FOUND(404)`

**POST /api/v1/workflow-runs/:id/cancel**
- Header: `Idempotency-Key` (optional)
- Body: `{reason, note}`
- 200: `{previous_status, current_status, operation_log_id}`
- Errors: `VALIDATION_ERROR(400)`, `NOT_FOUND(404)`, `CONFLICT(409)`, `IDEMPOTENCY_CONFLICT(409)`

**POST /api/v1/workflow-runs/:id/retry**
- Header: `Idempotency-Key` (optional)
- Body: `{reason, input_override}`
- 202: `{new_workflow_run_id, status: "pending"}`
- Errors: `VALIDATION_ERROR(400)`, `NOT_FOUND(404)`, `CONFLICT(409)`, `IDEMPOTENCY_CONFLICT(409)`

### AgentTask

**GET /api/v1/agent-tasks**
- Query: `workflow_run_id`、`step_run_id`、`agent_code`、`status`、`page`、`page_size`
- 200: `{items: [AgentTaskResponse], pagination: PaginationResponse}`
- Errors: `VALIDATION_ERROR(400)`

**GET /api/v1/agent-tasks/:id**
- 200: `AgentTaskDetailResponse`（含 input、output、error、step_run_id、llm_call_log_count、llm_call_log_ids []string）
- Errors: `NOT_FOUND(404)`

### LLMCallLog

**GET /api/v1/llm-call-logs**
- Query: `workflow_run_id`、`agent_task_id`、`provider`、`model`、`status`、`page`、`page_size`
- 200: `{items: [LLMCallLogResponse], pagination: PaginationResponse}`
- Errors: `VALIDATION_ERROR(400)`

**GET /api/v1/llm-call-logs/:id**
- 200: `LLMCallLogDetailResponse`
- Errors: `NOT_FOUND(404)`

### WorkflowSchedule（Contract Only）

**POST /api/v1/workflow-schedules**
- Body: `{template_version_id, project_id, schedule_type}`
- 201: `{schedule_id, status: "not_implemented"}`
- 说明：仅占位，不实现调度运行时

---

## 6. Module Design

### workflow 模块

```
internal/modules/workflow/
  dto.go      — 全部 Template/Version/StepTemplate/Run/StepRun DTO
  service.go  — Service interface + in-memory 实现
```

`workflow.Service` 接口方法：
- `ListTemplates(ctx, req ListWorkflowTemplatesRequest) (PagedWorkflowTemplatesResponse, error)`
- `CreateTemplate(ctx, req CreateWorkflowTemplateRequest) (CreateWorkflowTemplateResponse, error)`
- `GetTemplate(ctx, id string) (WorkflowTemplateResponse, error)`
- `ListVersions(ctx, templateID string, req PaginationRequest) (PagedVersionsResponse, error)`
- `CreateVersion(ctx, templateID string, req CreateVersionRequest) (CreateVersionResponse, error)`
- `GetVersion(ctx, id string) (WorkflowTemplateVersionDetailResponse, error)`
- `PublishVersion(ctx, id string, req PublishVersionRequest, idempotencyKey string) (PublishVersionResponse, error)`
- `ListRuns(ctx, req ListWorkflowRunsRequest) (PagedWorkflowRunsResponse, error)`
- `CreateRun(ctx, req CreateWorkflowRunRequest, idempotencyKey string) (CreateWorkflowRunResponse, error)`
- `GetRun(ctx, id string) (WorkflowRunDetailResponse, error)`
- `GetRunSteps(ctx, runID string) (ListStepRunsResponse, error)`
- `CancelRun(ctx, id string, req CancelRunRequest, idempotencyKey string) (CancelRunResponse, error)`
- `RetryRun(ctx, id string, req RetryRunRequest, idempotencyKey string) (RetryRunResponse, error)`

`workflow.EnginePort` 接口（exported，供 engine 包跨包调用，与 Service 接口分离）：
- `UpdateRunStatus(ctx context.Context, id, status string, output map[string]any, errMsg string) error`
- `CreateStepRun(ctx context.Context, req CreateStepRunRequest) (WorkflowStepRunResponse, error)`
- `UpdateStepRunStatus(ctx context.Context, id, status string, output map[string]any, errMsg string) error`
- `GetRunStepTemplates(ctx context.Context, templateVersionID string) ([]WorkflowStepTemplateResponse, error)`
- `GetRunForEngine(ctx context.Context, id string) (WorkflowRunResponse, error)`

in-memory 的 `workflowServiceImpl` 同时实现 `Service` 和 `EnginePort` 两个接口。

错误变量：`ErrValidation`, `ErrNotFound`, `ErrConflict`, `ErrIdempotencyConflict`

### agent 模块

```
internal/modules/agent/
  dto.go      — AgentTask DTO
  service.go  — Service interface + in-memory 实现
```

`agent.Service` 接口方法：
- `CreateTask(ctx, req CreateAgentTaskRequest) (AgentTaskResponse, error)` — 只供 engine 内部调用
- `UpdateTask(ctx, id string, update UpdateAgentTaskRequest) error`
- `ListTasks(ctx, req ListAgentTasksRequest) (PagedAgentTasksResponse, error)` — HTTP 查询
- `GetTask(ctx, id string) (AgentTaskDetailResponse, error)` — HTTP 查询

错误变量：`ErrValidation`, `ErrNotFound`, `ErrConflict`

### llm 模块扩展

扩展现有 `internal/modules/llm/` 模块：

新增 DTO（`dto.go`）：
- `CreateLLMCallLogRequest`、`LLMCallLogResponse`、`LLMCallLogDetailResponse`
- `ListLLMCallLogsRequest`、`PagedLLMCallLogsResponse`

扩展 `Service` 接口（`service.go`）新增方法：
- `CreateCallLog(ctx, req CreateLLMCallLogRequest) (LLMCallLogResponse, error)` — 只供 engine 调用
- `ListCallLogs(ctx, req ListLLMCallLogsRequest) (PagedLLMCallLogsResponse, error)` — HTTP 查询
- `GetCallLog(ctx, id string) (LLMCallLogDetailResponse, error)` — HTTP 查询

错误变量：`ErrValidation`, `ErrNotFound`, `ErrConflict`（`GetCallLog` 不存在时返回 `ErrNotFound`）

### schedule 模块（contract only）

```
internal/modules/schedule/
  dto.go      — CreateScheduleRequest, ScheduleResponse（占位）
  service.go  — Service interface + stub 实现
```

`schedule.Service` 接口方法：
- `CreateSchedule(ctx, req CreateScheduleRequest) (ScheduleResponse, error)` — 返回 not_implemented 状态

### engine 包

```
internal/engine/
  engine.go   — Engine struct + Submitter interface + Start/Submit 方法
```

```go
const engineChannelSize = 100  // 缓冲大小，避免 HTTP 线程阻塞

type Submitter interface { Submit(runID string) bool }

type Engine struct {
    runCh       chan string      // make(chan string, engineChannelSize)
    workflowSvc workflow.EnginePort
    agentSvc    agent.Service
    llmSvc      llm.Service
}

func New(wf workflow.EnginePort, ag agent.Service, lm llm.Service) *Engine
func (e *Engine) Start(ctx context.Context)   // 启动 goroutine worker
func (e *Engine) Submit(runID string) bool    // 非阻塞入队；channel 满时返回 false，不阻塞 HTTP 线程
```

WorkflowEngine 执行链路：
1. 从 channel 取出 runID → `workflowSvc.UpdateRunStatus(id, "running")`
2. 获取步骤列表 `workflowSvc.GetRunStepTemplates(templateVersionID)`
3. 逐步执行：创建 StepRun → 按 step_type 执行 → 更新 StepRun 状态
4. `step_type == "agent"` → `agentSvc.CreateTask()` → mock LLM → `llmSvc.CreateCallLog()` → `agentSvc.UpdateTask()`
5. 每步执行前调用 `workflowSvc.GetRunForEngine()` 检查 run 是否已 cancelled
6. 全部步骤完成 → `workflowSvc.UpdateRunStatus(id, "success"/"failed")`

### 新 HTTP Handlers

- `internal/http/handlers/workflow.go` — `WorkflowHandler`，处理 Template + Version + Run 所有端点
- `internal/http/handlers/agent.go` — `AgentHandler`，处理 AgentTask 端点
- `internal/http/handlers/llmlog.go` — `LLMLogHandler`，处理 LLMCallLog 端点
- `internal/http/handlers/schedule.go` — `ScheduleHandler`，处理 WorkflowSchedule 端点

Handler 错误映射：
- `workflow.ErrValidation` / `agent.ErrValidation` / `llm.ErrValidation` → 400 VALIDATION_ERROR
- `workflow.ErrNotFound` / `agent.ErrNotFound` / `llm.ErrNotFound` → 404 NOT_FOUND
- `workflow.ErrConflict` / `agent.ErrConflict` / `llm.ErrConflict` → 409 CONFLICT
- `workflow.ErrIdempotencyConflict` → 409 IDEMPOTENCY_CONFLICT

---

## 7. Output Contract

| API / 模块 | 输入 | 输出 | 产出类型 | 功能类型 | 是否跨组件 | 测试规范 |
|-----------|------|------|---------|---------|-----------|---------|
| GET/POST /workflow-templates | 分页/筛选参数 or 创建 Body | 模板列表/创建结果 | web-e2e | REST API (type id: web-e2e) | 是（Handler → workflow.Service） | standards/testing/web-e2e.md |
| GET /workflow-templates/:id/versions, POST /versions | ID or 创建 Body | 版本列表/创建结果 | web-e2e | REST API (type id: web-e2e) | 是（Handler → workflow.Service） | standards/testing/web-e2e.md |
| POST /workflow-template-versions/:id/publish | note + Idempotency-Key | previous/current status, oplog_id | web-e2e | REST API (type id: web-e2e) | 是（Handler → workflow.Service） | standards/testing/web-e2e.md |
| GET /workflow-template-versions/:id | ID | 版本详情含步骤 | web-e2e | REST API (type id: web-e2e) | 是（Handler → workflow.Service） | standards/testing/web-e2e.md |
| POST /workflow-runs | project/template/input + Idempotency-Key | run_id, status=pending | web-e2e | REST API + async dispatch (type id: web-e2e) | 是（Handler → workflow.Service + engine.Submitter） | standards/testing/web-e2e.md |
| GET /workflow-runs + /runs/:id + /runs/:id/steps | 分页/筛选 or ID | 列表/详情/步骤 | web-e2e | REST API (type id: web-e2e) | 是（Handler → workflow.Service） | standards/testing/web-e2e.md |
| POST /workflow-runs/:id/cancel | reason/note + Idempotency-Key | previous/current status, oplog_id | web-e2e | REST API (type id: web-e2e) | 是（Handler → workflow.Service） | standards/testing/web-e2e.md |
| POST /workflow-runs/:id/retry | reason/input_override + Idempotency-Key | new_run_id, status=pending | web-e2e | REST API + async dispatch (type id: web-e2e) | 是（Handler → workflow.Service + engine.Submitter） | standards/testing/web-e2e.md |
| GET /agent-tasks + /agent-tasks/:id | 分页筛选（workflow_run_id/step_run_id/agent_code/status）or ID | 列表/详情 | web-e2e | REST API (type id: web-e2e) | 是（Handler → agent.Service） | standards/testing/web-e2e.md |
| GET /llm-call-logs + /llm-call-logs/:id | 分页/筛选 or ID | 列表/详情 | web-e2e | REST API (type id: web-e2e) | 是（Handler → llm.Service） | standards/testing/web-e2e.md |
| POST /workflow-schedules | schedule request body | schedule_id 占位 | web-e2e | REST API contract only (type id: web-e2e) | 否 | standards/testing/web-e2e.md |
| WorkflowEngine 执行链路 | runID | Run/StepRun/AgentTask/LLMCallLog 完整链路 | integration | 跨模块执行 (type id: integration) | 是（engine → workflow + agent + llm） | standards/testing/integration.md |

---

## 8. Change Log

| 文件 | 变更类型 | 变更原因 |
|------|---------|---------|
| `internal/http/api/response.go` | 修改 | 新增 IDEMPOTENCY_CONFLICT、WORKFLOW_RUN_FAILED、AGENT_OUTPUT_INVALID、LLM_PROVIDER_ERROR、EXTERNAL_AUTOMATION_ERROR 错误码 |
| `internal/http/router.go` | 修改 | 注册 18 条新路由，初始化 workflow/agent/llm/schedule 服务和 WorkflowEngine；CORS 允许头中新增 `Idempotency-Key` |
| `internal/modules/llm/dto.go` | 修改 | 新增 LLMCallLog DTO 类型 |
| `internal/modules/llm/service.go` | 修改 | 扩展 Service 接口和 in-memory 实现，增加 LLMCallLog 相关方法 |
| `internal/modules/workflow/dto.go` | 新增 | Workflow 模块全部 DTO |
| `internal/modules/workflow/service.go` | 新增 | Workflow Service 接口和 in-memory 实现 |
| `internal/modules/agent/dto.go` | 新增 | AgentTask DTO |
| `internal/modules/agent/service.go` | 新增 | AgentTask Service 接口和 in-memory 实现 |
| `internal/modules/schedule/dto.go` | 新增 | WorkflowSchedule 占位 DTO |
| `internal/modules/schedule/service.go` | 新增 | WorkflowSchedule Service 接口和 stub 实现 |
| `internal/engine/engine.go` | 新增 | WorkflowEngine，协调 workflow/agent/llm 服务完成异步执行链路 |
| `internal/http/handlers/workflow.go` | 新增 | WorkflowHandler，处理 Template/Version/Run 所有 HTTP 端点 |
| `internal/http/handlers/agent.go` | 新增 | AgentHandler，处理 AgentTask 端点 |
| `internal/http/handlers/llmlog.go` | 新增 | LLMLogHandler，处理 LLMCallLog 端点 |
| `internal/http/handlers/schedule.go` | 新增 | ScheduleHandler，处理 WorkflowSchedule 端点 |
| `migrations/00003_create_workflow_tables.sql` | 新增 | Iteration 2 DB schema 迁移文件（goose 格式） |
| `apps/web-admin/lib/api.ts` | 修改 | 扩展 TypeScript API 类型定义，增加 WorkflowTemplate/WorkflowRun/AgentTask/LLMCallLog 相关类型和请求函数 |
| `apps/web-admin/app/workflow/templates/page.tsx` | 新增 | 工作流模板管理页：列表展示、筛选、新建、查看版本入口 |
| `apps/web-admin/app/workflow/templates/[id]/page.tsx` | 新增 | 工作流模板详情页：版本列表、创建版本、发布版本 |
| `apps/web-admin/app/workflow/runs/page.tsx` | 新增 | 工作流运行记录页：运行列表、筛选、手动触发入口 |
| `apps/web-admin/app/workflow/runs/[id]/page.tsx` | 新增 | 工作流运行详情页：运行状态、StepRun 列表、AgentTask 关联、LLM 日志关联 |
| `apps/web-admin/app/agent/tasks/page.tsx` | 新增 | AgentTask 列表页：按 workflow_run_id/step_run_id/agent_code/status 筛选 |
| `apps/web-admin/app/agent/tasks/[id]/page.tsx` | 新增 | AgentTask 详情页：input/output/错误/关联 LLM 日志 |
| `apps/web-admin/app/llm/logs/page.tsx` | 新增 | LLM 调用日志页：按 workflow_run_id/agent_task_id/provider/model/status 筛选 |

---

## 9. Development Tasks

- Task-01：新增 API 错误码（api/response.go）
  - 所属模块：internal/http/api
  - 简要描述：在 response.go 新增 IDEMPOTENCY_CONFLICT、WORKFLOW_RUN_FAILED、AGENT_OUTPUT_INVALID、LLM_PROVIDER_ERROR、EXTERNAL_AUTOMATION_ERROR 五个 ErrorCode 常量
  - 涉及接口/方法：无新方法，仅常量
  - 输入：无
  - 输出：新增 5 个 ErrorCode 常量，类型为 api.ErrorCode
  - 产出类型：none
  - 功能类型：API 基础设施常量扩展（type id: none）
  - 是否跨组件：否

- Task-02：workflow 模块 — Template/Version/StepTemplate 服务层
  - 所属模块：internal/modules/workflow
  - 简要描述：新建 workflow/dto.go（Template/Version/StepTemplate 相关 DTO）和 workflow/service.go（对应 Service 接口 + in-memory 实现），包含 ListTemplates/CreateTemplate/GetTemplate/ListVersions/CreateVersion/GetVersion/PublishVersion 方法，发布时生成 operation_log_id，重复 code 拒绝创建，已发布版本不允许再发布
  - 涉及接口/方法：workflow.Service（Template/Version 方法）
  - 输入：创建/查询/发布请求 DTO
  - 输出：对应 Response DTO；PublishVersion 返回 previous_status、current_status、operation_log_id
  - 产出类型：library
  - 功能类型：领域服务实现（type id: library）
  - 是否跨组件：否

- Task-03：workflow 模块 — Run/StepRun 服务层（状态机 + 幂等）
  - 所属模块：internal/modules/workflow
  - 简要描述：在 workflow/service.go 中新增 CreateRun/GetRun/ListRuns/GetRunSteps/CancelRun/RetryRun 方法，以及实现 `workflow.EnginePort` 接口方法（`UpdateRunStatus`/`CreateStepRun`/`UpdateStepRunStatus`/`GetRunStepTemplates`/`GetRunForEngine`）；CreateRun 和 CancelRun 支持 Idempotency-Key 去重；重试创建新 run 并关联 parent_run_id
  - 涉及接口/方法：workflow.Service（Run/StepRun 方法），内部 updateRunStatus、createStepRun、updateStepRunStatus、getRunStepTemplates
  - 输入：Run 创建/取消/重试请求 DTO + idempotencyKey string
  - 输出：对应 Response DTO；CancelRun 返回 operation_log_id；RetryRun 返回 new_workflow_run_id
  - 产出类型：library
  - 功能类型：领域服务实现（type id: library）
  - 是否跨组件：否

- Task-04：WorkflowTemplate + Version HTTP 端点（含路由注册）
  - 所属模块：internal/http/handlers, internal/http
  - 简要描述：新建 handlers/workflow.go 中 WorkflowHandler，实现 ListTemplates/CreateTemplate/GetTemplate/ListVersions/CreateVersion/GetVersionDetail/PublishVersion 七个 HTTP handler；在 router.go 中注册对应路由；解析 Idempotency-Key header
  - 涉及接口/方法：WorkflowHandler.ListTemplates()、CreateTemplate()、GetTemplate()、ListVersions()、CreateVersion()、GetVersionDetail()、PublishVersion()
  - 输入：HTTP request（query params / JSON body / path params / headers）
  - 输出：JSON 响应，遵循 api.Envelope 格式；201/200 成功，4xx 错误含错误码
  - 产出类型：web-e2e
  - 功能类型：REST API HTTP 端点（type id: web-e2e）
  - 是否跨组件：是（组件链路：WorkflowHandler → workflow.Service）

- Task-05：WorkflowRun HTTP 端点（触发 + 查询 + 取消 + 重试，含路由注册）
  - 所属模块：internal/http/handlers, internal/http
  - 简要描述：在 handlers/workflow.go 中实现 CreateRun/ListRuns/GetRun/GetRunSteps/CancelRun/RetryRun 六个 HTTP handler；POST /workflow-runs 返回 202；在 router.go 中注册对应路由；WorkflowHandler 持有 workflow.Service 和 engine.Submitter；**在 router.go 的 CORS 配置中将 `Idempotency-Key` 加入允许的请求头列表**
  - 涉及接口/方法：WorkflowHandler.CreateRun()、ListRuns()、GetRun()、GetRunSteps()、CancelRun()、RetryRun()
  - 输入：HTTP request
  - 输出：202/200 成功响应；创建/重试返回 run_id+status；取消返回 oplog_id
  - 产出类型：web-e2e
  - 功能类型：REST API HTTP 端点（type id: web-e2e）
  - 是否跨组件：是（组件链路：WorkflowHandler → workflow.Service + engine.Submitter）

- Task-06：agent 模块（Service 接口 + DTO + in-memory 实现）
  - 所属模块：internal/modules/agent
  - 简要描述：新建 agent/dto.go（AgentTask DTO）和 agent/service.go（Service 接口 + in-memory 实现），包含 CreateTask/UpdateTask/ListTasks/GetTask 方法；CreateTask 需要 step_run_id 参数，不允许外部直接调用（文档约束，无 HTTP 端点）
  - 涉及接口/方法：agent.Service
  - 输入：CreateAgentTaskRequest 需包含 workflow_run_id、step_run_id、agent_code
  - 输出：AgentTaskResponse；GetTask 返回含 llm_call_log_count 的 AgentTaskDetailResponse
  - 产出类型：library
  - 功能类型：领域服务实现（type id: library）
  - 是否跨组件：否

- Task-07：AgentTask HTTP 端点（含路由注册）
  - 所属模块：internal/http/handlers, internal/http
  - 简要描述：新建 handlers/agent.go 中 AgentHandler，实现 ListTasks/GetTask 两个 HTTP handler；在 router.go 中注册 GET /api/v1/agent-tasks 和 GET /api/v1/agent-tasks/:id
  - 涉及接口/方法：AgentHandler.ListTasks()、GetTask()
  - 输入：query params（workflow_run_id/step_run_id/agent_code/status/page/page_size）or path id
  - 输出：分页列表或详情 JSON
  - 产出类型：web-e2e
  - 功能类型：REST API HTTP 端点（type id: web-e2e）
  - 是否跨组件：是（组件链路：AgentHandler → agent.Service）

- Task-08：LLMCallLog 服务扩展（llm 模块）
  - 所属模块：internal/modules/llm
  - 简要描述：在 llm/dto.go 新增 LLMCallLog 相关 DTO；在 llm/service.go 扩展 Service 接口增加 CreateCallLog/ListCallLogs/GetCallLog 三个方法，并扩展 in-memory service 实现；现有 Provider 方法不变
  - 涉及接口/方法：llm.Service.CreateCallLog()、ListCallLogs()、GetCallLog()
  - 输入：CreateLLMCallLogRequest（需含 workflow_run_id、step_run_id、agent_task_id、provider、model、tokens、cost 等）
  - 输出：LLMCallLogResponse；ListCallLogs 返回分页列表
  - 产出类型：library
  - 功能类型：领域服务扩展（type id: library）
  - 是否跨组件：否

- Task-09：LLMCallLog HTTP 端点（含路由注册）
  - 所属模块：internal/http/handlers, internal/http
  - 简要描述：新建 handlers/llmlog.go 中 LLMLogHandler，实现 ListCallLogs/GetCallLog 两个 HTTP handler；在 router.go 注册 GET /api/v1/llm-call-logs 和 GET /api/v1/llm-call-logs/:id
  - 涉及接口/方法：LLMLogHandler.ListCallLogs()、GetCallLog()
  - 输入：query params（workflow_run_id/agent_task_id/provider/model/status/page/page_size）or path id
  - 输出：分页列表或详情 JSON
  - 产出类型：web-e2e
  - 功能类型：REST API HTTP 端点（type id: web-e2e）
  - 是否跨组件：是（组件链路：LLMLogHandler → llm.Service）

- Task-10：WorkflowSchedule 契约占位（模块 + 端点 + 路由注册）
  - 所属模块：internal/modules/schedule, internal/http/handlers, internal/http
  - 简要描述：新建 schedule/dto.go（CreateScheduleRequest + ScheduleResponse）和 schedule/service.go（stub 实现，返回 not_implemented 状态）；新建 handlers/schedule.go（ScheduleHandler）；在 router.go 注册 POST /api/v1/workflow-schedules
  - 涉及接口/方法：schedule.Service.CreateSchedule()、ScheduleHandler.CreateSchedule()
  - 输入：CreateScheduleRequest（template_version_id、project_id、schedule_type）
  - 输出：ScheduleResponse（schedule_id、status="not_implemented"）
  - 产出类型：web-e2e
  - 功能类型：REST API 契约占位（type id: web-e2e）
  - 是否跨组件：否

- Task-11：WorkflowEngine 异步执行链路
  - 所属模块：internal/engine
  - 简要描述：新建 engine/engine.go，实现 Engine struct（runCh channel、workflowSvc、agentSvc、llmSvc 字段）；Start() 启动 goroutine worker；Submit() 将 runID 送入 channel；worker 按 Flow Design 3.2 执行 Run → StepRun → AgentTask → LLMCallLog 链路；每步前检查 run 是否已 cancelled
  - 涉及接口/方法：Engine.Start()、Engine.Submit()、Submitter 接口
  - 输入：runID string（通过 Submit 入队）
  - 输出：workflow_run/step_run/agent_task/llm_call_log 在各自 Service 中的状态更新
  - 产出类型：integration
  - 功能类型：跨模块异步执行（type id: integration）
  - 是否跨组件：是（组件链路：engine → workflow.Service + agent.Service + llm.Service）

- Task-12：SQL migration（iteration 2 tables）
  - 所属模块：migrations
  - 简要描述：新建 migrations/00003_create_workflow_tables.sql（goose 格式），包含 workflow_template、workflow_template_version、workflow_step_template、workflow_run、workflow_step_run、agent_task、llm_call_log、workflow_schedule 八张表的 CREATE TABLE 语句和必要索引；Down 迁移包含对应 DROP TABLE
  - 涉及接口/方法：无
  - 输入：无
  - 输出：SQL 迁移文件，通过 migration 测试验证格式和必要约束
  - 产出类型：none
  - 功能类型：数据库 schema 定义（type id: none）
  - 是否跨组件：否

- Task-13：前端 API 类型扩展（apps/web-admin/lib/api.ts）
  - 所属模块：apps/web-admin
  - 简要描述：在 api.ts 中新增 WorkflowTemplate、WorkflowTemplateVersion、WorkflowStepTemplate、WorkflowRun、WorkflowStepRun、AgentTask、LLMCallLog 相关 TypeScript 类型和 fetch 请求函数；保持现有类型不变
  - 涉及接口/方法：前端调用所有 iteration 2 后端 API
  - 输入：API 请求参数类型
  - 输出：API 响应类型
  - 产出类型：none
  - 功能类型：前端基础设施扩展（type id: none）
  - 是否跨组件：否

- Task-14：工作流模板管理页（apps/web-admin/app/workflow/templates/）
  - 所属模块：apps/web-admin
  - 简要描述：新建 workflow/templates/page.tsx（模板列表、筛选 content_type/category/status、新建按钮）和 workflow/templates/[id]/page.tsx（版本列表、创建版本、发布版本按钮，发布含 Idempotency-Key）；操作失败展示错误码、错误信息、request_id；支持空状态、加载态、错误态
  - 涉及接口/方法：GET/POST /workflow-templates、GET/POST /workflow-templates/:id/versions、POST /workflow-template-versions/:id/publish
  - 输入：用户操作
  - 输出：API 调用和页面反馈
  - 产出类型：web-e2e
  - 功能类型：前端页面（type id: web-e2e）
  - 是否跨组件：是（前端 → api-server）

- Task-15：工作流运行记录页（apps/web-admin/app/workflow/runs/）
  - 所属模块：apps/web-admin
  - 简要描述：新建 workflow/runs/page.tsx（运行列表、筛选 project_id/template_version_id/status、手动触发按钮，触发含 Idempotency-Key）和 workflow/runs/[id]/page.tsx（运行状态、input/output/error、StepRun 列表、关联 AgentTask、关联 LLM 日志；支持取消和失败重试操作）；操作失败展示错误码、错误信息、request_id
  - 涉及接口/方法：GET/POST /workflow-runs、GET /workflow-runs/:id、GET /workflow-runs/:id/steps、POST /workflow-runs/:id/cancel、POST /workflow-runs/:id/retry
  - 输入：用户操作
  - 输出：API 调用和页面反馈
  - 产出类型：web-e2e
  - 功能类型：前端页面（type id: web-e2e）
  - 是否跨组件：是（前端 → api-server）

- Task-16：AgentTask 列表与详情页（apps/web-admin/app/agent/tasks/）
  - 所属模块：apps/web-admin
  - 简要描述：新建 agent/tasks/page.tsx（按 workflow_run_id/step_run_id/agent_code/status 筛选的列表页）和 agent/tasks/[id]/page.tsx（input/output/错误/started_at/finished_at/关联 LLM 日志 ID 列表）；支持空状态、加载态、错误态
  - 涉及接口/方法：GET /agent-tasks、GET /agent-tasks/:id
  - 输入：用户操作
  - 输出：API 调用和页面数据展示
  - 产出类型：web-e2e
  - 功能类型：前端页面（type id: web-e2e）
  - 是否跨组件：是（前端 → api-server）

- Task-17：LLM 调用日志页（apps/web-admin/app/llm/logs/）
  - 所属模块：apps/web-admin
  - 简要描述：新建 llm/logs/page.tsx（按 workflow_run_id/agent_task_id/provider/model/status 筛选的列表页，列展示 provider、model、input_tokens、output_tokens、cost、latency_ms、status）；支持空状态、加载态、错误态
  - 涉及接口/方法：GET /llm-call-logs、GET /llm-call-logs/:id
  - 输入：筛选条件
  - 输出：LLM 调用日志列表
  - 产出类型：web-e2e
  - 功能类型：前端页面（type id: web-e2e）
  - 是否跨组件：是（前端 → api-server）
