# Iteration 1：通用内容项目入口 PRD

## 1. 功能概述

本次迭代建立 AI Content Factory 的通用内容项目入口，覆盖系统首页、项目管理、项目详情壳层、项目模板管理、Prompt 模板管理与模型 Provider 管理。用户应能在前端页面查看系统概览、管理内容类型、创建与查看内容项目、维护 Prompt 模板、维护 LLM Provider 配置，并通过统一 API 完成页面操作。

本迭代同时要求 Go 后端交付对应 REST API、DTO 校验、OpenAPI 描述、统一响应结构、分页筛选排序、状态变更操作日志，以及 Provider API Key 的脱敏展示能力。Core 层必须保持内容类型无关，不得引入 Novel / Book / Chapter 等垂直内容包专属核心资源命名。

## 2. Functional Requirements

### 2.1 系统首页 / 系统大盘

- **FR-001（P0）系统大盘摘要**：用户进入首页时，可以看到项目数、待审稿数量、待发布数量、失败任务数量和今日成本。输入为无显式业务参数；输出为大盘摘要数据。加载失败时，页面显示错误码、错误信息和 request_id，并提供重试入口。
- **FR-002（P1）首页状态展示**：首页必须支持加载态、空状态、错误态和成功态。输入为用户访问首页；输出为对应视觉状态。无数据时显示空状态，不应显示为系统错误。

### 2.2 内容类型 / 项目模板管理

- **FR-003（P0）内容类型列表**：用户可以查看内容类型列表，并按启用状态筛选，列表支持分页、排序。输入包括 page、page_size、sort、order、enabled；输出为内容类型分页列表。参数不合法时返回 VALIDATION_ERROR。
- **FR-004（P0）新增内容类型**：用户可以创建新的内容类型，提交 code、name、project_schema。输出为 content_type_id。code 或 name 缺失、project_schema 非法时返回 VALIDATION_ERROR；code 冲突时返回 CONFLICT。
- **FR-005（P0）获取项目动态表单 Schema**：用户新建项目时，系统可以根据内容类型获取 project_schema，用于前端渲染动态表单。输入为内容类型 id；输出为 project_schema。内容类型不存在时返回 NOT_FOUND。
- **FR-006（P1）内容类型页面反馈**：内容类型页面新增成功后显示成功反馈并刷新关联列表；新增失败时显示错误码、错误信息和 request_id。

### 2.3 内容项目管理

- **FR-007（P0）项目列表**：用户可以查看内容项目列表，并按项目状态、内容类型筛选，列表支持分页、排序。输入包括 page、page_size、sort、order、status、content_type；输出为项目分页列表。筛选参数不合法时返回 VALIDATION_ERROR。
- **FR-008（P0）新建项目**：用户可以基于内容类型创建内容项目，提交 name、content_type_id、project_config。输出为 project_id 和初始 status。name 或 content_type_id 缺失、project_config 不符合对应 project_schema 时返回 VALIDATION_ERROR；内容类型不存在时返回 NOT_FOUND。
- **FR-009（P0）项目概览**：用户进入项目详情壳层时，可以查看项目进度、待处理事项和成本概览。输入为项目 id；输出为项目 overview。项目不存在时返回 NOT_FOUND。
- **FR-010（P0）暂停项目**：用户可以暂停项目，必须填写 reason，可选填写 note。输出为项目状态变更结果和 operation_log_id。项目不存在时返回 NOT_FOUND；当前状态不允许暂停时返回 CONFLICT；reason 缺失时返回 VALIDATION_ERROR。
- **FR-011（P0）项目状态变更留痕**：项目暂停等状态变更必须生成 operation_log。输入为状态变更动作及操作者上下文；输出为 operation_log_id。日志写入失败时状态变更不得被报告为成功。
- **FR-012（P1）项目管理页面反馈**：项目列表、新建项目、项目概览和暂停操作必须支持加载态、空状态、错误态、成功反馈。失败时页面显示错误码、错误信息和 request_id。

### 2.4 Prompt 模板管理

- **FR-013（P0）Prompt 模板列表**：用户可以查看 Prompt 模板列表，并按 agent_code 筛选，列表支持分页、排序。输入包括 page、page_size、sort、order、agent_code；输出为 Prompt 模板分页列表。参数不合法时返回 VALIDATION_ERROR。
- **FR-014（P0）新增 Prompt 模板**：用户可以创建 Prompt 模板，提交 code、template、variables。输出为 prompt_template_id。code 或 template 缺失、variables 非法时返回 VALIDATION_ERROR；code 冲突时返回 CONFLICT。
- **FR-015（P1）Prompt 页面反馈**：新增成功后显示成功反馈并刷新列表；失败时显示错误码、错误信息和 request_id。

### 2.5 模型 Provider 管理

- **FR-016（P0）Provider 列表**：用户可以查看 LLM Provider 配置列表，列表支持分页、排序。输入包括 page、page_size、sort、order；输出为 Provider 分页列表。列表中不得展示明文 api_key。
- **FR-017（P0）新增 Provider**：用户可以新增 Provider，提交 provider_type、base_url、api_key。输出为 provider_id 和 api_key_masked。provider_type、base_url 或 api_key 缺失时返回 VALIDATION_ERROR；Provider 配置冲突时返回 CONFLICT。
- **FR-018（P0）API Key 脱敏展示**：任何 Provider 查询或创建响应只能返回 api_key_masked，不得返回明文 api_key。输入为 Provider 查询或创建操作；输出为脱敏后的 key 展示值。
- **FR-019（P1）Provider 页面反馈**：新增成功后显示成功反馈并刷新列表；失败时显示错误码、错误信息和 request_id。

### 2.6 API 契约与前后端映射

- **FR-020（P0）统一 API 响应结构**：所有本迭代 API 必须使用 success、data、error、request_id 的统一响应结构。成功响应 error 为 null；失败响应 data 为 null，并包含 code、message、details。
- **FR-021（P0）统一错误码**：本迭代接口必须使用约定错误码表达失败场景，包括 VALIDATION_ERROR、UNAUTHORIZED、FORBIDDEN、NOT_FOUND、CONFLICT 和 INTERNAL_ERROR。
- **FR-022（P0）OpenAPI 描述**：本迭代所有接口必须进入 OpenAPI，并包含 summary、description、tags、operationId、parameters、requestBody、responses、security、examples。
- **FR-023（P0）Go DTO 与校验**：本迭代所有接口必须定义 Go request / response DTO，并对用户输入执行校验。校验失败时返回 VALIDATION_ERROR。
- **FR-024（P0）页面-接口一致性**：页面列表加载、新增、状态变更、查看详情必须调用本 PRD 对应 API，不得存在未定义接口的隐式调用。
- **FR-025（P1）认证占位**：接口契约必须保留 Bearer token 认证要求。未认证请求应返回 UNAUTHORIZED；无权限请求应返回 FORBIDDEN。

## 3. Non-Functional Requirements

- **安全性**：Provider API Key 不得在任何响应、日志或页面中明文展示；失败信息不得泄露凭证或内部敏感实现细节。
- **兼容性**：API 前缀统一为 `/api/v1`，响应结构必须与项目 API 契约一致，以便 Web Admin 和后续 Browser Extension 复用。
- **可观测性**：状态变更必须记录 operation_log；接口失败响应必须包含 request_id，便于用户反馈和排查。
- **可维护性**：Core 资源命名必须保持内容类型无关，禁止在核心项目、内容类型、内容单元相关命名中使用 Novel / Book / Chapter。
- **前端体验**：本迭代页面必须具备加载态、空状态、错误态、成功反馈；主要按钮必须有可点击反馈。

## 4. 验收标准

- **AC-001（对应 FR-001）**：访问首页时，页面能展示项目数、待审稿、待发布、失败任务和今日成本；接口失败时显示错误码、错误信息和 request_id。
- **AC-002（对应 FR-002）**：首页在加载中、无数据、请求失败和请求成功四类场景下均有明确 UI 状态。
- **AC-003（对应 FR-003）**：内容类型列表支持 page、page_size、sort、order、enabled 参数，并返回 items 与 pagination。
- **AC-004（对应 FR-004）**：提交合法内容类型后返回 content_type_id；缺失必填字段返回 VALIDATION_ERROR；重复 code 返回 CONFLICT。
- **AC-005（对应 FR-005）**：根据有效内容类型 id 能获取 project_schema；无效 id 返回 NOT_FOUND。
- **AC-006（对应 FR-006）**：内容类型新增成功显示成功反馈并刷新列表；失败时展示 request_id。
- **AC-007（对应 FR-007）**：项目列表支持 page、page_size、sort、order、status、content_type，并返回分页结构。
- **AC-008（对应 FR-008）**：基于合法 content_type_id 和 project_config 能创建项目并返回 project_id、status；非法配置返回 VALIDATION_ERROR。
- **AC-009（对应 FR-009）**：项目详情壳层能通过项目 id 加载项目进度、待处理和成本概览；无效 id 返回 NOT_FOUND。
- **AC-010（对应 FR-010）**：暂停项目必须提交 reason；成功后返回状态变更结果和 operation_log_id；状态冲突返回 CONFLICT。
- **AC-011（对应 FR-011）**：每次项目暂停成功后都能查询或验证对应 operation_log 记录存在。
- **AC-012（对应 FR-012）**：项目管理相关页面具备加载态、空状态、错误态和成功反馈。
- **AC-013（对应 FR-013）**：Prompt 模板列表支持分页、排序和 agent_code 筛选。
- **AC-014（对应 FR-014）**：提交合法 Prompt 模板后返回 prompt_template_id；缺失必填字段返回 VALIDATION_ERROR；重复 code 返回 CONFLICT。
- **AC-015（对应 FR-015）**：Prompt 模板新增成功显示成功反馈并刷新列表；失败时展示 request_id。
- **AC-016（对应 FR-016）**：Provider 列表支持分页、排序，且响应中不包含明文 api_key。
- **AC-017（对应 FR-017）**：提交合法 Provider 后返回 provider_id 和 api_key_masked；缺失必填字段返回 VALIDATION_ERROR。
- **AC-018（对应 FR-018）**：Provider 创建响应和列表响应都只出现 api_key_masked，不出现明文 api_key。
- **AC-019（对应 FR-019）**：Provider 新增成功显示成功反馈并刷新列表；失败时展示 request_id。
- **AC-020（对应 FR-020）**：所有本迭代 API 成功和失败响应均符合 success、data、error、request_id 结构。
- **AC-021（对应 FR-021）**：本迭代接口的参数错误、未认证、无权限、资源不存在、冲突和内部错误均使用约定错误码。
- **AC-022（对应 FR-022）**：OpenAPI 文档包含本迭代全部接口及必填描述字段。
- **AC-023（对应 FR-023）**：每个接口都有 Go request / response DTO；非法输入由校验返回 VALIDATION_ERROR。
- **AC-024（对应 FR-024）**：页面主要操作均能映射到本 PRD 中列出的 API；不存在未定义接口调用。
- **AC-025（对应 FR-025）**：接口保留 Bearer token 认证契约；未认证和无权限场景分别返回 UNAUTHORIZED 与 FORBIDDEN。
- **AC-026（全局约束）**：Core 代码和 API 命名中不出现 Novel / Book / Chapter 作为核心资源名。

## 5. Out of Scope

- 不实现超出首页、项目管理、项目详情壳层、项目模板管理、Prompt 模板管理、模型 Provider 管理之外的页面。
- 不实现 Workflow Engine、Agent Runtime、LLM 实际调用、WorkflowRun 生产链路和调度能力。
- 不实现 Novel Pack、Article Pack、Social Post Pack 等具体内容包业务逻辑。
- 不实现 n8n 核心编排；n8n 仍仅作为后续通知、Webhook、外部 API 同步和告警的外围能力。
- 不实现浏览器插件、发布队列、指标看板、策略建议和 Portfolio 管理。
- 不定义本迭代清单之外的隐式前端接口调用。

## 6. 依赖说明

- 依赖 Iteration 0 已建立的 Go API Server、Next.js Web Admin、基础路由、CI 和健康检查基线。
- 依赖 PostgreSQL 数据持久化能力以及迁移工具基线。
- 依赖统一 API 契约：`/api/v1` 前缀、统一响应结构、分页结构、错误码和 OpenAPI 要求。
- 依赖前端原型文件 `docs/requirements/ai-content-factory-clickable-prototype.html` 作为页面视觉和交互参考。
