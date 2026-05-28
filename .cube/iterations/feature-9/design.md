# Iteration 9 技术设计：策略建议与单类型业务闭环

## 1. 概述

本次设计在现有分层架构中新增 `strategy` 业务模块，围绕策略建议生成、列表筛选、人工确认与状态流转、策略执行与执行日志、操作审计构建项目级策略建议闭环。后端沿用 `handler -> module service -> DTO -> api.Envelope` 模式，前端沿用 `apps/web-admin/lib/api.ts` API client、项目工作区导航和现有管理台样式，OpenAPI 继续维护单一 `openapi/openapi.yaml`。

核心约束：

- Core 层保持内容类型无关，只使用 `project_id` 等通用关联，不引入 Novel / Book / Chapter 等内容类型专属资源命名。
- 策略建议生成必须读取项目已有指标能力（`metric_record`、`metric_summary_snapshot`），不接受前端直接传入指标结论。
- 状态机强制执行：`pending` → `confirmed` / `ignored`，`confirmed` → `executed` / `execution_failed`，`execution_failed` → `executed` / `execution_failed`；`ignored` 与 `executed` 为终态。
- `confirm` 只表示人工认可建议，不得自动修改项目配置、调度计划、Prompt、模型配置、发布状态或内容生产状态。
- 涉及跨模块变更的策略执行不直接改写其他模块状态，必须通过已有领域接口完成，并记录目标接口、目标资源和执行结果。
- 生成建议等触发型操作不阻塞 HTTP 请求等待最终处理完成，返回 `suggestion_run_id` 供后续查询。
- 所有状态变更操作写入 `operation_log`，操作日志写入失败时状态变更不得静默成功。
- 确认、忽略、执行、重试执行等状态变更接口支持 `Idempotency-Key`。

## 2. Impact Analysis

| 模块/文件 | 影响程度 | 说明 |
| --- | --- | --- |
| `apps/api-server/internal/modules/strategy/` | 新增 | 新增 DTO、错误常量、Service 接口与 Store 接口 |
| `apps/api-server/internal/http/handlers/strategy.go` | 新增 | 新增 HTTP Handler，负责参数解析、统一响应和 strategy 错误映射 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册策略建议路由（约 8 条） |
| `apps/api-server/migrations/00011_create_strategy_tables.sql` | 新增 | 新增 `strategy_suggestion`、`strategy_execution_log` 表 |
| `openapi/openapi.yaml` | 修改 | 增加本迭代所有接口 path、schema、响应和 examples |
| `apps/web-admin/lib/api.ts` | 修改 | 增加 strategy 类型与 API client 函数 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 修改 | 增加"策略建议"导航入口 |
| `apps/web-admin/app/projects/[projectId]/strategy-suggestions/page.tsx` | 新增 | 策略建议列表页 |
| `apps/web-admin/app/projects/[projectId]/strategy-suggestions/[suggestionId]/page.tsx` | 新增 | 建议详情页 |
| `apps/web-admin/app/projects/[projectId]/strategy-suggestions/[suggestionId]/actions/page.tsx` | 新增 | 建议操作页 |

对现有接口的兼容性分析：纯新增 API，不修改现有接口，向后兼容。

对现有数据的兼容性分析：新增 2 张表，复用 `operation_log` 和 `idempotency_record` 表（前者已有表结构，后者在 metrics 模块的 memoryStore 中以内存 map 实现，PostgreSQL 层面尚无独立表，幂等记录继续沿用各模块内存实现），不修改现有表结构。

## 3. Flow Design

### 3.1 策略建议生成流程

```
用户点击"生成建议"
  → POST /projects/{projectId}/strategy-suggestion-runs
  → Handler 解析参数（project_id, date_range, rule_codes?, metric_codes?, force_regenerate?）
  → 校验日期范围合法性，非法返回 VALIDATION_ERROR
  → 检查是否有正在进行的 suggestion_run（非 force_regenerate 时复用幂等结果或拒绝）
  → 创建 suggestion_run 记录，状态 generating，立即返回 suggestion_run_id
  → Service.GenerateSuggestions() 异步执行：
    1. 读取项目关联的 metric_record 和 metric_summary_snapshot
    2. 校验指标样本量是否充足，不足则生成低置信度建议或记录失败原因
    3. 按规则引擎逐条生成策略建议，每条包含类型/标题/触发原因/证据指标/影响范围/风险等级/置信度/建议动作/预期收益
    4. 建议类型必须属于枚举（keep/optimize/suspend/promote/cost_control），非法类型不得持久化
    5. 写入 strategy_suggestion 记录（状态 pending），关联 suggestion_run_id、项目、日期范围、触发规则、指标快照
    6. 更新 suggestion_run 状态为 completed 或 failed
```

### 3.2 策略建议列表与筛选流程

```
用户访问策略建议页
  → GET /projects/{projectId}/strategy-suggestions
  → Handler 解析筛选参数（status, suggestion_type, risk_level, confidence, date_from, date_to, page, page_size, sort, order）
  → Service.ListSuggestions() 查询并分页返回
  → 前端渲染：空态 / 加载态 / 错误态 / 成功态
```

### 3.3 人工确认流程

```
用户点击"确认"
  → POST /strategy-suggestions/{suggestionId}/confirm
  → Handler 读取 Idempotency-Key header
  → Service.ConfirmSuggestion()
    1. 查询建议，校验当前状态必须为 pending，否则返回 CONFLICT
    2. 幂等检查：同一 Idempotency-Key + 同请求体 → 返回已存结果；同键不同请求体 → IDEMPOTENCY_CONFLICT
    3. 更新状态 pending → confirmed
    4. 写入 operation_log（action=confirm, resource_type=strategy_suggestion）
    5. 返回 {suggestion_id, previous_status, current_status, operation_log_id}
  → confirm 仅改变建议状态并记录日志，不修改项目配置、调度计划、Prompt、模型配置、发布状态或内容生产状态
```

### 3.4 忽略流程

```
用户点击"忽略"
  → POST /strategy-suggestions/{suggestionId}/ignore
  → Handler 读取 Idempotency-Key header，解析 body（reason 必填，note 可选）
  → Service.IgnoreSuggestion()
    1. 查询建议，校验当前状态必须为 pending，否则返回 CONFLICT
    2. 校验 reason 非空，为空返回 VALIDATION_ERROR
    3. 幂等检查
    4. 更新状态 pending → ignored
    5. 写入 operation_log
    6. 返回 {suggestion_id, previous_status, current_status, operation_log_id}
```

### 3.5 执行流程

```
用户点击"执行"
  → POST /strategy-suggestions/{suggestionId}/execute
  → Handler 读取 Idempotency-Key header，解析 body（action_type, target_type, target_id, operator_note?）
  → Service.ExecuteSuggestion()
    1. 查询建议，校验当前状态必须为 confirmed，否则返回 CONFLICT
    2. 幂等检查
    3. 尝试执行：调用已有领域接口（如存在），缺少可用目标接口时执行失败
    4. 执行成功：更新状态 confirmed → executed，追加 strategy_execution_log（结果=success）
    5. 执行失败：更新状态 confirmed → execution_failed，追加 strategy_execution_log（结果=failed, failure_reason）
    6. 写入 operation_log
    7. 返回 {execution_log_id, suggestion_id, previous_status, current_status}
  → 策略模块不得绕过其他领域服务边界直接改写跨模块状态
```

### 3.6 重试执行流程

```
用户点击"重试"
  → POST /strategy-suggestions/{suggestionId}/retry
  → Handler 读取 Idempotency-Key header，解析 body（operator_note?）
  → Service.RetrySuggestion()
    1. 查询建议，校验当前状态必须为 execution_failed，否则返回 CONFLICT
    2. 幂等检查
    3. 重试执行：逻辑同 3.5 步骤 3-5
    4. 返回 {execution_log_id, suggestion_id, previous_status, current_status}
```

### 3.7 查看建议详情流程

```
用户点击建议卡片
  → GET /strategy-suggestions/{suggestionId}
  → Service.GetSuggestion() 返回完整建议详情
  → 包含建议类型、标题、触发原因、证据指标、影响范围、风险等级、置信度、建议动作、预期收益、metrics_snapshot、状态
  → 不存在时返回 NOT_FOUND
```

### 3.8 查看执行日志流程

```
用户在详情页查看执行记录
  → GET /strategy-suggestions/{suggestionId}/execution-logs
  → Service.ListExecutionLogs() 分页返回执行日志
  → 每条日志包含操作者、执行动作、目标资源、前后状态、结果、失败原因
```

## 4. Table Design

### 4.1 strategy_suggestion

```sql
CREATE TABLE IF NOT EXISTS strategy_suggestion (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    suggestion_run_id TEXT NOT NULL,
    suggestion_type TEXT NOT NULL CHECK (suggestion_type IN ('keep', 'optimize', 'suspend', 'promote', 'cost_control')),
    title TEXT NOT NULL,
    trigger_reason TEXT NOT NULL,
    evidence_metrics JSONB NOT NULL DEFAULT '[]'::jsonb,
    impact_scope TEXT NOT NULL DEFAULT '',
    risk_level TEXT NOT NULL CHECK (risk_level IN ('low', 'medium', 'high')),
    confidence TEXT NOT NULL CHECK (confidence IN ('low', 'medium', 'high')),
    suggested_action TEXT NOT NULL DEFAULT '',
    expected_benefit TEXT NOT NULL DEFAULT '',
    metrics_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL CHECK (status IN ('pending', 'confirmed', 'ignored', 'executed', 'execution_failed')) DEFAULT 'pending',
    ignored_reason TEXT NOT NULL DEFAULT '',
    ignored_note TEXT NOT NULL DEFAULT '',
    confirmed_at TIMESTAMPTZ,
    ignored_at TIMESTAMPTZ,
    executed_at TIMESTAMPTZ,
    date_from DATE NOT NULL,
    date_to DATE NOT NULL,
    triggered_rules JSONB NOT NULL DEFAULT '[]'::jsonb,
    generation_method TEXT NOT NULL DEFAULT 'rule',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_strategy_suggestion_project ON strategy_suggestion(project_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_strategy_suggestion_run ON strategy_suggestion(suggestion_run_id);
CREATE INDEX IF NOT EXISTS idx_strategy_suggestion_type_status ON strategy_suggestion(project_id, suggestion_type, status);
```

### 4.2 strategy_suggestion_run

```sql
CREATE TABLE IF NOT EXISTS strategy_suggestion_run (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    date_from DATE NOT NULL,
    date_to DATE NOT NULL,
    rule_codes JSONB NOT NULL DEFAULT '[]'::jsonb,
    metric_codes JSONB NOT NULL DEFAULT '[]'::jsonb,
    force_regenerate BOOLEAN NOT NULL DEFAULT FALSE,
    status TEXT NOT NULL CHECK (status IN ('generating', 'completed', 'failed')) DEFAULT 'generating',
    failure_reason TEXT NOT NULL DEFAULT '',
    suggestion_count INTEGER NOT NULL DEFAULT 0 CHECK (suggestion_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_strategy_suggestion_run_project ON strategy_suggestion_run(project_id, created_at DESC);
```

### 4.3 strategy_execution_log

```sql
CREATE TABLE IF NOT EXISTS strategy_execution_log (
    id TEXT PRIMARY KEY,
    suggestion_id TEXT NOT NULL REFERENCES strategy_suggestion(id),
    action_type TEXT NOT NULL,
    target_type TEXT NOT NULL DEFAULT '',
    target_id TEXT NOT NULL DEFAULT '',
    operator_note TEXT NOT NULL DEFAULT '',
    previous_status TEXT NOT NULL,
    current_status TEXT NOT NULL,
    result TEXT NOT NULL CHECK (result IN ('success', 'failed')),
    failure_reason TEXT NOT NULL DEFAULT '',
    target_interface TEXT NOT NULL DEFAULT '',
    target_resource TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_strategy_execution_log_suggestion ON strategy_execution_log(suggestion_id, created_at DESC);
```

字段说明：

- `strategy_suggestion.evidence_metrics`：JSON 数组，存储触发建议的证据指标列表（如 `[{"metric_code":"views","value":1200,"trend":"declining"}]`）
- `strategy_suggestion.metrics_snapshot`：JSON 对象，存储生成建议时的指标快照引用（如 `{"summary_snapshot_id":"metric-summary-snapshot-xxx"}`）
- `strategy_suggestion.triggered_rules`：JSON 数组，存储触发规则编码列表
- `strategy_execution_log.target_interface`：跨模块执行时调用的目标接口标识，未调用时为空
- `strategy_execution_log.target_resource`：跨模块执行时的目标资源标识

## 5. API Design

### 5.1 API 列表

| Method | Path | 说明 |
|--------|------|------|
| POST | `/projects/{projectId}/strategy-suggestion-runs` | 触发生成策略建议 |
| GET | `/projects/{projectId}/strategy-suggestions` | 查看策略建议分页列表 |
| GET | `/strategy-suggestions/{suggestionId}` | 查看建议详情 |
| POST | `/strategy-suggestions/{suggestionId}/confirm` | 确认建议 |
| POST | `/strategy-suggestions/{suggestionId}/ignore` | 忽略建议 |
| POST | `/strategy-suggestions/{suggestionId}/execute` | 执行建议 |
| POST | `/strategy-suggestions/{suggestionId}/retry` | 重试执行建议 |
| GET | `/strategy-suggestions/{suggestionId}/execution-logs` | 查看执行日志 |

### 5.2 POST /projects/{projectId}/strategy-suggestion-runs

请求参数：
```json
{
  "date_from": "2026-05-01",
  "date_to": "2026-05-25",
  "rule_codes": ["declining_views", "low_engagement"],
  "metric_codes": ["views", "likes"],
  "force_regenerate": false
}
```

响应（202 Accepted）：
```json
{
  "success": true,
  "data": {
    "suggestion_run_id": "strategy-run-project-1-20260525",
    "status": "generating"
  },
  "error": null,
  "request_id": "req_xxx"
}
```

错误码：`VALIDATION_ERROR`（日期范围缺失或非法）、`CONFLICT`（重复生成且不允许强制生成时复用幂等结果）

### 5.3 GET /projects/{projectId}/strategy-suggestions

查询参数：`status`, `suggestion_type`, `risk_level`, `confidence`, `date_from`, `date_to`, `page`, `page_size`, `sort`, `order`

响应：
```json
{
  "success": true,
  "data": {
    "items": [StrategySuggestionResponse],
    "pagination": { "page": 1, "page_size": 20, "total": 5, "has_next": false }
  },
  "error": null,
  "request_id": "req_xxx"
}
```

错误码：`VALIDATION_ERROR`（筛选值非法）

### 5.4 GET /strategy-suggestions/{suggestionId}

响应：
```json
{
  "success": true,
  "data": {
    "id": "strategy-suggestion-project-1-declining_views-20260525",
    "project_id": "project-1",
    "suggestion_run_id": "strategy-run-project-1-20260525",
    "suggestion_type": "optimize",
    "title": "阅读量持续下降",
    "trigger_reason": "近7天阅读量下降超过30%",
    "evidence_metrics": [{"metric_code":"views","value":1200,"trend":"declining"}],
    "impact_scope": "项目整体阅读表现",
    "risk_level": "medium",
    "confidence": "high",
    "suggested_action": "调整发布时间和频率",
    "expected_benefit": "预计可提升阅读量15-25%",
    "metrics_snapshot": {"summary_snapshot_id":"metric-summary-snapshot-xxx"},
    "status": "pending",
    "ignored_reason": "",
    "ignored_note": "",
    "confirmed_at": null,
    "ignored_at": null,
    "executed_at": null,
    "date_from": "2026-05-01",
    "date_to": "2026-05-25",
    "triggered_rules": ["declining_views"],
    "generation_method": "rule",
    "created_at": "2026-05-25T10:00:00Z",
    "updated_at": "2026-05-25T10:00:00Z"
  },
  "error": null,
  "request_id": "req_xxx"
}
```

错误码：`NOT_FOUND`

### 5.5 POST /strategy-suggestions/{suggestionId}/confirm

请求头：`Idempotency-Key: xxx`

请求参数：
```json
{
  "note": "同意此建议"
}
```

响应：
```json
{
  "success": true,
  "data": {
    "suggestion_id": "xxx",
    "previous_status": "pending",
    "current_status": "confirmed",
    "operation_log_id": "operation-log-xxx"
  },
  "error": null,
  "request_id": "req_xxx"
}
```

错误码：`CONFLICT`（非 pending 状态）、`IDEMPOTENCY_CONFLICT`（同键不同请求体）

### 5.6 POST /strategy-suggestions/{suggestionId}/ignore

请求头：`Idempotency-Key: xxx`

请求参数：
```json
{
  "reason": "当前不需要调整",
  "note": "下个迭代再考虑"
}
```

响应：
```json
{
  "success": true,
  "data": {
    "suggestion_id": "xxx",
    "previous_status": "pending",
    "current_status": "ignored",
    "operation_log_id": "operation-log-xxx"
  },
  "error": null,
  "request_id": "req_xxx"
}
```

错误码：`VALIDATION_ERROR`（缺少 reason）、`CONFLICT`（非 pending 状态或终态）、`IDEMPOTENCY_CONFLICT`

### 5.7 POST /strategy-suggestions/{suggestionId}/execute

请求头：`Idempotency-Key: xxx`

请求参数：
```json
{
  "action_type": "adjust_schedule",
  "target_type": "workflow_schedule",
  "target_id": "schedule-1",
  "operator_note": "调整发布时间到上午9点"
}
```

响应：
```json
{
  "success": true,
  "data": {
    "execution_log_id": "execution-log-xxx",
    "suggestion_id": "xxx",
    "previous_status": "confirmed",
    "current_status": "executed",
    "operation_log_id": "operation-log-xxx"
  },
  "error": null,
  "request_id": "req_xxx"
}
```

错误码：`CONFLICT`（非 confirmed 状态）、`IDEMPOTENCY_CONFLICT`

### 5.8 POST /strategy-suggestions/{suggestionId}/retry

请求头：`Idempotency-Key: xxx`

请求参数：
```json
{
  "operator_note": "目标接口已恢复，重试执行"
}
```

响应：同 execute 响应格式

错误码：`CONFLICT`（非 execution_failed 状态）、`IDEMPOTENCY_CONFLICT`

### 5.9 GET /strategy-suggestions/{suggestionId}/execution-logs

查询参数：`page`, `page_size`

响应：
```json
{
  "success": true,
  "data": {
    "items": [ExecutionLogResponse],
    "pagination": { "page": 1, "page_size": 20, "total": 2, "has_next": false }
  },
  "error": null,
  "request_id": "req_xxx"
}
```

错误码：`VALIDATION_ERROR`（分页参数非法）、`NOT_FOUND`（建议不存在）

## 6. Module Design

### 6.1 模块划分

| 模块 | 包路径 | 职责 |
|------|--------|------|
| strategy | `apps/api-server/internal/modules/strategy` | 策略建议业务逻辑、状态机、幂等处理、执行日志 |
| strategy handler | `apps/api-server/internal/http/handlers/strategy.go` | HTTP 参数解析、响应封装、错误映射 |
| strategy store | `apps/api-server/internal/modules/strategy`（同包） | 数据持久化接口与内存实现 |

### 6.2 Service 接口

```go
type Service interface {
    GenerateSuggestions(ctx context.Context, projectID string, req GenerateSuggestionsRequest, idempotencyKey string) (GenerateSuggestionsResponse, error)
    ListSuggestions(ctx context.Context, projectID string, req ListStrategySuggestionsRequest) (PagedStrategySuggestionsResponse, error)
    GetSuggestion(ctx context.Context, suggestionID string) (StrategySuggestionDetailResponse, error)
    ConfirmSuggestion(ctx context.Context, suggestionID string, req ConfirmSuggestionRequest, idempotencyKey string) (SuggestionStatusChangeResponse, error)
    IgnoreSuggestion(ctx context.Context, suggestionID string, req IgnoreSuggestionRequest, idempotencyKey string) (SuggestionStatusChangeResponse, error)
    ExecuteSuggestion(ctx context.Context, suggestionID string, req ExecuteSuggestionRequest, idempotencyKey string) (ExecuteSuggestionResponse, error)
    RetrySuggestion(ctx context.Context, suggestionID string, req RetrySuggestionRequest, idempotencyKey string) (ExecuteSuggestionResponse, error)
    ListExecutionLogs(ctx context.Context, suggestionID string, req ListExecutionLogsRequest) (PagedExecutionLogsResponse, error)
}
```

### 6.3 Store 接口

```go
type Store interface {
    InsertSuggestionRun(ctx context.Context, run StrategySuggestionRunResponse) error
    FindSuggestionRunByID(ctx context.Context, id string) (*StrategySuggestionRunResponse, error)
    UpdateSuggestionRunStatus(ctx context.Context, id, status, failureReason string, suggestionCount int) error

    InsertSuggestion(ctx context.Context, s StrategySuggestionDetailResponse) error
    FindSuggestionByID(ctx context.Context, id string) (*StrategySuggestionDetailResponse, error)
    UpdateSuggestionStatus(ctx context.Context, id, status string, fields map[string]any) error
    ListSuggestions(ctx context.Context, projectID string, req ListStrategySuggestionsRequest) ([]StrategySuggestionDetailResponse, int, error)

    InsertExecutionLog(ctx context.Context, log ExecutionLogResponse) error
    ListExecutionLogs(ctx context.Context, suggestionID string, page, pageSize int) ([]ExecutionLogResponse, int, error)

    CheckIdempotency(ctx context.Context, scope, endpoint, key, hash string) (refType string, refID string, conflict bool, err error)
    StoreIdempotency(ctx context.Context, scope, endpoint, key, hash, refType, refID string) error
}
```

### 6.4 模块间依赖关系

- strategy → content（复用 PaginationRequest / PaginationResponse）
- strategy → metrics（读取 metric_summary_snapshot 作为策略建议输入）
- strategy handler → api（复用 Envelope / WriteSuccess / WriteError）
- strategy 不直接依赖 workflow / agent / llm / publish / schedule 等模块；跨模块执行通过目标接口标识记录，不直接调用

### 6.5 与现有模块的集成方式

- **路由注册**：在 `router.go` 的 `/api/v1` 路由组中新增 8 条路由
- **Handler 创建**：在 `NewRouter` 中实例化 `StrategyHandler`，与现有 Handler 一致
- **Service 创建**：使用 `strategy.NewService()` 创建，与 metrics 模块一致的构造模式
- **前端导航**：在 `workspace-nav.tsx` 的 items 数组中添加"策略建议"入口
- **API Client**：在 `api.ts` 中添加 strategy 相关类型和 fetch 函数

## 7. Output Contract

### 7.1 API Output Contracts

| API | 输入 | 输出 | 产出类型 | 正确性规则 |
|-----|------|------|---------|-----------|
| GenerateSuggestions | projectID, date_from, date_to, rule_codes?, metric_codes?, force_regenerate?, Idempotency-Key | suggestion_run_id, status | web-e2e | date_from/date_to 必填且合法；返回 suggestion_run_id 立即响应；status 必须为 generating |
| ListSuggestions | projectID, status?, suggestion_type?, risk_level?, confidence?, date_from?, date_to?, page, page_size, sort, order | items[], pagination | web-e2e | 分页结构完整；筛选参数非法返回 VALIDATION_ERROR |
| GetSuggestion | suggestionID | 完整建议详情 | web-e2e | 包含全部可解释字段；不存在返回 NOT_FOUND |
| ConfirmSuggestion | suggestionID, note?, Idempotency-Key | suggestion_id, previous_status, current_status, operation_log_id | web-e2e | 状态必须从 pending→confirmed；非 pending 返回 CONFLICT；幂等键一致返回同一结果 |
| IgnoreSuggestion | suggestionID, reason(必填), note?, Idempotency-Key | suggestion_id, previous_status, current_status, operation_log_id | web-e2e | reason 必填；状态必须从 pending→ignored；终态忽略返回 CONFLICT |
| ExecuteSuggestion | suggestionID, action_type, target_type, target_id, operator_note?, Idempotency-Key | execution_log_id, suggestion_id, previous_status, current_status | web-e2e | 状态必须从 confirmed→executed 或 execution_failed；失败时记录 failure_reason |
| RetrySuggestion | suggestionID, operator_note?, Idempotency-Key | execution_log_id, suggestion_id, previous_status, current_status | web-e2e | 状态必须从 execution_failed→executed 或 execution_failed；非失败状态返回 CONFLICT |
| ListExecutionLogs | suggestionID, page, page_size | items[], pagination | web-e2e | 每条包含动作、目标、操作者、前后状态、结果、失败原因 |

### 7.2 Type Tests

根据 `workflow.yaml` 的 `project.features` 和本次迭代变更内容：

- 本次 API Design 有 HTTP endpoint → 触发 `web-e2e`
- 本次涉及前端页面/组件 → 触发 `frontend-ui`
- 本次有跨组件集成（Handler → Service → Store → PostgreSQL）→ 触发 `integration`

| type id | 业务描述 | 组件链路 | 测试规范 |
|---------|---------|---------|---------|
| web-e2e | 策略建议 API 全链路端到端验证 | StrategyHandler → StrategyService → Store → PostgreSQL | standards/testing/web-e2e.md |
| frontend-ui | 策略建议页面交互验证（列表、详情、操作） | api.ts → StrategyPage → StrategyActionsPage | standards/testing/frontend-ui.md |
| integration | 策略建议状态机集成验证 | StrategyHandler → StrategyService → Store | standards/testing/integration.md |

### 7.3 SQL Contract

本迭代策略建议列表查询 SQL：

目标方言：PostgreSQL

预期 SQL 模板（strategy_suggestion 列表查询）：
```sql
SELECT *
FROM strategy_suggestion
WHERE project_id = $1
  AND ($2::text = '' OR status = $2)
  AND ($3::text = '' OR suggestion_type = $3)
  AND ($4::text = '' OR risk_level = $4)
  AND ($5::text = '' OR confidence = $5)
  AND ($6::date IS NULL OR date_from >= $6)
  AND ($7::date IS NULL OR date_to <= $7)
ORDER BY %s %s
LIMIT $8 OFFSET $9
```

关键结构：
- 动态排序字段由 sort/order 参数映射到白名单列（created_at, risk_level, confidence, updated_at）
- 筛选参数使用 `($N::type IS NULL OR ...)` 模式匹配空值跳过

禁止模式：
- 不得将用户输入直接拼接到 ORDER BY 子句
- 不得在 WHERE 中使用未参数化的列名
- 不得使用 `SELECT *` 替代具体列名（骨架阶段允许，04 阶段实现时需明确列名）

执行日志查询 SQL：
```sql
SELECT *
FROM strategy_execution_log
WHERE suggestion_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
```

## 8. Change Log

| 文件 | 变更类型 | 变更原因 |
|------|---------|---------|
| `apps/api-server/internal/modules/strategy/dto.go` | 新增 | 策略建议 DTO 定义 |
| `apps/api-server/internal/modules/strategy/errors.go` | 新增 | 策略建议错误常量 |
| `apps/api-server/internal/modules/strategy/service.go` | 新增 | 策略建议 Service 接口与实现 |
| `apps/api-server/internal/modules/strategy/store.go` | 新增 | Store 接口定义 |
| `apps/api-server/internal/modules/strategy/memory_store.go` | 新增 | Store 内存实现 |
| `apps/api-server/internal/http/handlers/strategy.go` | 新增 | HTTP Handler |
| `apps/api-server/internal/http/router.go` | 修改 | 注册策略建议路由，添加 StrategyHandler 实例化和路由定义 |
| `apps/api-server/migrations/00011_create_strategy_tables.sql` | 新增 | 新增 strategy_suggestion_run、strategy_suggestion、strategy_execution_log 表 |
| `openapi/openapi.yaml` | 修改 | 增加本迭代所有接口 path、schema、响应和 examples |
| `apps/web-admin/lib/api.ts` | 修改 | 增加 strategy 类型定义和 API client 函数 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 修改 | 增加"策略建议"导航入口 |
| `apps/web-admin/app/projects/[projectId]/strategy-suggestions/page.tsx` | 新增 | 策略建议列表页 |
| `apps/web-admin/app/projects/[projectId]/strategy-suggestions/[suggestionId]/page.tsx` | 新增 | 建议详情页 |
| `apps/web-admin/app/projects/[projectId]/strategy-suggestions/[suggestionId]/actions/page.tsx` | 新增 | 建议操作页 |

## 9. Development Tasks

- Task-01：定义 strategy DTO、错误常量和 Service 接口
  - 任务类型：contract
  - 所属模块：api-server/strategy
  - 简要描述：定义 DTO、错误常量、Service 接口和 Store 接口，供测试编译引用。
  - 涉及接口/方法：strategy.Service、strategy.Store、NewService()
  - 输入：各 API request DTO
  - 输出：各 API response DTO 或 error
  - 依赖任务：无
  - 数据操作：无
  - 修改边界：只新增接口、DTO、错误常量和 panic 空实现
  - 禁止行为：不得写业务逻辑；不得访问数据库或外部系统
  - 产出类型：integration
  - 功能类型：后端模块接口契约（type id: integration）
  - 是否跨组件：否

- Task-02：创建 strategy 数据库迁移
  - 任务类型：migration
  - 所属模块：api-server/migrations
  - 简要描述：创建 strategy_suggestion_run、strategy_suggestion、strategy_execution_log 三张表的 DDL 迁移脚本。
  - 涉及接口/方法：无
  - 输入：无
  - 输出：00011_create_strategy_tables.sql
  - 依赖任务：无
  - 数据操作：写 strategy_suggestion_run 表；写 strategy_suggestion 表；写 strategy_execution_log 表
  - 修改边界：只新增迁移文件
  - 禁止行为：不得修改已有迁移文件；不得修改现有表结构
  - 产出类型：sql-query
  - 功能类型：数据库迁移脚本（type id: sql-query）
  - 是否跨组件：否

- Task-03：实现策略建议生成业务逻辑
  - 任务类型：business-implementation
  - 所属模块：api-server/strategy
  - 简要描述：校验日期范围、读取项目指标数据、按规则生成策略建议、写入 suggestion_run 和 suggestion 记录，立即返回 suggestion_run_id。
  - 涉及接口/方法：GenerateSuggestions()
  - 输入：GenerateSuggestionsRequest、projectID、idempotencyKey
  - 输出：GenerateSuggestionsResponse 或 error
  - 依赖任务：Task-01（Service 接口）、Task-02（表定义）
  - 数据操作：读 metric_summary_snapshot 表；读 metric_record 表；写 strategy_suggestion_run 表；写 strategy_suggestion 表；读写 idempotency 记录
  - 修改边界：只替换 GenerateSuggestions() 的空实现，不删除或重写 service.go
  - 禁止行为：不得使用内存存储替代数据库表；不得阻塞 HTTP 请求等待生成完成；不得接受前端直接传入的指标结论
  - 产出类型：integration
  - 功能类型：策略建议生成实现（type id: integration）
  - 是否跨组件：是（组件链路：StrategyHandler → StrategyService → MetricsStore → PostgreSQL）

- Task-04：实现策略建议列表与详情查询
  - 任务类型：business-implementation
  - 所属模块：api-server/strategy
  - 简要描述：按项目 ID 查询策略建议分页列表，支持状态/类型/风险/置信度/日期筛选和排序；按 ID 查询单条建议完整详情。
  - 涉及接口/方法：ListSuggestions()、GetSuggestion()
  - 输入：projectID、ListStrategySuggestionsRequest / suggestionID
  - 输出：PagedStrategySuggestionsResponse / StrategySuggestionDetailResponse 或 error
  - 依赖任务：Task-01（Service 接口）、Task-02（表定义）
  - 数据操作：读 strategy_suggestion 表
  - 修改边界：只替换 ListSuggestions() 和 GetSuggestion() 的空实现
  - 禁止行为：不得使用内存存储替代数据库表
  - 产出类型：web-e2e
  - 功能类型：策略建议查询实现（type id: web-e2e）
  - 是否跨组件：是（组件链路：StrategyHandler → StrategyService → PostgreSQL）

- Task-05：实现策略建议确认与忽略状态流转
  - 任务类型：business-implementation
  - 所属模块：api-server/strategy
  - 简要描述：校验当前状态为 pending 后执行确认（pending→confirmed）或忽略（pending→ignored），写入 operation_log，支持幂等。
  - 涉及接口/方法：ConfirmSuggestion()、IgnoreSuggestion()
  - 输入：suggestionID、ConfirmSuggestionRequest / IgnoreSuggestionRequest、idempotencyKey
  - 输出：SuggestionStatusChangeResponse 或 error
  - 依赖任务：Task-01（Service 接口）、Task-02（表定义）
  - 数据操作：读 strategy_suggestion 表；写 strategy_suggestion 表（状态更新）；写 operation_log 表；读写 idempotency 记录
  - 修改边界：只替换 ConfirmSuggestion() 和 IgnoreSuggestion() 的空实现
  - 禁止行为：确认后不得修改项目配置、调度计划、Prompt、模型配置、发布状态或内容生产状态；不得使用内存存储替代数据库表
  - 产出类型：web-e2e
  - 功能类型：策略建议状态流转实现（type id: web-e2e）
  - 是否跨组件：是（组件链路：StrategyHandler → StrategyService → Store → operation_log）

- Task-06：实现策略建议执行与重试执行
  - 任务类型：business-implementation
  - 所属模块：api-server/strategy
  - 简要描述：校验当前状态后执行建议（confirmed→executed/execution_failed）或重试（execution_failed→executed/execution_failed），追加 strategy_execution_log，写入 operation_log，支持幂等。
  - 涉及接口/方法：ExecuteSuggestion()、RetrySuggestion()
  - 输入：suggestionID、ExecuteSuggestionRequest / RetrySuggestionRequest、idempotencyKey
  - 输出：ExecuteSuggestionResponse 或 error
  - 依赖任务：Task-01（Service 接口）、Task-02（表定义）、Task-05（确认流转，执行依赖确认状态）
  - 数据操作：读 strategy_suggestion 表；写 strategy_suggestion 表（状态更新）；写 strategy_execution_log 表；写 operation_log 表；读写 idempotency 记录
  - 修改边界：只替换 ExecuteSuggestion() 和 RetrySuggestion() 的空实现
  - 禁止行为：不得绕过其他领域服务边界直接改写跨模块状态；不得使用内存存储替代数据库表；缺少可用目标接口时执行应失败并展示可解释原因
  - 产出类型：integration
  - 功能类型：策略建议执行实现（type id: integration）
  - 是否跨组件：是（组件链路：StrategyHandler → StrategyService → Store → operation_log → strategy_execution_log）

- Task-07：实现策略执行日志查询
  - 任务类型：business-implementation
  - 所属模块：api-server/strategy
  - 简要描述：按建议 ID 分页查询执行日志，每条包含操作者、执行动作、目标资源、前后状态、结果和失败原因。
  - 涉及接口/方法：ListExecutionLogs()
  - 输入：suggestionID、ListExecutionLogsRequest
  - 输出：PagedExecutionLogsResponse 或 error
  - 依赖任务：Task-01（Service 接口）、Task-02（表定义）
  - 数据操作：读 strategy_execution_log 表
  - 修改边界：只替换 ListExecutionLogs() 的空实现
  - 禁止行为：不得使用内存存储替代数据库表
  - 产出类型：web-e2e
  - 功能类型：执行日志查询实现（type id: web-e2e）
  - 是否跨组件：否

- Task-08：实现 Store 内存实现
  - 任务类型：business-implementation
  - 所属模块：api-server/strategy
  - 简要描述：实现 Store 接口的内存版本，包含 suggestion_run、suggestion、execution_log、idempotency 的 CRUD，供单元测试和非 PostgreSQL 环境使用。
  - 涉及接口/方法：strategy.Store 所有方法
  - 输入：各 Store 方法参数
  - 输出：各 Store 方法返回值
  - 依赖任务：Task-01（Store 接口定义）
  - 数据操作：无（内存 map）
  - 修改边界：只新增 memory_store.go 文件
  - 禁止行为：不得访问数据库或外部系统
  - 产出类型：integration
  - 功能类型：Store 内存实现（type id: integration）
  - 是否跨组件：否

- Task-09：注册策略建议路由与 Handler
  - 任务类型：api
  - 所属模块：api-server/http
  - 简要描述：在 router.go 中注册 8 条策略建议路由，实例化 StrategyHandler 并挂载到 /api/v1 路由组。
  - 涉及接口/方法：NewRouter()、NewStrategyHandler()
  - 输入：StrategyService、logger
  - 输出：8 条新路由注册
  - 依赖任务：Task-01（Service 接口）
  - 数据操作：无
  - 修改边界：在 router.go 中只新增 StrategyHandler 变量声明和路由注册代码块
  - 禁止行为：不得修改已有路由；不得删除或重写现有 Handler 实例化代码
  - 产出类型：web-e2e
  - 功能类型：路由注册（type id: web-e2e）
  - 是否跨组件：是（组件链路：Router → StrategyHandler → StrategyService）

- Task-10：实现策略建议 HTTP Handler
  - 任务类型：api
  - 所属模块：api-server/http/handlers
  - 简要描述：实现 StrategyHandler 的全部 HTTP 方法，负责请求解析、参数校验、Service 调用、统一响应封装和错误映射。
  - 涉及接口/方法：GenerateSuggestions()、ListSuggestions()、GetSuggestion()、ConfirmSuggestion()、IgnoreSuggestion()、ExecuteSuggestion()、RetrySuggestion()、ListExecutionLogs()（Handler 方法）
  - 输入：http.Request（路径参数、查询参数、请求体、Idempotency-Key header）
  - 输出：api.Envelope 统一响应
  - 依赖任务：Task-01（Service 接口、DTO）
  - 数据操作：无
  - 修改边界：只新增 handlers/strategy.go 文件
  - 禁止行为：不得在 Handler 中写业务逻辑；不得直接操作数据库
  - 产出类型：web-e2e
  - 功能类型：HTTP Handler 实现（type id: web-e2e）
  - 是否跨组件：是（组件链路：StrategyHandler → StrategyService → api.Envelope）

- Task-11：添加前端策略建议 API client
  - 任务类型：api
  - 所属模块：web-admin/lib
  - 简要描述：在 api.ts 中增加 strategy 相关的 TypeScript 类型定义和 fetch 函数。
  - 涉及接口/方法：fetchStrategySuggestions()、fetchStrategySuggestion()、generateStrategySuggestions()、confirmSuggestion()、ignoreSuggestion()、executeSuggestion()、retrySuggestion()、fetchExecutionLogs()
  - 输入：各函数参数
  - 输出：APIEnvelope<T> 响应
  - 依赖任务：Task-01（DTO 定义，用于类型对齐）
  - 数据操作：无
  - 修改边界：在 api.ts 中只新增类型定义和函数，不修改已有代码
  - 禁止行为：不得修改已有的 fetch 函数和类型定义
  - 产出类型：frontend-ui
  - 功能类型：前端 API client（type id: frontend-ui）
  - 是否跨组件：否

- Task-12：添加前端策略建议导航入口
  - 任务类型：ui
  - 所属模块：web-admin/workspace-nav
  - 简要描述：在 workspace-nav.tsx 的 items 数组中添加"策略建议"导航项，指向 strategy-suggestions。
  - 涉及接口/方法：无
  - 输入：无
  - 输出：新增导航项
  - 依赖任务：无
  - 数据操作：无
  - 修改边界：在 workspace-nav.tsx 的 items 数组中只新增一项
  - 禁止行为：不得修改已有导航项；不得删除现有导航入口
  - 产出类型：frontend-ui
  - 功能类型：导航入口（type id: frontend-ui）
  - 是否跨组件：否

- Task-13：实现前端策略建议列表页
  - 任务类型：ui
  - 所属模块：web-admin/strategy-suggestions
  - 简要描述：实现 /projects/[projectId]/strategy-suggestions 页面，包含建议卡片列表、筛选控件、分页、空态/加载态/错误态/成功态，样式与现有管理台一致。
  - 涉及接口/方法：fetchStrategySuggestions()、generateStrategySuggestions()
  - 输入：projectId、筛选参数
  - 输出：渲染页面
  - 依赖任务：Task-11（API client）
  - 数据操作：无
  - 修改边界：只新增 page.tsx 文件
  - 禁止行为：不得实现裸 HTML 或占位页；不得使用与现有管理台不一致的样式
  - 产出类型：frontend-ui
  - 功能类型：策略建议列表页（type id: frontend-ui）
  - 是否跨组件：是（组件链路：StrategyPage → api.ts → Go API → PostgreSQL）

- Task-14：实现前端建议详情页
  - 任务类型：ui
  - 所属模块：web-admin/strategy-suggestions/[suggestionId]
  - 简要描述：实现 /strategy-suggestions/[suggestionId] 页面，展示建议原因、影响范围、证据指标、metrics_snapshot、状态、置信度、风险说明和执行记录摘要，失败态展示 request_id。
  - 涉及接口/方法：fetchStrategySuggestion()、fetchExecutionLogs()
  - 输入：suggestionId
  - 输出：渲染详情页
  - 依赖任务：Task-11（API client）
  - 数据操作：无
  - 修改边界：只新增 page.tsx 文件
  - 禁止行为：不得实现裸 HTML 或占位页
  - 产出类型：frontend-ui
  - 功能类型：建议详情页（type id: frontend-ui）
  - 是否跨组件：是（组件链路：DetailPage → api.ts → Go API → PostgreSQL）

- Task-15：实现前端建议操作页
  - 任务类型：ui
  - 所属模块：web-admin/strategy-suggestions/[suggestionId]/actions
  - 简要描述：实现 /strategy-suggestions/[suggestionId]/actions 页面，提供确认、忽略、执行或重试执行的操作入口，按当前状态禁用不合法动作，表单字段缺失或状态冲突时展示错误反馈。
  - 涉及接口/方法：confirmSuggestion()、ignoreSuggestion()、executeSuggestion()、retrySuggestion()
  - 输入：suggestionId、操作表单数据
  - 输出：Toast/Alert 反馈、页面数据刷新
  - 依赖任务：Task-11（API client）
  - 数据操作：无
  - 修改边界：只新增 page.tsx 文件
  - 禁止行为：不得在非法状态下启用操作按钮；不得实现裸 HTML 或占位页
  - 产出类型：frontend-ui
  - 功能类型：建议操作页（type id: frontend-ui）
  - 是否跨组件：是（组件链路：ActionsPage → api.ts → Go API → PostgreSQL）
