# Iteration 2.1：接口契约与调度基线补齐

> 文件定位：本文件是 Iteration 2.1 的最新需求、技术、前端页面、接口契约和原型映射文档。  
> 蓝图约束：必须遵守 `00-product-blueprint.md`。  
> 技术栈：后端 Go / Golang；前端 Next.js + TypeScript。  
> 更新日期：2026-05-17  
> 重要规则：本项目不再保留单独前端迭代文档；前端需求已整合到本迭代。  
> 原型开发规则：本迭代前端页面必须基于原型页面实现，CSS / JS 均需可用，并接入可点击导航入口。

---

## 1. 迭代目标

补齐 WorkflowSchedule / ProductionPlan 运行时，支持每天生成 5 个 ContentItem；明确 n8n 外围集成边界；追溯补齐 Iteration 0 / 1 / 2 的 Web Admin 页面渲染，要求按原型实现，不允许裸 HTML 或仅路由可访问。

---

## 2. 蓝图对齐说明

| 蓝图约束 | 本迭代对齐方式 |
|---|---|
| Core 内容类型无关 | Core 不使用 Book / Chapter 作为核心资源 |
| 内容类型插件化 | Novel / Article / Social Post 通过 Content Pack 扩展 |
| 工作流自研 | 核心生产链路通过 Go Workflow Engine 承载 |
| Agent 可追踪 | AgentTask / LLMCallLog 必须落库 |
| 人工节点保留 | 审稿、发布、策略执行保留人工确认 |
| n8n 外围化 | n8n 只做通知、Webhook、外部 API 同步、告警 |
| 前后端整合 | 页面、接口、交互、验收在本迭代内闭环 |

---

## 3. 产品需求

- 提供本迭代对应的系统级或项目级操作入口。
- 页面操作必须能映射到明确 API。
- 异步动作必须返回运行记录 ID 或任务 ID。
- 状态变更必须记录操作日志。
- 页面必须支持空状态、加载态、错误态、成功反馈。
- 本迭代需补齐 Iteration 0 / 1 / 2 已交付页面的真实前端渲染：统一 Layout、样式系统、导航入口、卡片、表格、表单、状态标签等必须按原型实现。

---

## 4. Go 后端技术需求

- workflow_schedule
- production_plan
- schedule_trigger_log
- external_workflow_provider
- external_workflow_binding
- external_workflow_call_log

通用要求：

- 使用 Go struct 定义请求 / 响应 DTO。
- 使用 validator 做入参校验。
- 使用 sqlc + pgx 或等价方式访问 PostgreSQL。
- 使用 goose 或等价工具管理数据库迁移。
- 所有接口进入 OpenAPI。
- 状态变更写入 `operation_log`。
- 异步任务通过 worker / queue 执行，不阻塞 HTTP 请求。


## 4.1 后端需求

| 类别 | 要求 |
|---|---|
| 数据模型 | 完成本迭代数据模型、迁移、Repository / Store、Service 边界，覆盖 `workflow_schedule`、`production_plan`、`schedule_trigger_log`、`external_workflow_provider`、`external_workflow_binding`、`external_workflow_call_log`。 |
| API 契约 | 第 6 节接口必须具备 Go request / response DTO、validator 校验、统一响应结构、统一错误结构和 OpenAPI 描述。 |
| 状态与审计 | 状态变更接口必须校验状态流转合法性，并写入 `operation_log`；失败时返回统一错误码和 `request_id`。 |
| 异步执行 | 触发型接口不得阻塞 HTTP 请求，必须返回 `run_id` / `job_id` / 业务记录 ID，并通过运行记录或详情接口查询结果。 |
| 幂等与重试 | 创建运行、触发执行、发布回填、确认类接口按 `api-contract-standard.md` 要求支持 `Idempotency-Key` 或明确说明不需要幂等。 |
| 前端联调支撑 | 后端需提供可用于页面空态、加载态、错误态、成功态联调的数据与错误响应；不得只返回无法渲染的占位数据。 |

---

## 5. 前端页面范围

| 页面 / 组件 | 路由建议 | 交互 |
|---|---|---|
| 生产计划 / 调度管理 | `/workflow/schedules` | 调度计划列表、新建、启用、停用、试跑 |
| 运行记录 | `/workflow/runs` | 查看由调度触发的 WorkflowRun |
| 外部自动化 / n8n | `/external-automation/n8n` | n8n Provider、Binding、回调配置 |
| 成本汇总 | `/llm/cost-summary` | 按项目、日期、模型查看 LLM 调用成本汇总 |


## 5.1 前端需求

| 类别 | 要求 |
|---|---|
| 页面范围 | 必须实现第 5 节定义的全部页面 / 组件：生产计划 / 调度管理、运行记录、外部自动化 / n8n、成本汇总。 |
| 导航入口 | 系统级页面必须接入首页或全局侧边栏；项目级页面必须接入项目详情 / 项目工作区菜单；不允许核心页面只能手输 URL 访问。 |
| 数据联调 | 页面列表、详情、表单提交、状态变更、异步触发必须绑定第 6 节对应 API；允许开发期 mock，但验收必须能切换真实 API。 |
| 交互完整性 | 筛选、分页、查看详情、新增 / 编辑、状态操作、异步触发、弹窗确认、Toast 提示均需可点击并有反馈。 |
| 状态展示 | 每个页面必须实现空态、加载态、错误态、成功态；失败态必须展示错误码、错误信息和 `request_id`。 |
| 路由可用 | 页面刷新后可直接访问，不出现 404；当前路由必须有导航高亮或当前位置提示。 |
| 测试要求 | Playwright / e2e 至少覆盖从导航进入页面、主要按钮点击、接口成功渲染、接口失败渲染。 |

## 5.2 页面渲染需求（CSS / JS / 原型）

本迭代所有前端页面必须基于 `prototype/ai-content-factory-clickable-prototype.html` 中对应原型开发。CSS 和 JS 必须同时可用：页面不仅要能返回 HTML，还必须具备真实管理台视觉样式、组件布局和可执行交互。

| 页面 / 组件 | 路由建议 | CSS 渲染要求 | JS 交互要求 |
|---|---|---|---|
| 生产计划 / 调度管理 | `/workflow/schedules` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 调度计划列表、新建、启用、停用、试跑；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 运行记录 | `/workflow/runs` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 查看由调度触发的 WorkflowRun；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 外部自动化 / n8n | `/external-automation/n8n` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | n8n Provider、Binding、回调配置；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 成本汇总 | `/llm/cost-summary` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 按项目、日期、模型查看 LLM 调用成本汇总；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |

### 5.2.1 统一渲染基线

- 必须加载全局样式和所选 UI 方案（Tailwind / Ant Design / shadcn/ui 等），不允许展示浏览器默认裸 HTML。
- 必须使用统一 AppLayout，包括顶部栏或侧边栏、主内容区、页面标题区、导航高亮。
- 页面结构、视觉层级、核心卡片、列表、表单、弹窗、状态标签应参照原型；如因 API 契约调整导致偏差，需在文档或代码注释中说明。
- JS 交互必须可用，包括导航跳转、筛选、分页、表单校验、弹窗开关、提交、重试、详情跳转、Toast / Alert 反馈。
- 失败态必须样式化展示统一错误结构中的 `error.code`、`error.message`、`request_id`。
- Playwright 截图或人工验收不得出现链接、按钮、标题、纯文本直接堆叠的裸页面。

### 5.2.2 页面渲染不通过判定

以下任一情况视为本迭代前端验收不通过：

- 页面呈现浏览器默认裸 HTML 样式。
- CSS 未加载，导航、按钮、表格、卡片、表单没有管理台样式。
- JS 不可用，按钮点击、筛选、表单提交、弹窗、路由跳转无反馈。
- 核心页面只能通过手输 URL 访问。
- 页面刷新后 404、状态丢失或导航高亮丢失。
- 错误态不展示 `request_id`。
- 页面主结构与原型明显不一致，且没有说明偏差原因。

---

## 6. 后端接口交付清单

| 方法 | 接口 | 输入 | 输出 | 原型页面映射 |
|---|---|---|---|---|
| GET | `/api/v1/workflow-schedules` | project_id、enabled、page | 调度计划列表 | 生产计划页面 |
| POST | `/api/v1/workflow-schedules` | project_id、cron_expression、production_plan | schedule_id、next_run_at | 新建生产计划 |
| POST | `/api/v1/workflow-schedules/:id/enable` | note | previous_enabled、current_enabled、next_run_at | 启用计划 |
| POST | `/api/v1/workflow-schedules/:id/disable` | reason、note | 状态变更、operation_log_id | 停用计划 |
| POST | `/api/v1/workflow-schedules/:id/test-run` | input_override | workflow_run_id、status | 试跑计划 |
| GET | `/api/v1/workflow-schedules/:id/triggers` | page | 调度触发记录 | 调度详情 |
| GET | `/api/v1/llm-call-logs/summary` | project_id、date_from、date_to | calls、tokens、cost、by_model | 成本汇总 |
| POST | `/api/v1/external-automation/providers` | provider_type、base_url、token | provider_id | n8n Provider 配置 |
| POST | `/api/v1/external-automation/bindings` | trigger_event、provider_id、webhook_url | binding_id | n8n 绑定 |

---

## 7. 页面-接口映射规则

| 页面动作 | 接口调用要求 |
|---|---|
| 列表加载 | 必须调用 GET 列表接口，支持分页、筛选、排序 |
| 新增 / 编辑 | 必须调用 POST / PATCH 接口，表单字段与 DTO 对齐 |
| 状态变更 | 必须调用动作接口，并传 reason / note |
| 异步触发 | 必须立即返回 run_id / job_id，并跳转或提示查看运行记录 |
| 查看详情 | 必须调用详情接口，不依赖列表缓存 |
| 成功反馈 | 页面显示 Toast，并刷新关联数据 |
| 失败反馈 | 页面显示错误码、错误信息和 request_id |

---

## 8. 原型页面映射

本迭代页面已映射到原型文件：

```text
prototype/ai-content-factory-clickable-prototype.html
```

对应页面：

- 生产计划 / 调度管理
- 运行记录
- 外部自动化 / n8n
- 成本汇总

---

## 8.1 基于原型的前端开发要求

本迭代前端实现必须以 `prototype/ai-content-factory-clickable-prototype.html` 中对应原型页面为开发基准，而不是只实现可访问的空路由或临时页面。

| 要求 | 说明 |
|---|---|
| 页面布局 | 页面结构、信息层级、核心卡片、列表、表单、弹窗需参考原型实现 |
| 关键交互 | 原型中的主要按钮、筛选、查看详情、状态变更、触发动作必须可点击 |
| 导航入口 | 系统级页面接入首页或全局侧边栏；项目级页面接入项目详情 / 项目工作区菜单 |
| 路由访问 | 不允许核心页面只能通过手输 URL 访问；刷新页面后仍可正常访问 |
| 状态反馈 | 必须实现空态、加载态、错误态、成功反馈，并展示统一错误码与 request_id |
| 接口绑定 | 页面数据、按钮动作、详情跳转必须绑定本迭代定义的 API，不使用不可追踪的 mock 替代真实联调 |
| 原型偏差处理 | 如原型字段与 API DTO 不一致，以本迭代接口契约为准，并同步更新原型映射说明 |

---

## 8.2 Iteration 0 / 1 / 2 页面渲染追溯补齐

> 本节是 Iteration 2.1 的补充开发范围，用于修复前置迭代已交付页面“仅输出裸 HTML、未按原型完成渲染”的问题。该补充不改变 Iteration 0 / 1 / 2 的业务边界，只补齐 Web Admin 前端渲染、样式、导航和原型一致性。

### 8.2.1 统一渲染基线

| 要求 | 说明 | 验收方式 |
|---|---|---|
| 全局样式生效 | Next.js 全局 CSS、Tailwind / Ant Design / shadcn/ui 等实际选型必须正确接入 | 页面不再显示浏览器默认裸 HTML |
| 管理台 Layout | 必须有统一 AppLayout，包括顶部栏或侧边栏、主内容区、页面标题区、导航高亮 | 首页与子页面截图一致 |
| 原型一致性 | 页面结构、信息层级、卡片、表格、表单、按钮、弹窗参考 `prototype/ai-content-factory-clickable-prototype.html` | Playwright 截图对比 / 人工验收 |
| 导航闭环 | 系统级页面从首页或全局侧边栏进入；项目级页面从项目详情 / 工作区进入 | 不允许核心页面只能手输 URL |
| 状态渲染 | 空态、加载态、错误态、成功反馈必须样式化展示 | e2e 覆盖 loading / empty / error |
| 错误展示 | 失败态必须展示统一错误码、错误信息、request_id | mock 500 或后端错误验证 |
| API 联调 | 页面展示数据必须来自对应 API；允许空数据，但不允许用不可追踪静态假数据冒充联调 | Network / e2e 验证 |

### 8.2.2 Iteration 0 页面渲染补齐

| 页面 | 路由建议 | 原型要求 | 接口 | 验收要求 |
|---|---|---|---|---|
| 系统默认页 / 健康检查页 | `/` 或 `/system/health` | 管理台壳层、服务状态卡片、数据库状态卡片、构建信息卡片 | `GET /api/v1/health`、`GET /api/v1/system/db-check`、`GET /api/v1/system/info` | 不允许裸 HTML；健康状态用状态标签展示 |
| Swagger / OpenAPI 入口页 | `/swagger-openapi` | OpenAPI 入口卡片、文档跳转按钮、接口分组说明 | OpenAPI 静态文档或 `/swagger` 入口 | 可从导航进入，刷新不 404 |
| 系统配置检查页 | `/system/config-check` | 配置项检查列表、缺失项错误提示、重试按钮 | `GET /api/v1/system/config-check`、`GET /api/v1/system/migration-status` | 错误态展示 request_id |

### 8.2.3 Iteration 1 页面渲染补齐

| 页面 | 路由建议 | 原型要求 | 接口 | 验收要求 |
|---|---|---|---|---|
| 首页 / 系统大盘 | `/` | 统计卡片、进行中项目、待处理事项、快捷入口 | `GET /api/v1/dashboard/summary` | 不允许只有文字堆叠；统计项必须卡片化 |
| 项目管理 | `/projects` | 项目列表、筛选、分页、新建项目入口 | `GET /api/v1/projects`、`POST /api/v1/projects` | 列表、空态、错误态均有样式 |
| 项目详情壳层 | `/projects/:id` | 项目概览、工作区导航、后续项目级页面入口 | `GET /api/v1/projects/:id/overview` | 项目工作区菜单可承载 Iteration 3+ 页面 |
| 项目模板管理 | `/content-types` | ContentType 列表、Schema 查看、新增模板 | `GET /api/v1/content-types`、`POST /api/v1/content-types` | 表格、表单、弹窗按原型渲染 |
| Prompt 模板管理 | `/prompt-templates` | Prompt 列表、Agent 过滤、新增 Prompt | `GET /api/v1/prompt-templates`、`POST /api/v1/prompt-templates` | 过滤和新增按钮可点击 |
| 模型 Provider 管理 | `/llm/providers` | Provider 列表、新增 Provider、API Key 脱敏展示 | `GET /api/v1/llm-providers`、`POST /api/v1/llm-providers` | 密钥不得明文展示 |

### 8.2.4 Iteration 2 页面渲染补齐

| 页面 | 路由建议 | 原型要求 | 接口 | 验收要求 |
|---|---|---|---|---|
| 工作流模板管理 | `/workflow/templates` | 模板列表、版本管理、发布状态、创建模板入口 | `GET /api/v1/workflow-templates`、`POST /api/v1/workflow-templates`、`POST /api/v1/workflow-templates/:id/versions`、`POST /api/v1/workflow-template-versions/:id/publish` | 可从导航进入；发布状态有标签 |
| 运行记录 | `/workflow/runs` | WorkflowRun 列表、状态筛选、详情入口、步骤概览 | `POST /api/v1/workflow-runs`、`GET /api/v1/workflow-runs/:id`、`GET /api/v1/workflow-runs/:id/steps` | 运行状态有视觉区分 |
| AgentTask 管理 | `/agent/tasks` | AgentTask 列表、状态、所属 WorkflowRun、详情入口 | `GET /api/v1/agent-tasks`、`GET /api/v1/agent-tasks/:id` | 输入输出区域格式化展示 |
| AgentTask 详情 | `/agent/tasks/:id` | 输入、输出、错误、关联 LLMCallLog | `GET /api/v1/agent-tasks/:id` | 错误与结构化输出校验结果可见 |
| LLM 调用日志 | `/llm/logs` | 调用日志列表、模型、Token、成本、耗时、错误 | `GET /api/v1/llm-call-logs` | 成本 / Token 字段表格化展示 |

### 8.2.5 页面渲染缺陷判定标准

以下情况均视为 Iteration 2.1 验收不通过：

- 页面呈现浏览器默认裸 HTML 样式。
- 导航链接直接堆叠，无统一导航栏或侧边栏。
- 页面只有 `<a>`、`button`、`h1`、纯文本，没有卡片、表格、表单等管理台组件。
- 核心页面只能通过手输 URL 访问。
- 页面刷新后 404 或当前导航状态丢失。
- 错误态不展示 request_id。
- 页面与原型主结构明显不一致，且未在文档中说明偏差原因。

## 9. 验收标准

- [ ] Go 后端接口均有 DTO、校验和 OpenAPI 描述。
- [ ] 前端页面可以完成本迭代定义的主要交互。
- [ ] 页面上的主要按钮均有可点击反馈。
- [ ] 页面-接口映射已与本文件一致。
- [ ] 列表接口支持分页、筛选、排序。
- [ ] 状态变更接口写入 operation_log。
- [ ] 异步接口返回 run_id / job_id。
- [ ] Core 层没有引入 Novel / Book / Chapter 作为核心资源命名。
- [ ] 本迭代新增页面已按原型页面完成主要布局和交互，不是空白页、占位页或仅可访问路由。
- [ ] 本迭代新增页面已接入现有导航体系：系统级页面可从首页 / 全局侧边栏进入，项目级页面可从项目详情 / 项目工作区进入。
- [ ] 不允许本迭代核心页面只能通过手输 URL 访问。
- [ ] 新增页面刷新后可直接访问，并保持当前导航高亮或当前位置提示。
- [ ] e2e 或集成测试覆盖从导航入口进入页面、触发主要按钮、展示接口返回结果和错误态。
- [ ] Iteration 0 / 1 / 2 页面已完成基于原型的真实渲染，不再出现浏览器默认裸 HTML。
- [ ] Web Admin 已具备统一 AppLayout、全局样式、导航入口、当前路由高亮和主内容区布局。
- [ ] Iteration 0 的健康检查、OpenAPI 入口、配置检查页面已按原型渲染并可从导航访问。
- [ ] Iteration 1 的首页 / 系统大盘、项目管理、项目详情壳层、项目模板、Prompt 模板、Provider 管理页面已按原型渲染。
- [ ] Iteration 2 的 Workflow Templates、Workflow Runs、Agent Tasks、AgentTask 详情、LLM Logs 页面已按原型渲染。
- [ ] 0 / 1 / 2 所有核心页面刷新后可直接访问，不出现 404。
- [ ] 0 / 1 / 2 所有核心页面具备空态、加载态、错误态和成功反馈。
- [ ] Playwright e2e 覆盖从首页或全局导航进入 Iteration 0 / 1 / 2 核心页面。
- [ ] 前端页面已按原型完成真实渲染，不出现浏览器默认裸 HTML。
- [ ] CSS 已生效：统一 AppLayout、导航、卡片、表格、表单、按钮、状态标签可见。
- [ ] JS 已生效：导航、筛选、分页、表单提交、弹窗、详情跳转、状态操作可点击并有结果反馈。
- [ ] 本迭代核心页面可从导航入口进入，刷新后可直接访问且不出现 404。
- [ ] 页面失败态展示统一错误码、错误信息和 request_id。
- [ ] e2e 覆盖页面渲染、导航进入、主要交互和接口失败态。
- [ ] 本迭代完成后可以支撑下一迭代。

---

## 10. 本迭代明确不做

- 不做超出本迭代页面范围的业务功能。
- 不做未定义接口的隐式前端调用。
- 不做绕过 WorkflowRun / AgentTask / LLMCallLog 的核心生产链路。
- 不做 n8n 核心编排。
- 不借页面渲染补齐扩大 Iteration 0 / 1 / 2 的业务功能范围；只补样式、Layout、导航、原型一致性和已定义接口联调。
