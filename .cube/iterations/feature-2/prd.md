# PRD：Iteration 2 — Workflow Engine 与多 Agent 架构

> 基于需求文档：`docs/requirements/iteration-2-multi-agent-architecture.md`  
> 撰写日期：2026-05-15  
> 需求确认日期：2026-05-15

---

## 1. 功能概述

Iteration 2 的目标是建立**可运行、可追踪、可审计**的 Go 自研 Workflow Engine 与 Agent Runtime。它负责把 WorkflowTemplateVersion 转换为 WorkflowRun / WorkflowStepRun / AgentTask / LLMCallLog 的完整执行链路，为 Iteration 3 之后的 Novel Pack 内容规划和生成提供底座。调度运行时、生产计划和成本汇总放到 Iteration 2.1，不在本迭代实现。

### 三项核心约束

| 约束 | 说明 |
|------|------|
| **WorkflowRun 是唯一运行入口** | AgentTask 不能被前端或外部系统单独触发，必须由 WorkflowStepRun 派生，保证链路可追踪 |
| **本迭代只支持轻量顺序工作流** | 不做复杂 DAG、条件分支、并行编排；步骤按模板版本定义顺序执行 |
| **LLMCallLog 是基础数据资产** | 每次 LLM 调用都必须绑定 agent_task_id / workflow_run_id / provider / model / token / cost / error，为后续成本统计、失败追踪、模型质量分析提供基础 |

包含以下功能点：

1. **工作流模板管理**：创建、查询、版本化、发布工作流模板。
2. **工作流运行管理**：手动触发、查询运行详情、查询步骤、取消、失败后整体重试。
3. **Agent 执行记录**：AgentTask 创建、状态跟踪、输入输出和错误记录。
4. **LLM 调用日志**：记录每次模型调用的模型、Provider、Token、成本、耗时、错误和关联信息。
5. **前端管理页面**：工作流模板管理、运行记录、AgentTask 详情、LLM 调用日志。
6. **调度接口占位**：WorkflowSchedule / ProductionPlan 本迭代仅提供接口契约，不实现运行时。

---

## 2. Functional Requirements

### 工作流模板管理

| 编号 | 功能描述 | 输入 | 输出 | 异常处理 | 优先级 |
|------|---------|------|------|----------|--------|
| FR-001 | 查看工作流模板列表，支持按内容类型、分类、状态筛选 | 筛选条件（content_type、category、status）、分页参数 | 模板分页列表 | 参数非法返回 `VALIDATION_ERROR`；列表为空展示空状态 | P0 |
| FR-002 | 创建工作流模板 | 模板名称、code、内容类型、分类、描述 | 新建 template_id 和初始状态 | 名称或 code 重复返回 `CONFLICT`；必填字段缺失返回 `VALIDATION_ERROR` | P0 |
| FR-003 | 查看工作流模板详情 | template_id | 模板详情 | 不存在返回 `NOT_FOUND` | P0 |
| FR-004 | 查看模板版本列表 | template_id、分页参数 | 版本分页列表 | 模板不存在返回 `NOT_FOUND` | P0 |
| FR-005 | 创建模板版本（包含 StepTemplate 列表） | input_schema、output_schema、steps（含 step_type 和 order_index） | 新建 template_version_id、step 数量、初始状态 | step 格式非法返回 `VALIDATION_ERROR` | P0 |
| FR-006 | 查看模板版本详情（含步骤列表） | template_version_id | 版本详情 + steps | 不存在返回 `NOT_FOUND` | P0 |
| FR-007 | 发布模板版本；发布动作写入操作日志 | template_version_id、发布备注 | 变更前后状态、operation_log_id | 版本状态不是 draft 返回 `CONFLICT`；已发布版本不允许重复发布 | P0 |
| FR-008 | WorkflowStepTemplate 支持 `agent`、`human_review`、`condition`、`system_task` 四类节点 | step_type 字段 | 合法节点记录 | 不在枚举范围内返回 `VALIDATION_ERROR` | P0 |
| FR-009 | 已发布的 WorkflowTemplateVersion 不允许直接修改；修改必须创建新版本 | — | — | 尝试修改已发布版本返回 `CONFLICT` | P0 |

### 工作流运行管理

| 编号 | 功能描述 | 输入 | 输出 | 异常处理 | 优先级 |
|------|---------|------|------|----------|--------|
| FR-010 | 查看 WorkflowRun 列表，支持按 project_id、template_version_id、status 筛选 | 筛选条件、分页参数 | 运行分页列表 | 参数非法返回 `VALIDATION_ERROR` | P0 |
| FR-011 | 手动触发 WorkflowRun；HTTP 请求立即返回 run_id 和初始状态，实际执行由异步 worker 完成 | project_id、template_version_id、input | workflow_run_id、初始 status（pending） | 模板版本未发布返回 `CONFLICT`；必填字段缺失返回 `VALIDATION_ERROR` | P0 |
| FR-012 | 查看 WorkflowRun 详情（含 input、output、error） | workflow_run_id | 运行详情 | 不存在返回 `NOT_FOUND` | P0 |
| FR-013 | 查看 WorkflowRun 下的 StepRun 列表 | workflow_run_id | StepRun 列表 | 不存在返回 `NOT_FOUND` | P0 |
| FR-014 | 取消运行中的 WorkflowRun；取消动作写入操作日志 | workflow_run_id、reason 或 note | 变更前后状态、operation_log_id | 非运行中状态返回 `CONFLICT` | P0 |
| FR-015 | 对失败的 WorkflowRun 发起整体重试；重试创建新 run，保留原 run | workflow_run_id、reason、可选 input_override | 新 workflow_run_id、初始 status | 原 run 非失败状态返回 `CONFLICT` | P0 |

### Agent 执行记录

| 编号 | 功能描述 | 输入 | 输出 | 异常处理 | 优先级 |
|------|---------|------|------|----------|--------|
| FR-016 | 查看 AgentTask 列表，支持按 workflow_run_id、step_run_id、agent_code、status 筛选 | 筛选条件、分页参数 | AgentTask 分页列表 | 参数非法返回 `VALIDATION_ERROR` | P0 |
| FR-017 | 查看 AgentTask 详情（含 input、output、status、error、关联 StepRun 和 LLM 日志） | agent_task_id | AgentTask 详情 | 不存在返回 `NOT_FOUND` | P0 |
| FR-018 | AgentTask 必须由 WorkflowStepRun 派生；不允许绕过 WorkflowRun 单独创建 AgentTask | — | — | 直接创建 AgentTask 的请求应被拒绝（仅内部创建） | P0 |
| FR-019 | AgentTask 输出结构化校验失败时状态置为 `failed`，错误码使用 `AGENT_OUTPUT_INVALID` | — | — | 校验失败时记录错误信息 | P0 |

### LLM 调用日志

| 编号 | 功能描述 | 输入 | 输出 | 异常处理 | 优先级 |
|------|---------|------|------|----------|--------|
| FR-020 | 查看 LLM 调用日志列表，支持按 workflow_run_id、agent_task_id、provider、model、status 筛选 | 筛选条件、分页参数 | LLM 日志分页列表 | 参数非法返回 `VALIDATION_ERROR` | P0 |
| FR-021 | 查看 LLM 调用日志详情（含 provider、model、Token、成本、耗时、错误、request_id） | llm_call_log_id | 日志详情 | 不存在返回 `NOT_FOUND` | P0 |

### 前端管理页面

| 编号 | 功能描述 | 输入 | 输出 | 异常处理 | 优先级 |
|------|---------|------|------|----------|--------|
| FR-022 | 工作流模板管理页：列表展示、筛选、新建、查看版本、创建版本、发布版本 | 用户操作 | 对应 API 调用和反馈 | 操作失败展示错误码、错误信息和 request_id | P0 |
| FR-023 | 工作流运行记录页：运行列表、筛选、手动运行、查看详情、取消、失败重试 | 用户操作 | 对应 API 调用和反馈 | 操作失败展示错误码、错误信息和 request_id | P0 |
| FR-024 | 工作流运行详情页：运行状态、输入输出、StepRun 列表、关联 AgentTask、关联 LLM 日志 | workflow_run_id | 综合详情展示 | 数据加载失败展示错误态和 request_id | P0 |
| FR-025 | AgentTask 列表页：按 workflow_run_id、status、agent_code 筛选 | 筛选条件 | AgentTask 列表 | 加载失败展示错误态 | P0 |
| FR-026 | AgentTask 详情页：查看 input、output、错误、关联 StepRun、关联 LLM 日志 | agent_task_id | 详情展示 | 加载失败展示错误态和 request_id | P0 |
| FR-027 | LLM 调用日志页：按 workflow_run_id、agent_task_id、provider、model、status 筛选 | 筛选条件 | LLM 日志列表 | 加载失败展示错误态 | P0 |
| FR-028 | 所有页面支持空状态、加载态、错误态和成功操作反馈 | — | — | 失败必须展示错误码、错误信息和 request_id | P0 |

### 调度接口占位

| 编号 | 功能描述 | 输入 | 输出 | 异常处理 | 优先级 |
|------|---------|------|------|----------|--------|
| FR-029 | 提供 `POST /api/v1/workflow-schedules` 接口契约占位；不实现调度运行时 | 调度参数（Contract only） | schedule_id skeleton | 返回占位响应，不做实际调度 | P1 |

---

## 3. Non-Functional Requirements

- **幂等性**：创建 WorkflowRun、取消 WorkflowRun、重试 WorkflowRun、发布模板版本等接口必须支持 `Idempotency-Key` 保护，重复请求不得产生副作用。
- **异步执行**：WorkflowRun 触发后由异步 worker 执行，HTTP 请求响应时间 < 500ms。
- **操作审计**：所有状态变更（发布、取消、重试）必须写入 `operation_log`，不得遗漏。
- **可追踪性**：AgentTask 必须关联 WorkflowRun 和 StepRun；LLMCallLog 必须关联 AgentTask。
- **API 规范**：所有接口遵守项目统一 `success/data/error/request_id` 响应格式；分页接口支持 `page`、`page_size`、`sort`、`order`。

---

## 4. 验收标准

- WorkflowTemplate 可以创建、查询、筛选；重复 code 不能创建成功。
- WorkflowTemplateVersion 可以创建、查询、发布；已发布版本拒绝直接修改。
- WorkflowStepTemplate 支持 `agent`、`human_review`、`condition`、`system_task` 四类节点；非法 step_type 被拒绝。
- 手动触发 WorkflowRun 时 HTTP 请求立即返回 `workflow_run_id` 和 `pending` 状态，实际执行由 worker 异步完成。
- WorkflowRun 可以查询详情、StepRun 列表、关联 AgentTask、关联 LLMCallLog。
- WorkflowRun 可以取消（仅运行中状态），取消后 operation_log 中有记录。
- 失败 WorkflowRun 可以整体重试，新建 run 有 parent_run_id 指向原 run，原 run 保留。
- AgentTask 只能由 WorkflowStepRun 内部派生，不能通过 API 直接创建。
- AgentTask 详情包含 input、output、status、error、started_at、finished_at。
- LLMCallLog 详情包含 provider、model、input_tokens、output_tokens、cost、currency、latency_ms、status、error、request_id。
- 列表接口支持分页、多条件筛选、排序。
- 创建 WorkflowRun、发布版本、取消、重试接口在相同 Idempotency-Key 下重复调用不产生副作用。
- 前端页面可以完成工作流模板管理、运行记录查看、AgentTask 详情、LLM 日志查询的主要交互。
- 页面上的主要操作按钮均有加载态和结果反馈；失败反馈展示错误码、错误信息和 request_id。
- Core 层代码中不出现 Novel / Book / Chapter 作为核心资源命名。
- 本迭代完成后，Iteration 3 可以直接基于 WorkflowTemplate / WorkflowRun / AgentTask / LLMCallLog 编排 Novel Pack 新书规划流程。

---

## 5. Out of Scope

- 不做 Novel 新书规划、世界观、人物、大纲等业务能力。
- 不做章节正文生成。
- 不做审稿、发布、指标、策略业务。
- 不做复杂 DAG、并行网关、补偿事务、长事务编排。
- 不做 WorkflowSchedule / ProductionPlan 运行时（仅 Contract only 占位）。
- 不做 n8n 核心编排（n8n 仅用于后续通知、Webhook、外部 API 同步）。
- 不做 LLM 成本汇总统计接口（放入 Iteration 2.1）。
- 不做绕过 WorkflowRun / AgentTask / LLMCallLog 追踪链路的核心生产链路。

---

## 6. 依赖说明

- 复用 Iteration 1 已建立的 PromptTemplate 和 LLM Provider 能力，本迭代不重复实现。
- 工作流异步执行依赖 worker / queue 能力（使用 asynq 或等价方案，具体技术选型在 02-design 阶段确认）。
