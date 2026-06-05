# PRD：Iteration 12 — Article Pack 内容类型扩展

## 功能概述

本迭代以 Content Pack 插件化方式将 Article（公众号 / 知乎 / SEO 文章）接入 AI Content Factory Core。核心交付包含三个功能点：

1. **Article Pack 注册**：将 Article 注册为 ContentType，绑定默认 WorkflowTemplate、Prompt 配置和指标模板，形成可用于创建项目的内容包。
2. **Article 项目生产**：在项目工作区提供 Article 专属扩展配置和文章生成能力，生成结果写入通用 ContentItem / content_version，并通过 Workflow Engine 和 Agent Runtime 执行与追踪。
3. **Article 指标配置**：为项目启用 / 停用 Article 默认指标模板，为后续指标录入和策略建议提供配置基础。

前端交付三类管理台页面，页面样式和交互必须基于原型实现，并接入现有导航体系。

---

## Functional Requirements

### FR-1 Article Pack 注册

**FR-001（P0）** 注册 Article Pack  
- 输入：schema（项目配置字段、生成输入字段、生成输出字段、指标模板字段）、workflows（引用 Workflow Engine，支持资料整理 / 大纲生成 / 正文生成 / 质检四类步骤）、metrics（Article 默认指标列表，如 views、likes、shares 等）  
- 输出：content_pack_id、content_type_id、registered_workflow_version_ids、metric_template_ids  
- 异常：同一 Idempotency-Key 且输入相同时幂等返回相同结果；输入不一致返回 IDEMPOTENCY_CONFLICT  
- 约束：注册成功后必须完成三类绑定：content_type.code = article、Article 默认工作流模板版本、Article 默认指标模板；注册接口不得直接创建业务项目

**FR-002（P1）** 查看 Article Pack 注册状态  
- 输入：无  
- 输出：注册状态、schema 摘要、默认 workflow、默认 metrics 列表  
- 异常：未注册时返回空状态或 NOT_FOUND

---

### FR-2 Article 项目扩展配置

**FR-003（P0）** 获取项目 Article 扩展配置  
- 输入：projectId  
- 输出：topic_style、audience_profile、seo_config、source_policy、structure_policy、default_workflow_template_version_id、enabled_metric_codes、version  
- 异常：项目不存在返回 NOT_FOUND；项目 ContentType 非 Article 返回 FORBIDDEN

**FR-004（P0）** 更新项目 Article 扩展配置  
- 输入：topic_style、audience_profile、seo_config（SEO 配置）、source_policy（资料来源策略）、structure_policy（文章结构策略）、default_workflow_template_version_id、Idempotency-Key  
- 输出：version_id、operation_log_id  
- 异常：字段校验失败返回 VALIDATION_ERROR；状态不合法返回 CONFLICT  
- 约束：配置变更必须写入 operation_log，返回配置版本号

---

### FR-3 Article 生成运行

**FR-005（P0）** 创建文章生成运行  
- 输入：topic、audience、source_refs（资料或已有内容资产引用）、seo_keywords、outline_required、target_platform、generation_config、Idempotency-Key  
- 输出：generation_run_id、workflow_run_id、初始 status、查询详情入口  
- 异常：项目不存在 / ContentType 非 Article / Article 扩展配置未启用 / 默认工作流版本无效时返回对应错误；同一幂等键输入变化返回 IDEMPOTENCY_CONFLICT  
- 约束：接口只启动异步运行，不同步等待正文；生成运行必须经 Workflow Engine 和 Agent Runtime 执行；AgentTask / LLMCallLog 必须落库

**FR-006（P0）** 获取生成列表  
- 输入：status、topic、target_platform、page、page_size、sort、order  
- 输出：生成运行列表（generation_run_id、workflow_run_id、status、topic、created_at 等）、分页信息  
- 异常：参数不合法返回 VALIDATION_ERROR

**FR-007（P0）** 获取生成详情  
- 输入：id（generation_run_id）  
- 输出：生成状态、Article 快照（title、summary、outline、seo_metadata、source_refs、quality_summary）、content_item_id、content_version_id、workflow_run_id、agent_task_refs  
- 异常：不存在返回 NOT_FOUND  
- 约束：展示正文时优先读取版本化内容，不仅依赖运行结果缓存

**FR-008（P1）** 失败重试  
- 输入：id（失败的 generation_run_id）、reason、input_override（可选）、Idempotency-Key  
- 输出：新的 generation_run_id、workflow_run_id、status  
- 异常：运行非失败状态时返回 CONFLICT；不存在返回 NOT_FOUND  
- 约束：重试原因必须写入 operation_log

**FR-009（P1）** 获取 Article 内容快照  
- 输入：itemId（content_item_id）  
- 输出：title、summary、outline、seo_metadata、source_refs、latest_content_version_id  
- 异常：不存在返回 NOT_FOUND

---

### FR-4 Article 指标配置

**FR-010（P1）** 获取项目 Article 指标配置  
- 输入：platform（可选）、enabled（可选）、page  
- 输出：Article 指标模板列表（metric_code、name、unit、value_type、platform）和项目启用状态  
- 异常：项目不存在返回 NOT_FOUND

**FR-011（P1）** 更新项目 Article 指标配置  
- 输入：enabled_metric_codes、platform_overrides、note、Idempotency-Key  
- 输出：version_id、operation_log_id  
- 异常：指标码不合法返回 VALIDATION_ERROR  
- 约束：变更必须写入 operation_log，返回配置版本号；不直接录入 MetricRecord；支持按平台启用不同指标集合

---

### FR-5 前端页面

**FR-012（P0）** 项目模板管理 / Article Pack 页（路由：`/content-packs/article`）  
- 展示 Article Pack 注册状态、schema 摘要、默认 workflow、默认 metrics  
- 支持注册 / 重新注册 Article Pack、查看注册结果、查看错误详情  
- 接入首页或全局侧边栏导航

**FR-013（P0）** 项目工作区 / Article 内容规划与生产页（路由：`/projects/:projectId/article`）  
- 展示项目 Article 扩展配置、生成输入表单、运行列表、生成详情、关联 ContentItem / content_version、失败重试入口  
- 必须接入项目详情 / 项目工作区菜单  
- 明确区分"生成运行状态"（来自 Article generation / WorkflowRun）和"内容审稿 / 发布状态"（来自 Review / Publish 模块）

**FR-014（P1）** Article 指标配置页（路由：`/projects/:projectId/article/metrics`）  
- 展示默认指标模板、平台差异、项目启用状态、最近配置版本  
- 支持启用 / 停用指标并展示审计反馈  
- 接入项目工作区菜单

---

## Non-Functional Requirements

**性能**  
- 列表接口响应时间 < 500ms（P95）；生成运行为异步，HTTP 接口须在 2s 内返回 run_id  

**幂等与审计**  
- FR-001、FR-004、FR-005、FR-008、FR-011 必须支持 `Idempotency-Key`；所有配置变更、生成触发、重试、取消、指标启停须写入 `operation_log`  

**可追踪性**  
- Article 生成链路的 AgentTask / LLMCallLog 必须落库，支持从 generation_run_id 追溯到具体 agent_task_id  

**安全**  
- 所有接口须 Bearer token 鉴权  
- 跨项目访问必须鉴权，无权限返回 FORBIDDEN  

**前端渲染基线**  
- 页面必须加载统一 AppLayout 和全局样式，不允许出现浏览器默认裸 HTML  
- JS 交互全部可用：导航、筛选、分页、表单提交、弹窗、状态操作  

---

## 验收标准

- [ ] `POST /api/v1/content-packs/article/register` 幂等：相同输入重复提交返回相同结果；输入变化返回 IDEMPOTENCY_CONFLICT
- [ ] Article Pack 注册后可用于创建 ContentType 为 article 的项目
- [ ] Article 扩展配置 PATCH 接口变更写入 operation_log，返回版本号
- [ ] Article 生成接口返回 generation_run_id 和 workflow_run_id，不同步等待正文
- [ ] 生成成功后写入 ContentItem 和 content_version，可通过 content_item_id 查询 Article 快照
- [ ] 生成失败时保留 error_code、error_message、request_id，并可按原始输入重试
- [ ] Article 指标配置 PATCH 接口变更写入 operation_log，返回版本号
- [ ] 全部 11 条后端接口有 Go DTO、validator 校验和 OpenAPI 描述
- [ ] 列表接口支持 page、page_size、sort、order 和按 status / topic / target_platform 筛选
- [ ] 前端三类页面均基于原型完成真实渲染，不出现裸 HTML
- [ ] CSS 已生效：统一 AppLayout、导航、卡片、表格、表单、按钮、状态标签可见
- [ ] JS 已生效：导航、筛选、分页、表单提交、弹窗、详情跳转可点击并有反馈
- [ ] 系统级页面可从首页 / 全局侧边栏进入，项目级页面可从项目工作区菜单进入
- [ ] 核心页面刷新后可直接访问，不出现 404，导航高亮保持
- [ ] 失败态展示统一 error.code、error.message 和 request_id
- [ ] e2e 覆盖导航进入页面、主要按钮点击、接口成功渲染、接口失败渲染
- [ ] Core 层未引入 Article / Novel / Book / Chapter 作为核心资源命名
- [ ] 本迭代完成后可支撑 Iteration 13 Social Post Pack 复用内容包扩展模式

---

## Out of Scope

- 不做独立审稿、独立发布队列、独立指标记录和独立策略建议系统（复用 Iteration 5 / 7 / 8 / 9 / 11）
- 不做绕过 WorkflowRun / AgentTask / LLMCallLog 的核心生产链路
- 不做绕过 ContentItem / content_version 的独立文章内容主表
- 不做外部资料自动抓取 / 爬虫能力
- 不做平台自动发布
- 不做 n8n 核心编排
- 不做超出本迭代页面范围的业务功能
- 不做 Social Post 等其他内容包的专属能力

---

## 依赖说明

| 依赖迭代 | 使用方式 |
|---|---|
| Iteration 1：ContentProject / ContentType / Prompt / Provider | Article Pack 注册为 ContentType；项目创建、Prompt 模板、Provider 均复用 |
| Iteration 2：WorkflowRun / AgentTask / LLMCallLog | Article 生成经 Workflow Engine 与 Agent Runtime 执行，并记录模型调用日志 |
| Iteration 4：ContentItem / generation_run | Article 生成复用通用内容生成运行与 ContentItem |
| Iteration 5：content_review / content_version | 生成后的草稿进入审稿，审核通过后形成可发布版本 |
| Iteration 7：PublishJob | Article 发布任务复用发布队列与手动回填 |
| Iteration 8：MetricTemplate / MetricRecord | 本迭代注册 Article 指标模板；MetricRecord 录入、趋势和缺失提醒复用指标中心 |
| Iteration 9：StrategySuggestion | Article 指标后续可驱动策略建议，本迭代不新增策略规则引擎 |
| Iteration 11：Platform Adapter / Browser Extension | 插件在后续读取已审稿文章辅助发布 / 回填；本迭代只提供稳定内容与配置数据 |
