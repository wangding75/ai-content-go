# PRD：Iteration 13 — Social Post Pack 内容类型扩展

## 功能概述

本次迭代新增 Social Post Pack，作为 Content Pack 扩展注册进入系统，并交付以下四个功能点：

1. **Social Post Pack 注册与查看**：将 Social Post 内容类型以 Content Pack 方式注册到平台，支持查看 schema、workflows、metrics 配置。
2. **项目 Social 内容生成**：在项目工作区配置 Social 生成策略，触发短内容异步生成，生成多版本候选文案。
3. **多版本文案管理**：查看候选文案列表，人工选择主版本，主版本绑定 content_version 后进入审稿/发布链路。
4. **标签与封面文案生成**：基于指定文案版本异步生成标签和封面文案，并查看资产结果。

前端同步交付上述四个页面，并接入现有导航体系。

---

## Functional Requirements

### F1：Social Post Pack 注册与查看

**FR-001** [P0] 查看 Social Post Pack 配置
- 输入：无（GET 请求）
- 输出：schema 定义、workflows 列表、metrics 列表、current_version
- 异常：Pack 未注册时返回 NOT_FOUND，前端展示空态提示

**FR-002** [P0] 注册 Social Post Pack
- 输入：schema（内容类型定义）、workflows（工作流配置）、metrics（指标定义）、version（版本号）
- 输出：content_pack_id、content_type_id、registered_version
- 幂等：以 `content_pack_code = social_post` 做唯一约束；重复注册同一 schema 版本返回已有结果；schema 版本冲突返回 `CONFLICT`
- 异常：参数缺失或格式错误返回 VALIDATION_ERROR

---

### F2：项目 Social 内容生成

**FR-010** [P0] 查看项目 Social 配置
- 输入：projectId
- 输出：target_platforms、variant_policy（default_variant_count、caption_length_policy）、hashtag_policy、cover_copy_policy、tone_style、forbidden_terms
- 异常：项目不存在返回 NOT_FOUND；配置未初始化时返回空配置默认值

**FR-011** [P0] 更新项目 Social 配置
- 输入：target_platforms、default_variant_count、caption_length_policy、hashtag_policy、cover_copy_policy、tone_style、forbidden_terms
- 输出：version_id、operation_log_id
- 约束：支持 Idempotency-Key；状态变更写入 operation_log；仅保存内容生成策略，不保存平台账号敏感凭证
- 异常：字段校验失败返回 VALIDATION_ERROR

**FR-012** [P0] 触发短内容生成
- 输入：topic（主题描述）、source_content_item_id（可选，引用已有内容单元）、platform（目标平台）、version_count（版本数量，默认 3，最大 10）、tone_style、asset_options（是否同步触发标签/封面文案）
- 输出：generation_run_id、workflow_run_id、status（初始状态 running）
- 约束：
  - 必须先创建或关联通用 ContentItem，再生成多个 social_post_variant
  - 必须创建 WorkflowRun / AgentTask / LLMCallLog，全程可追踪
  - 接口不阻塞 HTTP 请求，异步执行
  - version_count 服务端上限为 10
- 异常：version_count 超限返回 VALIDATION_ERROR；LLM 输出解析失败返回 AGENT_OUTPUT_INVALID 并保留失败任务日志

**FR-013** [P0] 查询生成状态/详情
- 输入：generation_run_id
- 输出：status（running/completed/failed）、content_item_id、variants（文案列表）、error（失败信息）、workflow_run_id
- 前端通过轮询或刷新此接口获取生成结果
- 异常：run_id 不存在返回 NOT_FOUND

---

### F3：多版本文案管理

**FR-020** [P0] 查看多版本文案列表
- 输入：content_item_id、status（过滤）、platform（过滤）、page、page_size
- 输出：文案列表（id、variant_index、platform、title、body、hashtags、cover_copy、tone_style、status、created_at）、pagination（total、page、page_size）
- 约束：支持按状态、平台过滤，支持分页

**FR-021** [P0] 选择主版本
- 输入：content_item_id、variantId、note（选择备注）
- 输出：selected_variant_id、content_version_id（绑定 content_version 进入审稿链路）、operation_log_id
- 约束：
  - 同一 ContentItem 同一时间只能有一个主选版本（status = selected）
  - 选择后当前版本 status 变为 selected，原主选版本变为 rejected/archived
  - 状态变更写入 operation_log，记录操作者、来源状态、目标状态、原因
  - 主选版本绑定 content_version，后续进入审稿/发布不再允许修改
- 异常：variantId 不存在返回 NOT_FOUND；状态流转非法返回 VALIDATION_ERROR

---

### F4：标签与封面文案生成

**FR-030** [P0] 触发标签生成
- 输入：content_item_id、platform、variant_id（关联文案版本）、count（标签数量）、style（风格）
- 输出：generation_run_id、workflow_run_id、status
- 约束：
  - 异步执行，不阻塞 HTTP 请求
  - 必须创建 WorkflowRun / AgentTask / LLMCallLog
  - 结果关联 content_item_id、variant_id、目标平台、generation_run_id
- 异常：LLM 输出解析失败返回 AGENT_OUTPUT_INVALID

**FR-031** [P0] 触发封面文案生成
- 输入：content_item_id、platform、variant_id、style（文案风格）、count（版本数）
- 输出：generation_run_id、workflow_run_id、status
- 约束：同 FR-030

**FR-032** [P1] 查看标签与封面文案资产
- 输入：content_item_id、platform（过滤）、variant_id（过滤）
- 输出：tags（含 tags[]、platform、source_variant_id、generation_run_id、created_at）、cover_copy（含 cover_copy、style、platform、source_variant_id、generation_run_id、created_at）、asset_suggestions、source_runs
- 异常：无资产时返回空列表

---

### F5：数据模型约束

**FR-040** [P0] social_post_extension（项目 Social 配置）
- 字段：id、project_id、target_platforms、default_variant_count、caption_length_policy、hashtag_policy、cover_copy_policy、tone_style、forbidden_terms、config_version、created_at、updated_at
- 约束：每个项目最多一条配置记录；只存内容生成策略，不存平台账号凭证

**FR-041** [P0] social_post_variant（多版本文案）
- 字段：id、content_item_id、generation_run_id、workflow_run_id、variant_index、platform、title、body、hashtags、cover_copy、tone_style、status、content_version_id、selected_at、created_at
- 状态流：generated → selected / rejected / archived
- 约束：重新生成只新增新记录，不覆盖历史版本

**FR-042** [P1] metric_template（Social 默认指标模板）
- 字段：id、content_type_code、platform、metric_code、name、unit、aggregation_type、enabled、created_at
- 约束：只注册指标定义（曝光、点击、点赞、收藏、评论、转发、关注转化），不直接写入 MetricRecord

---

## Non-Functional Requirements

**NFR-001 可追踪性**：所有 LLM 生成动作（短内容、多版本文案、标签、封面文案）必须创建 WorkflowRun / AgentTask / LLMCallLog，全程可通过 run_id 查询。

**NFR-002 幂等性**：注册 Pack、配置更新、触发生成等关键接口必须支持 Idempotency-Key 或有明确幂等策略说明。

**NFR-003 成本控制**：version_count 服务端强制上限为 10，防止单次请求造成不可控 LLM 成本。

**NFR-004 审计**：所有状态变更操作写入 operation_log，至少记录操作者、来源状态、目标状态、原因和关联资源 ID。

**NFR-005 安全隔离**：social_post_extension 不保存平台账号凭证；平台差异通过扩展配置字段表达，不写入 Core 模型。

**NFR-006 前端渲染**：所有页面必须加载全局样式和统一 AppLayout，不允许出现浏览器默认裸 HTML；失败态必须展示 error.code、error.message 和 request_id。

---

## 验收标准

### 后端
- [ ] 3 张数据表（social_post_extension、social_post_variant、metric_template）完成迁移，字段与本 PRD 一致
- [ ] 11 个接口均有 Go DTO、validator 校验、统一响应结构和 OpenAPI 描述
- [ ] 状态变更接口写入 operation_log
- [ ] 异步触发接口返回 run_id / workflow_run_id，不阻塞 HTTP 请求
- [ ] 幂等接口支持 Idempotency-Key
- [ ] Core 层无 Novel / Book / Chapter 等非通用命名侵入
- [ ] LLM 输出解析失败返回 AGENT_OUTPUT_INVALID 并保留失败日志

### 前端
- [ ] 四个页面可从导航入口进入，刷新后直接访问不出现 404
- [ ] CSS 已生效：统一 AppLayout、导航、卡片、表格、表单、按钮、状态标签可见
- [ ] JS 已生效：导航跳转、筛选、分页、表单提交、弹窗、状态操作可点击并有反馈
- [ ] 每个页面均有空态、加载态、错误态、成功态展示
- [ ] 失败态展示 error.code、error.message 和 request_id
- [ ] 页面数据绑定本迭代 API，不使用不可追踪的 mock 替代真实联调
- [ ] E2E 覆盖：导航进入页面、主要按钮点击、接口成功渲染、接口失败渲染

---

## Out of Scope

- 不做图片生成、图片上传、图片素材库
- 不做平台自动发布、平台账号登录、浏览器插件自动填充
- 不做外部平台指标采集（由 Iteration 8 MetricRecord 承接）
- 不做 Platform Adapter / Browser Extension 平台格式转换（由 Iteration 11 承接）
- 不做 n8n 核心编排
- 不做超出本迭代页面范围的业务功能

---

## 依赖说明

| 依赖 | 使用方式 |
|---|---|
| Iteration 1：ContentProject / ContentType | Social Post Pack 注册为内容类型扩展，项目通过 content_type_code = social_post 启用 |
| Iteration 2：WorkflowRun / AgentTask / LLMCallLog | 所有 LLM 生成动作必须落库 |
| Iteration 4：ContentItem / generation_run | 短内容以 ContentItem 为主资源 |
| Iteration 5：content_review / content_version | 人工选择主版本后绑定 content_version 进入审稿 |
| Iteration 7：PublishJob | 审稿通过后的内容进入发布队列（本迭代不直接完成发布） |
| Iteration 8：MetricTemplate / MetricRecord | 本迭代只注册默认指标模板，指标录入由 Iteration 8 承担 |
| Iteration 12：Article Pack | 复用 Content Pack 注册模式 |
