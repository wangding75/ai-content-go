# Iteration 2：Workflow Engine 与多 Agent 架构

> 文件定位：本文件是 Iteration 2 的最新需求、技术、前端页面、接口契约和原型映射文档。  
> 蓝图约束：必须遵守 `00-product-blueprint.md`。  
> 技术栈：后端 Go / Golang；前端 Next.js + TypeScript。  
> 更新日期：2026-05-15  
> 评审日期：2026-05-15  
> 评审结论摘要：评审通过。Iteration 2 与产品蓝图无背离；本次评审将原有泛化描述补充为可开发、可验收的 Workflow Engine、Agent Runtime、LLMCallLog、接口契约和前端页面范围，并明确 WorkflowSchedule / ProductionPlan 在本迭代仅做 Contract only，运行时能力进入 Iteration 2.1。  
> 是否需要更新蓝图：否。  
> 重要规则：本项目不再保留单独前端迭代文档；前端需求已整合到本迭代。

---

## 1. 迭代目标

建立 Go 自研轻量 Workflow Engine、Agent Runtime、LLMCallLog，为后续 Novel Pack 的新书规划、内容生成、审稿、记忆系统提供统一执行底座。

本迭代必须完成：

- 工作流模板管理：创建、查询、版本化、发布。
- 工作流运行管理：手动触发、查询运行详情、查询步骤、取消、失败后整体重试。
- Agent Runtime：基于 StepRun 创建 AgentTask，执行 Prompt 渲染、调用 LLM Provider、结构化输出校验，并记录状态。
- LLM 调用日志：记录模型、Provider、Token、成本、耗时、错误、请求关联关系。
- 前端管理页面：工作流模板管理、运行记录、AgentTask 详情、LLM 调用日志。
- WorkflowSchedule / ProductionPlan：本迭代只提供接口契约占位，不实现调度运行时；调度运行时进入 Iteration 2.1。

本迭代完成后，Iteration 3 可以基于 `WorkflowTemplate` / `WorkflowRun` / `AgentTask` / `LLMCallLog` 编排 Novel Pack 新书规划流程，而不需要重复实现核心工作流或 Agent 执行能力。

---

## 2. 蓝图对齐说明

| 蓝图约束 | 本迭代对齐方式 |
|---|---|
| Core 内容类型无关 | Core 仅使用 WorkflowTemplate、WorkflowRun、AgentTask、ContentProject、ContentType 等通用资源；不得出现 Book / Chapter 作为 Core 资源 |
| 内容类型插件化 | 本迭代只提供通用工作流和 Agent 执行底座；Novel / Article / Social Post 的业务语义由 Content Pack 在后续迭代扩展 |
| 工作流自研 | 核心生产链路由 Go Workflow Engine 承载，不依赖 n8n 执行核心 Agent 编排 |
| Agent 可追踪 | 每次 Agent 执行必须落 AgentTask；每次模型调用必须落 LLMCallLog |
| 人工节点保留 | WorkflowStepTemplate 必须支持 `human_review` 类型，但本迭代只实现节点定义和状态，不实现审稿业务 |
| n8n 外围化 | n8n 只作为后续通知、Webhook、外部 API 同步、告警集成；本迭代不做 n8n 核心编排 |
| 前后端整合 | 工作流模板、运行记录、AgentTask、LLM 日志页面必须有明确接口映射 |
| Go 后端统一 | API Server、Workflow Engine、Agent Runtime、Worker 均采用 Go 实现 |

### 2.1 初步对齐检查

| 检查项 | 初始需求描述 | 蓝图定义 | 状态 |
|---|---|---|---|
| Workflow Engine | 建立 Go 自研轻量 Workflow Engine | 核心生产链路由自研 Workflow Engine 承载 | ✅ 对齐 |
| Agent Runtime | 建立 Agent Runtime、AgentTask、LLMCallLog | AgentTask、LLMCallLog 必须记录输入、输出、模型、Token、成本、错误 | ✅ 对齐 |
| Core 内容类型无关 | Core 不使用 Book / Chapter 作为核心资源 | Core 层不得写死 Novel / Book / Chapter | ✅ 对齐 |
| n8n 边界 | n8n 只做通知、Webhook、同步、告警 | n8n 不承载核心 Agent 编排 | ✅ 对齐 |
| WorkflowSchedule / ProductionPlan | 本迭代只做 Contract only | 每天生成 5 个 ContentItem 由 WorkflowSchedule + ProductionPlan 承载 | ✅ 对齐，运行时放入 Iteration 2.1 |
| 页面-接口映射 | 初始需求已有页面和接口清单，但路由与交互描述偏泛 | 每个迭代必须包含前端页面、后端接口、页面-接口映射 | ⚠️ 已补充为明确页面和 API |
| 接口契约 | 初始接口只列输入输出摘要，缺少状态、幂等、错误和详情接口要求 | API 必须有 DTO、OpenAPI、统一错误、分页、幂等 | ⚠️ 已补充到最终需求 |

结论：不存在蓝图冲突；存在需求细化不足，已在本文件内修订。

---

## 3. 产品需求

- 用户可以查看工作流模板列表，并按内容类型、分类、状态筛选。
- 用户可以创建工作流模板，并为模板创建版本。
- 用户可以配置 WorkflowStepTemplate，支持 `agent`、`human_review`、`condition`、`system_task` 四类基础节点。
- 用户可以发布工作流模板版本；发布动作必须写入 `operation_log`。
- 用户可以基于已发布模板版本手动触发一次 WorkflowRun。
- 用户可以查看 WorkflowRun 列表、运行详情、StepRun 列表、AgentTask 列表、LLMCallLog 列表。
- 用户可以取消运行中的 WorkflowRun；取消动作必须写入 `operation_log`。
- 用户可以对失败的 WorkflowRun 发起整体重试；重试必须创建新的 WorkflowRun，并关联原始 run。
- 用户可以查看 AgentTask 的 input、output、状态、错误、重试来源和关联 StepRun。
- 用户可以查看 LLM 调用日志，包括 Provider、模型、Token、成本、耗时、错误和 request_id。
- 页面必须支持空状态、加载态、错误态、成功反馈。
- 所有失败反馈必须展示错误码、错误信息和 request_id。

---

## 4. Go 后端技术需求

### 4.1 实现范围

- `workflow_template`
- `workflow_template_version`
- `workflow_step_template`
- `workflow_run`
- `workflow_step_run`
- `agent_task`
- `llm_call_log`
- `operation_log`

### 4.2 Workflow Engine 要求

- 采用 Go 实现轻量顺序执行引擎。
- 本迭代只支持线性步骤执行；`condition` 节点只允许简单条件跳过，不实现复杂 DAG。
- WorkflowTemplate 负责定义工作流元信息。
- WorkflowTemplateVersion 负责保存 input_schema、output_schema、steps 和版本状态。
- WorkflowRun 负责记录一次执行实例，必须关联 `project_id`、`template_version_id`、`status`、`input`、`output`、`error`。
- WorkflowStepRun 负责记录单步执行状态，必须关联 `workflow_run_id`、`step_template_id`、`status`、`input`、`output`、`error`。
- 发布后的 WorkflowTemplateVersion 不允许直接修改；修改必须创建新版本。
- WorkflowRun 必须通过异步 worker 执行，HTTP 请求只返回 `workflow_run_id` 和初始状态。

### 4.3 Agent Runtime 要求

- AgentTask 必须由 WorkflowStepRun 派生，不允许绕过 WorkflowRun 单独执行核心 Agent。
- AgentTask 必须记录：`workflow_run_id`、`step_run_id`、`agent_code`、`prompt_template_id`、`input`、`output`、`status`、`error`、`started_at`、`finished_at`。
- PromptTemplate 和 LLM Provider 复用 Iteration 1 已建立的能力，不在本迭代重复定义。
- Agent 输出必须支持结构化校验；校验失败状态为 `failed`，错误码使用 `AGENT_OUTPUT_INVALID`。
- LLM Provider 调用失败状态为 `failed`，错误码使用 `LLM_PROVIDER_ERROR`。

### 4.4 LLMCallLog 要求

- 每次模型调用必须记录 LLMCallLog。
- LLMCallLog 必须关联 `workflow_run_id`、`step_run_id`、`agent_task_id`。
- 必须记录 provider、model、input_tokens、output_tokens、cost、currency、latency_ms、status、error、request_id。
- 本迭代提供明细查询；成本汇总接口放入 Iteration 2.1。

### 4.5 通用实现要求

- 使用 Go struct 定义 request / response DTO。
- 使用 validator 做入参校验。
- 使用 sqlc + pgx 或等价方式访问 PostgreSQL。
- 使用 goose 或等价工具管理数据库迁移。
- 所有接口进入 OpenAPI。
- 列表接口必须支持 `page`、`page_size`、`sort`、`order`。
- 创建 WorkflowRun、取消 WorkflowRun、重试 WorkflowRun、发布模板版本必须支持 `Idempotency-Key` 或等价幂等保护。
- 状态变更必须写入 `operation_log`。
- 异步任务通过 worker / queue 执行，不阻塞 HTTP 请求。

---

## 5. 数据模型

| 模型 | 核心字段 | 说明 |
|---|---|---|
| workflow_template | id、code、name、content_type、category、status | 工作流模板主表 |
| workflow_template_version | id、template_id、version、input_schema、output_schema、status | 工作流版本，不可直接修改已发布版本 |
| workflow_step_template | id、template_version_id、step_code、step_type、agent_code、order_index、input_mapping、output_mapping | 工作流步骤模板 |
| workflow_run | id、project_id、template_version_id、status、input、output、error、source、parent_run_id | 一次工作流运行实例 |
| workflow_step_run | id、workflow_run_id、step_template_id、status、input、output、error、started_at、finished_at | 单个步骤运行记录 |
| agent_task | id、workflow_run_id、step_run_id、agent_code、prompt_template_id、status、input、output、error | Agent 执行记录 |
| llm_call_log | id、workflow_run_id、step_run_id、agent_task_id、provider、model、tokens、cost、latency_ms、status、error | 模型调用日志 |
| operation_log | id、resource_type、resource_id、action、reason、note、operator_id、created_at | 状态变更记录 |

### 5.1 状态枚举

| 对象 | 状态 |
|---|---|
| workflow_template | draft、active、archived |
| workflow_template_version | draft、published、deprecated |
| workflow_run | pending、running、success、failed、cancelled |
| workflow_step_run | pending、running、success、failed、skipped、cancelled |
| agent_task | pending、running、success、failed、cancelled |
| llm_call_log | success、failed |

---

## 6. 前端页面范围

| 页面 / 组件 | 路由建议 | 交互 |
|---|---|---|
| 工作流模板管理 | `/workflow/templates` | 模板列表、筛选、新建模板、查看版本、创建版本、发布版本 |
| 工作流运行记录 | `/workflow/runs` | 运行列表、筛选、手动运行、查看详情、取消、失败重试 |
| 工作流运行详情 | `/workflow/runs/[id]` | 查看运行状态、输入输出、StepRun、关联 AgentTask、关联 LLM 日志 |
| AgentTask 列表 | `/agent/tasks` | 按 workflow_run_id、status、agent_code 筛选 |
| AgentTask 详情 | `/agent/tasks/[id]` | 查看 input、output、错误、关联 StepRun、关联 LLM 日志 |
| LLM 调用日志 | `/llm/logs` | 按 workflow_run_id、agent_task_id、provider、model、status 筛选 |

---

## 7. 后端接口交付清单

| 方法 | 接口 | 输入 | 输出 | 原型页面映射 |
|---|---|---|---|---|
| GET | `/api/v1/workflow-templates` | page、page_size、sort、order、content_type、category、status | 模板分页列表 | 工作流模板管理 |
| POST | `/api/v1/workflow-templates` | code、name、content_type、category、description | workflow_template_id、status | 新增模板 |
| GET | `/api/v1/workflow-templates/:id` | id | 模板详情 | 工作流模板详情 |
| GET | `/api/v1/workflow-templates/:id/versions` | id、page、page_size | 模板版本列表 | 模板版本管理 |
| POST | `/api/v1/workflow-templates/:id/versions` | input_schema、output_schema、steps | template_version_id、step_count、status | 创建模板版本 |
| GET | `/api/v1/workflow-template-versions/:id` | id | 版本详情、steps | 模板版本详情 |
| POST | `/api/v1/workflow-template-versions/:id/publish` | note | previous_status、current_status、operation_log_id | 发布模板版本 |
| GET | `/api/v1/workflow-runs` | project_id、template_version_id、status、page、page_size | 运行分页列表 | 运行记录 |
| POST | `/api/v1/workflow-runs` | project_id、template_version_id、input | workflow_run_id、status | 手动运行 |
| GET | `/api/v1/workflow-runs/:id` | id | 运行详情、input、output、error | 运行记录详情 |
| GET | `/api/v1/workflow-runs/:id/steps` | id | StepRun 列表 | 运行步骤 |
| POST | `/api/v1/workflow-runs/:id/cancel` | reason、note | previous_status、current_status、operation_log_id | 取消运行 |
| POST | `/api/v1/workflow-runs/:id/retry` | reason、input_override | new_workflow_run_id、status | 失败重试 |
| GET | `/api/v1/agent-tasks` | workflow_run_id、step_run_id、agent_code、status、page、page_size | AgentTask 分页列表 | Agent 管理 / 运行记录 |
| GET | `/api/v1/agent-tasks/:id` | id | input、output、status、error、关联日志 | AgentTask 详情 |
| GET | `/api/v1/llm-call-logs` | workflow_run_id、agent_task_id、provider、model、status、page、page_size | LLM 调用日志列表 | LLM 日志 |
| GET | `/api/v1/llm-call-logs/:id` | id | LLM 调用日志详情 | LLM 日志详情 |
| POST | `/api/v1/workflow-schedules` | Contract only | schedule_id skeleton | 调度接口契约占位 |

### 7.1 Contract only 说明

`POST /api/v1/workflow-schedules` 在本迭代只用于提前稳定接口形态，不要求实现调度器、cron 解析、触发记录或 ProductionPlan 运行时。完整运行时进入 Iteration 2.1。

---

## 8. 页面-接口映射规则

| 页面动作 | 接口调用要求 |
|---|---|
| 模板列表加载 | 调用 `GET /api/v1/workflow-templates`，支持分页、筛选、排序 |
| 新建模板 | 调用 `POST /api/v1/workflow-templates`，字段与 DTO 对齐 |
| 创建模板版本 | 调用 `POST /api/v1/workflow-templates/:id/versions`，steps 必须包含 step_type 和 order_index |
| 发布模板版本 | 调用 `POST /api/v1/workflow-template-versions/:id/publish`，传 note，写 operation_log |
| 手动运行 | 调用 `POST /api/v1/workflow-runs`，立即返回 workflow_run_id |
| 查看运行详情 | 调用 `GET /api/v1/workflow-runs/:id`，不依赖列表缓存 |
| 查看运行步骤 | 调用 `GET /api/v1/workflow-runs/:id/steps` |
| 取消运行 | 调用 `POST /api/v1/workflow-runs/:id/cancel`，必须传 reason 或 note |
| 失败重试 | 调用 `POST /api/v1/workflow-runs/:id/retry`，创建新 run，不覆盖旧 run |
| 查看 AgentTask | 调用 `GET /api/v1/agent-tasks/:id` |
| 查看 LLM 日志 | 调用 `GET /api/v1/llm-call-logs` 或 `GET /api/v1/llm-call-logs/:id` |
| 成功反馈 | 页面显示 Toast，并刷新关联数据 |
| 失败反馈 | 页面显示错误码、错误信息和 request_id |

---

## 9. 原型页面映射

本迭代页面已映射到原型文件：

```text
prototype/ai-content-factory-clickable-prototype.html
```

对应页面：

- 工作流模板管理
- Agent 管理
- 运行记录
- AgentTask 详情
- LLM 调用日志

---

## 10. 需求评审与讨论

| 维度 | 评审结论 | 修订建议 |
|---|---|---|
| 完整性 | 初始需求覆盖 Workflow Engine、Agent Runtime、LLMCallLog 主线，但产品需求、状态流转、页面行为、接口详情偏泛 | 已补充模板版本、发布、手动运行、取消、重试、StepRun、AgentTask、LLM 日志明细 |
| 合理性 | Go 自研轻量 Workflow Engine 合理；本迭代不应实现复杂 DAG、调度器、Novel 业务 | 明确只支持线性步骤和简单 condition，复杂调度进入 Iteration 2.1 |
| 一致性 | 与 Iteration 0 的工程底座、Iteration 1 的项目 / Prompt / Provider 能力、Iteration 2.1 的调度基线不存在冲突 | 本迭代复用 Iteration 1 的 PromptTemplate / LLM Provider，不重复实现；调度运行时交给 2.1 |
| 风险点 | 若 AgentTask 可被直接调用，会绕过 WorkflowRun 追踪链路；若 WorkflowSchedule 在本迭代实现，会造成范围膨胀 | 要求 AgentTask 必须由 StepRun 派生；WorkflowSchedule 仅 Contract only |
| 验收风险 | 原验收项无法验证“可追踪”和“工作流自研”是否真正落地 | 增加状态枚举、日志关联、幂等、operation_log、OpenAPI、页面错误展示等验收项 |

---

## 11. 最终需求变更说明

| 变更项 | 变更内容 | 原因 |
|---|---|---|
| 评审结论 | 增加评审日期、结论摘要、是否更新蓝图 | 满足迭代评审流程 |
| 迭代目标 | 从“建立引擎”细化为模板、版本、运行、StepRun、AgentTask、LLMCallLog、前端页面闭环 | 原目标过泛，难以验收 |
| 产品需求 | 增加模板发布、手动运行、取消、失败重试、日志查询、错误反馈 | 补齐用户操作场景 |
| 技术需求 | 增加 Workflow Engine、Agent Runtime、LLMCallLog 的执行规则和边界 | 防止实现偏离蓝图 |
| 数据模型 | 增加核心字段和状态枚举 | 便于数据库迁移和 API DTO 实现 |
| 接口清单 | 增加详情、列表、取消、重试、日志详情接口 | 支撑前端页面闭环和运行可追踪 |
| 2.1 边界 | 明确 WorkflowSchedule / ProductionPlan 在本迭代只做 Contract only | 避免 Iteration 2 范围膨胀，与 Iteration 2.1 分工清晰 |
| 验收标准 | 增加可执行、可追踪、可查询、可幂等、可审计验收项 | 将需求转化为可验证交付 |

---

## 12. 最终对齐验证

| 检查项 | 最终需求 | 蓝图定义 | 状态 |
|---|---|---|---|
| Core 内容类型无关 | 使用 Workflow / Agent / LLM 通用模型，不引入 Book / Chapter 作为 Core 资源 | Core 不写死小说专属概念 | ✅ 无背离 |
| 工作流自研 | Go Workflow Engine 承载 WorkflowTemplate / WorkflowRun / StepRun | 核心生产链路由自研 Workflow Engine 承载 | ✅ 无背离 |
| Agent 可追踪 | AgentTask 必须由 StepRun 派生，LLMCallLog 必须关联 AgentTask | AgentTask、LLMCallLog 必须记录输入、输出、模型、Token、成本、错误 | ✅ 无背离 |
| n8n 外围化 | 不实现 n8n 编排核心 Agent，只保留后续外围集成边界 | n8n 只做通知、Webhook、同步、告警 | ✅ 无背离 |
| 人工节点保留 | 支持 human_review step_type，但不实现审稿业务 | 审稿、发布、策略执行保留人工确认 | ✅ 无背离 |
| 前后端一体验收 | 页面、接口、映射、验收标准均在本文件闭环 | 每个迭代必须包含前端页面、后端接口、页面-接口映射 | ✅ 无背离 |
| Iteration 2.1 分工 | 调度运行时放入 2.1，本迭代只做 Contract only | WorkflowSchedule + ProductionPlan 承载每日生产计划 | ✅ 无背离 |

结论：最终需求与蓝图无背离，不需要更新 `00-product-blueprint.md`。

---

## 13. 验收标准

- [ ] Go 后端接口均有 request / response DTO、validator 校验和 OpenAPI 描述。
- [ ] WorkflowTemplate 可以创建、查询、筛选。
- [ ] WorkflowTemplateVersion 可以创建、查询、发布；已发布版本不可直接修改。
- [ ] WorkflowStepTemplate 支持 `agent`、`human_review`、`condition`、`system_task` 四类基础节点。
- [ ] WorkflowRun 可以手动触发，HTTP 请求立即返回 `workflow_run_id` 和初始状态。
- [ ] WorkflowRun 可以查询详情、StepRun、AgentTask、LLMCallLog。
- [ ] WorkflowRun 可以取消；取消动作写入 `operation_log`。
- [ ] 失败 WorkflowRun 可以整体重试；重试创建新 run，并保留原 run。
- [ ] AgentTask 必须由 WorkflowStepRun 派生，不允许绕过 WorkflowRun 执行核心 Agent。
- [ ] AgentTask 记录 input、output、status、error、started_at、finished_at。
- [ ] LLMCallLog 记录 provider、model、Token、成本、耗时、错误和 request_id。
- [ ] 列表接口支持分页、筛选、排序。
- [ ] 创建 WorkflowRun、发布版本、取消、重试等接口支持幂等保护。
- [ ] 状态变更接口写入 `operation_log`。
- [ ] 页面可以完成模板管理、运行记录、AgentTask 详情、LLM 日志查询主要交互。
- [ ] 页面上的主要按钮均有可点击反馈。
- [ ] 页面-接口映射已与本文件一致。
- [ ] 失败反馈显示错误码、错误信息和 request_id。
- [ ] Core 层没有引入 Novel / Book / Chapter 作为核心资源命名。
- [ ] 本迭代完成后可以支撑 Iteration 3 的 Novel Pack 新书规划流程。

---

## 14. 本迭代明确不做

- 不做 Novel 新书规划、世界观、人物、大纲等业务能力。
- 不做章节正文生成。
- 不做审稿、发布、指标、策略业务。
- 不做复杂 DAG、并行网关、补偿事务、长事务编排。
- 不做 WorkflowSchedule / ProductionPlan 运行时；只保留 Contract only，占位进入 Iteration 2.1。
- 不做 n8n 核心编排。
- 不做未定义接口的隐式前端调用。
- 不做绕过 WorkflowRun / AgentTask / LLMCallLog 的核心生产链路。
