# Iteration 4：内容单元生成闭环

> 文件定位：本文件是 Iteration 4 的最新需求、技术、前端页面、接口契约和原型映射文档。  
> 蓝图约束：必须遵守 `00-product-blueprint.md`。  
> 技术栈：后端 Go / Golang；前端 Next.js + TypeScript。  
> 更新日期：2026-05-17  
> 评审日期：2026-05-18  
> 评审结论摘要：与 `00-product-blueprint.md` 无背离；补强生成运行状态模型、前置规划依赖、WorkflowRun / AgentTask / LLMCallLog 追踪、幂等与重试、列表查询与页面联调要求。  
> 是否需要更新蓝图：否。  
> 重要规则：本项目不再保留单独前端迭代文档；前端需求已整合到本迭代。  
> 原型开发规则：本迭代前端页面必须基于原型页面实现，CSS / JS 均需可用，并接入可点击导航入口。

---

## 1. 迭代目标

实现 Novel 内容单元脚本与正文生成；前端整合到项目工作区的内容生产页。

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

- 提供项目级内容生产入口，支持从项目工作区进入手动生成、批量生成、生成状态查看、失败重试和 ContentItem 查看。
- 内容生成必须基于 Iteration 3 已确认的规划资产，包括 confirmed_topic、worldview、characters、arc / outline 等；未完成规划确认的项目不得直接生成正文。
- `generation_run` 是内容生产业务运行记录；每次手动生成、批量生成、失败重试都必须关联或创建 `workflow_run`，不得绕过 Workflow Engine 直接调用 Agent。
- 生成链路必须产生可追踪的 `AgentTask` 与 `LLMCallLog`，用于查看输入、输出、模型、Token、成本、错误和结构化校验结果。
- `content_item` 表示通用内容单元；Novel 章节相关字段只能进入 `novel_chapter_extension`，不得把 Book / Chapter 作为 Core 资源名。
- 内容单元状态至少覆盖 `planned`、`generating`、`generated`、`generation_failed`、`pending_review`，为 Iteration 5 审稿中心提供稳定入口。
- 页面操作必须能映射到明确 API；异步动作必须立即返回 `generation_run_id` / `workflow_run_id`，由详情或列表接口查询最终结果。
- 状态变更必须记录操作日志；页面必须支持空状态、加载态、错误态、成功反馈，并展示统一错误结构中的 `request_id`。

---

## 4. Go 后端技术需求

- content_item
- novel_chapter_extension
- generation_run

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
| 数据模型 | 完成本迭代数据模型、迁移、Repository / Store、Service 边界，覆盖 `content_item`、`novel_chapter_extension`、`generation_run`。 |
| 前置依赖 | 创建 `generation_run` 前必须校验项目、ContentType、已确认规划结果、可用工作流模板版本和 LLM Provider；缺失时返回统一错误结构。 |
| 状态模型 | 明确 `generation_run` 状态：`pending`、`running`、`succeeded`、`failed`、`retrying`；明确 `content_item` 状态：`planned`、`generating`、`generated`、`generation_failed`、`pending_review`。 |
| Workflow 集成 | 手动生成、批量生成、失败重试必须关联 `workflow_run_id`；生成步骤必须派生 `workflow_step_run`、`agent_task`、`llm_call_log`，并可从生成详情追踪。 |
| API 契约 | 第 6 节接口必须具备 Go request / response DTO、validator 校验、统一响应结构、统一错误结构和 OpenAPI 描述。 |
| 状态与审计 | 状态变更接口必须校验状态流转合法性，并写入 `operation_log`；失败时返回统一错误码和 `request_id`。 |
| 异步执行 | 触发型接口不得阻塞 HTTP 请求，必须返回 `generation_run_id`、`workflow_run_id` 或业务记录 ID，并通过运行记录或详情接口查询结果。 |
| 幂等与重试 | 手动生成、批量生成、失败重试属于重复提交敏感接口，必须支持 `Idempotency-Key`；同一幂等键请求体不一致时返回 `IDEMPOTENCY_CONFLICT`。 |
| 前端联调支撑 | 后端需提供可用于页面空态、加载态、错误态、成功态联调的数据与错误响应；不得只返回无法渲染的占位数据。 |

---

## 5. 前端页面范围

| 页面 / 组件 | 路由建议 | 交互 |
|---|---|---|
| 项目工作区 / 内容生产 | `/projects/:projectId/production` | 手动生成、批量生成、查看生成状态 |
| 生成运行详情 | `/generation-runs/:runId` | 查看生成运行详情、状态、输出 ContentItem |
| ContentItem 列表 | `/projects/:projectId/content-items` | 内容单元列表、筛选、查看详情 |
| 失败重试 | `/generation-runs/:runId/retry` | 对失败生成运行发起重试 |


## 5.1 前端需求

| 类别 | 要求 |
|---|---|
| 页面范围 | 必须实现第 5 节定义的全部页面 / 组件：项目工作区 / 内容生产、生成运行详情、ContentItem 列表、失败重试。 |
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
| 项目工作区 / 内容生产 | `/projects/:projectId/production` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 手动生成、批量生成、查看生成状态；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 生成运行详情 | `/generation-runs/:runId` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 查看生成运行详情、状态、输出 ContentItem；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| ContentItem 列表 | `/projects/:projectId/content-items` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 内容单元列表、筛选、查看详情；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 失败重试 | `/generation-runs/:runId/retry` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 对失败生成运行发起重试；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |

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
| POST | `/api/v1/projects/:projectId/generation-runs` | `confirmed_topic_id`、`worldview_version_id`、`arc_id`、`target_count`、`start_sequence_no`、`generation_config`；Header：`Idempotency-Key` | `generation_run_id`、`workflow_run_id`、`status` | 手动生成 |
| POST | `/api/v1/projects/:projectId/generation-runs/batch` | `range`、`batch_size`、`generation_config`；Header：`Idempotency-Key` | `generation_run_ids[]`、`workflow_run_ids[]`、`accepted_count` | 批量生成 |
| GET | `/api/v1/projects/:projectId/generation-runs` | `status`、`page`、`page_size`、`sort`、`order` | 生成运行列表、分页信息 | 内容生产 / 生成状态 |
| GET | `/api/v1/generation-runs/:id` | `id` | `generation_run_id`、`workflow_run_id`、状态、步骤摘要、输出 `content_items[]`、错误信息 | 生成详情 |
| POST | `/api/v1/generation-runs/:id/retry` | `reason`、`input_override`；Header：`Idempotency-Key` | `new_generation_run_id`、`workflow_run_id`、`operation_log_id` | 失败重试 |
| GET | `/api/v1/projects/:projectId/content-items` | `status`、`page`、`page_size`、`sort`、`order` | ContentItem 列表、分页信息 | 内容生产列表 |
| GET | `/api/v1/content-items/:id` | `id` | 正文、扩展字段、版本、来源 `generation_run_id` | 内容详情 |

### 6.1 接口补充规则

- `POST /generation-runs`、`POST /generation-runs/batch`、`POST /generation-runs/:id/retry` 必须支持 `Idempotency-Key`。
- 所有列表接口必须返回统一分页结构，并支持 `page`、`page_size`、`sort`、`order`。
- 生成详情必须能向下追踪 `workflow_run`、`workflow_step_run`、`agent_task`、`llm_call_log`；可以通过嵌入摘要或关联 ID 实现。
- 失败重试必须保留原失败运行记录，不允许覆盖原 `generation_run`；新运行通过 `retry_of_generation_run_id` 关联原运行。
- 生成成功后的 `content_item` 默认进入 `pending_review` 或等价待审状态，但审稿动作由 Iteration 5 承接。

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

- 项目工作区 / 内容生产
- 生成运行详情
- ContentItem 列表
- 失败重试

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
- [ ] 手动生成、批量生成、失败重试接口均支持 `Idempotency-Key`，重复提交不会产生重复内容单元。
- [ ] 生成详情可以追踪到 `workflow_run_id`、步骤摘要、AgentTask / LLMCallLog 关联信息和失败原因。
- [ ] `content_item` 状态流转可进入 `pending_review`，能够被 Iteration 5 审稿中心接入。
- [ ] Novel 章节字段仅落在 `novel_chapter_extension`，Core API 与 Core 数据模型不引入 `book` / `chapter` 作为核心资源。
- [ ] 本迭代完成后可以支撑下一迭代。

---

## 10. 本迭代明确不做

- 不做超出本迭代页面范围的业务功能。
- 不做未定义接口的隐式前端调用。
- 不做绕过 WorkflowRun / AgentTask / LLMCallLog 的核心生产链路。
- 不做 n8n 核心编排。
- 不做审稿通过、审稿打回、编辑后通过等审稿状态机能力；这些由 Iteration 5 承接。
- 不做发布、指标、策略建议能力；这些由 Iteration 7、8、9 承接。
- 不把 `novel_chapter_extension` 上升为 Core 模型；其仅作为 Novel Pack 扩展表。

---

## 11. 评审记录（2026-05-18）

### 11.1 初步对齐检查

| 检查项 | 初始需求描述 | 蓝图定义 | 状态 |
|---|---|---|---|
| 产品定位 | 实现 Novel 内容单元脚本与正文生成 | Novel Pack 是 AI Content Factory 的第一个垂直内容包，Core 面向多内容形态 | ✅ 对齐 |
| Core 命名边界 | Core 使用 `content_item`，Novel 字段进入 `novel_chapter_extension` | Core 不得写死 Novel / Book / Chapter，ContentItem 替代 Chapter | ✅ 对齐 |
| 工作流边界 | 生成运行关联 `workflow_run_id` | 核心生产链路由自研 Workflow Engine 承载 | ✅ 对齐 |
| Agent 可追踪 | 生成过程必须产生 AgentTask / LLMCallLog | AgentTask、LLMCallLog 必须记录输入、输出、模型、Token、成本、错误 | ✅ 对齐 |
| n8n 边界 | 不做 n8n 核心编排 | n8n 只做通知、Webhook、外部 API 同步、告警 | ✅ 对齐 |
| 前后端一体验收 | 页面、接口、原型映射已在本迭代内定义 | 每个迭代必须包含页面、接口、页面-接口映射与验收标准 | ✅ 对齐 |
| API 契约 | 初始接口已定义，但列表查询、幂等、重试输出不够明确 | 列表接口支持分页排序；创建运行等接口支持 Idempotency-Key；异步任务返回 run_id / job_id | ⚠️ 已修订 |
| 前置依赖 | 初始需求未明确必须使用 Iteration 3 规划结果 | Iteration 3 产出规划资产，Iteration 4 承接内容生成 | ⚠️ 已修订 |

### 11.2 评审意见与修订说明

| 问题 | 影响 | 修订结果 |
|---|---|---|
| 缺少生成运行列表接口 | 内容生产页只能跳转详情，无法稳定查看历史生成状态 | 新增 `GET /api/v1/projects/:projectId/generation-runs` |
| 重试接口输出过于模糊 | 无法区分原失败运行和新重试运行 | 输出改为 `new_generation_run_id`、`workflow_run_id`、`operation_log_id`，并要求 `retry_of_generation_run_id` 关联 |
| 幂等要求未落到具体接口 | 重复点击手动生成 / 批量生成 / 重试可能产生重复内容 | 明确三类触发接口必须支持 `Idempotency-Key` |
| 前置规划依赖不明确 | 可能绕过选题、世界观、大纲直接生成正文 | 明确创建生成运行前必须校验 confirmed_topic、worldview、characters、arc / outline 等规划资产 |
| 状态模型不完整 | Iteration 5 审稿中心缺少稳定接入状态 | 明确 `content_item` 生成后进入 `pending_review` 或等价待审状态 |

### 11.3 最终对齐验证

结论：修订后的 Iteration 4 与 `00-product-blueprint.md` 无背离，不需要更新蓝图。  
本次只更新 `iteration-4-content-generation-loop.md`，用于补强接口契约、状态流转、前置依赖和验收标准。
