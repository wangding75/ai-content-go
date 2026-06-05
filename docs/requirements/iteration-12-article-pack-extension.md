# Iteration 12：Article Pack 内容类型扩展

> 文件定位：本文件是 Iteration 12 的最新需求、技术、前端页面、接口契约和原型映射文档。  
> 蓝图约束：必须遵守 `00-product-blueprint.md`。  
> 技术栈：后端 Go / Golang；前端 Next.js + TypeScript。  
> 更新日期：2026-06-04  
> 评审日期：2026-06-04  
> 评审结论摘要：评审通过。Iteration 12 与产品蓝图无背离，可进入开发；需补齐 Article Pack 注册边界、Article 项目配置版本、资料 / SEO 输入、生成快照、ContentItem / content_version 关联、指标模板边界、审稿发布衔接和幂等审计要求。  
> 是否需要更新蓝图：否。  
> 重要规则：本项目不再保留单独前端迭代文档；前端需求已整合到本迭代。  
> 原型开发规则：本迭代前端页面必须基于原型页面实现，CSS / JS 均需可用，并接入可点击导航入口。

---

## 1. 迭代目标

新增 Article Pack；前端整合项目模板与项目工作区文章扩展页面。

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
| Content Pack 定位 | 新增 Article Pack，并提供项目模板和项目工作区文章页面 | 蓝图明确 Article Pack 是 AI Content Factory Core 的垂直内容包之一 | ✅ 对齐 |
| Core 内容类型无关 | 初始模型使用 `article_extension`、`article_generation_snapshot`，未直接改造 Core 命名 | Core 层不得写死 Novel / Book / Chapter 等专属概念，扩展能力应放在 Content Pack | ✅ 对齐 |
| 内容类型插件化 | 通过 `/content-packs/article/register` 注册 schema / workflows / metrics | Novel / Article / Social Post 通过 Content Pack 扩展 | ✅ 对齐 |
| Workflow / Agent 边界 | 初始需求有 Article 生成接口，但未明确必须落 `WorkflowRun`、`AgentTask`、`LLMCallLog` | 核心生产链路由 Workflow Engine 承载，Agent 调用必须可追踪 | ⚠️ 偏差：已补齐生成链路、快照和可追踪要求 |
| 内容资产与版本边界 | 初始需求只返回“文章内容”，未明确 `ContentItem` / `content_version` 关联 | Core 以 `ContentItem` 表示内容单元，审稿与发布依赖内容版本 | ⚠️ 偏差：已补齐生成结果写入通用内容单元和版本流转 |
| 人工节点保留 | 初始需求未说明生成文章如何进入审稿、发布与策略确认 | 审稿、发布、策略执行必须支持人工确认 | ⚠️ 偏差：已明确复用 Iteration 5 / 7 / 11 的人工闭环 |
| 指标闭环 | 初始需求包含 Article 指标配置，但未说明与 `MetricTemplate` / `MetricRecord` 的关系 | Metrics & Strategy 负责 MetricRecord、MetricTemplate、StrategySuggestion | ⚠️ 偏差：已补齐 Article 默认指标模板与 MetricRecord 边界 |
| API 契约 | 初始接口较少，缺少列表、重试、指标配置、幂等和错误边界 | API 契约要求 `/api/v1`、DTO、统一响应、分页、幂等、OpenAPI | ⚠️ 偏差：已补齐接口清单和幂等审计要求 |
| 前端原型与导航 | 已定义三类 Article 页面和路由 | 每个迭代必须包含前端页面、接口、页面-接口映射与验收标准 | ✅ 对齐 |

结论：初始需求与蓝图方向一致，无需更新蓝图；评审重点是把 Article Pack 从“单独文章生成页面”收敛为“复用 Core 内容单元、Workflow、审稿、发布、指标闭环的内容类型扩展”。

---

## 3. 产品需求

- 提供本迭代对应的系统级或项目级操作入口。
- 页面操作必须能映射到明确 API。
- 异步动作必须返回运行记录 ID 或任务 ID。
- 状态变更必须记录操作日志。
- 页面必须支持空状态、加载态、错误态、成功反馈。

### 3.1 评审后补充产品需求

- Article Pack 必须作为 Content Pack 扩展接入现有 Core，而不是新增一套独立文章系统。注册后应形成可用于创建项目的 `ContentType`、默认 WorkflowTemplate / WorkflowTemplateVersion、默认 Prompt / Agent 配置和默认指标模板。
- Article 项目必须保留项目级扩展配置，至少覆盖选题风格、目标受众、目标平台、SEO 配置、资料来源策略、文章结构策略、默认字数范围、默认工作流版本和指标模板启用状态。
- Article 生成必须产出通用 `ContentItem`，并通过 Article 扩展表保存文章专属字段；Core 层仍只识别 `ContentItem`、`content_version`、`generation_run` 等通用资源。
- Article 生成链路必须经过 Workflow Engine 和 Agent Runtime，生成运行需要返回 `generation_run_id` 与 `workflow_run_id`，具体 Agent 调用必须落 `AgentTask` / `LLMCallLog`。
- Article 生成输入必须支持人工提供资料或引用已有内容资产，至少包含 `topic`、`audience`、`source_refs`、`seo_keywords`、`outline_required`、`target_platform`、`generation_config`；不得隐式依赖未声明的数据来源。
- 生成结果至少包含标题、摘要、大纲、正文、SEO 元信息、引用来源、生成配置摘要、质量检查摘要和关联 `content_item_id` / `content_version_id`。
- 生成后的文章草稿必须进入现有审稿链路；是否发布、复制、回填、采集指标，不在 Article Pack 内自建状态机，应复用 Iteration 5 / 7 / 8 / 11 的能力。
- Article 指标配置只负责注册和启用 Article 默认指标模板；真实指标录入、趋势、缺失提醒和策略建议继续复用 Iteration 8 / 9。
- Article Pack 页面必须分别覆盖：内容包注册与 schema 查看、项目级 Article 配置与生成、Article 指标配置；项目级入口必须接入项目工作区菜单。
- 本迭代完成后应为 Iteration 13 的 Social Post Pack 复用内容包扩展模式提供参考，但不得提前实现 Social Post 专属能力。

---

## 4. Go 后端技术需求

- article_extension
- article_generation_snapshot
- metric_template

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
| 数据模型 | 完成本迭代数据模型、迁移、Repository / Store、Service 边界，覆盖 `article_extension`、`article_generation_snapshot`、`metric_template`。 |
| API 契约 | 第 6 节接口必须具备 Go request / response DTO、validator 校验、统一响应结构、统一错误结构和 OpenAPI 描述。 |
| 状态与审计 | 状态变更接口必须校验状态流转合法性，并写入 `operation_log`；失败时返回统一错误码和 `request_id`。 |
| 异步执行 | 触发型接口不得阻塞 HTTP 请求，必须返回 `run_id` / `job_id` / 业务记录 ID，并通过运行记录或详情接口查询结果。 |
| 幂等与重试 | 创建运行、触发执行、发布回填、确认类接口按 `api-contract-standard.md` 要求支持 `Idempotency-Key` 或明确说明不需要幂等。 |
| 前端联调支撑 | 后端需提供可用于页面空态、加载态、错误态、成功态联调的数据与错误响应；不得只返回无法渲染的占位数据。 |

## 4.2 数据模型细化

| 模型 | 必要字段 / 约束 | 说明 |
|---|---|---|
| `article_extension` | `id`、`project_id`、`content_type_id`、`topic_style`、`audience_profile`、`seo_config`、`source_policy`、`structure_policy`、`default_workflow_template_version_id`、`enabled_metric_codes`、`version`、`created_at`、`updated_at` | 项目级 Article 配置表；只保存文章内容类型专属配置，不污染 Core 项目表。 |
| `article_generation_snapshot` | `id`、`project_id`、`generation_run_id`、`workflow_run_id`、`content_item_id`、`content_version_id`、`topic`、`outline`、`title`、`summary`、`seo_metadata`、`source_refs`、`quality_summary`、`status`、`error_code`、`created_at`、`completed_at` | Article 生成快照；用于页面详情、失败排查和可追溯，不替代 `generation_run` / `WorkflowRun`。 |
| `metric_template` | `id`、`content_type`、`metric_code`、`name`、`unit`、`value_type`、`platform`、`enabled` | 复用 Iteration 8 指标模板模型；本迭代只负责为 Article Pack 注册默认模板，不新增独立指标系统。 |

## 4.3 Article Pack 注册规则

- `content-packs/article/register` 必须幂等：同一 `Idempotency-Key` 和相同 schema / workflows / metrics 输入重复提交返回同一 `content_pack_id`；输入不一致返回 `IDEMPOTENCY_CONFLICT`。
- 注册成功后必须至少完成三类绑定：`content_type.code = article`、Article 默认工作流模板版本、Article 默认指标模板。
- 注册接口不得直接创建业务项目；业务项目仍通过 Iteration 1 的项目创建接口创建，并选择 Article ContentType。
- schema 必须声明项目配置字段、生成输入字段、生成输出字段和指标模板字段；OpenAPI example 中必须给出最小可运行示例。
- workflows 应引用现有 Workflow Engine 能力，至少支持“资料整理 / 大纲生成 / 正文生成 / 质量检查”四类步骤，可在实现上合并步骤，但追踪上必须能定位 AgentTask。
- metrics 应注册 Article 默认指标，例如 `views`、`click_through_rate`、`read_completion_rate`、`likes`、`comments`、`shares`、`saves`、`conversion_count`；是否启用由项目级配置控制。

## 4.4 Article 生成规则

- 创建 Article 生成运行时，必须校验项目存在、项目 ContentType 为 Article、Article 扩展配置已启用、默认工作流版本有效。
- 生成接口只启动异步运行，不同步等待最终文章正文；响应返回 `generation_run_id`、`workflow_run_id`、初始 `status` 和查询详情入口。
- `article_generation_snapshot.status` 至少包含 `queued`、`running`、`succeeded`、`failed`、`canceled`；状态变更必须可从 `generation_run` 或 `WorkflowRun` 追踪。
- 生成成功后必须写入或关联一个 `ContentItem`，并创建可审稿的 `content_version`；页面展示正文时应优先读取版本化内容，避免只依赖运行结果缓存。
- 生成失败时必须保留失败阶段、错误码、错误信息、`request_id`、相关 `workflow_run_id` / `agent_task_id`，并支持按原始输入重试。
- 同一幂等键下重复创建生成运行不得产生多个实际运行；若请求体变化必须返回 `IDEMPOTENCY_CONFLICT`。

## 4.5 Article 指标配置规则

- Article 指标配置页展示项目已启用的 Article 指标模板，不直接录入 MetricRecord。
- 指标模板必须支持按平台启用，例如公众号、知乎、SEO 站点可以启用不同指标集合。
- 指标配置变更必须写入 `operation_log`，并返回配置版本号，便于后续指标趋势和策略建议解释。
- 指标数据来源可以是人工录入、浏览器插件回填或 n8n 外围同步，但最终必须落 `MetricRecord`，不得只保存在外部系统。

## 4.6 与前后迭代依赖

| 依赖来源 | 本迭代使用方式 |
|---|---|
| Iteration 1：ContentProject / ContentType / Prompt / Provider | Article Pack 注册为 ContentType；项目创建、Prompt 模板、模型 Provider 均复用已有能力。 |
| Iteration 2：WorkflowRun / AgentTask / LLMCallLog | Article 生成必须经 Workflow Engine 与 Agent Runtime 执行，并记录模型调用日志。 |
| Iteration 4：ContentItem / generation_run | Article 生成复用通用内容生成运行与 ContentItem，不新增平行内容主表。 |
| Iteration 5：content_review / content_version | 生成后的文章草稿进入审稿，审核通过后形成可发布版本。 |
| Iteration 6：Knowledge Memory | 资料、风格、近期内容可作为上下文来源，但 Article Pack 不直接改造记忆系统。 |
| Iteration 7：PublishJob | Article 发布任务复用发布队列与手动回填，不在本迭代实现自动发布状态机。 |
| Iteration 8：MetricTemplate / MetricRecord | 本迭代注册 Article 指标模板；指标录入、趋势和缺失提醒复用指标中心。 |
| Iteration 9：StrategySuggestion | Article 指标后续可驱动策略建议，但本迭代不新增策略规则引擎。 |
| Iteration 11：Platform Adapter / Browser Extension | 插件可在后续读取已审稿文章并辅助发布 / 回填；本迭代只提供稳定内容与配置数据。 |
| Iteration 13：Social Post Pack | Article Pack 的扩展模式为 Social Post Pack 提供参考，但不得提前实现 Social 专属字段。 |

## 4.7 幂等、审计与错误处理

- `POST /api/v1/content-packs/article/register`、`PATCH /api/v1/projects/:projectId/article/extension`、`POST /api/v1/projects/:projectId/article/generation-runs`、`POST /api/v1/projects/:projectId/article/generation-runs/:id/retry`、`PATCH /api/v1/projects/:projectId/article/metrics` 必须支持 `Idempotency-Key` 或明确说明幂等策略。
- 所有配置变更、生成触发、重试、取消、指标启停都必须写入 `operation_log`。
- 所有接口必须使用统一响应结构；失败时展示 `error.code`、`error.message`、`details` 和 `request_id`。
- 列表接口必须支持 `page`、`page_size`、`sort`、`order`，并支持按 `status`、`created_at`、`topic`、`target_platform` 筛选。

---

## 5. 前端页面范围

| 页面 / 组件 | 路由建议 | 交互 |
|---|---|---|
| 项目模板管理 / Article Pack | `/content-packs/article` | 注册 Article Pack、查看 schema/workflows/metrics |
| 项目工作区 / Article 内容规划与生产 | `/projects/:projectId/article` | 文章配置、文章生成、生成详情 |
| Article 指标配置 | `/projects/:projectId/article/metrics` | Article 指标模板与项目指标配置 |


## 5.1 前端需求

| 类别 | 要求 |
|---|---|
| 页面范围 | 必须实现第 5 节定义的全部页面 / 组件：项目模板管理 / Article Pack、项目工作区 / Article 内容规划与生产、Article 指标配置。 |
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
| 项目模板管理 / Article Pack | `/content-packs/article` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 注册 Article Pack、查看 schema/workflows/metrics；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 项目工作区 / Article 内容规划与生产 | `/projects/:projectId/article` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 文章配置、文章生成、生成详情；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| Article 指标配置 | `/projects/:projectId/article/metrics` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | Article 指标模板与项目指标配置；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |

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

## 5.3 页面交互补充

| 页面 / 组件 | 补充交互要求 |
|---|---|
| 项目模板管理 / Article Pack | 展示 Article Pack 注册状态、schema 摘要、默认 workflow、默认 metrics；支持重新注册 / 查看注册结果 / 查看错误详情。 |
| 项目工作区 / Article 内容规划与生产 | 展示项目 Article 配置、生成输入表单、运行列表、生成详情、关联 ContentItem / content_version、失败重试入口。 |
| Article 指标配置 | 展示默认指标模板、平台差异、项目启用状态、最近配置版本；支持启用 / 停用指标并展示审计反馈。 |

页面必须明确区分“生成运行状态”和“内容审稿 / 发布状态”：前者来自 Article generation / WorkflowRun，后者来自既有 Review / Publish 模块。

---

## 6. 后端接口交付清单

| 方法 | 接口 | 输入 | 输出 | 原型页面映射 |
|---|---|---|---|---|
| POST | `/api/v1/content-packs/article/register` | schema、workflows、metrics | content_pack_id、content_type_id、registered_workflow_version_ids、metric_template_ids | 注册 Article Pack |
| GET | `/api/v1/content-packs/article` | 无 | 注册状态、schema、workflows、metrics | Article Pack 详情 |
| GET | `/api/v1/projects/:projectId/article/extension` | projectId | 文章扩展配置、version、enabled_metric_codes | Article 项目配置 |
| PATCH | `/api/v1/projects/:projectId/article/extension` | topic_style、audience_profile、seo_config、source_policy、structure_policy、default_workflow_template_version_id | version_id、operation_log_id | 更新文章配置 |
| GET | `/api/v1/projects/:projectId/article/generation-runs` | status、topic、target_platform、page、page_size、sort、order | 生成运行列表、分页信息 | Article 生成列表 |
| POST | `/api/v1/projects/:projectId/article/generation-runs` | topic、audience、source_refs、seo_keywords、outline_required、target_platform、generation_config | generation_run_id、workflow_run_id、status | 文章生成 |
| GET | `/api/v1/projects/:projectId/article/generation-runs/:id` | id | 生成状态、Article 快照、content_item_id、content_version_id、workflow_run_id、agent_task_refs | 生成详情 |
| POST | `/api/v1/projects/:projectId/article/generation-runs/:id/retry` | reason、input_override | generation_run_id、workflow_run_id、status | 失败重试 |
| GET | `/api/v1/content-items/:itemId/article/snapshot` | itemId | title、summary、outline、seo_metadata、source_refs、latest_content_version_id | Article 内容详情 |
| GET | `/api/v1/projects/:projectId/article/metrics` | platform、enabled、page | Article 指标模板和项目启用状态 | Article 指标配置 |
| PATCH | `/api/v1/projects/:projectId/article/metrics` | enabled_metric_codes、platform_overrides、note | version_id、operation_log_id | 更新 Article 指标配置 |

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

### 7.1 本迭代页面-接口映射补充

| 页面动作 | 必须调用的接口 | 说明 |
|---|---|---|
| 查看 Article Pack 注册状态 | `GET /api/v1/content-packs/article` | 用于项目模板管理页加载 schema / workflows / metrics 状态。 |
| 注册 / 重新注册 Article Pack | `POST /api/v1/content-packs/article/register` | 必须使用 `Idempotency-Key`，成功后刷新注册状态。 |
| 加载 Article 项目配置 | `GET /api/v1/projects/:projectId/article/extension` | 项目工作区 Article 页首屏必须调用。 |
| 保存 Article 项目配置 | `PATCH /api/v1/projects/:projectId/article/extension` | 保存后显示 Toast，并展示新版本号。 |
| 提交文章生成 | `POST /api/v1/projects/:projectId/article/generation-runs` | 只创建异步运行，页面跳转或提示查看生成详情。 |
| 查看生成列表 | `GET /api/v1/projects/:projectId/article/generation-runs` | 支持分页、筛选、排序。 |
| 查看生成详情 | `GET /api/v1/projects/:projectId/article/generation-runs/:id` | 不依赖列表缓存，展示快照、ContentItem、content_version 和 Workflow 关联。 |
| 重试失败生成 | `POST /api/v1/projects/:projectId/article/generation-runs/:id/retry` | 仅允许失败运行重试，必须记录原因。 |
| 查看 Article 内容详情 | `GET /api/v1/content-items/:itemId/article/snapshot` | 用于从 ContentItem 进入文章专属详情。 |
| 加载 / 保存指标配置 | `GET /api/v1/projects/:projectId/article/metrics`、`PATCH /api/v1/projects/:projectId/article/metrics` | 只处理模板启停，不直接录入 MetricRecord。 |

---

## 8. 原型页面映射

本迭代页面已映射到原型文件：

```text
prototype/ai-content-factory-clickable-prototype.html
```

对应页面：

- 项目模板管理 / Article Pack
- 项目工作区 / Article 内容规划与生产
- Article 指标配置

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
- 不做绕过 ContentItem / content_version 的独立文章内容主表。
- 不做独立审稿、独立发布队列、独立指标记录和独立策略建议系统。
- 不做外部资料自动抓取 / 爬虫能力；本迭代只接受用户输入、已有内容资产或显式 `source_refs`。
- 不做平台自动发布；发布协作继续由 PublishJob / Browser Extension / Platform Adapter 承接。
- 不做 n8n 核心编排。

---

## 11. 评审结论记录

| 项目 | 结论 |
|---|---|
| 产品评审结论 | 通过。Iteration 12 能在不改变 Core 架构的前提下引入 Article Pack，满足内容类型扩展阶段目标。 |
| 技术评审结论 | 有条件通过。需按本次评审补齐 Article Pack 注册幂等、项目扩展配置版本、生成快照、ContentItem / content_version 关联、指标模板边界和前后迭代复用关系。 |
| 是否允许进入开发 | 允许进入开发。 |
| 评审人 | ChatGPT |
| 评审日期 | 2026-06-04 |
| 是否需要更新蓝图 | 否。当前改动属于蓝图内 Article Pack 细化，不改变系统架构边界。 |
| 主要风险 | Article 生成绕过通用 Workflow / ContentItem；Article 指标配置与 MetricRecord 系统重复；生成结果未版本化导致审稿发布不可追溯；注册接口重复执行导致 schema / workflow / metric 重复。 |
| 调整项 | 补齐初步对齐检查、评审后产品需求、数据模型细化、注册规则、生成规则、指标配置规则、依赖关系、接口清单、页面-接口映射、明确不做范围和最终对齐验证。 |

---

## 12. 最终需求变更说明

| 变更项 | 变更内容 | 原因 |
|---|---|---|
| Article Pack 边界 | 明确 Article Pack 只作为 Content Pack 扩展，注册为 ContentType / Workflow / Metrics，不新增独立文章系统 | 保持 Core 内容类型无关，避免和蓝图冲突 |
| 生成链路 | 明确 Article 生成必须返回 `generation_run_id` / `workflow_run_id`，并落 AgentTask / LLMCallLog | 符合自研 Workflow Engine 与 Agent 可追踪原则 |
| 内容落库 | 明确生成成功后关联 `ContentItem` 和 `content_version`，Article 专属信息写入扩展快照 | 保证后续审稿、发布、指标都能复用通用链路 |
| 指标边界 | 明确本迭代只注册和启用 Article 指标模板，不直接录入 MetricRecord | 避免重复建设 Iteration 8 指标中心 |
| 审稿发布衔接 | 明确文章草稿进入 Iteration 5 审稿，发布复用 Iteration 7 / 11 | 保留人工节点并复用已有状态机 |
| API 清单 | 增加注册状态、生成列表、失败重试、Article 内容快照、指标配置接口 | 支撑前端三类页面完整联调和可追溯 |
| 幂等与审计 | 明确注册、配置、生成、重试、指标启停需要幂等或审计 | 符合 API 契约规范和重复提交防护要求 |
| 明确不做 | 增加不做独立审稿 / 发布 / 指标 / 策略、不做爬虫、不做自动发布 | 收敛迭代范围，减少实现风险 |

---

## 13. 最终对齐验证

| 检查项 | 最终需求定义 | 蓝图定义 | 状态 |
|---|---|---|---|
| Core 内容类型无关 | Core 继续使用 `ContentProject`、`ContentType`、`ContentItem`、`content_version`、`generation_run` 等通用模型，Article 字段进入扩展表 | Core 层不得写死 Novel / Book / Chapter 等专属概念 | ✅ 对齐 |
| 内容类型插件化 | Article Pack 通过注册 schema、workflow、metrics 接入项目模板体系 | Novel / Article / Social Post 通过 Content Pack 扩展 | ✅ 对齐 |
| Workflow 自研 | Article 生成只创建异步 Workflow / Generation Run，不同步直出不可追踪正文 | 核心生产链路由自研 Workflow Engine 承载 | ✅ 对齐 |
| Agent 可追踪 | 资料整理、大纲、正文、质检等 Agent 调用必须落 AgentTask / LLMCallLog | AgentTask、LLMCallLog 必须记录输入、输出、模型、Token、成本、错误 | ✅ 对齐 |
| 人工节点保留 | 生成后的文章草稿进入审稿，发布和策略执行继续人工确认 | 审稿、发布、策略执行必须支持人工确认 | ✅ 对齐 |
| Metrics & Strategy 边界 | Article 只注册指标模板，MetricRecord、趋势和策略建议复用既有模块 | Metrics & Strategy 负责 MetricRecord、MetricTemplate、StrategySuggestion | ✅ 对齐 |
| n8n 外围化 | n8n 只可做通知或外部同步，核心生成 / 状态不落在 n8n execution | n8n 只做通知、Webhook、外部 API 同步、告警 | ✅ 对齐 |
| 前后端一体验收 | 三类页面、接口、页面映射、原型渲染、错误态和 e2e 均纳入本迭代 | 每个迭代必须包含前端页面、后端接口、页面-接口映射 | ✅ 对齐 |

结论：最终需求与产品蓝图无背离；无需更新 `00-product-blueprint.md`。本次评审已直接更新 `iteration-12-article-pack-extension.md`。
