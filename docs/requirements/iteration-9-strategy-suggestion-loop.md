# Iteration 9：策略建议与单类型业务闭环

> 文件定位：本文件是 Iteration 9 的最新需求、技术、前端页面、接口契约和原型映射文档。  
> 蓝图约束：必须遵守 `00-product-blueprint.md`。  
> 技术栈：后端 Go / Golang；前端 Next.js + TypeScript。  
> 更新日期：2026-05-27  
> 评审日期：2026-05-27  
> 评审结论摘要：评审通过。Iteration 9 与产品蓝图无背离，可进入开发；需补齐策略建议生成依据、状态机、人工确认边界、执行日志、幂等与前后迭代依赖。  
> 是否需要更新蓝图：否。  
> 重要规则：本项目不再保留单独前端迭代文档；前端需求已整合到本迭代。  
> 原型开发规则：本迭代前端页面必须基于原型页面实现，CSS / JS 均需可用，并接入可点击导航入口。

---

## 1. 迭代目标

基于指标生成策略建议；前端整合到项目工作区的策略建议页。

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

### 2.1 初步对齐检查

| 检查项 | 初始需求描述 | 蓝图定义 | 状态 |
|---|---|---|---|
| Metrics & Strategy 边界 | 基于项目指标生成 `strategy_suggestion`，并提供确认、忽略、执行链路 | 蓝图定义 Metrics & Strategy 覆盖 MetricRecord、MetricTemplate、StrategySuggestion | ✅ 对齐 |
| Core 内容类型无关 | 使用 `strategy_rule`、`strategy_suggestion`、`strategy_execution_log`，不引入 Novel / Book / Chapter 作为核心资源 | Core 层不得写死 Novel / Book / Chapter 等小说专属概念 | ✅ 对齐 |
| 指标依赖 | 初始需求依赖 Iteration 8 的 MetricRecord / Metrics Dashboard，但未明确指标快照与缺失数据处理 | 策略建议属于数据反馈和策略优化系统，指标应可追溯、可解释 | ⚠️ 偏差：已补齐 `metrics_snapshot`、缺失指标处理与建议生成依据 |
| 人工节点保留 | 初始需求提供确认 / 忽略 / 执行操作 | 蓝图要求审稿、发布、策略执行必须支持人工确认 | ✅ 对齐 |
| 工作流与 Agent 边界 | 初始需求未明确策略执行是否可直接改生产配置或触发生成 | 蓝图要求核心生产链路由 Workflow Engine / Agent Runtime 承载，AgentTask / LLMCallLog 可追踪 | ⚠️ 偏差：已补齐执行边界，不允许绕过 WorkflowRun / AgentTask / LLMCallLog |
| n8n 边界 | 初始需求未说明策略通知或执行是否依赖 n8n | 蓝图规定 n8n 只做通知、Webhook、外部 API 同步、告警，不保存核心状态 | ✅ 对齐：本迭代不使用 n8n 承载策略状态机 |
| API 契约 | 初始接口具备列表、详情、确认、忽略、执行，但缺少状态机、幂等、详情字段与执行日志查询 | API 契约要求统一响应、分页、状态日志、幂等敏感接口 | ⚠️ 偏差：已补齐状态机、幂等、审计、执行日志和列表排序要求 |
| 前端原型 | 策略建议页、详情页、操作页已映射原型 | 每个迭代必须包含前端页面、接口、页面-接口映射与验收 | ✅ 对齐 |

---

## 3. 产品需求

- 提供本迭代对应的系统级或项目级操作入口。
- 页面操作必须能映射到明确 API。
- 异步动作必须返回运行记录 ID 或任务 ID。
- 状态变更必须记录操作日志。
- 页面必须支持空状态、加载态、错误态、成功反馈。

### 3.1 评审后补充产品需求

- 策略建议必须基于可追溯的项目指标生成，至少关联 `project_id`、`date_range`、触发规则、指标快照、生成时间和生成方式。
- 策略建议不得只输出一句自然语言结论；详情页必须展示建议类型、触发原因、证据指标、影响范围、风险说明、建议动作和预期收益。
- 建议类型必须收敛为可枚举值，首期至少支持 `keep`、`optimize`、`suspend`、`promote`、`cost_control`。
- 建议状态必须有明确人工流转：新生成建议进入 `pending`，人工确认后进入 `confirmed`，忽略后进入 `ignored`，执行后进入 `executed`，执行失败进入 `execution_failed`。
- `confirm` 只表示人工认可建议，不得自动修改项目配置、调度计划、Prompt 或模型配置。
- `execute` 必须记录实际执行动作；若动作会影响工作流、调度、发布或内容生产，只能通过既有领域接口或后续明确接口完成，不得在策略模块内直接改写其他模块状态。
- 执行建议必须可追踪：写入 `strategy_execution_log` 和 `operation_log`，记录操作者、执行动作、目标资源、前后状态、结果和失败原因。
- 建议生成时如存在指标缺失、样本过小或日期范围不足，必须降级为低置信度建议或返回可解释的失败原因，不允许生成无法解释的强结论。
- 本迭代只完成单项目、单内容类型的策略闭环；多项目组合优先级和跨项目策略由 Iteration 10 Portfolio 承接。

---

## 4. Go 后端技术需求

- strategy_rule
- strategy_suggestion
- strategy_execution_log

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
| 数据模型 | 完成本迭代数据模型、迁移、Repository / Store、Service 边界，覆盖 `strategy_rule`、`strategy_suggestion`、`strategy_execution_log`。 |
| API 契约 | 第 6 节接口必须具备 Go request / response DTO、validator 校验、统一响应结构、统一错误结构和 OpenAPI 描述。 |
| 状态与审计 | 状态变更接口必须校验状态流转合法性，并写入 `operation_log`；失败时返回统一错误码和 `request_id`。 |
| 异步执行 | 触发型接口不得阻塞 HTTP 请求，必须返回 `run_id` / `job_id` / 业务记录 ID，并通过运行记录或详情接口查询结果。 |
| 幂等与重试 | 创建运行、触发执行、发布回填、确认类接口按 `api-contract-standard.md` 要求支持 `Idempotency-Key` 或明确说明不需要幂等。 |
| 前端联调支撑 | 后端需提供可用于页面空态、加载态、错误态、成功态联调的数据与错误响应；不得只返回无法渲染的占位数据。 |

## 4.2 数据模型细化

| 模型 | 必要字段 / 约束 | 说明 |
|---|---|---|
| `strategy_rule` | `id`、`project_id`、`content_type`、`rule_code`、`name`、`description`、`trigger_condition`、`suggestion_type`、`severity`、`enabled`、`created_at`、`updated_at` | 策略规则配置；首期可以内置规则，也要落库或以可查询方式暴露，便于解释建议来源。 |
| `strategy_suggestion` | `id`、`project_id`、`rule_id`、`suggestion_type`、`status`、`title`、`reason`、`evidence`、`impact_scope`、`risk_level`、`confidence`、`metrics_snapshot`、`date_range`、`source_run_id`、`confirmed_at`、`ignored_at`、`executed_at`、`created_at`、`updated_at` | 策略建议主表；`evidence`、`impact_scope`、`metrics_snapshot` 使用结构化 JSON，必须能支撑详情页解释。 |
| `strategy_execution_log` | `id`、`suggestion_id`、`project_id`、`action_type`、`target_type`、`target_id`、`operator_id`、`operator_note`、`before_snapshot`、`after_snapshot`、`status`、`error_message`、`created_at` | 策略执行流水；每次执行或执行失败都必须追加记录，不覆盖历史。 |

## 4.3 策略建议生成规则

- 生成入口必须接收 `date_range`，并允许可选 `rule_codes`、`metric_codes`、`force_regenerate`。
- 生成过程必须读取 Iteration 8 的 `metric_record` / 指标汇总能力，不直接从前端传入指标结论。
- 每条建议必须保留 `metrics_snapshot`，包含触发指标、当前值、对比基线、阈值、样本量和缺失数据说明。
- 当同一项目、同一规则、同一日期范围已存在未处理建议时，默认不得重复生成；如允许强制生成，必须通过幂等键和版本字段避免重复处理。
- 建议置信度 `confidence` 建议使用 `low`、`medium`、`high`，并在详情页展示置信度来源。
- 规则首期可以是确定性规则，不强制引入 LLM；如使用 LLM 解释或生成文案，仍必须记录 AgentTask / LLMCallLog。

## 4.4 策略建议状态机

| 当前状态 | 允许动作 | 下一状态 | 说明 |
|---|---|---|---|
| `pending` | 确认 | `confirmed` | 人工认可建议，必须提交 `note` 或允许空备注。 |
| `pending` | 忽略 | `ignored` | 必须提交 `reason`，用于后续规则优化。 |
| `confirmed` | 执行 | `executed` / `execution_failed` | 执行动作成功进入 `executed`；失败进入 `execution_failed` 并记录错误。 |
| `execution_failed` | 重新执行 | `executed` / `execution_failed` | 允许重试，但必须追加新的 `strategy_execution_log`。 |
| `ignored` | 确认 / 执行 | 不允许 | 忽略为终态；如需恢复，后续迭代单独定义 reopen。 |
| `executed` | 确认 / 忽略 / 重复执行 | 不允许 | 执行完成为终态，避免重复改动项目配置或计划。 |

## 4.5 执行边界与审计规则

- `execute` 首期只允许执行可审计的轻量动作，例如记录人工执行结果、生成后续待办、标记建议已执行。
- 如建议动作涉及暂停项目、调整生产计划、修改 Prompt、切换模型、提升发布频率等跨模块变更，必须调用对应模块已定义接口，并记录目标接口、目标资源和结果。
- 策略模块不得直接绕过 ContentProject、WorkflowSchedule、PromptTemplate、LLM Provider 等模块的服务边界写库。
- `confirm`、`ignore`、`execute`、`retry-execute` 必须支持 `Idempotency-Key`，重复请求返回同一状态变更结果。
- 所有状态变更必须写入 `operation_log`；所有执行尝试必须写入 `strategy_execution_log`。
- 非法状态流转返回 `CONFLICT`；指标不足或建议依据缺失返回 `VALIDATION_ERROR` 或业务错误，并带 `request_id`。

## 4.6 与前后迭代依赖

| 依赖来源 | 本迭代使用方式 |
|---|---|
| Iteration 1：ContentProject / ContentType | 策略建议以 `project_id` 为边界，首期只处理单项目、单内容类型。 |
| Iteration 2：WorkflowRun / AgentTask / LLMCallLog | 策略生成如触发 Agent 或 LLM 解释，必须记录 AgentTask / LLMCallLog；异步生成返回运行记录或业务运行 ID。 |
| Iteration 4：ContentItem / generation_run | 内容生产表现可作为影响范围和建议证据，但本迭代不直接重试生成。 |
| Iteration 5：content_review / review_report | 审稿质量、打回率、质检问题可作为策略建议证据来源。 |
| Iteration 7：PublishJob | 发布完成与回填结果可作为指标关联基础；策略模块不负责发布状态机。 |
| Iteration 8：MetricRecord / MetricTemplate | 本迭代的主要输入；缺失指标必须显式展示并影响建议置信度。 |
| Iteration 10：ProjectPortfolio | 多项目优先级、组合健康、跨项目资源分配不在本迭代实现，由 Portfolio 承接。 |

---

## 5. 前端页面范围

| 页面 / 组件 | 路由建议 | 交互 |
|---|---|---|
| 项目工作区 / 策略建议 | `/projects/:projectId/strategy-suggestions` | 生成建议、查看建议列表、筛选状态 |
| 建议详情 | `/strategy-suggestions/:suggestionId` | 查看建议原因、影响范围和上下文 |
| 确认 / 忽略 / 执行 | `/strategy-suggestions/:suggestionId/actions` | 确认、忽略或执行策略建议 |


## 5.1 前端需求

| 类别 | 要求 |
|---|---|
| 页面范围 | 必须实现第 5 节定义的全部页面 / 组件：项目工作区 / 策略建议、建议详情、确认 / 忽略 / 执行。 |
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
| 项目工作区 / 策略建议 | `/projects/:projectId/strategy-suggestions` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 生成建议、查看建议列表、筛选状态；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 建议详情 | `/strategy-suggestions/:suggestionId` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 查看建议原因、影响范围和上下文；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 确认 / 忽略 / 执行 | `/strategy-suggestions/:suggestionId/actions` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 确认、忽略或执行策略建议；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |

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
| POST | `/api/v1/projects/:projectId/strategy-suggestions/generate` | date_range、rule_codes、metric_codes、force_regenerate | suggestion_run_id、status | 生成建议 |
| GET | `/api/v1/projects/:projectId/strategy-suggestions` | status、suggestion_type、risk_level、confidence、date_from、date_to、page、page_size、sort、order | 建议分页列表 | 策略页 |
| GET | `/api/v1/strategy-suggestions/:id` | id | 建议原因、影响范围、证据指标、metrics_snapshot、状态、执行记录摘要 | 建议详情 |
| POST | `/api/v1/strategy-suggestions/:id/confirm` | note | suggestion_id、previous_status、current_status、operation_log_id | 确认建议 |
| POST | `/api/v1/strategy-suggestions/:id/ignore` | reason、note | suggestion_id、previous_status、current_status、operation_log_id | 忽略建议 |
| POST | `/api/v1/strategy-suggestions/:id/execute` | action_type、target_type、target_id、operator_note | execution_log_id、suggestion_id、previous_status、current_status | 执行建议 |
| POST | `/api/v1/strategy-suggestions/:id/retry-execute` | operator_note | execution_log_id、current_status | 重新执行失败建议 |
| GET | `/api/v1/strategy-suggestions/:id/execution-logs` | page、page_size | 执行日志列表 | 建议详情 / 执行记录 |

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

- 项目工作区 / 策略建议
- 建议详情
- 确认 / 忽略 / 执行

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
- [ ] 前端页面已按原型完成真实渲染，不出现浏览器默认裸 HTML。
- [ ] CSS 已生效：统一 AppLayout、导航、卡片、表格、表单、按钮、状态标签可见。
- [ ] JS 已生效：导航、筛选、分页、表单提交、弹窗、详情跳转、状态操作可点击并有结果反馈。
- [ ] 本迭代核心页面可从导航入口进入，刷新后可直接访问且不出现 404。
- [ ] 页面失败态展示统一错误码、错误信息和 request_id。
- [ ] e2e 覆盖页面渲染、导航进入、主要交互和接口失败态。
- [ ] 策略建议生成结果包含可解释的 `metrics_snapshot`、触发规则、影响范围、风险等级和置信度。
- [ ] `pending → confirmed / ignored → executed / execution_failed` 状态机校验完整，非法流转返回 `CONFLICT`。
- [ ] `confirm`、`ignore`、`execute`、`retry-execute` 支持幂等键并写入 `operation_log`。
- [ ] 每次执行或执行失败均追加 `strategy_execution_log`，详情页可查看执行历史。
- [ ] 缺失指标、样本不足、规则禁用等场景有明确错误或低置信度展示，不生成无法解释的强结论。
- [ ] 策略执行不得绕过其他领域服务直接改写跨模块状态。
- [ ] 本迭代完成后可以支撑下一迭代。

---

## 10. 本迭代明确不做

- 不做超出本迭代页面范围的业务功能。
- 不做未定义接口的隐式前端调用。
- 不做绕过 WorkflowRun / AgentTask / LLMCallLog 的核心生产链路。
- 不做 n8n 核心编排。
- 不做多项目 Portfolio 级策略排序和资源分配。
- 不做自动改写 Prompt、模型配置、调度计划或项目状态的隐式执行。
- 不做没有指标依据、没有证据快照的黑盒策略建议。
