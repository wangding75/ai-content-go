# Iteration 1：通用内容项目入口

> 文件定位：本文件是 Iteration 1 的最新需求、技术、前端页面、接口契约和原型映射文档。  
> 蓝图约束：必须遵守 `00-product-blueprint.md`。  
> 技术栈：后端 Go / Golang；前端 Next.js + TypeScript。  
> 更新日期：2026-05-14  
> 重要规则：本项目不再保留单独前端迭代文档；前端需求已整合到本迭代。

---

## 1. 迭代目标

建立项目管理、内容类型、Prompt、LLM Provider 管理能力，并提供系统首页与项目管理页面。

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

---

## 4. Go 后端技术需求

- content_project
- content_type
- content_item
- prompt_template
- llm_provider_config

通用要求：

- 使用 Go struct 定义请求 / 响应 DTO。
- 使用 validator 做入参校验。
- 使用 sqlc + pgx 或等价方式访问 PostgreSQL。
- 使用 goose 或等价工具管理数据库迁移。
- 所有接口进入 OpenAPI。
- 状态变更写入 `operation_log`。
- 异步任务通过 worker / queue 执行，不阻塞 HTTP 请求。

---

## 5. 前端页面范围

| 页面 / 组件 | 路由建议 | 交互 |
|---|---|---|
| 首页 / 系统大盘 | `/` | 查看、筛选、编辑、触发动作、查看详情 |
| 项目管理 | `/` | 查看、筛选、编辑、触发动作、查看详情 |
| 项目详情壳层 | `/` | 查看、筛选、编辑、触发动作、查看详情 |
| 项目模板管理 | `/` | 查看、筛选、编辑、触发动作、查看详情 |
| Prompt 模板管理 | `/prompt` | 查看、筛选、编辑、触发动作、查看详情 |
| 模型 Provider 管理 | `/provider` | 查看、筛选、编辑、触发动作、查看详情 |

---

## 6. 后端接口交付清单

| 方法 | 接口 | 输入 | 输出 | 原型页面映射 |
|---|---|---|---|---|
| GET | `/api/v1/dashboard/summary` | 无 | 项目数、待审稿、待发布、失败任务、今日成本 | 首页卡片 |
| GET | `/api/v1/content-types` | page、enabled | 内容类型列表 | 项目模板管理 |
| POST | `/api/v1/content-types` | code、name、project_schema | content_type_id | 新增项目模板 |
| GET | `/api/v1/content-types/:id/project-schema` | id | project_schema | 新建项目动态表单 |
| GET | `/api/v1/projects` | page、status、content_type | 项目列表 | 项目管理 |
| POST | `/api/v1/projects` | name、content_type_id、project_config | project_id、status | 新建项目 |
| GET | `/api/v1/projects/:id/overview` | id | 项目进度、待处理、成本 | 项目工作区概览 |
| POST | `/api/v1/projects/:id/pause` | reason、note | 状态变更、operation_log_id | 项目暂停 |
| GET | `/api/v1/prompt-templates` | page、agent_code | Prompt 列表 | Prompt 模板管理 |
| POST | `/api/v1/prompt-templates` | code、template、variables | prompt_template_id | 新增 Prompt |
| GET | `/api/v1/llm-providers` | page | Provider 列表 | 模型管理 |
| POST | `/api/v1/llm-providers` | provider_type、base_url、api_key | provider_id、api_key_masked | 新增 Provider |

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

- 首页 / 系统大盘
- 项目管理
- 项目详情壳层
- 项目模板管理
- Prompt 模板管理
- 模型 Provider 管理

---

## 9. 验收标准

- [ ] Go 后端接口均有 DTO、校验和 OpenAPI 描述。
- [ ] 前端页面可以完成本迭代定义的主要交互。
- [ ] 页面上的主要按钮均有可点击反馈。
- [ ] 页面-接口映射已与本文件一致。
- [ ] 列表接口支持分页、筛选、排序。
- [ ] 状态变更接口写入 operation_log。
- [ ] 异步接口返回 run_id / job_id。
- [ ] Core 层没有引入 Novel / Book / Chapter 作为核心资源命名。
- [ ] 本迭代完成后可以支撑下一迭代。

---

## 10. 本迭代明确不做

- 不做超出本迭代页面范围的业务功能。
- 不做未定义接口的隐式前端调用。
- 不做绕过 WorkflowRun / AgentTask / LLMCallLog 的核心生产链路。
- 不做 n8n 核心编排。
