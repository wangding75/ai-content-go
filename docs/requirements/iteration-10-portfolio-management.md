# Iteration 10：Project Portfolio 多项目管理

> 文件定位：本文件是 Iteration 10 的最新需求、技术、前端页面、接口契约和原型映射文档。  
> 蓝图约束：必须遵守 `00-product-blueprint.md`。  
> 技术栈：后端 Go / Golang；前端 Next.js + TypeScript。  
> 更新日期：2026-05-28  
> 评审日期：2026-05-28  
> 评审结论摘要：评审通过。Iteration 10 与产品蓝图无背离，可进入开发；需补齐 Portfolio 详情接口、项目成员管理规则、健康快照计算口径、跨项目成本/策略汇总边界、幂等与审计要求。  
> 是否需要更新蓝图：否。  
> 重要规则：本项目不再保留单独前端迭代文档；前端需求已整合到本迭代。  
> 原型开发规则：本迭代前端页面必须基于原型页面实现，CSS / JS 均需可用，并接入可点击导航入口。

---

## 1. 迭代目标

多项目组合管理；前端整合系统级 Portfolio 页面。

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
| Core 内容类型无关 | 使用 `project_portfolio`、`portfolio_project`、`portfolio_status_snapshot` 聚合项目，不引入 Novel / Book / Chapter 核心命名 | Core 层不得写死 Novel / Book / Chapter 等小说专属概念 | ✅ 对齐 |
| 系统级 Portfolio 页面 | 初始需求将 Portfolio 管理放在系统级页面 | 蓝图页面体系包含系统级 `Portfolio 管理` | ✅ 对齐 |
| 前后端一体验收 | 初始需求包含 Portfolio 管理、详情、项目优先级、健康 / 成本汇总页面和 API | 每个迭代必须包含前端页面、后端接口、页面-接口映射 | ✅ 对齐 |
| 与 Metrics / Strategy 边界 | 初始需求提到健康 / 成本汇总，但未明确数据来源与计算口径 | 蓝图将 MetricRecord、MetricTemplate、StrategySuggestion 归属 Metrics & Strategy 模块 | ⚠️ 偏差：已补齐只读聚合、快照与来源引用规则，不在本迭代重建指标或策略引擎 |
| 与人工节点边界 | 初始需求含项目优先级调整，但未说明优先级是否自动执行策略 | 蓝图要求策略执行保留人工确认 | ⚠️ 偏差：已明确 Portfolio 只提供资源分配视角，不自动执行策略建议 |
| API 契约 | 初始接口缺少 Portfolio 详情、成员移除、快照列表和幂等说明 | API 契约要求 DTO、统一错误、分页、状态日志、幂等敏感接口 | ⚠️ 偏差：已补齐详情接口、成员管理、快照查询、幂等与审计规则 |
| n8n 边界 | 初始需求未使用 n8n | 蓝图要求 n8n 只做通知、Webhook、外部 API 同步、告警 | ✅ 对齐 |

---

## 3. 产品需求

- 提供本迭代对应的系统级或项目级操作入口。
- 页面操作必须能映射到明确 API。
- 异步动作必须返回运行记录 ID 或任务 ID。
- 状态变更必须记录操作日志。
- 页面必须支持空状态、加载态、错误态、成功反馈。

### 3.1 评审后补充产品需求

- Portfolio 是跨项目管理视角，用于按主题、业务线、平台、内容类型或增长目标聚合多个 `ContentProject`，不得替代项目工作区内的生产、审稿、发布、指标录入和策略执行流程。
- 一个项目可以加入多个 Portfolio；同一 Portfolio 内同一项目只能存在一条有效成员关系。
- Portfolio 成员必须支持优先级、权重、目标角色和备注，优先级用于人工资源排序，不得自动触发策略执行或项目状态变更。
- 健康汇总必须基于已有项目状态、发布状态、MetricRecord、LLMCallLog、StrategySuggestion 等来源做只读聚合，并记录来源时间范围。
- 成本汇总必须复用 LLMCallLog / 成本汇总能力，不在本迭代新增独立成本账本。
- `portfolio_status_snapshot` 用于保存某个时间点的聚合结果，必须记录 `date_range`、计算时间、来源版本 / 来源时间、健康评分、风险项和成本摘要。
- Portfolio 页面可以展示跨项目待办和策略建议摘要，但确认 / 忽略 / 执行策略仍回到 Iteration 9 的策略建议接口，不在 Portfolio 中绕过人工确认。
- Portfolio 只做管理与观察，不负责创建 ContentItem、不触发 WorkflowRun、不直接调用 AgentRuntime。

---

## 4. Go 后端技术需求

- project_portfolio
- portfolio_project
- portfolio_status_snapshot

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
| 数据模型 | 完成本迭代数据模型、迁移、Repository / Store、Service 边界，覆盖 `project_portfolio`、`portfolio_project`、`portfolio_status_snapshot`。 |
| API 契约 | 第 6 节接口必须具备 Go request / response DTO、validator 校验、统一响应结构、统一错误结构和 OpenAPI 描述。 |
| 状态与审计 | 状态变更接口必须校验状态流转合法性，并写入 `operation_log`；失败时返回统一错误码和 `request_id`。 |
| 异步执行 | 触发型接口不得阻塞 HTTP 请求，必须返回 `run_id` / `job_id` / 业务记录 ID，并通过运行记录或详情接口查询结果。 |
| 幂等与重试 | 创建运行、触发执行、发布回填、确认类接口按 `api-contract-standard.md` 要求支持 `Idempotency-Key` 或明确说明不需要幂等。 |
| 前端联调支撑 | 后端需提供可用于页面空态、加载态、错误态、成功态联调的数据与错误响应；不得只返回无法渲染的占位数据。 |

## 4.2 数据模型细化

| 模型 | 必要字段 / 约束 | 说明 |
|---|---|---|
| `project_portfolio` | `id`、`name`、`description`、`scope_type`、`owner_id`、`status`、`health_policy`、`created_at`、`updated_at` | Portfolio 主表；`scope_type` 可表示内容类型、平台、业务线或自定义组合；`health_policy` 存储阈值配置。 |
| `portfolio_project` | `id`、`portfolio_id`、`project_id`、`priority`、`weight`、`role`、`note`、`enabled`、`created_at`、`updated_at` | Portfolio 成员关系；同一 `portfolio_id + project_id` 只能有一条有效关系。 |
| `portfolio_status_snapshot` | `id`、`portfolio_id`、`date_range`、`health_score`、`health_status`、`project_count`、`risk_summary`、`cost_summary`、`strategy_summary`、`source_refs`、`calculated_at` | 聚合快照表；用于健康 / 成本看板和历史对比，不作为指标事实表。 |

## 4.3 Portfolio 健康计算规则

- 健康状态最低支持 `healthy`、`warning`、`critical`、`unknown`。
- 健康分数建议按 0-100 输出，计算依据至少包含生产进度、失败运行、待审稿、待发布、指标缺失、成本异常和待确认策略建议。
- 数据不足时不得伪造健康结论，必须返回 `unknown`，并在 `risk_summary` 中说明缺失来源。
- 健康快照必须保存 `source_refs`，引用参与计算的项目、指标区间、成本区间、策略建议集合或运行状态集合。
- 重新计算健康快照可以异步执行；HTTP 接口只返回 `snapshot_id` 或 `job_id`，最终结果通过详情或快照列表查询。

## 4.4 成员管理与优先级规则

- `POST /api/v1/portfolios/:id/projects` 必须校验 `project_id` 存在、项目未归档或允许被观察、当前用户有读取项目权限。
- 重复加入同一项目应返回 `CONFLICT`，或在相同 `Idempotency-Key` 下返回既有成员关系。
- 调整优先级必须写入 `operation_log`，记录原优先级、新优先级、操作者和原因 / 备注。
- 支持从 Portfolio 移除项目；移除只影响成员关系，不删除 `ContentProject`，不改变项目状态。
- 权重用于健康 / 成本汇总的加权展示，默认值为 1；不得用于自动暂停项目或自动执行策略建议。

## 4.5 与前置迭代依赖

| 依赖来源 | 本迭代使用方式 |
|---|---|
| Iteration 1：ContentProject / ContentType | Portfolio 成员以 `project_id` 关联已有项目，不重新定义项目生命周期。 |
| Iteration 2 / 2.1：WorkflowRun / LLMCallLog / Schedule | 健康与成本汇总读取运行状态、失败记录和模型成本，不直接触发核心工作流。 |
| Iteration 7：PublishJob | 跨项目发布阻塞、待回填数量可作为 Portfolio 风险项。 |
| Iteration 8：MetricRecord | 健康趋势、缺失数据和表现指标从指标中心读取，不重复保存指标事实。 |
| Iteration 9：StrategySuggestion | Portfolio 可展示策略建议摘要和待处理数量，但策略确认 / 忽略 / 执行仍由策略建议接口处理。 |
| Iteration 11：Platform Adapter / Browser Extension | 本迭代不处理平台插件、Adapter 配置或自动采集，只消费其后续回写数据。 |

## 4.6 幂等、审计与错误处理

- `POST /api/v1/portfolios`、`POST /api/v1/portfolios/:id/projects`、成员移除、优先级调整、健康快照重算必须明确是否要求 `Idempotency-Key`；创建和重算类接口建议支持。
- 所有变更类接口必须写入 `operation_log`，包含 Portfolio、项目成员、优先级、权重、健康策略等变更。
- 列表接口必须支持 `page`、`page_size`、`sort`、`order`，并支持按 `status`、`scope_type`、`owner_id` 查询。
- 查询健康 / 成本汇总时必须返回 `date_range`、`calculated_at` 和 `source_refs`，避免前端误判数据实时性。
- 非法项目、无权限项目、重复加入、已禁用 Portfolio、非法优先级必须使用统一错误结构返回，并带 `request_id`。

---

## 5. 前端页面范围

| 页面 / 组件 | 路由建议 | 交互 |
|---|---|---|
| Portfolio 管理 | `/portfolios` | 组合列表、新增组合 |
| Portfolio 详情 | `/portfolios/:portfolioId` | 查看组合项目、状态、汇总信息 |
| 项目优先级 | `/portfolios/:portfolioId/projects` | 加入项目、调整优先级 |
| 健康 / 成本汇总 | `/portfolios/:portfolioId/health` | 查看组合健康和成本汇总 |


## 5.1 前端需求

| 类别 | 要求 |
|---|---|
| 页面范围 | 必须实现第 5 节定义的全部页面 / 组件：Portfolio 管理、Portfolio 详情、项目优先级、健康 / 成本汇总。 |
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
| Portfolio 管理 | `/portfolios` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 组合列表、新增组合；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| Portfolio 详情 | `/portfolios/:portfolioId` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 查看组合项目、状态、汇总信息；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 项目优先级 | `/portfolios/:portfolioId/projects` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 加入项目、调整优先级；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 健康 / 成本汇总 | `/portfolios/:portfolioId/health` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 查看组合健康和成本汇总；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |

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
| POST | `/api/v1/portfolios` | name、description、scope_type、health_policy | portfolio_id | 新增组合 |
| GET | `/api/v1/portfolios` | page、page_size、status、scope_type、owner_id、sort、order | 组合列表 | Portfolio 管理 |
| GET | `/api/v1/portfolios/:id` | id | Portfolio 基本信息、项目数量、健康摘要、成本摘要 | Portfolio 详情 |
| PATCH | `/api/v1/portfolios/:id` | name、description、status、health_policy | 更新后的 Portfolio | 编辑组合 |
| POST | `/api/v1/portfolios/:id/projects` | project_id、priority、weight、role、note | portfolio_project_id | 加入项目 |
| GET | `/api/v1/portfolios/:id/projects` | page、page_size、priority、status、sort、order | Portfolio 项目成员列表 | 项目优先级 |
| PATCH | `/api/v1/portfolios/:id/projects/:projectId/priority` | priority、weight、note | 状态变更、operation_log_id | 调整优先级 |
| DELETE | `/api/v1/portfolios/:id/projects/:projectId` | reason、note | 状态变更、operation_log_id | 移除项目 |
| POST | `/api/v1/portfolios/:id/status-snapshots/recalculate` | date_range、force | snapshot_id 或 job_id | 重新计算健康快照 |
| GET | `/api/v1/portfolios/:id/status-snapshots` | date_range、page、page_size | 快照列表 | 健康历史 |
| GET | `/api/v1/portfolios/:id/health-summary` | date_range、snapshot_id | 健康汇总、risk_summary、source_refs | 健康看板 |
| GET | `/api/v1/portfolios/:id/cost-summary` | date_range、snapshot_id | 成本汇总、by_project、by_model、source_refs | 成本看板 |
| GET | `/api/v1/portfolios/:id/strategy-summary` | status、date_range | 策略建议摘要、待处理数量、source_refs | Portfolio 详情 / 健康看板 |

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

- Portfolio 管理
- Portfolio 详情
- 项目优先级
- 健康 / 成本汇总

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
| 产品评审结论 | 通过。Iteration 10 能在数据闭环后提供跨项目组合管理、优先级排序、健康与成本汇总，符合规模化阶段需求。 |
| 技术评审结论 | 有条件通过。需按本次评审补齐 Portfolio 详情接口、成员移除与权重规则、健康快照来源追踪、策略摘要边界、幂等与审计要求。 |
| 是否允许进入开发 | 允许进入开发。 |
| 评审人 | ChatGPT |
| 评审日期 | 2026-05-28 |
| 是否需要更新蓝图 | 否。当前改动属于蓝图内细化，不改变系统架构边界。 |
| 主要风险 | Portfolio 健康计算口径不清导致数据不可解释；跨项目聚合重复保存 MetricRecord / LLMCallLog 事实数据；优先级被误用为自动策略执行；缺少详情与成员列表接口导致前端只能依赖列表缓存；健康快照无来源引用导致无法追溯。 |
| 调整项 | 补齐初步对齐检查、评审后产品需求、数据模型细化、健康计算规则、成员管理规则、前后迭代依赖、API 清单、幂等审计和最终对齐验证。 |

---

## 12. 最终需求变更说明

| 变更项 | 变更内容 | 原因 |
|---|---|---|
| Portfolio 定位 | 明确 Portfolio 是跨项目管理和观察视角，不替代项目工作区生产流程 | 避免 Portfolio 越权触发 WorkflowRun、AgentRuntime 或策略执行 |
| 成员关系 | 增加项目可加入多个 Portfolio、同一组合内唯一、支持 priority / weight / role / note | 支撑真实多业务线管理与资源排序 |
| 健康快照 | 增加 `portfolio_status_snapshot` 字段、状态枚举、计算来源和 `source_refs` | 保证健康 / 成本汇总可追踪、可解释、可复算 |
| 成本边界 | 明确成本汇总复用 LLMCallLog，不新增独立成本账本 | 避免重复事实数据和口径冲突 |
| 策略边界 | Portfolio 只展示 StrategySuggestion 摘要，不执行确认 / 忽略 / 执行动作 | 遵守人工确认原则，并复用 Iteration 9 状态机 |
| API 清单 | 增加详情、编辑、成员列表、移除项目、快照重算、快照列表、策略摘要接口 | 补齐前端页面联调与详情页数据来源 |
| 幂等审计 | 明确创建、加入成员、优先级调整、移除、快照重算的幂等与 operation_log 要求 | 符合 API 契约和可审计要求 |

---

## 13. 最终对齐验证

| 检查项 | 最终需求描述 | 蓝图定义 | 状态 |
|---|---|---|---|
| Core 内容类型无关 | Portfolio 只引用 ContentProject，不使用 Novel / Book / Chapter 作为核心资源 | Core 层不得写死小说专属概念 | ✅ 对齐 |
| 系统级页面 | Portfolio 管理、详情、优先级、健康 / 成本看板均为系统级管理入口 | 页面体系包含系统级 Portfolio 管理 | ✅ 对齐 |
| Metrics & Strategy 边界 | 健康 / 成本 / 策略只读聚合已有指标、成本和策略建议，不重建事实表 | Metrics & Strategy 负责 MetricRecord、MetricTemplate、StrategySuggestion | ✅ 对齐 |
| 人工节点 | 优先级只用于排序，策略执行仍走 Iteration 9 人工确认 | 审稿、发布、策略执行必须支持人工确认 | ✅ 对齐 |
| Workflow / Agent 边界 | Portfolio 不触发核心生产链路、不绕过 WorkflowRun / AgentTask / LLMCallLog | 核心生产链路由 Workflow Engine 与 Agent Runtime 承载 | ✅ 对齐 |
| n8n 边界 | 本迭代不引入 n8n 编排，后续只可消费外部回写数据 | n8n 只做外围自动化，不承载核心状态 | ✅ 对齐 |
| API 契约 | 最终接口补齐 DTO、分页、错误、幂等、审计和来源引用要求 | 每个迭代必须包含接口输入输出、页面映射和验收标准 | ✅ 对齐 |

结论：最终需求与产品蓝图无背离，本次评审不需要更新 `00-product-blueprint.md`。

