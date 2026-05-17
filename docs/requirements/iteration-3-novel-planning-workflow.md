# Iteration 3：Novel Pack 新书规划流程

> 文件定位：本文件是 Iteration 3 的最新需求、技术、前端页面、接口契约和原型映射文档。  
> 蓝图约束：必须遵守 `00-product-blueprint.md`。  
> 技术栈：后端 Go / Golang；前端 Next.js + TypeScript。  
> 更新日期：2026-05-17  
> 重要规则：本项目不再保留单独前端迭代文档；前端需求已整合到本迭代。  
> 原型开发规则：本迭代前端页面必须基于原型页面实现，CSS / JS 均需可用，并接入可点击导航入口。
> 评审日期：2026-05-17  
> 评审结论：无蓝图背离；本次评审补强 Novel Pack 扩展边界、候选选题持久化、人工确认审计、接口分页 / 幂等与 WorkflowRun / AgentTask / LLMCallLog 追踪要求。

---

## 1. 迭代目标

基于通用 Workflow Engine，实现小说项目的选题、世界观、人物、大纲规划；前端整合到项目工作区的内容规划页。

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
- 本迭代是 Novel Pack 的业务扩展，不改变 Core 的内容类型无关边界；Novel 专属概念只能出现在 Novel Pack 扩展模型、接口 tag、页面和 workflow input/output 中。
- 新书规划必须通过已发布的 WorkflowTemplateVersion 触发 WorkflowRun，由 StepRun 派生 AgentTask，并记录 LLMCallLog。
- 候选选题必须持久化为可确认资源，支持人工确认；确认动作必须写入 operation_log，并保留确认前后的状态。
- 世界观、人物、大纲等规划产物必须关联 ContentProject / ContentAsset，不允许以 Book / Chapter 作为 Core 资源名。

---

## 4. Go 后端技术需求

- content_asset
- planning_run
- planning_snapshot
- novel_topic_candidate
- novel_worldview
- novel_character
- novel_arc

通用要求：

- 使用 Go struct 定义请求 / 响应 DTO。
- 使用 validator 做入参校验。
- 使用 sqlc + pgx 或等价方式访问 PostgreSQL。
- 使用 goose 或等价工具管理数据库迁移。
- 所有接口进入 OpenAPI。
- 状态变更写入 `operation_log`。
- 异步任务通过 worker / queue 执行，不阻塞 HTTP 请求。
- `planning_run` 作为 Novel Pack 业务运行记录，必须关联 `workflow_run_id`。
- `novel_topic_candidate`、`novel_worldview`、`novel_character`、`novel_arc` 只能作为 Novel Pack 扩展表，不得进入 Core 通用模型命名。


## 4.1 后端需求

| 类别 | 要求 |
|---|---|
| 数据模型 | 完成本迭代数据模型、迁移、Repository / Store、Service 边界，覆盖 `content_asset`、`planning_run`、`planning_snapshot`、`novel_topic_candidate`、`novel_worldview`、`novel_character`、`novel_arc`。 |
| Novel Pack 边界 | `novel_*` 表、DTO、API tag 和页面只属于 Novel Pack 扩展；Core 层只感知 `content_project`、`content_asset`、`workflow_run`、`agent_task` 等通用模型。 |
| Workflow 衔接 | 启动规划时必须校验项目存在、内容类型为 Novel Pack、WorkflowTemplateVersion 已发布；成功后创建 `planning_run` 并关联 `workflow_run_id`。 |
| Agent 追踪 | 规划流程中的选题、世界观、人物、大纲生成步骤必须由 StepRun 派生 AgentTask，并记录 LLMCallLog、输入、输出、错误和 Token / 成本。 |
| 候选选题与人工确认 | 候选选题必须持久化到 `novel_topic_candidate`，具备 `candidate_id`、`status`、`score`、`reason`、`snapshot_id`；确认动作从 `candidate` 流转为 `confirmed`，并写入 `operation_log`。 |
| API 契约 | 第 6 节接口必须具备 Go request / response DTO、validator 校验、统一响应结构、统一错误结构和 OpenAPI 描述。 |
| 状态与审计 | 状态变更接口必须校验状态流转合法性，并写入 `operation_log`；失败时返回统一错误码和 `request_id`。 |
| 异步执行 | 触发型接口不得阻塞 HTTP 请求，必须返回 `run_id` / `job_id` / 业务记录 ID，并通过运行记录或详情接口查询结果。 |
| 幂等与重试 | 创建运行、触发执行、发布回填、确认类接口按 `api-contract-standard.md` 要求支持 `Idempotency-Key` 或明确说明不需要幂等。 |
| 前端联调支撑 | 后端需提供可用于页面空态、加载态、错误态、成功态联调的数据与错误响应；不得只返回无法渲染的占位数据。 |

---

## 5. 前端页面范围

| 页面 / 组件 | 路由建议 | 交互 |
|---|---|---|
| 项目工作区 / 内容规划 | `/projects/:projectId/planning` | 启动规划、查看规划运行状态、查看候选选题 |
| 候选选题确认弹窗 | `/projects/:projectId/planning/topics` | 查看候选选题并确认 |
| 世界观编辑 | `/projects/:projectId/novel/worldview` | 查看和编辑世界观、禁止项、版本 |
| 人物管理 | `/projects/:projectId/novel/characters` | 人物列表、新增人物、查看详情 |
| 大纲管理 | `/projects/:projectId/novel/arcs` | 查看弧线大纲和规划结果 |


## 5.1 前端需求

| 类别 | 要求 |
|---|---|
| 页面范围 | 必须实现第 5 节定义的全部页面 / 组件：项目工作区 / 内容规划、候选选题确认弹窗、世界观编辑、人物管理、大纲管理。 |
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
| 项目工作区 / 内容规划 | `/projects/:projectId/planning` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 启动规划、查看规划运行状态、查看候选选题；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 候选选题确认弹窗 | `/projects/:projectId/planning/topics` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 查看候选选题并确认；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 世界观编辑 | `/projects/:projectId/novel/worldview` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 查看和编辑世界观、禁止项、版本；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 人物管理 | `/projects/:projectId/novel/characters` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 人物列表、新增人物、查看详情；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 大纲管理 | `/projects/:projectId/novel/arcs` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 查看弧线大纲和规划结果；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |

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
| POST | `/api/v1/projects/:projectId/novel/planning-runs` | genre、audience、count、template_version_id、input_override | planning_run_id、workflow_run_id、status | 启动规划 |
| GET | `/api/v1/projects/:projectId/novel/planning-runs` | page、page_size、sort、order、status | planning_run 列表、分页信息 | 内容规划页 / 规划历史 |
| GET | `/api/v1/projects/:projectId/novel/planning-runs/:runId` | runId | 候选选题、状态、workflow_run_id、step_runs 摘要 | 规划运行详情 |
| POST | `/api/v1/projects/:projectId/novel/topics/:topicId/confirm` | note | confirmed_topic_id、previous_status、current_status、operation_log_id | 确认选题 |
| GET | `/api/v1/projects/:projectId/novel/worldview` | projectId | 世界观、禁止项、版本 | 世界观页 |
| PATCH | `/api/v1/projects/:projectId/novel/worldview` | worldview、forbidden_rules、note | version_id、operation_log_id | 编辑世界观 |
| GET | `/api/v1/projects/:projectId/novel/characters` | page、page_size、sort、order、role | 人物列表、分页信息 | 人物管理 |
| POST | `/api/v1/projects/:projectId/novel/characters` | name、role、profile、note | character_id、operation_log_id | 新增人物 |
| GET | `/api/v1/projects/:projectId/novel/arcs` | page、page_size、sort、order | 弧线大纲列表、分页信息 | 大纲管理 |

接口补充要求：

- `POST /api/v1/projects/:projectId/novel/planning-runs` 必须支持 `Idempotency-Key`，不得同步等待完整规划结果。
- `POST /api/v1/projects/:projectId/novel/topics/:topicId/confirm` 是人工确认状态变更接口，必须校验候选选题属于当前项目且尚未确认。
- 所有列表接口必须返回统一分页结构；所有失败响应必须返回统一错误码、错误信息和 `request_id`。
- 页面如果展示 Agent 生成过程或失败原因，必须从 WorkflowRun / StepRun / AgentTask / LLMCallLog 查询，不得直接读取不可追踪的临时结果。

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

- 项目工作区 / 内容规划
- 候选选题确认弹窗
- 世界观编辑
- 人物管理
- 大纲管理

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
- [ ] `planning_run` 已关联 `workflow_run_id`，规划结果可追踪到 StepRun / AgentTask / LLMCallLog。
- [ ] 候选选题已持久化为 `novel_topic_candidate`，确认选题接口返回 `operation_log_id` 和确认前后状态。
- [ ] Novel 专属模型只在 Novel Pack 扩展边界内实现，Core 模块未新增 `novel_*`、Book、Chapter 等核心资源。
- [ ] 规划运行、确认选题等重复提交敏感接口已按 API 契约支持 `Idempotency-Key`。
- [ ] 本迭代完成后可以支撑下一迭代。

---

## 10. 本迭代明确不做

- 不做超出本迭代页面范围的业务功能。
- 不做未定义接口的隐式前端调用。
- 不做绕过 WorkflowRun / AgentTask / LLMCallLog 的核心生产链路。
- 不把 `novel_*` 模型上移为 Core 通用模型。
- 不使用 Book / Chapter 作为 Core API、Core 表名或 Core DTO 命名。
- 不做 n8n 核心编排。

---

## 11. 评审记录

| 日期 | 结论 | 变更 | 原因 |
|---|---|---|---|
| 2026-05-17 | 无蓝图背离，已修订本迭代需求 | 增加 `planning_run`、`novel_topic_candidate`；补强 WorkflowRun / AgentTask / LLMCallLog 追踪；确认选题返回 `operation_log_id`；列表接口补充分页参数；明确 Novel Pack 扩展边界。 | 原需求目标与蓝图一致，但候选选题资源、人工确认审计、接口契约和 Core / Novel Pack 边界需要更明确，避免实现阶段把小说概念写入 Core 或绕过工作流追踪。 |
