# Iteration 4 PRD：内容单元生成闭环

## 1. 功能概述

本次迭代实现 Novel Pack 的内容单元生成闭环，支持用户在项目工作区中基于已确认的新书规划资产触发内容生成、批量生成、查看生成运行状态、追踪生成详情、查看生成后的 ContentItem，并对失败的生成运行发起重试。

本迭代范围完整覆盖 Go 后端、前端页面和联调验收：

- Go 后端提供生成运行、内容单元、Novel 章节扩展、异步生成、追踪日志、幂等与错误响应能力。
- 前端实现内容生产、生成运行详情、ContentItem 列表、失败重试 4 个页面，并接入导航、API、状态反馈和 e2e 测试。
- 生成成功后的 `content_item.status` 固定为 `pending_review`，作为 Iteration 5 审稿中心的入口状态。
- 本迭代 P0 不要求验收时真实调用外部 LLM Provider；但必须具备可替换 Provider 接口，并产生可追踪的 AgentTask / LLMCallLog。

## 2. Functional Requirements

### 2.1 生成前置条件与规划资产校验

**FR-001（P0）项目必须完成规划确认后才能生成内容**

- 功能描述：用户只能在项目已具备已确认规划资产时触发内容生成。
- 输入：项目 ID、已确认选题、世界观版本、人物设定、弧线 / 大纲等规划资产引用。
- 输出：校验通过后允许创建生成运行；校验失败时拒绝生成。
- 异常处理：缺少项目、ContentType、已确认规划结果、可用工作流模板版本或 LLM Provider 时，返回统一错误结构，并包含 `request_id`。

**FR-002（P0）生成动作不得绕过 Workflow Engine**

- 功能描述：手动生成、批量生成和失败重试必须关联或创建 WorkflowRun，用户可以通过生成详情追踪运行过程。
- 输入：生成触发请求或重试请求。
- 输出：立即返回 `generation_run_id`、`workflow_run_id` 或对应业务记录 ID。
- 异常处理：工作流不可用或创建失败时，返回统一错误结构，不创建不可追踪的生成任务。

### 2.2 手动生成内容单元

**FR-003（P0）用户可以发起手动生成**

- 功能描述：用户在项目工作区 / 内容生产页中选择已确认规划资产和生成配置，触发一个生成运行。
- 输入：`projectId`、`confirmed_topic_id`、`worldview_version_id`、`arc_id`、`target_count`、`start_sequence_no`、`generation_config`，以及 `Idempotency-Key`。
- 输出：`generation_run_id`、`workflow_run_id`、初始状态。
- 异常处理：输入不合法、幂等冲突、规划资产缺失、无可用工作流或 Provider 时，返回统一错误结构，并在前端展示错误码、错误信息和 `request_id`。

**FR-004（P0）手动生成必须支持幂等**

- 功能描述：用户重复点击或网络重试时，同一 `Idempotency-Key` 和相同请求体不得产生重复内容单元。
- 输入：重复提交的手动生成请求。
- 输出：相同请求返回同一业务结果；请求体不一致时返回冲突错误。
- 异常处理：同一幂等键请求体不一致时返回 `IDEMPOTENCY_CONFLICT`。

### 2.3 批量生成内容单元

**FR-005（P0）用户可以发起批量生成**

- 功能描述：用户在内容生产页中按范围或批量参数触发多个内容单元生成运行。
- 输入：`projectId`、`range`、`batch_size`、`generation_config`，以及 `Idempotency-Key`。
- 输出：`generation_run_ids[]`、`workflow_run_ids[]`、`accepted_count`。
- 异常处理：批量参数非法、无可生成范围、幂等冲突或前置条件缺失时，返回统一错误结构。

**FR-006（P0）批量生成必须异步执行**

- 功能描述：批量生成请求不得阻塞 HTTP 请求等待最终正文生成完成。
- 输入：批量生成请求。
- 输出：立即返回已受理运行记录，用户通过列表或详情查看最终结果。
- 异常处理：部分任务无法受理时，应清晰返回已受理数量和失败原因，不生成无法追踪的后台任务。

### 2.4 生成运行列表与详情追踪

**FR-007（P0）用户可以查看项目生成运行列表**

- 功能描述：用户在内容生产页查看项目下的生成运行历史和当前状态。
- 输入：`projectId`、`status`、`page`、`page_size`、`sort`、`order`。
- 输出：生成运行列表和统一分页信息。
- 异常处理：项目不存在、筛选参数非法或请求失败时，返回统一错误结构，前端展示错误态和 `request_id`。

**FR-008（P0）用户可以查看生成运行详情**

- 功能描述：用户进入生成运行详情页，查看运行状态、步骤摘要、输出 ContentItem、失败原因和追踪信息。
- 输入：`generation_run_id`。
- 输出：`generation_run_id`、`workflow_run_id`、运行状态、步骤摘要、输出 `content_items[]`、错误信息、AgentTask / LLMCallLog 关联追踪信息。
- 异常处理：运行不存在、无权限或详情加载失败时，返回统一错误结构，前端展示错误码、错误信息和 `request_id`。

**FR-009（P0）生成运行状态必须覆盖完整生命周期**

- 功能描述：系统必须展示并维护生成运行状态：`pending`、`running`、`succeeded`、`failed`、`retrying`。
- 输入：生成运行状态查询或状态变化事件。
- 输出：用户可在列表和详情中看到当前状态。
- 异常处理：非法状态流转不得成功，并返回统一错误结构。

### 2.5 ContentItem 列表与详情

**FR-010（P0）用户可以查看项目 ContentItem 列表**

- 功能描述：用户在 ContentItem 列表页查看生成后的内容单元，并按状态分页、筛选、排序。
- 输入：`projectId`、`status`、`page`、`page_size`、`sort`、`order`。
- 输出：ContentItem 列表和统一分页信息。
- 异常处理：项目不存在、筛选参数非法或接口失败时，前端展示统一错误结构中的错误码、错误信息和 `request_id`。

**FR-011（P0）用户可以查看 ContentItem 详情**

- 功能描述：用户可以查看内容正文、扩展字段、版本信息和来源生成运行。
- 输入：`content_item_id`。
- 输出：正文、扩展字段、版本、来源 `generation_run_id`。
- 异常处理：内容单元不存在、无权限或详情加载失败时，返回统一错误结构。

**FR-012（P0）ContentItem 必须具备完整生成状态集合**

- 功能描述：内容单元状态必须至少覆盖 `planned`、`generating`、`generated`、`generation_failed`、`pending_review`，用于支撑生成过程展示、失败处理和后续审稿接入。
- 输入：内容单元创建、生成中、生成成功、生成失败和进入待审的状态变化。
- 输出：用户可在 ContentItem 列表、详情和生成运行详情中看到明确状态。
- 异常处理：非法状态不得被展示为成功结果，应保留失败原因和请求追踪信息。

**FR-013（P0）生成成功后 ContentItem 必须进入待审状态**

- 功能描述：生成成功后的内容单元必须进入 `pending_review` 状态，供下一迭代审稿中心接入。
- 输入：生成成功事件。
- 输出：`content_item.status = pending_review`。
- 异常处理：状态更新失败时，生成运行不得被标记为不可追踪成功，应保留失败原因和请求追踪信息。

**FR-014（P0）Core 层不得引入 Novel 专属核心资源命名**

- 功能描述：通用内容单元使用 ContentItem 表达，Novel 章节相关字段只能作为 Novel Pack 扩展信息出现。
- 输入：ContentItem 创建、列表、详情和生成结果。
- 输出：Core API 与 Core 数据模型不使用 Book / Chapter 作为核心资源名。
- 异常处理：不适用。

### 2.6 失败重试

**FR-015（P0）用户可以对失败生成运行发起重试**

- 功能描述：用户可以从生成运行详情页通过弹窗或抽屉触发重试，也可以直接访问 `/generation-runs/:runId/retry` 路由完成重试。
- 输入：失败的 `generation_run_id`、`reason`、`input_override`，以及 `Idempotency-Key`。
- 输出：`new_generation_run_id`、`workflow_run_id`、`operation_log_id`。
- 异常处理：原运行不存在、原运行非失败状态、幂等冲突或请求参数非法时，返回统一错误结构。

**FR-016（P0）失败重试不得覆盖原失败运行**

- 功能描述：失败重试必须创建新的生成运行，保留原失败运行记录，并建立重试关联。
- 输入：失败重试请求。
- 输出：新的生成运行记录和原失败运行关联关系。
- 异常处理：创建新运行失败时，原失败运行保持不变。

**FR-017（P0）失败重试路由必须可访问并可刷新**

- 功能描述：`/generation-runs/:runId/retry` 必须可直接访问，刷新后不出现 404，并展示导航或当前位置提示。
- 输入：浏览器访问重试路由。
- 输出：可用的失败重试页面或复用 Retry 组件的页面。
- 异常处理：运行不存在或不可重试时，展示样式化错误态和 `request_id`。

### 2.7 AgentTask / LLMCallLog 追踪

**FR-018（P0）生成链路必须产生 AgentTask**

- 功能描述：每次生成运行必须记录 Agent 任务，用户或系统可追踪输入、输出、状态和错误。
- 输入：生成运行中的 Agent 执行过程。
- 输出：可从生成详情追踪到 AgentTask 摘要或关联信息。
- 异常处理：Agent 输出不合法时，生成运行失败并返回可追踪错误。

**FR-019（P0）生成链路必须产生 LLMCallLog**

- 功能描述：每次模型调用必须记录 Provider、模型、输入输出摘要、Token、成本、错误和结构化校验结果。
- 输入：LLM Provider 调用或开发期 Provider / stub 调用。
- 输出：可从生成详情追踪到 LLMCallLog 摘要或关联信息。
- 异常处理：Provider 调用失败或输出校验失败时，记录错误并返回统一错误结构。

**FR-020（P0）验收允许使用开发期 Provider / stub**

- 功能描述：本迭代 P0 验收不强制真实调用外部 LLM Provider，但必须保证生成闭环、日志闭环和错误闭环成立。
- 输入：开发期 Provider / stub 的生成请求。
- 输出：可追踪的生成结果、AgentTask、LLMCallLog 和错误信息。
- 异常处理：stub 返回失败时，系统应按真实 Provider 失败同等展示错误态。

### 2.8 后端 API 与统一契约

**FR-021（P0）后端必须提供本迭代 API 能力**

- 功能描述：后端必须提供手动生成、批量生成、生成运行列表、生成详情、失败重试、ContentItem 列表和 ContentItem 详情能力。
- 输入：各页面和异步动作对应请求。
- 输出：统一响应结构、统一错误结构、分页结构和业务数据。
- 异常处理：所有失败响应必须包含 `error.code`、`error.message`、`error.details` 和 `request_id`。

**FR-022（P0）所有本迭代接口必须具备 Go DTO、入参校验和 OpenAPI 描述**

- 功能描述：接口契约必须可被前后端联调和自动化测试使用。
- 输入：接口请求、Header、Path、Query、Body。
- 输出：明确的 request / response DTO、validator 校验结果和 OpenAPI 文档。
- 异常处理：校验失败返回 `VALIDATION_ERROR`。

**FR-023（P0）状态变更必须记录 operation_log**

- 功能描述：生成运行创建、状态变化、失败重试等关键操作必须记录操作日志。
- 输入：状态变更或动作请求。
- 输出：可追踪的操作日志记录；失败重试返回 `operation_log_id`。
- 异常处理：操作日志写入失败时，不应让状态变化处于不可审计状态。

### 2.9 前端页面与交互

**FR-024（P0）实现项目工作区 / 内容生产页**

- 功能描述：用户可以从项目工作区导航进入内容生产页，执行手动生成、批量生成、查看生成状态。
- 输入：用户导航、筛选、分页、生成表单提交。
- 输出：生成运行列表、生成触发反馈、成功 Toast、错误态和详情跳转。
- 异常处理：空数据展示空态；接口失败展示错误码、错误信息和 `request_id`。

**FR-025（P0）实现生成运行详情页**

- 功能描述：用户可以查看生成运行详情、步骤摘要、输出内容单元、失败原因，并从失败运行触发重试。
- 输入：`runId` 路由参数。
- 输出：运行详情、状态标签、追踪信息、输出 ContentItem、重试入口。
- 异常处理：加载中展示加载态；失败展示样式化错误态。

**FR-026（P0）实现 ContentItem 列表页**

- 功能描述：用户可以从项目工作区导航进入内容单元列表，查看、筛选、分页并进入详情。
- 输入：项目 ID、筛选和分页参数。
- 输出：ContentItem 列表、状态标签、详情入口。
- 异常处理：空列表展示空态；接口失败展示统一错误结构。

**FR-027（P0）实现失败重试页面 / 组件**

- 功能描述：失败重试可以作为详情页弹窗 / 抽屉实现，但必须同时支持 `/generation-runs/:runId/retry` 独立路由访问。
- 输入：失败运行 ID、重试原因、输入覆盖。
- 输出：新生成运行 ID、WorkflowRun ID、操作日志 ID 和成功反馈。
- 异常处理：不可重试时展示错误态，不允许静默失败。

**FR-028（P0）前端页面必须基于原型完成真实渲染**

- 功能描述：4 个页面必须参考 `docs/requirements/ai-content-factory-clickable-prototype.html` 的对应页面布局和交互，不允许裸 HTML 或占位页面。
- 输入：浏览器访问页面。
- 输出：统一 AppLayout、导航、标题区、卡片、表格、表单、弹窗、状态标签和 Toast / Alert。
- 异常处理：CSS 未加载、JS 不可用、页面刷新 404 或核心页面只能手输 URL 均视为验收不通过。

**FR-029（P0）前端必须绑定本迭代 API**

- 功能描述：页面列表、详情、表单提交、重试和异步触发必须绑定本迭代 API。
- 输入：页面加载、筛选、分页、提交、详情跳转、重试。
- 输出：真实 API 响应或可切换到真实 API 的开发期联调响应。
- 异常处理：接口失败必须展示统一错误结构中的 `request_id`。

### 2.10 测试与联调验收

**FR-030（P0）必须覆盖 e2e / 联调验收**

- 功能描述：测试必须覆盖从导航进入页面、主要按钮点击、接口成功渲染、接口失败渲染和错误态展示。
- 输入：Playwright / e2e 或等价联调测试。
- 输出：页面渲染、导航、主要交互、成功态和失败态的测试证据。
- 异常处理：缺少关键页面或关键交互覆盖时，本迭代不得通过验收。

**FR-031（P0）后端必须覆盖关键业务验收**

- 功能描述：后端测试必须覆盖 DTO 校验、幂等冲突、前置依赖缺失、异步受理、状态流转、operation_log 和统一错误结构。
- 输入：后端接口测试、服务测试或集成测试。
- 输出：可证明本迭代后端闭环成立的测试结果。
- 异常处理：测试未覆盖 P0 错误路径时，本迭代不得通过验收。

## 3. Non-Functional Requirements

**NFR-001（P0）统一 API 契约**

所有 API 必须遵守 `/api/v1` REST JSON、统一成功响应、统一失败响应、统一分页结构、统一错误码和 `request_id` 要求。

**NFR-002（P0）异步响应性能**

手动生成、批量生成和失败重试请求不得等待正文最终生成完成；HTTP 请求必须在任务受理后及时返回运行记录 ID。

**NFR-003（P0）幂等安全**

手动生成、批量生成和失败重试必须支持 `Idempotency-Key`，防止重复点击、浏览器重试或网络重放导致重复内容单元。

**NFR-004（P0）可追踪性**

生成运行必须能追踪到 WorkflowRun、步骤摘要、AgentTask、LLMCallLog、输出 ContentItem、失败原因和操作日志。

**NFR-005（P0）前端可用性**

所有新增页面必须具备真实管理台视觉样式、可执行 JS 交互、导航入口、刷新可访问能力，以及空态、加载态、错误态和成功反馈。

**NFR-006（P0）内容类型边界**

Core 层必须保持内容类型无关，不得将 Novel / Book / Chapter 固化为核心资源命名。

**NFR-007（P0）验收稳定性**

外部 LLM 真实调用不作为 P0 阻塞验收；开发期 Provider / stub 必须能够稳定验证生成闭环、日志闭环和错误闭环。

## 4. 验收标准

- [ ] 用户无法在缺少已确认规划资产的项目中触发内容生成，并能看到统一错误结构和 `request_id`。
- [ ] 手动生成接口返回 `generation_run_id`、`workflow_run_id` 和初始状态，且不阻塞等待最终正文。
- [ ] 批量生成接口返回 `generation_run_ids[]`、`workflow_run_ids[]` 和 `accepted_count`，且不产生不可追踪任务。
- [ ] 手动生成、批量生成和失败重试均支持 `Idempotency-Key`。
- [ ] 同一幂等键请求体不一致时返回 `IDEMPOTENCY_CONFLICT`。
- [ ] 生成运行列表支持状态筛选、分页、排序，并返回统一分页结构。
- [ ] 生成运行详情展示 `workflow_run_id`、状态、步骤摘要、输出 ContentItem、失败原因、AgentTask / LLMCallLog 关联信息。
- [ ] 生成运行状态覆盖 `pending`、`running`、`succeeded`、`failed`、`retrying`。
- [ ] ContentItem 列表支持状态筛选、分页、排序，并可进入详情。
- [ ] ContentItem 详情展示正文、扩展字段、版本和来源 `generation_run_id`。
- [ ] ContentItem 状态至少覆盖 `planned`、`generating`、`generated`、`generation_failed`、`pending_review`。
- [ ] 生成成功后 `content_item.status` 固定为 `pending_review`。
- [ ] Core API 与 Core 数据模型不引入 `book` / `chapter` 作为核心资源。
- [ ] Novel 章节相关字段只作为 Novel Pack 扩展信息出现。
- [ ] 失败重试创建新的生成运行，不覆盖原失败运行。
- [ ] 失败重试返回 `new_generation_run_id`、`workflow_run_id`、`operation_log_id`。
- [ ] `/generation-runs/:runId/retry` 可直接访问，刷新后不 404，并展示导航或当前位置提示。
- [ ] 生成链路产生 AgentTask，并能从生成详情追踪。
- [ ] 生成链路产生 LLMCallLog，并能追踪输入、输出、模型、Token、成本、错误和结构化校验结果。
- [ ] 使用开发期 Provider / stub 时，生成闭环、日志闭环和错误闭环仍可验收。
- [ ] 所有本迭代接口均有 Go request / response DTO。
- [ ] 所有本迭代接口均有 validator 校验。
- [ ] 所有本迭代接口均进入 OpenAPI。
- [ ] 状态变更和失败重试写入 `operation_log`。
- [ ] 内容生产页可从项目工作区导航进入，并支持手动生成、批量生成、查看生成状态。
- [ ] 生成运行详情页可展示状态、输出、追踪信息和失败重试入口。
- [ ] ContentItem 列表页可从项目工作区导航进入，并支持筛选、分页、详情跳转。
- [ ] 失败重试可以从详情页弹窗 / 抽屉触发，也可以通过独立路由访问。
- [ ] 4 个前端页面均基于原型完成真实管理台渲染，不出现裸 HTML。
- [ ] 4 个前端页面均具备 CSS 样式、JS 交互、导航高亮或当前位置提示。
- [ ] 4 个前端页面均实现空态、加载态、错误态、成功态。
- [ ] 页面失败态展示错误码、错误信息和 `request_id`。
- [ ] 页面刷新后可直接访问，不出现 404。
- [ ] e2e / 联调测试覆盖从导航入口进入页面、触发主要按钮、接口成功渲染和接口失败渲染。
- [ ] 后端测试覆盖 DTO 校验、幂等冲突、前置依赖缺失、异步受理、状态流转、operation_log 和统一错误结构。
- [ ] 本迭代完成后可以支撑 Iteration 5 审稿中心从 `pending_review` 状态接入。

## 5. Out of Scope

- 不做审稿通过、审稿打回、编辑后通过等审稿状态机能力；这些由 Iteration 5 承接。
- 不做发布、指标、策略建议能力；这些由后续迭代承接。
- 不做 n8n 核心编排；n8n 仍只作为外围自动化。
- 不允许绕过 WorkflowRun / AgentTask / LLMCallLog 直接调用 Agent 生成内容。
- 不把 `novel_chapter_extension` 上升为 Core 模型。
- 不要求 P0 验收时真实调用外部 LLM Provider。
- 不做超出内容生产、生成运行详情、ContentItem 列表、失败重试 4 个页面范围之外的新增业务页面。
- 不做未定义接口的隐式前端调用。

## 6. 依赖说明

- 依赖 Iteration 3 已确认的规划资产，包括 confirmed_topic、worldview、characters、arc / outline 等。
- 依赖既有 API 契约：统一响应结构、统一错误结构、统一分页结构、统一错误码和 `request_id`。
- 依赖 Workflow Engine 能创建并追踪 WorkflowRun 和步骤摘要。
- 依赖 Agent Runtime 和 LLM Provider 抽象产生 AgentTask / LLMCallLog。
- 依赖前端原型 `docs/requirements/ai-content-factory-clickable-prototype.html` 作为页面布局和交互基准。
