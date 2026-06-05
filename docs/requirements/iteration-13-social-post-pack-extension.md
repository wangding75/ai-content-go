# Iteration 13：Social Post Pack 内容类型扩展

> 文件定位：本文件是 Iteration 13 的最新需求、技术、前端页面、接口契约和原型映射文档。  
> 蓝图约束：必须遵守 `00-product-blueprint.md`。  
> 技术栈：后端 Go / Golang；前端 Next.js + TypeScript。  
> 更新日期：2026-06-05  
> 评审日期：2026-06-05  
> 评审结论摘要：评审通过。Iteration 13 与产品蓝图无背离，可进入开发；需补齐 Social Post Pack 注册边界、短内容生成运行详情、多版本文案状态、标签 / 封面文案异步生成、与审稿 / 发布 / 指标链路的衔接，以及幂等与审计要求。  
> 是否需要更新蓝图：否。  
> 重要规则：本项目不再保留单独前端迭代文档；前端需求已整合到本迭代。  
> 原型开发规则：本迭代前端页面必须基于原型页面实现，CSS / JS 均需可用，并接入可点击导航入口。

---

## 1. 迭代目标

新增 Social Post Pack；前端整合项目模板与项目工作区短内容扩展页面。

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
| Core 内容类型无关 | 使用 `social_post_extension`、`social_post_variant`、`metric_template` 承载 Social Post Pack 扩展能力 | Core 层不得写死 Novel / Book / Chapter 等内容类型专属概念 | ✅ 对齐 |
| 内容类型插件化 | 新增 Social Post Pack，并通过项目模板、schema、workflow、metrics 注册进入系统 | Novel / Article / Social Post 通过 Content Pack 扩展 | ✅ 对齐 |
| 工作流自研 | 初始需求要求短内容生成、标签生成、封面文案生成，但未明确必须经 WorkflowRun / AgentTask | 核心生产链路由自研 Workflow Engine 承载，AgentTask / LLMCallLog 必须可追踪 | ⚠️ 偏差：已补齐所有 LLM 生成动作必须创建 WorkflowRun / AgentTask / LLMCallLog |
| 多版本文案 | 初始需求有多版本文案页面，但缺少版本状态、选择与追踪规则 | ContentItem / content_version 是通用内容产物，人工节点必须保留 | ⚠️ 偏差：已补齐 `social_post_variant` 状态、选择动作和进入审稿 / 发布链路的规则 |
| 标签与封面文案 | 初始需求仅定义生成入口，未说明同步 / 异步边界和结果查询 | API 契约要求异步任务返回 run_id / job_id，结果通过详情接口查询 | ⚠️ 偏差：已调整为异步生成并补充结果查询接口 |
| Platform Adapter 边界 | Social Post Pack 面向小红书 / 微博等图文平台，容易与平台发布适配混淆 | Platform Adapter / Browser Extension 负责平台格式转换、插件协作和数据采集 | ⚠️ 偏差：已明确本迭代只生成短内容、标签、封面文案，不做自动发布、平台登录、插件填充和指标采集 |
| 指标体系 | 初始需求复用 `metric_template`，未说明与 Iteration 8 MetricRecord 的关系 | Metrics & Strategy 负责 MetricRecord、MetricTemplate、StrategySuggestion | ✅ 对齐：本迭代只注册 Social 默认指标模板，不直接采集指标记录 |
| 前端原型 | 已包含项目模板、Social 内容生成、多版本文案、标签与封面文案页面 | 每个迭代必须包含前端页面、接口、页面-接口映射和验收标准 | ✅ 对齐 |

---

## 3. 产品需求

- 提供本迭代对应的系统级或项目级操作入口。
- 页面操作必须能映射到明确 API。
- 异步动作必须返回运行记录 ID 或任务 ID。
- 状态变更必须记录操作日志。
- 页面必须支持空状态、加载态、错误态、成功反馈。


### 3.1 评审后补充产品需求

- Social Post Pack 必须作为 Content Pack 扩展注册，不得把小红书 / 微博等平台专属字段写入 Core 模型；平台差异通过扩展配置、Platform Adapter 或后续插件层处理。
- 项目级 Social 配置必须支持目标平台、默认文案长度、默认版本数量、标签策略、封面文案策略、语气风格、禁用表达等字段，并允许人工修改。
- 短内容生成必须基于通用 `ContentItem` 创建内容单元，再通过 `social_post_variant` 保存多版本候选文案；不得绕过 ContentItem 直接生成不可追踪的孤立文案。
- 多版本文案至少支持 `generated`、`selected`、`rejected`、`archived` 状态；同一 ContentItem 同一时间只能有一个主选版本进入审稿 / 发布链路。
- 标签与封面文案属于 Social Post Pack 的生成资产，必须关联 `content_item_id`、可选 `variant_id`、目标平台和生成运行，便于追溯。
- 所有涉及 LLM 的生成动作，包括短内容、多版本文案、标签、封面文案，必须创建 WorkflowRun / AgentTask / LLMCallLog，并通过详情或资产查询接口读取最终结果。
- Social Post Pack 默认指标模板只负责注册指标定义，例如曝光、点击、点赞、收藏、评论、转发、关注转化；真实指标录入和趋势分析继续复用 Iteration 8 的 MetricRecord 能力。
- 本迭代必须保留人工选择与确认节点：生成多个文案后，由人工选择主版本，后续再进入审稿、发布队列或手动发布回填。
- 本迭代不做图片生成、图片上传、平台自动发布、平台登录、浏览器插件自动填充、外部指标采集；这些能力分别由后续媒体资产能力、Iteration 7 / 11 / 8 承接。

---

## 4. Go 后端技术需求

- social_post_extension
- social_post_variant
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
| 数据模型 | 完成本迭代数据模型、迁移、Repository / Store、Service 边界，覆盖 `social_post_extension`、`social_post_variant`、`metric_template`。 |
| API 契约 | 第 6 节接口必须具备 Go request / response DTO、validator 校验、统一响应结构、统一错误结构和 OpenAPI 描述。 |
| 状态与审计 | 状态变更接口必须校验状态流转合法性，并写入 `operation_log`；失败时返回统一错误码和 `request_id`。 |
| 异步执行 | 触发型接口不得阻塞 HTTP 请求，必须返回 `run_id` / `job_id` / 业务记录 ID，并通过运行记录或详情接口查询结果。 |
| 幂等与重试 | 创建运行、触发执行、发布回填、确认类接口按 `api-contract-standard.md` 要求支持 `Idempotency-Key` 或明确说明不需要幂等。 |
| 前端联调支撑 | 后端需提供可用于页面空态、加载态、错误态、成功态联调的数据与错误响应；不得只返回无法渲染的占位数据。 |


## 4.2 数据模型细化

| 模型 | 必要字段 / 约束 | 说明 |
|---|---|---|
| `social_post_extension` | `id`、`project_id`、`target_platforms`、`default_variant_count`、`caption_length_policy`、`hashtag_policy`、`cover_copy_policy`、`tone_style`、`forbidden_terms`、`config_version`、`created_at`、`updated_at` | 项目级 Social Post 配置；只保存内容生成策略，不保存平台账号敏感凭证。 |
| `social_post_variant` | `id`、`content_item_id`、`generation_run_id`、`workflow_run_id`、`variant_index`、`platform`、`title`、`body`、`hashtags`、`cover_copy`、`tone_style`、`status`、`content_version_id`、`selected_at`、`created_at` | 多版本文案候选；主选版本可绑定 `content_version_id` 进入审稿 / 发布链路。 |
| `metric_template` | `id`、`content_type_code`、`platform`、`metric_code`、`name`、`unit`、`aggregation_type`、`enabled`、`created_at` | Social Post Pack 默认指标模板；不直接写入 MetricRecord。 |

## 4.3 短内容生成与多版本规则

- `version_count` 必须有服务端上限，建议默认 3、最大 10，避免一次请求造成不可控 LLM 成本。
- 短内容生成必须先创建或关联通用 `ContentItem`，再生成多个 `social_post_variant`，保证后续审稿、发布、指标都能沿用通用内容单元。
- 每次生成运行必须返回 `generation_run_id`、`workflow_run_id` 和初始状态；前端通过生成详情接口轮询或刷新结果。
- `social_post_variant.status` 默认是 `generated`；人工选择后变为 `selected`，被替换或废弃时变为 `rejected` / `archived`。
- 主选版本进入审稿时，应生成或绑定 `content_version`，避免后续再次编辑导致发布内容不可追溯。
- 重新生成只新增新的生成运行和候选版本，不覆盖历史 `social_post_variant`；历史版本保留用于对比和审计。

## 4.4 标签、封面文案与平台适配边界

- 标签与封面文案生成必须支持按平台生成，例如小红书偏标签和封面标题，微博偏正文话题和短标题，但平台差异只保存在扩展配置或结果字段中。
- 标签生成结果至少包含 `tags[]`、`platform`、`source_variant_id`、`generation_run_id`、`created_at`。
- 封面文案生成结果至少包含 `cover_copy`、`style`、`platform`、`source_variant_id`、`generation_run_id`、`created_at`。
- 本迭代只产出文本资产和结构化建议，不生成图片、不上传图片、不处理图片素材库。
- 平台格式转换、插件自动填充、平台发布回填、采集日志由 Iteration 11 的 Platform Adapter / Browser Extension 承接，本迭代不得重复实现。

## 4.5 幂等、审计与错误处理

- `content-packs/social-post/register`、项目 Social 配置更新、短内容生成、标签生成、封面文案生成、选择主版本等接口必须支持 `Idempotency-Key` 或明确说明幂等策略。
- 注册 Social Post Pack 必须以 `content_pack_code = social_post` 做唯一约束；重复注册同一 schema 版本应返回同一结果，schema 冲突返回 `CONFLICT`。
- 状态变更接口必须写入 `operation_log`，至少记录操作者、来源状态、目标状态、原因和关联资源。
- 所有列表接口必须支持 `page`、`page_size`、`sort`、`order`，并按项目、平台、状态过滤。
- LLM 输出必须做结构化校验；多版本文案、标签、封面文案解析失败时返回 `AGENT_OUTPUT_INVALID`，并保留失败 AgentTask 与 LLMCallLog。
- 前端失败态必须展示统一错误结构中的 `error.code`、`error.message`、`request_id`。

## 4.6 与前后迭代依赖

| 依赖来源 | 本迭代使用方式 |
|---|---|
| Iteration 1：ContentProject / ContentType | Social Post Pack 注册为内容类型扩展，项目通过 `content_type_code = social_post` 启用。 |
| Iteration 2：WorkflowRun / AgentTask / LLMCallLog | 所有 LLM 生成动作必须落 WorkflowRun / AgentTask / LLMCallLog。 |
| Iteration 4：ContentItem / generation_run | 短内容生成以 ContentItem 为主资源，生成运行可复用通用 generation_run 语义。 |
| Iteration 5：content_review / content_version | 人工选择的主版本应绑定 content_version，后续进入审稿流程。 |
| Iteration 7：PublishJob | 通过审稿后的 Social 内容进入发布队列，本迭代不直接完成发布。 |
| Iteration 8：MetricRecord / MetricTemplate | 本迭代只注册 Social 默认指标模板，指标录入与趋势仍由 Iteration 8 承担。 |
| Iteration 11：Platform Adapter / Browser Extension | 平台适配、插件填充、采集日志由 Iteration 11 承接，本迭代只提供可发布文本资产。 |
| Iteration 12：Article Pack | 复用 Content Pack 注册方式，但 Social Post 增加多版本文案、标签和封面文案资产规则。 |

---

## 5. 前端页面范围

| 页面 / 组件 | 路由建议 | 交互 |
|---|---|---|
| 项目模板管理 / Social Post Pack | `/content-packs/social-post` | 注册 Social Post Pack、查看 schema/workflows/metrics |
| 项目工作区 / Social 内容生成 | `/projects/:projectId/social-post` | 短内容生成、版本数量配置、生成状态 |
| 多版本文案 | `/content-items/:itemId/social-post/variants` | 查看多版本文案和生成结果 |
| 标签与封面文案 | `/content-items/:itemId/social-post/assets` | 生成标签、封面文案并查看结果 |


## 5.1 前端需求

| 类别 | 要求 |
|---|---|
| 页面范围 | 必须实现第 5 节定义的全部页面 / 组件：项目模板管理 / Social Post Pack、项目工作区 / Social 内容生成、多版本文案、标签与封面文案。 |
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
| 项目模板管理 / Social Post Pack | `/content-packs/social-post` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 注册 Social Post Pack、查看 schema/workflows/metrics；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 项目工作区 / Social 内容生成 | `/projects/:projectId/social-post` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 短内容生成、版本数量配置、生成状态；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 多版本文案 | `/content-items/:itemId/social-post/variants` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 查看多版本文案和生成结果；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 标签与封面文案 | `/content-items/:itemId/social-post/assets` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 生成标签、封面文案并查看结果；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |

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
| GET | `/api/v1/content-packs/social-post` | 无 | schema、workflows、metrics、current_version | 查看 Social Post Pack 配置 |
| POST | `/api/v1/content-packs/social-post/register` | schema、workflows、metrics、version | content_pack_id、content_type_id、registered_version | 注册 Social Post Pack |
| GET | `/api/v1/projects/:projectId/social-post/extension` | projectId | target_platforms、variant_policy、hashtag_policy、cover_copy_policy | Social 项目配置 |
| PATCH | `/api/v1/projects/:projectId/social-post/extension` | target_platforms、default_variant_count、caption_length_policy、hashtag_policy、cover_copy_policy、tone_style、forbidden_terms | version_id、operation_log_id | 更新 Social 项目配置 |
| POST | `/api/v1/projects/:projectId/social-post/generation-runs` | topic、source_content_item_id、platform、version_count、tone_style、asset_options | generation_run_id、workflow_run_id、status | 生成短内容 |
| GET | `/api/v1/projects/:projectId/social-post/generation-runs/:id` | id | status、content_item_id、variants、error、workflow_run_id | 生成状态 / 生成详情 |
| GET | `/api/v1/content-items/:id/social-post/variants` | status、platform、page、page_size | 多版本文案列表、pagination | 多版本文案 |
| POST | `/api/v1/content-items/:id/social-post/variants/:variantId/select` | note | selected_variant_id、content_version_id、operation_log_id | 选择主版本 |
| POST | `/api/v1/content-items/:id/social-post/tags/generate` | platform、variant_id、count、style | generation_run_id、workflow_run_id、status | 生成标签 |
| POST | `/api/v1/content-items/:id/social-post/cover-copy/generate` | platform、variant_id、style、count | generation_run_id、workflow_run_id、status | 生成封面文案 |
| GET | `/api/v1/content-items/:id/social-post/assets` | platform、variant_id | tags、cover_copy、asset_suggestions、source_runs | 标签与封面文案结果 |

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

- 项目模板管理 / Social Post Pack
- 项目工作区 / Social 内容生成
- 多版本文案
- 标签与封面文案

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

## 11. 需求评审结论

| 评审维度 | 结论 | 改进建议 |
|---|---|---|
| 完整性 | 初始需求覆盖 Social Post Pack 注册、短内容生成、多版本文案、标签与封面文案页面，但缺少生成详情、资产结果查询、版本选择和依赖链路 | 已补齐生成详情、variants 查询、主版本选择、assets 查询和依赖说明。 |
| 合理性 | Social Post Pack 作为 Content Pack 扩展合理；但标签和封面文案属于 LLM 生成动作，不应同步直接返回最终文本 | 已改为异步返回 `generation_run_id` / `workflow_run_id`，结果通过查询接口获取。 |
| 一致性 | 与 Iteration 1 / 2 / 4 / 5 / 7 / 8 / 11 / 12 基本一致；初始需求未明确与审稿、发布、指标、平台适配的边界 | 已补齐与前后迭代依赖，避免绕过审稿、直接发布或重复实现平台插件能力。 |
| 风险点 | 多版本文案可能导致版本不可追踪；平台差异可能侵入 Core；LLM 输出可能不稳定；一次生成过多版本可能造成成本失控 | 已新增版本状态、主选版本规则、平台边界、结构化校验和 `version_count` 上限要求。 |
| 前端可验收性 | 页面范围明确，但初始接口不足以支撑“生成状态”“多版本文案”“标签与封面文案结果”完整交互 | 已补充必要查询与状态接口，前端可覆盖 loading / empty / error / success。 |

---

## 12. 最终需求变更说明

| 变更项 | 变更内容 | 原因 |
|---|---|---|
| 文档头部 | 增加评审日期、评审结论摘要、是否需要更新蓝图 | 方便后续追踪本轮评审结论。 |
| 初步对齐检查 | 新增 2.1 对齐表，标明内容类型插件化、工作流追踪、多版本文案、平台边界等检查项 | 满足需求评审流程，并暴露初始需求中的偏差点。 |
| 产品需求 | 新增 3.1，明确 Social Post Pack 注册边界、人工选择、多版本、标签 / 封面文案、指标和不做范围 | 降低 Core 污染、重复实现和不可追溯风险。 |
| 数据模型 | 新增 4.2，细化 `social_post_extension`、`social_post_variant`、`metric_template` 必要字段 | 初始需求只有表名，缺少可落地字段约束。 |
| 生成规则 | 新增 4.3，定义短内容生成、多版本文案、主选版本和历史版本保留规则 | 防止多版本覆盖、成本失控和后续审稿 / 发布链路断裂。 |
| 资产边界 | 新增 4.4，明确标签、封面文案的结果字段和与 Platform Adapter 的边界 | 避免本迭代越界实现图片、插件或平台发布能力。 |
| 幂等审计 | 新增 4.5，明确注册、配置、生成、选择主版本等动作的幂等、审计和错误处理 | 对齐 API 契约和可追踪要求。 |
| 迭代依赖 | 新增 4.6，明确与 Iteration 1 / 2 / 4 / 5 / 7 / 8 / 11 / 12 的关系 | 防止 Social Pack 越界替代通用内容、审稿、发布或指标能力。 |
| 接口清单 | 扩展第 6 节，增加 Pack 查询、配置更新、生成详情、多版本查询、主版本选择、资产查询；将标签 / 封面文案生成改为异步返回运行 ID | 支撑前端完整联调和符合异步任务契约。 |

---

## 13. 最终对齐验证

| 检查项 | 最终需求定义 | 蓝图定义 | 状态 |
|---|---|---|---|
| Core 内容类型无关 | Core 仍使用 ContentProject、ContentItem、content_version、WorkflowRun 等通用模型；Social 专属字段仅在 Social Post Pack 扩展模型中 | Core 层不得写死 Novel / Book / Chapter 等内容类型专属概念 | ✅ 对齐 |
| 内容类型插件化 | Social Post Pack 通过 `/content-packs/social-post/register` 注册 schema、workflows、metrics | Novel / Article / Social Post 通过 Content Pack 扩展 | ✅ 对齐 |
| 工作流自研 | 短内容、标签、封面文案生成均必须创建 WorkflowRun / AgentTask / LLMCallLog | 核心生产链路由自研 Workflow Engine 承载，Agent 可追踪 | ✅ 对齐 |
| 人工节点保留 | 多版本文案生成后必须人工选择主版本，后续才能进入审稿 / 发布链路 | 审稿、发布、策略执行必须支持人工确认 | ✅ 对齐 |
| n8n 外围化 | 本迭代不使用 n8n 做核心编排，只允许后续通知或外围同步 | n8n 只做通知、Webhook、外部 API 同步、告警 | ✅ 对齐 |
| Platform Adapter 边界 | 本迭代不做平台登录、自动发布、插件填充、采集日志，只输出可发布文本资产 | Platform Adapter / Browser Extension 负责平台格式转换、插件协作和采集回填 | ✅ 对齐 |
| Metrics & Strategy | 本迭代只注册 Social 默认指标模板，不直接写入指标记录或生成策略建议 | Metrics & Strategy 负责 MetricRecord、MetricTemplate、StrategySuggestion | ✅ 对齐 |
| 前后端一体验收 | 已补齐页面、接口、页面-接口映射、原型渲染和测试要求 | 每个迭代必须包含前端页面、后端接口、页面-接口映射 | ✅ 对齐 |

最终判断：Iteration 13 修订后的需求与产品蓝图无背离，不需要更新 `00-product-blueprint.md`。
