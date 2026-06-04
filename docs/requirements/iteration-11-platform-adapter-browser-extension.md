# Iteration 11：平台适配器与浏览器插件

> 文件定位：本文件是 Iteration 11 的最新需求、技术、前端页面、接口契约和原型映射文档。  
> 蓝图约束：必须遵守 `00-product-blueprint.md`。  
> 技术栈：后端 Go / Golang；前端 Next.js + TypeScript。  
> 更新日期：2026-06-01  
> 评审日期：2026-06-01  
> 评审结论摘要：评审通过。Iteration 11 与产品蓝图无背离，可进入开发；需补齐 Platform Adapter 边界、插件认证与权限、发布任务拉取 / 填充 / 回填状态机、采集日志结构、n8n 回调幂等与审计、与 Iteration 7 发布队列及 Iteration 8 指标回填的衔接说明。  
> 是否需要更新蓝图：否。  
> 重要规则：本项目不再保留单独前端迭代文档；前端需求已整合到本迭代。  
> 原型开发规则：本迭代前端页面必须基于原型页面实现，CSS / JS 均需可用，并接入可点击导航入口。

---

## 1. 迭代目标

建立 Platform Adapter、Chrome 插件和 n8n 外围集成；前端整合平台 Adapter 和外部自动化页面。

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
| Core 内容类型无关 | 使用 `platform_adapter_config`、`plugin_client`、`platform_collect_log`、`external_workflow_provider` 等通用模型 | Core 层不得写死 Novel / Book / Chapter 等小说专属概念 | ✅ 对齐 |
| Browser Extension 边界 | 插件用于平台半自动发布、页面填充、数据采集回填 | 蓝图定义 Browser Extension 负责平台半自动发布、页面填充、数据采集回填 | ✅ 对齐 |
| Platform Adapter 职责 | 初始需求覆盖 Adapter 配置与插件拉取任务，但缺少平台格式转换、任务锁定、回填状态机 | 蓝图定义 Platform Adapter 负责 PublishTarget、PublishJob、平台格式转换、插件协作 | ⚠️ 偏差：已补齐 Adapter 能力边界、任务领取 / 填充 / 回填规则 |
| 人工节点保留 | 插件辅助填充和回填，不直接替代人工确认发布 | 审稿、发布、策略执行必须支持人工确认 | ✅ 对齐 |
| n8n 外围化 | 初始需求包含 n8n 回调，但未明确只能做通知、Webhook、外部同步、告警 | n8n 不承载核心 Agent 编排或核心发布状态机 | ⚠️ 偏差：已补齐 n8n 回调边界、幂等、审计和禁止承载核心编排 |
| 与 Iteration 7 发布队列衔接 | 插件接口拉取待发布任务并回填状态 | Iteration 7 已定义手动发布队列、复制载荷、发布回填与状态机 | ⚠️ 偏差：初始需求未明确复用 `publish_job`，已补齐不得另建发布状态机 |
| 与 Iteration 8 指标回填衔接 | 采集日志用于平台表现数据回填 | Metrics & Strategy 层负责 MetricRecord、MetricTemplate、StrategySuggestion | ⚠️ 偏差：已补齐采集日志到指标录入的关联字段和非强自动入库边界 |
| API 契约 | 初始接口基本符合 `/api/v1`，但插件认证、分页、幂等、统一错误、状态日志不足 | API 契约要求 DTO、统一错误、分页、状态日志、幂等敏感接口和 OpenAPI | ⚠️ 偏差：已补齐插件 Token、Idempotency-Key、列表分页和错误结构要求 |
| 前端原型 | 已要求 Adapter、插件客户端、n8n、采集日志页面按原型实现 | 每个迭代必须包含前端页面、接口、页面-接口映射和验收 | ✅ 对齐 |

---

## 3. 产品需求

- 提供本迭代对应的系统级或项目级操作入口。
- 页面操作必须能映射到明确 API。
- 异步动作必须返回运行记录 ID 或任务 ID。
- 状态变更必须记录操作日志。
- 页面必须支持空状态、加载态、错误态、成功反馈。

### 3.1 评审后补充产品需求

- Platform Adapter 必须作为平台能力配置层，统一管理平台标识、发布模式、页面填充规则、字段映射、采集规则和敏感凭证引用；不得在 Core 业务逻辑中散落平台专属判断。
- Chrome 插件只承担“辅助人工发布”的客户端能力：拉取待发布任务、展示复制 / 填充载荷、在用户确认后执行页面填充、回填外部链接和采集结果；不得绕过审稿、发布队列或人工确认。
- 插件发布任务必须复用 Iteration 7 的 `publish_job` 状态机，不得新增一套与 `publish_job` 不一致的插件发布状态。
- 插件客户端必须支持注册、禁用、密钥轮换、版本记录、最后活跃时间和权限范围；插件密钥只允许在创建 / 轮换时返回明文，其余查询必须脱敏。
- 插件认证应使用短期 `access_token`，服务端必须校验插件状态、版本兼容性、权限范围和可访问项目 / 发布目标。
- 插件拉取发布任务时必须支持锁定 / 领取机制，避免多个插件客户端重复填充同一个发布任务。锁定超时后允许释放或重新领取。
- 插件回填发布结果必须支持幂等；相同 `Idempotency-Key` 重复提交返回相同结果，不得重复写状态变更日志。
- 平台采集日志必须记录采集来源、平台、目标账号、外部 URL、指标原始值、采集时间、解析状态、错误信息和关联 `publish_job_id` / `content_item_id`。
- 平台采集结果在本迭代只完成采集日志和可人工确认的回填入口；是否自动写入 `metric_record` 必须受配置控制，默认不得无人工确认直接污染指标中心。
- n8n 只允许作为外围回调入口，用于外部通知、Webhook、第三方同步、告警和轻量事件接入；不得承载核心发布状态机、Agent 编排或 Workflow Engine。
- Adapter 配置、插件客户端状态变更、插件回填、采集确认、n8n 回调处理均必须记录 `operation_log` 或领域日志，保证可追踪。
- 本迭代不要求覆盖所有真实平台 DOM 自动化细节，但必须提供可扩展的 Adapter 配置结构和至少一个示例平台配置用于联调。

---

## 4. Go 后端技术需求

- platform_adapter_config
- plugin_client
- platform_collect_log
- external_workflow_provider

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
| 数据模型 | 完成本迭代数据模型、迁移、Repository / Store、Service 边界，覆盖 `platform_adapter_config`、`plugin_client`、`platform_collect_log`、`external_workflow_provider`。 |
| API 契约 | 第 6 节接口必须具备 Go request / response DTO、validator 校验、统一响应结构、统一错误结构和 OpenAPI 描述。 |
| 状态与审计 | 状态变更接口必须校验状态流转合法性，并写入 `operation_log`；失败时返回统一错误码和 `request_id`。 |
| 异步执行 | 触发型接口不得阻塞 HTTP 请求，必须返回 `run_id` / `job_id` / 业务记录 ID，并通过运行记录或详情接口查询结果。 |
| 幂等与重试 | 创建运行、触发执行、发布回填、确认类接口按 `api-contract-standard.md` 要求支持 `Idempotency-Key` 或明确说明不需要幂等。 |
| 前端联调支撑 | 后端需提供可用于页面空态、加载态、错误态、成功态联调的数据与错误响应；不得只返回无法渲染的占位数据。 |

## 4.2 数据模型细化

| 模型 | 必要字段 / 约束 | 说明 |
|---|---|---|
| `platform_adapter_config` | `id`、`platform`、`display_name`、`mode`、`target_type`、`field_mapping`、`fill_rules`、`collect_rules`、`credential_ref`、`enabled`、`version`、`created_at`、`updated_at` | 平台适配配置。`mode` 至少区分 `manual_copy`、`extension_fill`、`collect_only`；敏感凭证只能保存引用，不保存明文。 |
| `plugin_client` | `id`、`client_name`、`client_type`、`version`、`api_key_hash`、`status`、`scope`、`last_seen_at`、`revoked_at`、`created_at`、`updated_at` | 插件客户端注册表。密钥只存哈希；`scope` 控制可访问项目、平台和动作。 |
| `plugin_session` | `id`、`plugin_client_id`、`access_token_hash`、`expires_at`、`ip`、`user_agent`、`created_at`、`revoked_at` | 插件短期认证会话。可用 Redis 或数据库实现，但接口契约需明确过期时间。 |
| `plugin_publish_lock` | `id`、`publish_job_id`、`plugin_client_id`、`locked_until`、`status`、`created_at`、`released_at` | 发布任务领取 / 锁定记录，防止多插件重复填充。 |
| `platform_collect_log` | `id`、`project_id`、`platform`、`target_id`、`publish_job_id`、`content_item_id`、`external_url`、`raw_payload`、`parsed_metrics`、`status`、`error`、`collected_at`、`created_at` | 平台采集日志；原始采集数据与解析后指标分开保存。 |
| `external_workflow_provider` | `id`、`provider_type`、`name`、`base_url`、`credential_ref`、`enabled`、`created_at`、`updated_at` | 外围自动化 Provider；与 Iteration 2.1 的外部自动化 Provider 保持同一模型或兼容迁移。 |
| `external_workflow_callback_log` | `id`、`provider_id`、`binding_id`、`event_type`、`idempotency_key`、`payload`、`status`、`error`、`created_at`、`processed_at` | n8n 或外部自动化回调日志；用于幂等、追踪和错误排查。 |

## 4.3 Platform Adapter 配置规则

- `platform` 必须使用稳定枚举或注册表编码，例如 `fanqie`、`weixin_mp`、`zhihu`、`xiaohongshu`，不得在业务代码中使用页面标题或中文名做判断。
- `field_mapping` 用于定义系统字段到平台页面字段的映射，例如标题、正文、标签、封面文案、发布时间。
- `fill_rules` 用于描述插件页面填充策略，例如选择器、等待条件、字段顺序、是否需要人工确认；不得保存用户账号密码明文。
- `collect_rules` 用于描述平台数据采集策略，例如外部 URL 匹配、指标字段、解析方式和采集频率建议。
- Adapter 配置必须版本化。插件拉取任务时应返回所使用的 adapter version，便于回填和问题排查。
- 禁用 Adapter 后，插件不得继续领取该平台的新任务；已领取但未回填的任务需提示用户重新确认或释放锁。

## 4.4 插件认证与权限规则

| 场景 | 要求 |
|---|---|
| 插件注册 | 返回 `plugin_client_id` 和一次性 `api_key`；服务端仅保存 `api_key_hash`。 |
| 插件认证 | `POST /api/v1/plugin/auth` 使用 `api_key` 换取短期 `access_token`、`expires_at` 和 `scope`。 |
| 权限控制 | 插件请求必须校验 `plugin_client.status = active`、Token 未过期、scope 覆盖请求动作。 |
| 密钥轮换 | 必须提供轮换接口或在客户端管理页支持重新生成密钥；旧密钥立即失效。 |
| 禁用客户端 | 禁用后新认证失败，已有 session 可立即失效或在短 TTL 后失效，策略需写入接口说明。 |
| 审计 | 注册、认证失败、密钥轮换、禁用、任务领取、回填都必须写审计或领域日志。 |

## 4.5 插件发布任务协作规则

- 插件拉取任务来源必须是 Iteration 7 的 `publish_job`，且只返回允许插件处理的状态，例如 `queued`、`copied` 或明确允许插件接管的状态。
- `GET /api/v1/plugin/publish-jobs` 必须支持 `project_id`、`platform`、`status`、`page`、`page_size`，并只返回当前插件 scope 内的任务。
- 插件领取任务时应创建或刷新 `plugin_publish_lock`，返回 `lock_id`、`locked_until` 和填充载荷。
- 填充载荷必须绑定 `content_version_id`、`payload_hash`、`adapter_config_id`、`adapter_version`，避免正文后续变更导致不可追溯。
- 页面填充完成不等于发布完成。插件应支持回传 `filled` / `published` / `failed` 等事件，但只有用户确认发布结果后才能将 `publish_job` 推进到已发布状态。
- 插件回填发布结果时必须提交 `external_url`、`published_at`、`note`、`lock_id`、`payload_hash`；服务端校验锁和状态后再更新 `publish_job`。
- 插件回填失败必须记录 `reason`、`retryable` 和平台错误摘要，不得吞掉失败原因。

## 4.6 平台采集与指标回填规则

- 采集日志可以由插件主动提交，也可以由外部自动化回调写入，但都必须统一进入 `platform_collect_log`。
- `parsed_metrics` 至少支持平台原始指标键、标准指标编码、数值、单位、采集时间和置信状态。
- 采集日志默认不直接写入 `metric_record`；只有当项目或 Adapter 配置显式开启自动回填，且指标模板匹配成功时，才允许生成 MetricRecord。
- 自动回填写入 MetricRecord 时必须记录来源 `platform_collect_log_id`，避免重复采集造成重复指标。
- 采集失败、解析失败、指标模板缺失必须可在采集日志页筛选，并提供人工重试或标记忽略入口。
- 本迭代不承诺完成所有平台反爬、登录态维护和复杂 DOM 适配，只要求建立可扩展采集链路和日志闭环。

## 4.7 n8n 外围回调规则

- n8n 回调接口只接受外围事件，例如通知发送结果、外部同步结果、告警事件、第三方采集结果。
- n8n 回调不得直接创建 WorkflowRun、推进核心 Agent 编排、绕过发布状态机或直接修改内容正文。
- 回调必须支持 `Idempotency-Key` 或从 payload 中提取稳定事件 ID 作为幂等键。
- 回调处理应先写入 `external_workflow_callback_log`，再异步处理或同步返回 `accepted`；失败必须保留错误摘要和 `request_id`。
- 回调 payload 必须做 schema 校验，未知事件类型返回统一错误结构，不得静默成功。

## 4.8 与前后迭代依赖

| 依赖来源 | 本迭代使用方式 |
|---|---|
| Iteration 1：ContentProject / ContentType | Adapter、插件权限、采集日志均以项目或内容类型作为边界，避免跨项目数据污染。 |
| Iteration 2.1：external_workflow_provider / binding / call_log | n8n Provider 与 Binding 应复用或兼容 2.1 的外部自动化模型，不重复造概念。 |
| Iteration 5：content_review / content_version | 插件填充载荷必须绑定可发布的审核通过版本。 |
| Iteration 7：publish_target / publish_job / publish_log | 插件拉取、填充、回填复用发布队列和状态机，不另建插件发布主表。 |
| Iteration 8：metric_template / metric_record | 平台采集日志为指标录入提供来源，自动写入需匹配指标模板且可追溯。 |
| Iteration 10：Portfolio | 本迭代不直接操作 Portfolio，但采集和发布状态可被 Portfolio 健康汇总读取。 |
| Iteration 12 / 13：Article Pack / Social Post Pack | Adapter 字段映射需支持不同 Content Pack 的发布字段和平台差异。 |

---

## 5. 前端页面范围

| 页面 / 组件 | 路由建议 | 交互 |
|---|---|---|
| 平台 Adapter 管理 | `/platform-adapters` | Adapter 配置列表、新增平台配置 |
| 插件客户端 | `/plugin-clients` | 插件注册、认证信息、客户端列表 |
| 外部自动化 / n8n | `/external-automation/n8n` | n8n 回调、绑定、调用日志入口 |
| 采集日志 | `/platform-collect-logs` | 平台采集日志列表和详情 |


## 5.1 前端需求

| 类别 | 要求 |
|---|---|
| 页面范围 | 必须实现第 5 节定义的全部页面 / 组件：平台 Adapter 管理、插件客户端、外部自动化 / n8n、采集日志。 |
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
| 平台 Adapter 管理 | `/platform-adapters` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | Adapter 配置列表、新增平台配置；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 插件客户端 | `/plugin-clients` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 插件注册、认证信息、客户端列表；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 外部自动化 / n8n | `/external-automation/n8n` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | n8n 回调、绑定、调用日志入口；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |
| 采集日志 | `/platform-collect-logs` | 按原型渲染页面布局、信息层级、卡片 / 表格 / 表单 / 弹窗 / 状态标签，不允许裸 HTML。 | 平台采集日志列表和详情；按钮、筛选、详情跳转、提交动作必须可执行并有反馈。 |

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
| POST | `/api/v1/platform-adapter-configs` | platform、display_name、mode、target_type、field_mapping、fill_rules、collect_rules、credential_ref | adapter_config_id、version | 新增平台配置 |
| GET | `/api/v1/platform-adapter-configs` | platform、mode、enabled、page、page_size、sort、order | 配置列表、pagination | Adapter 管理 |
| GET | `/api/v1/platform-adapter-configs/:id` | id | Adapter 配置详情、版本、规则摘要 | Adapter 详情 |
| PATCH | `/api/v1/platform-adapter-configs/:id` | display_name、field_mapping、fill_rules、collect_rules、enabled、reason | version、operation_log_id | 编辑 / 启停 Adapter |
| POST | `/api/v1/plugin-clients/register` | client_name、client_type、version、scope | plugin_client_id、api_key、api_key_once | 插件注册 |
| GET | `/api/v1/plugin-clients` | status、client_type、page、page_size | 插件客户端列表、pagination | 插件客户端 |
| PATCH | `/api/v1/plugin-clients/:id` | status、scope、reason | plugin_client_id、status、operation_log_id | 禁用 / 启用插件 |
| POST | `/api/v1/plugin-clients/:id/rotate-key` | reason | plugin_client_id、api_key_once、operation_log_id | 密钥轮换 |
| POST | `/api/v1/plugin/auth` | api_key、client_version | access_token、expires_at、scope | 插件认证 |
| GET | `/api/v1/plugin/publish-jobs` | project_id、platform、status、page、page_size | 插件可处理发布任务、pagination | 插件拉取任务 |
| POST | `/api/v1/plugin/publish-jobs/:id/lock` | lock_ttl_seconds | lock_id、locked_until、payload、adapter_version | 插件领取任务 |
| POST | `/api/v1/plugin/publish-jobs/:id/filled` | lock_id、payload_hash、note | event_id、status | 插件标记已填充 |
| POST | `/api/v1/plugin/publish-jobs/:id/published` | lock_id、external_url、published_at、payload_hash、note | publish_job_id、status、operation_log_id | 插件回填已发布 |
| POST | `/api/v1/plugin/publish-jobs/:id/failed` | lock_id、reason、retryable、platform_error | publish_job_id、status、operation_log_id | 插件回填失败 |
| POST | `/api/v1/platform-collect-logs` | project_id、platform、target_id、publish_job_id、external_url、raw_payload、parsed_metrics、collected_at | collect_log_id、status | 提交采集日志 |
| GET | `/api/v1/platform-collect-logs` | project_id、platform、status、page、page_size、sort、order | 采集日志列表、pagination | 采集日志 |
| GET | `/api/v1/platform-collect-logs/:id` | id | 原始数据、解析指标、关联内容、错误信息 | 采集日志详情 |
| POST | `/api/v1/platform-collect-logs/:id/confirm-metrics` | metric_records、note | metric_record_ids、operation_log_id | 人工确认写入指标 |
| POST | `/api/v1/external-automation/callbacks/n8n/:bindingId` | payload、event_type、idempotency_key | accepted、callback_log_id | n8n 回调 |

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

- 平台 Adapter 管理
- 插件客户端
- 外部自动化 / n8n
- 采集日志

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
- [ ] Adapter 配置支持字段映射、填充规则、采集规则、凭证引用、启停和版本化。
- [ ] 插件客户端支持注册、短期认证、禁用、密钥轮换、scope 权限和最后活跃时间展示。
- [ ] 插件发布任务复用 `publish_job`，支持任务锁定、填充事件、已发布回填、失败回填和幂等处理。
- [ ] 插件回填和 n8n 回调均支持 `Idempotency-Key` 或稳定事件 ID，重复提交不会重复推进状态。
- [ ] 平台采集日志支持列表、详情、失败筛选、原始 payload、解析指标和人工确认写入 MetricRecord。
- [ ] n8n 回调只处理外围事件，不创建核心 WorkflowRun、不推进 Agent 编排、不绕过发布状态机。
- [ ] 本迭代完成后可以支撑下一迭代。

---

## 10. 本迭代明确不做

- 不做超出本迭代页面范围的业务功能。
- 不做未定义接口的隐式前端调用。
- 不做绕过 WorkflowRun / AgentTask / LLMCallLog 的核心生产链路。
- 不做 n8n 核心编排。
- 不做完全无人值守的平台自动发布；插件只做辅助填充、人工确认和结果回填。
- 不在数据库中保存平台账号密码、插件 api_key 明文或其他敏感凭证明文。
- 不承诺覆盖所有平台 DOM 自动化、反爬、登录态维护和复杂采集适配；本迭代只交付可扩展配置、示例平台和日志闭环。
- 不默认把采集结果自动写入 MetricRecord；自动回填必须显式配置且保留来源。

---

## 11. 需求评审结论

### 11.1 完整性评审

| 评审项 | 结论 | 改进建议 |
|---|---|---|
| Adapter 配置 | 初始需求只有配置列表和新增平台配置，缺少字段映射、填充规则、采集规则、版本化与启停规则。 | 新增 4.2、4.3，明确 Adapter 配置结构和版本化要求。 |
| 插件客户端 | 初始需求包含注册和认证，但缺少禁用、密钥轮换、权限范围、短期 Token。 | 新增 4.4，并扩展第 6 节接口。 |
| 发布协作 | 初始需求能拉取任务和回填，但未明确与 Iteration 7 的 `publish_job` 状态机一致。 | 新增 4.5，要求复用发布队列和锁定机制。 |
| 平台采集 | 初始需求有采集日志页面，但缺少采集字段、解析结果和指标回填边界。 | 新增 4.6，区分日志、解析、人工确认和自动回填。 |
| n8n 回调 | 初始需求只有回调接口，未说明幂等、schema 校验、处理日志和边界。 | 新增 4.7，限制 n8n 为外围自动化。 |
| 前端联调 | 页面范围基本完整。 | 保留原型渲染要求，并在验收中强调插件任务、采集日志、错误态和导航入口。 |

### 11.2 合理性评审

- 采用 Platform Adapter + Browser Extension 的方案合理，符合“人工发布保留 + 插件辅助填充”的产品定位。
- 插件直接自动发布到平台风险较高，且与 Iteration 7 的手动发布闭环冲突，因此本迭代限定为辅助填充、人工确认和结果回填。
- 通过 `plugin_publish_lock` 解决多插件重复领取同一发布任务的问题，是保证状态一致性的必要约束。
- 平台采集默认不直接写入 `metric_record`，可以降低错误解析、重复采集和平台口径变化对指标中心的污染风险。
- n8n 回调先落日志再处理，符合外围自动化定位，也方便失败重放和排查。

### 11.3 与已完成迭代一致性评审

| 前置迭代 | 一致性结论 | 说明 |
|---|---|---|
| Iteration 1 | ✅ 一致 | 继续以 ContentProject / ContentType 作为项目和内容类型边界。 |
| Iteration 2 / 2.1 | ✅ 一致 | 不绕过 Workflow Engine；n8n Provider / Binding 与 2.1 外部自动化模型兼容。 |
| Iteration 5 | ✅ 一致 | 插件填充使用审核通过的 content_version，不直接修改正文。 |
| Iteration 7 | ✅ 一致 | 插件发布任务复用 publish_job、publish_log、发布状态机。 |
| Iteration 8 | ✅ 一致 | 采集日志为 MetricRecord 提供来源，但默认需人工确认。 |
| Iteration 10 | ✅ 一致 | 不改变 Portfolio，仅为后续健康汇总提供发布 / 采集状态来源。 |

### 11.4 风险点与处理

| 风险 | 影响 | 处理要求 |
|---|---|---|
| 插件密钥泄露 | 可能导致发布任务和采集数据被非法访问 | api_key 仅一次性返回，服务端只存 hash，支持禁用与轮换。 |
| 多客户端重复回填 | 可能造成发布状态反复跳变或重复日志 | 引入锁定机制与 Idempotency-Key。 |
| 平台 DOM 变化 | 自动填充 / 采集失败 | Adapter 配置版本化，采集失败可追踪并支持重试。 |
| n8n 越界承载核心流程 | 破坏自研 Workflow Engine 边界 | 明确 n8n 不得推进核心 Agent 编排和发布状态机。 |
| 指标污染 | 错误采集直接进入指标中心 | 默认只写采集日志，自动写 MetricRecord 需显式开启并保留来源。 |
| 平台凭证明文存储 | 安全风险 | `credential_ref` 只保存密钥引用，不保存明文凭证。 |

---

## 12. 最终需求变更说明

| 变更项 | 变更内容 | 原因 |
|---|---|---|
| 文档头部 | 增加评审日期、评审结论摘要、是否需要更新蓝图 | 方便后续追踪本轮评审结论。 |
| 初步对齐检查 | 新增 2.1，对 Core、Browser Extension、Platform Adapter、n8n、Iteration 7 / 8 依赖、API 契约做对齐 | 满足评审流程，并暴露初始需求偏差点。 |
| 产品需求 | 新增 3.1，明确插件辅助人工发布、复用发布队列、客户端权限、采集回填和 n8n 边界 | 降低越界实现和状态不一致风险。 |
| 数据模型 | 新增 4.2，细化 Adapter、插件客户端、插件会话、任务锁、采集日志、n8n 回调日志字段 | 初始需求只有表名，缺少可落地字段约束。 |
| Adapter 规则 | 新增 4.3，明确平台枚举、字段映射、填充规则、采集规则和版本化 | 支撑多平台扩展和插件联调。 |
| 插件认证 | 新增 4.4，明确注册、短期 Token、权限、轮换、禁用和审计 | 控制插件访问权限和密钥泄露风险。 |
| 发布协作 | 新增 4.5，明确任务来源、锁定、载荷绑定、填充事件和发布回填 | 与 Iteration 7 发布队列保持一致。 |
| 采集回填 | 新增 4.6，明确采集日志、解析指标、MetricRecord 写入边界和失败处理 | 避免指标污染并支撑 Iteration 8 数据闭环。 |
| n8n 回调 | 新增 4.7，明确外围事件、幂等、schema 校验和日志 | 对齐 n8n 外围化原则。 |
| 迭代依赖 | 新增 4.8，明确与 Iteration 1 / 2.1 / 5 / 7 / 8 / 10 / 12 / 13 的关系 | 防止重复造概念或越界实现。 |
| 接口清单 | 扩展第 6 节接口，补齐 Adapter 详情 / 编辑、插件管理、锁定、填充、失败、采集确认等接口 | 支撑真实前后端联调和验收。 |

---

## 13. 最终对齐验证

| 检查项 | 最终需求定义 | 蓝图定义 | 状态 |
|---|---|---|---|
| Core 内容类型无关 | Core 仅使用 platform、adapter、plugin、publish、collect、external workflow 等通用概念 | Core 不得写死 Novel / Book / Chapter | ✅ 对齐 |
| Browser Extension 职责 | 插件负责半自动发布、页面填充、数据采集回填，不直接替代人工确认 | Browser Extension 负责平台半自动发布、页面填充、数据采集回填 | ✅ 对齐 |
| Platform Adapter 边界 | Adapter 负责 PublishJob 协作、平台格式转换、字段映射、插件协作 | Platform Adapter 负责 PublishTarget、PublishJob、平台格式转换、插件协作 | ✅ 对齐 |
| 人工节点保留 | 插件填充后仍需用户确认发布结果；自动写指标默认关闭或需确认 | 审稿、发布、策略执行必须支持人工确认 | ✅ 对齐 |
| n8n 外围化 | n8n 仅处理外围回调、通知、同步和告警，不承载核心状态机 | n8n 只做通知、Webhook、外部 API 同步、告警 | ✅ 对齐 |
| API 契约 | 使用 `/api/v1`、DTO、统一响应、分页、幂等、operation_log、OpenAPI | API 契约要求统一前缀、响应结构、分页、幂等和 OpenAPI | ✅ 对齐 |
| 前后端一体验收 | Adapter、插件客户端、n8n、采集日志页面与接口、原型、e2e 同步验收 | 每个迭代必须包含前端页面、后端接口、页面-接口映射 | ✅ 对齐 |

结论：最终需求与产品蓝图无背离，不需要更新 `00-product-blueprint.md`。

