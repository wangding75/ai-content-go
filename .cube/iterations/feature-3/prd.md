# Iteration 3 PRD：Novel Pack 新书规划流程

## 1. 功能概述

本次迭代在 AI Content Factory 的通用 Core 能力之上，交付 Novel Pack 新书规划流程。用户可以在项目工作区进入内容规划页，启动小说新书规划，查看规划运行与候选选题，通过人工确认选题，并继续查看或维护世界观、人物与大纲规划产物。

本迭代完整覆盖后端、前端、OpenAPI、数据库迁移、操作审计、WorkflowRun / StepRun / AgentTask / LLMCallLog 追踪、页面-接口联调、e2e 与基于原型的渲染验收。

原型基准使用当前仓库实际路径：`docs/requirements/ai-content-factory-clickable-prototype.html`。

## 2. Functional Requirements

### 2.1 Novel Pack 规划入口与项目工作区

- **FR-001（P0）项目工作区内容规划入口**  
  用户必须能从项目详情或项目工作区导航进入 `/projects/:projectId/planning`。页面输入为 `projectId`；输出为内容规划页，包含规划启动入口、规划运行状态、规划历史与候选选题入口。若项目不存在、无权限或接口失败，页面必须展示统一错误码、错误信息与 `request_id`。

- **FR-002（P0）页面状态展示**  
  内容规划页必须支持空态、加载态、错误态、成功态。空态提示用户启动新书规划；加载态展示规划运行或数据加载中；错误态展示 `error.code`、`error.message`、`request_id`；成功态展示规划运行、候选选题和可继续操作入口。

- **FR-003（P0）基于原型渲染**  
  内容规划页、候选选题确认弹窗、世界观编辑页、人物管理页和大纲管理页必须参考 `docs/requirements/ai-content-factory-clickable-prototype.html` 的管理台视觉与交互实现。输出页面不得是裸 HTML、空白页、占位页或仅可手输 URL 访问的页面。

### 2.2 启动新书规划

- **FR-004（P0）启动规划运行**  
  用户在内容规划页填写或选择 `genre`、`audience`、`count`、`template_version_id`、`input_override` 后，可以发起新书规划。系统通过 `POST /api/v1/projects/:projectId/novel/planning-runs` 创建规划运行，立即返回 `planning_run_id`、`workflow_run_id`、`status`，不得同步等待完整规划结果。

- **FR-005（P0）启动规划前置校验**  
  系统必须校验项目存在、用户有权限、项目内容类型属于 Novel Pack、`WorkflowTemplateVersion` 已发布、请求字段合法。校验失败时返回统一错误响应；内容类型不匹配或模板版本不可用时不得创建规划运行。

- **FR-006（P0）启动规划幂等**  
  启动规划接口必须支持 `Idempotency-Key`。相同幂等键与相同请求体重复提交时返回同一规划运行结果；相同幂等键但请求体不一致时返回 `IDEMPOTENCY_CONFLICT`。

### 2.3 规划运行与追踪

- **FR-007（P0）规划运行列表**  
  用户可以在内容规划页查看规划运行历史。系统通过 `GET /api/v1/projects/:projectId/novel/planning-runs` 支持 `page`、`page_size`、`sort`、`order`、`status` 查询，输出统一分页结构，包含 `items` 与 `pagination`。

- **FR-008（P0）规划运行详情**  
  用户可以查看单次规划运行详情。系统通过 `GET /api/v1/projects/:projectId/novel/planning-runs/:runId` 输出运行状态、`workflow_run_id`、候选选题、StepRun 摘要、Agent 执行状态与失败原因。若运行不存在或不属于当前项目，返回 `NOT_FOUND` 或 `FORBIDDEN`。

- **FR-009（P0）Workflow 与 Agent 追踪**  
  新书规划必须通过已发布的 `WorkflowTemplateVersion` 触发 `WorkflowRun`，由 StepRun 派生 AgentTask，并记录 LLMCallLog。选题、世界观、人物、大纲生成步骤必须可追踪输入、输出、模型、Token、成本、错误与状态。

- **FR-010（P0）规划结果持久化**  
  系统必须持久化规划运行、规划快照、候选选题、世界观、人物和大纲规划产物。输出资源必须关联 `ContentProject` / `ContentAsset` 或 Novel Pack 扩展资源，不得只保存在不可追踪的临时结果中。

### 2.4 候选选题与人工确认

- **FR-011（P0）候选选题展示**  
  用户可以在内容规划页查看候选选题摘要，并打开候选选题确认弹窗。候选选题包含 `candidate_id`、标题或主题信息、`status`、`score`、`reason`、`snapshot_id`。候选为空时展示空态。

- **FR-012（P0）候选选题确认弹窗**  
  候选选题确认以 `/projects/:projectId/planning` 页面内弹窗为主。`/projects/:projectId/planning/topics` 仅作为可直达和刷新恢复的路由入口，进入后仍应回到同一确认体验。弹窗必须支持查看候选详情、输入确认 `note`、取消、提交和反馈。

- **FR-013（P0）确认候选选题**  
  用户确认候选选题时，系统调用 `POST /api/v1/projects/:projectId/novel/topics/:topicId/confirm`，输入 `note`，输出 `confirmed_topic_id`、`previous_status`、`current_status`、`operation_log_id`。系统必须校验候选选题属于当前项目且尚未确认。

- **FR-014（P0）确认动作幂等与审计**  
  确认候选选题接口必须支持 `Idempotency-Key`。确认成功后候选状态从 `candidate` 流转为 `confirmed`，并写入 `operation_log`，保留确认前后状态、操作者、时间与备注。重复确认或非法状态流转必须返回统一错误响应。

### 2.5 世界观编辑

- **FR-015（P1）查看世界观**  
  用户可以从项目工作区导航进入 `/projects/:projectId/novel/worldview` 查看世界观、禁止项和版本信息。系统通过 `GET /api/v1/projects/:projectId/novel/worldview` 输出当前世界观数据。无世界观时展示空态和创建/编辑入口。

- **FR-016（P1）编辑世界观**  
  用户可以编辑 `worldview`、`forbidden_rules` 并填写 `note`。系统通过 `PATCH /api/v1/projects/:projectId/novel/worldview` 保存变更，输出 `version_id`、`operation_log_id`。保存失败时保留用户输入并展示统一错误信息。

### 2.6 人物管理

- **FR-017（P1）人物列表**  
  用户可以进入 `/projects/:projectId/novel/characters` 查看人物列表。系统通过 `GET /api/v1/projects/:projectId/novel/characters` 支持 `page`、`page_size`、`sort`、`order`、`role` 查询，输出统一分页结构。

- **FR-018（P1）新增人物**  
  用户可以新增人物，输入 `name`、`role`、`profile`、`note`。系统通过 `POST /api/v1/projects/:projectId/novel/characters` 创建人物，输出 `character_id`、`operation_log_id`。字段不合法、项目不存在或无权限时返回统一错误响应。

### 2.7 大纲管理

- **FR-019（P1）大纲列表**  
  用户可以进入 `/projects/:projectId/novel/arcs` 查看弧线大纲和规划结果。系统通过 `GET /api/v1/projects/:projectId/novel/arcs` 支持 `page`、`page_size`、`sort`、`order` 查询，输出统一分页结构。

- **FR-020（P1）规划产物关联**  
  世界观、人物与大纲产物必须能关联到对应项目、规划运行或规划快照。页面展示规划来源时必须来自可追踪的 `planning_run`、`planning_snapshot`、WorkflowRun / StepRun / AgentTask / LLMCallLog 记录。

### 2.8 API、OpenAPI 与统一契约

- **FR-021（P0）统一 API 响应**  
  本迭代所有接口必须使用统一响应结构：成功响应包含 `success: true`、`data`、`error: null`、`request_id`；失败响应包含 `success: false`、`data: null`、`error.code`、`error.message`、`error.details`、`request_id`。

- **FR-022（P0）Go DTO 与校验**  
  本迭代所有接口必须有 Go request / response DTO，并通过 struct tag 与 validator 校验输入。校验失败返回 `VALIDATION_ERROR`，并提供字段级错误详情。

- **FR-023（P0）OpenAPI 交付**  
  本迭代所有接口必须进入 OpenAPI 3.0 文档，包含 `summary`、`description`、`tags`、`operationId`、`parameters`、`requestBody`、`responses`、`security`、`examples`。

- **FR-024（P0）列表分页契约**  
  规划运行、人物、大纲等列表接口必须支持分页、筛选与排序，并返回统一 `items` 与 `pagination` 结构。

### 2.9 数据持久化与边界

- **FR-025（P0）本迭代数据资源**  
  系统必须支持本迭代所需资源：`content_asset`、`planning_run`、`planning_snapshot`、`novel_topic_candidate`、`novel_worldview`、`novel_character`、`novel_arc`。这些资源需要满足页面展示、接口查询、状态变更和审计追踪需求。

- **FR-026（P0）Novel Pack 扩展边界**  
  Novel 专属概念只能出现在 Novel Pack 扩展模型、接口 tag、页面和 workflow input/output 中。Core 层不得新增 `novel_*`、Book、Chapter 等核心资源命名；Core 只感知 `ContentProject`、`ContentAsset`、`WorkflowRun`、`AgentTask` 等通用模型。

- **FR-027（P0）操作日志**  
  候选选题确认、世界观编辑、人物新增等状态变更或人工操作必须写入 `operation_log`。接口输出必须在适用场景返回 `operation_log_id`。

### 2.10 前端交互与联调

- **FR-028（P0）页面-接口绑定**  
  本迭代页面的数据加载、列表分页、筛选排序、表单提交、状态变更、异步触发、详情查看必须绑定第 6 节接口，不得使用不可追踪的临时 mock 替代验收联调。

- **FR-029（P0）导航、刷新与高亮**  
  所有新增页面必须接入导航入口，刷新后可直接访问且不出现 404。当前路由必须有导航高亮或当前位置提示。

- **FR-030（P0）交互反馈**  
  页面按钮、筛选、分页、表单提交、弹窗开关、确认动作、重试、详情跳转必须可点击并有结果反馈。成功时展示 Toast 或等价反馈并刷新关联数据；失败时展示统一错误信息。

## 3. Non-Functional Requirements

- **NFR-001（P0）安全与鉴权**  
  所有 API 遵循 Bearer Token 鉴权约定；无认证返回 `UNAUTHORIZED`，无权限返回 `FORBIDDEN`。不得在 GET 参数、日志或错误信息中泄露凭据。

- **NFR-002（P0）异步响应**  
  启动规划等触发型接口不得阻塞等待完整 Agent 生成结果，必须快速返回运行记录 ID，并通过运行记录或详情接口查询后续状态。

- **NFR-003（P0）可观测性**  
  WorkflowRun、StepRun、AgentTask、LLMCallLog 必须记录规划链路中的输入、输出、模型、Token、成本、错误和状态，支持页面展示运行状态和失败原因。

- **NFR-004（P0）审计性**  
  人工确认和状态变更必须记录操作日志，包含操作对象、操作者、前后状态、备注和时间，便于追溯。

- **NFR-005（P0）兼容统一 API 契约**  
  本迭代 API 必须保持 `/api/v1` REST JSON 风格，成功、失败、分页与错误码遵循项目 API 契约规范。

- **NFR-006（P0）前端渲染质量**  
  新增页面必须加载全局样式与统一 AppLayout，具备管理台视觉层级、导航、卡片、表格、表单、按钮、状态标签和弹窗样式，不得出现浏览器默认裸 HTML。

- **NFR-007（P1）可测试性**  
  后端接口、DTO 校验、状态流转、幂等、操作日志、分页查询、错误响应、前端主要交互和失败态必须可通过自动化测试验证。e2e 至少覆盖导航进入、主要按钮点击、接口成功渲染和接口失败渲染。

## 4. 验收标准

- **AC-001 对应 FR-001 / FR-002 / FR-029**：用户可以从项目详情或项目工作区导航进入 `/projects/:projectId/planning`；刷新后页面可访问，导航高亮或当前位置提示正确；页面具备空态、加载态、错误态、成功态。
- **AC-002 对应 FR-003 / NFR-006**：内容规划、候选确认、世界观、人物、大纲页面按 `docs/requirements/ai-content-factory-clickable-prototype.html` 完成主要布局和交互，不出现裸 HTML、空白页或占位页。
- **AC-003 对应 FR-004 / FR-005 / FR-006 / NFR-002**：启动规划接口校验项目、Novel Pack 类型与已发布模板版本；成功时立即返回 `planning_run_id`、`workflow_run_id`、`status`；支持 `Idempotency-Key`；重复幂等请求行为符合契约。
- **AC-004 对应 FR-007 / FR-024**：规划运行列表接口支持分页、筛选、排序，并返回统一分页结构。
- **AC-005 对应 FR-008 / FR-009 / NFR-003**：规划运行详情可展示 `workflow_run_id`、候选选题、StepRun 摘要、Agent 状态和失败原因，且数据可追踪到 WorkflowRun / StepRun / AgentTask / LLMCallLog。
- **AC-006 对应 FR-010 / FR-025**：规划运行、规划快照、候选选题、世界观、人物、大纲均持久化为可查询资源，不依赖临时不可追踪结果。
- **AC-007 对应 FR-011 / FR-012**：内容规划页可打开候选选题确认弹窗；`/projects/:projectId/planning/topics` 可作为直达和刷新恢复入口；候选为空时展示空态。
- **AC-008 对应 FR-013 / FR-014 / FR-027 / NFR-004**：确认候选选题接口支持 `Idempotency-Key`，成功返回 `confirmed_topic_id`、`previous_status`、`current_status`、`operation_log_id`，并写入操作日志；非法重复确认返回统一错误响应。
- **AC-009 对应 FR-015 / FR-016**：世界观页可查看和编辑世界观、禁止项与版本；保存成功返回 `version_id`、`operation_log_id`；失败时保留输入并展示统一错误。
- **AC-010 对应 FR-017 / FR-018**：人物页支持分页、筛选、排序查看人物列表，并可新增人物；新增成功返回 `character_id`、`operation_log_id`。
- **AC-011 对应 FR-019 / FR-020**：大纲页支持分页、排序查看弧线大纲，且规划来源可关联到项目、规划运行或规划快照。
- **AC-012 对应 FR-021 / FR-022 / FR-023 / NFR-005**：所有接口均有 Go DTO、validator 校验、统一成功/失败响应、OpenAPI 3.0 描述与示例。
- **AC-013 对应 FR-026**：Core 层没有新增 `novel_*`、Book、Chapter 作为核心资源命名；Novel 专属资源只位于 Novel Pack 扩展边界。
- **AC-014 对应 FR-028 / FR-030**：页面列表、详情、表单提交、状态变更、异步触发和失败态均绑定对应 API；主要按钮和交互可点击并有成功或失败反馈。
- **AC-015 对应 NFR-001**：未认证请求返回 `UNAUTHORIZED`；无权限请求返回 `FORBIDDEN`；日志和错误响应不泄露凭据。
- **AC-016 对应 NFR-007**：e2e 或集成测试覆盖从导航入口进入页面、触发启动规划/确认选题等主要动作、展示接口成功结果和统一失败态。

## 5. Out of Scope

- 不做超出本迭代页面范围的业务功能。
- 不做内容单元正文生成闭环；该能力属于后续内容生成迭代。
- 不做审稿、发布、指标、策略建议、Portfolio、浏览器插件等后续迭代能力。
- 不做未定义接口的隐式前端调用。
- 不绕过 WorkflowRun / StepRun / AgentTask / LLMCallLog 保存核心生产链路结果。
- 不把 `novel_*` 模型上移为 Core 通用模型。
- 不使用 Book / Chapter 作为 Core API、Core 表名或 Core DTO 命名。
- 不使用 n8n 承载核心规划编排；n8n 仅可作为外围通知、Webhook、外部 API 同步或告警。
- 不在 PRD 阶段规定具体表结构、字段类型、索引、内部包结构或代码实现方案。

## 6. 依赖说明

- 依赖通用 `ContentProject` / `ContentAsset` 资源能力。
- 依赖已发布的 `WorkflowTemplateVersion`、WorkflowRun / StepRun 执行链路和 AgentTask / LLMCallLog 追踪能力。
- 依赖统一 API 响应、错误码、分页、鉴权、OpenAPI 与操作日志契约。
- 依赖前端项目工作区导航、统一 AppLayout、全局样式、路由刷新能力和错误态展示能力。
- 原型验收依赖 `docs/requirements/ai-content-factory-clickable-prototype.html`。
