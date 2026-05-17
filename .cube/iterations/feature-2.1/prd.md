# Iteration 2.1 PRD：接口契约与调度基线补齐

## 1. 功能概述

本次迭代包含两个交付方向：

1. 补齐 WorkflowSchedule / ProductionPlan 运行时能力，支撑每天生成 5 个 ContentItem，并补齐调度、外部自动化、成本汇总相关 API 与 Web Admin 页面。
2. 修复 Web Admin 历史 E2E 遗留问题，使 Iteration 1 UI E2E 可在长生命周期内存后端上重复运行，不依赖手动重启或清空数据。

本迭代同时追溯补齐 Iteration 0 / 1 / 2 已交付页面的真实前端渲染、统一 Layout、导航入口、状态反馈和原型一致性，但不扩大这些历史迭代的业务边界。

## 2. Functional Requirements

### 调度与生产计划

- FR-001（P0）：用户可以查看 WorkflowSchedule 列表。
  - 输入：项目 ID、启用状态、分页、排序等筛选条件。
  - 输出：调度计划列表、分页信息、每条计划的启用状态、下一次执行时间、关联生产计划摘要。
  - 异常处理：筛选参数无效时展示 `VALIDATION_ERROR`；服务异常时展示统一错误码、错误信息和 `request_id`。

- FR-002（P0）：用户可以创建 WorkflowSchedule 与 ProductionPlan，用于配置每日生成 ContentItem 的生产计划。
  - 输入：项目 ID、cron 表达式、每日生成数量等生产计划配置；每日生成数量默认值为 5，用户可配置。
  - 输出：`schedule_id`、`next_run_at`、每日生成数量和创建结果反馈。
  - 异常处理：必填字段缺失、cron 表达式无效、每日生成数量无效、项目不存在时返回统一错误结构；重复提交需按接口契约支持幂等或明确不可幂等原因。

- FR-003（P0）：用户可以启用、停用调度计划。
  - 输入：调度计划 ID、启用备注或停用原因。
  - 输出：变更前后启用状态、下一次执行时间或操作日志 ID。
  - 异常处理：非法状态流转返回 `CONFLICT`；计划不存在返回 `NOT_FOUND`；失败态展示 `request_id`。

- FR-004（P0）：用户可以对调度计划发起试跑。
  - 输入：调度计划 ID、可选输入覆盖参数。
  - 输出：立即返回 `workflow_run_id` 和初始状态，不等待最终执行完成。
  - 异常处理：计划不存在、输入覆盖不合法、工作流触发失败时返回统一错误结构。

- FR-005（P1）：用户可以查看调度触发记录。
  - 输入：调度计划 ID、分页参数。
  - 输出：触发日志列表、触发时间、状态、关联 WorkflowRun、错误摘要。
  - 异常处理：计划不存在返回 `NOT_FOUND`；空数据展示样式化空态。

### 运行记录与成本汇总

- FR-006（P0）：用户可以查看由调度触发的 WorkflowRun 运行记录。
  - 输入：运行状态、项目 ID、分页等筛选条件。
  - 输出：WorkflowRun 列表、状态、触发来源、步骤概览和详情入口。
  - 异常处理：接口失败时展示统一错误结构，不依赖无法追踪的静态假数据。

- FR-007（P1）：用户可以查看 LLM 调用成本汇总。
  - 输入：项目 ID、起止日期、模型或 Provider 筛选条件。
  - 输出：调用次数、Token 数、成本总计、按模型聚合数据。
  - 异常处理：日期范围无效返回 `VALIDATION_ERROR`；无数据展示空态。

### 外部自动化 / n8n

- FR-008（P0）：用户可以配置外部自动化 Provider。
  - 输入：Provider 类型、基础 URL、Token 等连接信息。
  - 输出：`provider_id` 和配置成功反馈。
  - 异常处理：Token 不得明文展示；连接信息无效或保存失败时展示统一错误结构。

- FR-009（P0）：用户可以配置外部自动化 Binding。
  - 输入：触发事件、Provider ID、Webhook URL。
  - 输出：`binding_id` 和绑定成功反馈。
  - 异常处理：Provider 不存在返回 `NOT_FOUND`；Webhook URL 无效返回 `VALIDATION_ERROR`。

- FR-010（P0）：系统必须明确 n8n 只承载外围自动化，不承载核心工作流编排。
  - 输入：用户配置或触发外部自动化动作。
  - 输出：外部调用记录或绑定结果，不改变 Workflow Engine、AgentTask、LLMCallLog 的核心追踪边界。
  - 异常处理：外部自动化失败返回 `EXTERNAL_AUTOMATION_ERROR` 并保留可追踪调用记录。

### Web Admin 页面与导航

- FR-011（P0）：Web Admin 必须提供生产计划 / 调度管理页面。
  - 输入：用户从导航进入页面，执行筛选、新建、启用、停用、试跑等操作。
  - 输出：按原型渲染的管理台页面、列表、表单、弹窗、状态标签、Toast 或 Alert 反馈。
  - 异常处理：加载失败、提交失败和异步触发失败均展示错误码、错误信息和 `request_id`。

- FR-012（P0）：Web Admin 必须提供外部自动化 / n8n 页面。
  - 输入：用户配置 Provider、Binding 和回调信息。
  - 输出：按原型渲染的配置页面、脱敏 Token 展示、成功或失败反馈。
  - 异常处理：Token 不得明文暴露；错误态样式化展示统一错误结构。

- FR-013（P1）：Web Admin 必须提供成本汇总页面。
  - 输入：项目、日期、模型筛选条件。
  - 输出：成本统计卡片、按模型聚合表格或列表。
  - 异常处理：无数据展示空态；接口失败展示统一错误结构。

- FR-014（P0）：本迭代新增页面必须接入导航体系。
  - 输入：用户从首页、全局侧边栏或相关工作区入口点击进入。
  - 输出：页面可访问、刷新不 404、当前路由有高亮或当前位置提示。
  - 异常处理：不得存在只能手输 URL 访问的核心页面。

### 历史页面渲染追溯补齐

- FR-015（P0）：Iteration 0 / 1 / 2 已交付核心页面必须按原型完成真实渲染。
  - 输入：用户访问首页、健康检查、OpenAPI 入口、配置检查、项目管理、项目详情、项目模板、Prompt 模板、Provider 管理、工作流模板、运行记录、AgentTask、LLM Logs 等页面。
  - 输出：统一 AppLayout、全局样式、导航入口、卡片、表格、表单、按钮、状态标签和主内容区布局。
  - 异常处理：不得出现浏览器默认裸 HTML、导航链接堆叠、页面刷新 404 或错误态缺失 `request_id`。

- FR-016（P0）：页面必须具备可执行 JS 交互。
  - 输入：用户点击导航、筛选、分页、表单提交、弹窗、详情跳转、状态操作。
  - 输出：可见反馈、数据刷新、跳转或错误提示。
  - 异常处理：按钮无反馈、筛选无效、弹窗打不开、路由跳转失败均视为验收不通过。

### Web Admin E2E 遗留问题修复

- FR-017（P0）：Iteration 1 UI E2E 必须可重复运行。
  - 输入：针对长生命周期内存后端重复执行 `iteration1-ui.spec.ts`。
  - 输出：测试不依赖手动重启后端或清空内存状态。
  - 异常处理：不得因既有项目、Prompt 模板、LLM Provider 等残留数据导致冲突或误选对象。

- FR-018（P0）：项目管理 E2E 流程必须使用当前测试运行创建的项目。
  - 输入：测试创建带唯一标识的项目后进入项目详情或执行暂停。
  - 输出：定位并进入当前测试创建的项目，而不是 fallback 项目或历史 seed 项目。
  - 异常处理：如果项目已暂停或不是当前测试创建对象，不得继续执行会导致 `CONFLICT` 的断言路径。

- FR-019（P0）：Prompt 模板和 LLM Provider E2E 测试数据必须每次运行唯一。
  - 输入：每次测试运行生成唯一 projectName、promptCode、providerSecret、providerURL。
  - 输出：创建、查询和断言均使用本次运行数据。
  - 异常处理：不得因固定数据重复造成 201 断言失败、唯一性冲突或误匹配历史记录。

- FR-020（P1）：E2E 请求等待必须避免竞态。
  - 输入：点击会触发网络请求的按钮或导航动作。
  - 输出：先注册 `waitForResponse`，再点击，并等待预期响应。
  - 异常处理：不得因本地响应过快导致测试错过响应并超时。

- FR-021（P1）：空态、Provider 脱敏和 Dashboard 错误注入测试必须稳定。
  - 输入：空态通过显式 fixture 访问；Provider 断言使用列表安全定位；Dashboard 错误路由在导航前注册。
  - 输出：空态断言与数据变更流分离；密钥不明文展示；错误态展示 `INTERNAL_ERROR` 和 `request_id`。
  - 异常处理：不得因多个 provider 元素造成 strict mode violation，不得因延迟加载导致错误断言不稳定。

## 3. Non-Functional Requirements

- NFR-001（P0）：所有 API 必须遵守 `/api/v1` REST JSON、统一响应结构、统一错误结构和 `request_id` 规范。
- NFR-002（P0）：所有列表接口必须支持分页、筛选和排序。
- NFR-003（P0）：所有状态变更必须写入 `operation_log`。
- NFR-004（P0）：异步触发接口不得阻塞 HTTP 请求，必须立即返回运行记录 ID、任务 ID 或业务记录 ID。
- NFR-005（P0）：创建运行、触发执行、确认类或可能重复提交的接口必须支持 `Idempotency-Key`，或在设计阶段明确说明不需要幂等。
- NFR-006（P0）：Core 领域命名不得引入 Novel / Book / Chapter 等小说专属概念。
- NFR-007（P0）：前端页面必须加载全局样式和统一 AppLayout，不得出现裸 HTML。
- NFR-008（P0）：页面失败态必须展示错误码、错误信息和 `request_id`。
- NFR-009（P0）：LLM Provider Token 或密钥类信息必须脱敏展示，不得出现在页面正文、日志或测试输出中。
- NFR-010（P1）：Playwright / E2E 必须覆盖导航进入、主要按钮点击、接口成功渲染和接口失败渲染。

## 4. 验收标准

- AC-001：可以通过 API 获取、创建、启用、停用、试跑 WorkflowSchedule，并查询调度触发记录。
- AC-002：创建调度计划时可以配置每日生成 ContentItem 数量，默认值为 5。
- AC-003：本迭代先完成手动触发和试跑闭环，调度试跑接口立即返回 `workflow_run_id` 和初始状态。
- AC-004：状态变更接口产生可追踪 `operation_log`。
- AC-005：外部自动化 Provider 与 Binding 可配置，Token 脱敏展示，失败返回 `EXTERNAL_AUTOMATION_ERROR`。
- AC-006：LLM 成本汇总页面可按项目、日期、模型展示 calls、tokens、cost 和 by_model 数据。
- AC-007：生产计划 / 调度管理、运行记录、外部自动化 / n8n、成本汇总页面均可从导航进入，刷新不 404。
- AC-008：新增和追溯补齐页面均具备统一 AppLayout、样式化卡片 / 表格 / 表单 / 状态标签，不出现裸 HTML。
- AC-009：页面空态、加载态、错误态、成功反馈均可见；错误态包含错误码、错误信息和 `request_id`。
- AC-010：Iteration 0 / 1 / 2 核心页面均可通过首页或全局导航进入，并保持当前路由高亮或当前位置提示。
- AC-011：`iteration1-ui.spec.ts` 可针对长生命周期内存后端重复执行，不需要手动重启或清空数据。
- AC-012：E2E 项目流只进入本次运行创建的项目，不误入 seed 项目或历史项目。
- AC-013：Prompt 模板、LLM Provider 测试数据每次运行唯一，不与历史运行冲突。
- AC-014：Provider 密钥脱敏断言在存在多个 Provider 时仍稳定，且页面不展示明文 secret。
- AC-015：Dashboard 错误注入测试能稳定等待 mocked 500 响应并断言错误态。
- AC-016：Iteration 2 navigation E2E 在修复 Iteration 1 UI E2E 后仍保持通过。
- AC-017：`npm --prefix apps/web-admin run lint` 通过。
- AC-018：`WEB_BASE_URL=http://127.0.0.1:3000 npm --prefix apps/web-admin run test:ui -- e2e/iteration1-ui.spec.ts` 通过。
- AC-019：`WEB_BASE_URL=http://127.0.0.1:3000 npm --prefix apps/web-admin run test:ui -- e2e/iteration2-navigation.spec.ts` 通过。
- AC-020：Go 后端相关测试和构建命令通过，且接口均有 DTO、校验和 OpenAPI 描述。

## 5. Out of Scope

- 不实现超出本迭代页面范围的业务功能。
- 不新增未定义接口的隐式前端调用。
- 不绕过 WorkflowRun / AgentTask / LLMCallLog 建立不可追踪的核心生产链路。
- 不让 n8n 承载核心 Agent 编排或核心工作流状态。
- 不借历史页面渲染补齐扩大 Iteration 0 / 1 / 2 的业务范围，只补样式、Layout、导航、原型一致性、状态反馈和已定义接口联调。
- 不在 PRD 阶段规定数据库表结构、内部代码分层、具体组件库选型或测试实现细节；这些进入后续设计与测试阶段确认。
- 不在本迭代实现真实 cron 自动触发；本迭代先完成手动触发、试跑和触发记录闭环。

## 6. 依赖说明

- 必须遵守 `docs/requirements/00-product-blueprint.md` 中的 Core 内容类型无关、Workflow Engine 自研、n8n 外围化、Agent 可追踪和前后端一体验收原则。
- 必须遵守 `docs/requirements/api-contract-standard.md` 中的统一 API 契约。
- 前端页面必须参考 `docs/requirements/ai-content-factory-clickable-prototype.html` 的对应原型页面。
- E2E 修复范围来自 `docs/requirements/web-admin-e2e-followups-next-iteration.md`。

## 7. 已确认边界

- 本迭代先做手动触发、试跑和触发记录闭环，不实现真实 cron 自动触发。
- ProductionPlan 的每日生成数量默认值为 5，并允许用户配置。
- Iteration 0 / 1 / 2 历史页面渲染追溯补齐必须全部在本迭代完成。
- E2E 遗留问题允许同步调整页面 fixture 行为，以支撑稳定测试。
