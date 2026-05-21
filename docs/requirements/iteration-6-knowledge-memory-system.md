# Iteration 6：Knowledge Memory 记忆系统

> 文件定位：本文件是 Iteration 6 的最新需求、技术、前端页面、接口契约和原型映射文档。  
> 蓝图约束：必须遵守 `00-product-blueprint.md`。  
> 技术栈：后端 Go / Golang；前端 Next.js + TypeScript。  
> 更新日期：2026-05-19  
> 评审日期：2026-05-19  
> 评审结论摘要：评审通过。Iteration 6 与产品蓝图无背离，可进入开发；需补齐记忆模型边界、上下文装配快照、一致性报告详情、人工修正入口、幂等与审计要求。  
> 是否需要更新蓝图：否。  
> 重要规则：本项目不再保留单独前端迭代文档；前端需求已整合到本迭代。  
> 原型开发规则：本迭代前端页面必须基于原型页面实现，CSS / JS 均需可用，并接入可点击导航入口。

---

## 1. 迭代目标

建立通用知识与记忆系统；前端整合到项目工作区的记忆与一致性页。

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
| Core 内容类型无关 | 以 `knowledge_memory`、`memory_snapshot`、`consistency_report` 承载记忆能力 | Core 层不得写死 Novel / Book / Chapter 等小说专属概念 | ✅ 对齐 |
| Knowledge Memory 职责 | 初始需求覆盖 StaticContext、DynamicState、上下文预览 | 蓝图定义 Knowledge Memory 负责 StaticContext、DynamicState、RecentContentWindow、StyleGuide | ⚠️ 偏差：初始需求未明确 RecentContentWindow 与 StyleGuide，已补齐 |
| 与 Workflow / Agent 边界 | 上下文装配供 AgentRuntime 使用，不直接替代 WorkflowRun / AgentTask | 核心生产链路必须经 Workflow Engine 与 Agent Runtime，并记录 AgentTask / LLMCallLog | ✅ 对齐 |
| 与前置迭代依赖 | 依赖项目、内容单元、规划资产、审稿版本等上游数据 | ContentProject、ContentItem、ContentAsset、content_version 已在 Iteration 1 / 3 / 4 / 5 建立 | ✅ 对齐 |
| 一致性检查 | 初始需求仅有报告列表与触发接口 | 项目工作区需要支持记忆与一致性页，报告应可追踪、可查看问题项 | ⚠️ 偏差：缺少报告详情与问题项结构，已补齐 |
| API 契约 | 初始接口基本符合 `/api/v1`，但未明确幂等与详情接口 | API 契约要求 DTO、统一错误、分页、状态日志、幂等敏感接口 | ⚠️ 偏差：已补充 mutating 接口幂等、报告详情、快照列表 |
| 前端原型 | 已要求项目级页面接入项目工作区，CSS / JS 可用 | 每个迭代必须包含前端页面、接口、映射和验收标准 | ✅ 对齐 |

---

## 3. 产品需求

- 提供本迭代对应的系统级或项目级操作入口。
- 页面操作必须能映射到明确 API。
- 异步动作必须返回运行记录 ID 或任务 ID。
- 状态变更必须记录操作日志。
- 页面必须支持空状态、加载态、错误态、成功反馈。

### 3.1 评审后补充产品需求

- 记忆系统必须同时覆盖 `StaticContext`、`DynamicState`、`RecentContentWindow`、`StyleGuide` 四类上下文，不得只展示静态 / 动态两类字段。
- `StaticContext` 用于项目长期稳定资料，例如世界观、人物、设定、禁用规则、风格规则摘要；允许人工修正，修正必须写入 `operation_log`。
- `DynamicState` 用于最近内容推进状态，例如当前剧情进展、人物关系变化、未解决伏笔、最新状态变更；必须由 ContentItem / content_version / review_report 等来源驱动。
- `RecentContentWindow` 用于上下文装配时拉取最近 N 个 ContentItem / content_version 摘要，必须受 Token 预算限制。
- `StyleGuide` 用于保存内容风格、语气、叙事约束、禁用表达等通用风格信息，供后续生成和审稿复用。
- 上下文装配必须生成可追踪 `memory_snapshot`，记录来源、Token 预算、截断策略、装配结果摘要和触发来源。
- 一致性检查必须输出结构化问题项，至少包含问题类型、严重级别、关联内容、证据摘要、建议处理方式。
- 本迭代不负责真实修复内容正文；只负责发现一致性问题、记录报告、支持人工查看和后续迭代引用。

---

## 4. Go 后端技术需求

- knowledge_memory
- memory_snapshot
- consistency_report

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
| 数据模型 | 完成本迭代数据模型、迁移、Repository / Store、Service 边界，覆盖 `knowledge_memory`、`memory_snapshot`、`consistency_report`。 |
| API 契约 | 第 6 节接口必须具备 Go request / response DTO、validator 校验、统一响应结构、统一错误结构和 OpenAPI 描述。 |
| 状态与审计 | 状态变更接口必须校验状态流转合法性，并写入 `operation_log`；失败时返回统一错误码和 `request_id`。 |
| 异步执行 | 触发型接口不得阻塞 HTTP 请求，必须返回 `run_id` / `job_id` / 业务记录 ID，并通过运行记录或详情接口查询结果。 |
| 幂等与重试 | 创建运行、触发执行、发布回填、确认类接口按 `api-contract-standard.md` 要求支持 `Idempotency-Key` 或明确说明不需要幂等。 |
| 前端联调支撑 | 后端需提供可用于页面空态、加载态、错误态、成功态联调的数据与错误响应；不得只返回无法渲染的占位数据。 |

## 4.2 数据模型细化

| 模型 | 必要字段 / 约束 | 说明 |
|---|---|---|
| `knowledge_memory` | `id`、`project_id`、`static_context`、`dynamic_state`、`recent_window_policy`、`style_guide`、`version`、`updated_at` | 项目级当前记忆主表；不得保存完整正文，只保存摘要、规则和结构化状态。 |
| `memory_snapshot` | `id`、`project_id`、`content_item_id`、`source_type`、`source_id`、`assembled_context`、`source_refs`、`token_budget`、`estimated_tokens`、`truncation_policy`、`created_at` | 每次上下文装配或动态状态更新的可追踪快照。 |
| `consistency_report` | `id`、`project_id`、`range`、`status`、`issue_count`、`severity_summary`、`issues`、`source_snapshot_id`、`created_at`、`completed_at` | 一致性检查报告；`issues` 必须为结构化 JSON。 |

## 4.3 上下文装配规则

- 装配顺序固定为：`StaticContext` → `StyleGuide` → `DynamicState` → `RecentContentWindow` → 当前任务输入。
- 装配必须接受 `purpose` 与 `budget`，根据用途决定上下文裁剪策略。
- 当上下文超过预算时，优先保留禁用规则、当前状态、最近内容摘要，再裁剪历史摘要。
- 装配结果必须可复现：相同 project、purpose、content_item_id、budget、源版本在同一幂等键下不得生成冲突快照。
- 装配不得绕过 AgentRuntime；它只提供上下文输入，不直接完成内容生成。

## 4.4 一致性检查规则

| 检查类型 | 最低要求 |
|---|---|
| 设定一致性 | 检查世界观、人物设定、禁用规则与内容摘要的冲突。 |
| 状态连续性 | 检查 DynamicState 与最近 ContentItem / content_version 摘要是否断裂。 |
| 风格一致性 | 检查正文摘要与 StyleGuide 是否明显冲突。 |
| 人物关系 | 检查人物状态、关系、身份、立场是否前后矛盾。 |
| 未解决伏笔 | 识别 DynamicState 中标记为 open 的伏笔是否被错误关闭或遗忘。 |

## 4.5 与前置迭代依赖

| 依赖来源 | 本迭代使用方式 |
|---|---|
| Iteration 1：ContentProject / ContentType | 记忆以 `project_id` 作为边界，不跨项目污染上下文。 |
| Iteration 2：WorkflowRun / AgentTask / LLMCallLog | 上下文装配供 AgentRuntime 使用，相关调用仍需落 AgentTask / LLMCallLog。 |
| Iteration 3：ContentAsset / Novel 扩展设定 | 世界观、人物、大纲等可进入 StaticContext，但 Novel 专属结构不得进入 Core 字段命名。 |
| Iteration 4：ContentItem / generation_run | RecentContentWindow 和 DynamicState 更新以内容单元产物为主要来源。 |
| Iteration 5：content_version / review_report | 审稿通过版本可作为可信来源；AI 质检报告可作为一致性检查输入。 |

---

## 5. 前端页面范围

| 页面 / 组件 | 路由建议 | 交互 |
|---|---|---|
| 项目工作区 / 记忆上下文 | `/projects/:projectId/memory` | 查看并人工修正 StaticContext、DynamicState、StyleGuide，查看记忆快照 |
| 一致性报告 | `/projects/:projectId/consistency-reports` | 查看一致性报告列表、状态、严重级别汇总和问题项 |
| 上下文预览 | `/projects/:projectId/memory/context-preview` | 预览上下文来源、Token 预算、截断策略和装配结果 |


## 5.1 前端需求

| 类别 | 要求 |
|---|---|
| 页面范围 | 必须实现第 5 节定义的全部页面 / 组件：项目工作区 / 记忆上下文、一致性报告、上下文预览。 |
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
| 项目工作区 / 记忆上下文 | `/projects/:projectId/memory` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 查看并人工修正 StaticContext、DynamicState、StyleGuide，查看记忆快照；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 一致性报告 | `/projects/:projectId/consistency-reports` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 查看一致性报告列表、状态、严重级别汇总和问题项；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 上下文预览 | `/projects/:projectId/memory/context-preview` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 预览上下文来源、Token 预算、截断策略和装配结果；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |

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
| GET | `/api/v1/projects/:projectId/knowledge-memory` | projectId | StaticContext、DynamicState、RecentContentWindow、StyleGuide、version | 记忆页 |
| PATCH | `/api/v1/projects/:projectId/knowledge-memory/static-context` | static_context、note | version、operation_log_id | 人工修正静态记忆 |
| PATCH | `/api/v1/projects/:projectId/knowledge-memory/style-guide` | style_guide、note | version、operation_log_id | 人工修正风格规则 |
| GET | `/api/v1/projects/:projectId/knowledge-memory/snapshots` | content_item_id、page | snapshot 列表、来源、Token、创建时间 | 记忆快照列表 |
| GET | `/api/v1/projects/:projectId/knowledge-memory/context-preview` | content_item_id、purpose、budget | 上下文来源、Token 预算、截断策略、预览文本 | 上下文预览 |
| POST | `/api/v1/projects/:projectId/knowledge-memory/assemble-context` | purpose、budget、content_item_id | context_snapshot_id、estimated_tokens、truncation_policy | 装配上下文 |
| POST | `/api/v1/content-items/:id/update-dynamic-state` | summary、changes、source_version_id | memory_snapshot_id、dynamic_state_version | 更新动态状态 |
| POST | `/api/v1/projects/:projectId/consistency-reports` | range、scope、severity_threshold | report_id、status | 一致性检查 |
| GET | `/api/v1/projects/:projectId/consistency-reports` | status、page | 报告列表、状态、issue_count、severity_summary | 一致性报告页 |
| GET | `/api/v1/projects/:projectId/consistency-reports/:reportId` | reportId | 报告详情、问题项、来源快照、建议处理方式 | 一致性报告详情 |

### 6.1 接口补充要求

- 所有 mutating 接口（PATCH / POST）必须写入 `operation_log` 或对应状态日志。
- `assemble-context`、`update-dynamic-state`、`consistency-reports` 创建接口必须支持 `Idempotency-Key`，避免页面重试产生重复快照或重复报告。
- 列表接口必须支持 `page`、`page_size`、`sort`、`order`，并返回统一分页结构。
- 报告详情中的 `issues` 必须是结构化数组，不允许只返回不可解析的长文本。
- 失败响应必须使用统一错误结构，并返回 `request_id`。

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

- 项目工作区 / 记忆上下文
- 一致性报告
- 上下文预览

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
- [ ] 本迭代完成后可以支撑下一迭代。

---

## 10. 本迭代明确不做

- 不做超出本迭代页面范围的业务功能。
- 不做未定义接口的隐式前端调用。
- 不做绕过 WorkflowRun / AgentTask / LLMCallLog 的核心生产链路。
- 不做 n8n 核心编排。

---

## 11. 评审结论记录

| 项目 | 结论 |
|---|---|
| 产品评审结论 | 通过。Iteration 6 能补齐内容生产后的项目级记忆、一致性检查与上下文装配能力，符合 AI Content Factory 的持续生产需求。 |
| 技术评审结论 | 有条件通过。需按本次评审补齐 RecentContentWindow、StyleGuide、记忆快照、报告详情、幂等与审计要求。 |
| 是否允许进入开发 | 允许进入开发。 |
| 评审人 | ChatGPT |
| 评审日期 | 2026-05-19 |
| 是否需要更新蓝图 | 否。当前改动属于蓝图内细化，不改变系统架构边界。 |
| 主要风险 | 记忆系统边界过宽导致重复保存正文；动态状态来源不可追踪；上下文装配缺少 Token 预算导致提示词膨胀；一致性报告只有文本无法用于后续策略闭环。 |
| 调整项 | 补齐四类记忆模型、快照追踪、上下文装配规则、一致性问题结构、人工修正接口、幂等与审计要求。 |

---

## 12. 最终需求变更说明

| 变更项 | 变更内容 | 原因 |
|---|---|---|
| 记忆模型范围 | 从 StaticContext / DynamicState 扩展为 StaticContext / DynamicState / RecentContentWindow / StyleGuide | 与蓝图中 Knowledge Memory 职责保持一致。 |
| 记忆主表约束 | 明确 `knowledge_memory` 不保存完整正文，只保存摘要、规则和结构化状态 | 避免与 ContentItem、content_version、ContentAsset 重复。 |
| 快照追踪 | 强制 `memory_snapshot` 记录来源、Token、截断策略和装配结果 | 保证上下文装配可审计、可复现。 |
| 上下文装配规则 | 明确装配顺序、预算裁剪、幂等复现和 AgentRuntime 边界 | 防止提示词膨胀与绕过核心工作流。 |
| 一致性报告 | 增加报告详情接口和结构化 issue 要求 | 支撑前端查看、后续策略建议和人工处理。 |
| 人工修正入口 | 增加 StaticContext / StyleGuide 修正接口，并要求写 operation_log | 满足人工节点保留和长期知识维护需求。 |
| API 幂等 | 对创建快照、动态状态更新、一致性报告创建增加 Idempotency-Key 要求 | 防止页面重试造成重复快照或重复报告。 |
| 前端交互 | 记忆页增加 StyleGuide、快照列表、截断策略、严重级别汇总展示 | 保证页面不是只读占位页，满足原型驱动验收。 |

---

## 13. 最终对齐验证

| 检查项 | 最终需求定义 | 蓝图定义 | 状态 |
|---|---|---|---|
| Core 内容类型无关 | Core 仅使用 project、content_item、memory、snapshot、report 等通用概念 | Core 不得写死 Novel / Book / Chapter | ✅ 对齐 |
| Knowledge Memory 职责 | 覆盖 StaticContext、DynamicState、RecentContentWindow、StyleGuide | Knowledge Memory 层负责四类上下文 | ✅ 对齐 |
| Workflow / Agent 边界 | 装配上下文只作为 AgentRuntime 输入，不绕过 WorkflowRun / AgentTask / LLMCallLog | 核心生产链路由 Workflow Engine + Agent Runtime 承载 | ✅ 对齐 |
| 人工节点 | StaticContext / StyleGuide 支持人工修正并写 operation_log | 审稿、发布、策略执行等关键动作保留人工确认 | ✅ 对齐 |
| API 契约 | 使用 `/api/v1`、DTO、统一错误、分页、幂等、OpenAPI | API 契约要求统一前缀、响应结构、分页、幂等和 OpenAPI | ✅ 对齐 |
| 前后端一体验收 | 页面、接口、原型映射、CSS / JS、e2e 均纳入本迭代 | 每个迭代必须包含页面、接口、页面-接口映射与验收标准 | ✅ 对齐 |
| n8n 边界 | 不使用 n8n 承载记忆装配或一致性核心逻辑 | n8n 只做外围通知、Webhook、同步、告警 | ✅ 对齐 |

结论：最终需求与产品蓝图无背离，不需要更新 `00-product-blueprint.md`。

