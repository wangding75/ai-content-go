# AI Content Factory 产品蓝图 v2.0（Go 技术栈版）

> 文档定位：本蓝图是 AI Content Factory 后续迭代需求、技术设计、前后端联调、原型验收的最高约束。  
> 更新日期：2026-05-14  
> 本次更新：后端开发语言统一调整为 **Go / Golang**；前端需求不再单独成文档，而是整合到对应迭代；每个迭代必须包含页面、接口、页面-接口映射与验收标准。

---

## 1. 产品定位

AI Content Factory 是面向多内容形态的 AI 内容生产、审核、分发、数据反馈和策略优化系统。

Novel Factory 不再作为系统本体，而是 AI Content Factory 的第一个垂直内容包：**Novel Pack**。

```text
AI Content Factory Core
├── Novel Pack：AI 小说工厂
├── Article Pack：公众号 / 知乎 / SEO 文章
├── Short Video Pack：短视频脚本
├── Social Post Pack：小红书 / 微博图文
├── Marketing Pack：营销文案 / 落地页 / 广告文案
└── Course Pack：课程内容
```

---

## 2. 核心原则

| 原则 | 要求 |
|---|---|
| Core 内容类型无关 | Core 层不得写死 Novel / Book / Chapter 等小说专属概念 |
| 内容类型插件化 | Novel / Article / Social Post 通过 Content Pack 扩展 |
| 工作流自研 | 核心生产链路由自研 Workflow Engine 承载 |
| n8n 外围化 | n8n 只做通知、Webhook、外部 API 同步、告警，不承载核心 Agent 编排 |
| Agent 可追踪 | AgentTask、LLMCallLog 必须记录输入、输出、模型、Token、成本、错误 |
| 人工节点保留 | 审稿、发布、策略执行必须支持人工确认 |
| 前后端一体验收 | 每个迭代必须包含前端页面、后端接口、页面-接口映射 |
| Go 后端统一 | 后端服务、领域服务、任务 worker、调度器统一采用 Go |

---

## 3. 系统分层架构

```text
Web Admin / Browser Extension
  ↓
API Gateway / Go HTTP API
  ↓
Content Business Service
  ↓
Workflow Engine
  ↓
Agent Runtime
  ↓
Knowledge Memory
  ↓
LLM Router
  ↓
PostgreSQL / Redis / Object Storage
```

| 层级 | 职责 |
|---|---|
| Web Admin | 系统大盘、项目管理、工作流、Agent、审稿、发布、指标、策略 |
| Browser Extension | 平台半自动发布、页面填充、数据采集回填 |
| Go API Server | REST API、鉴权、DTO 校验、OpenAPI、操作日志 |
| Content Business | ContentProject、ContentType、ContentItem、ContentAsset |
| Workflow Engine | WorkflowTemplate、WorkflowRun、StepRun、Schedule、ProductionPlan |
| Agent Runtime | AgentTask、Prompt 渲染、结构化输出校验 |
| Knowledge Memory | StaticContext、DynamicState、RecentContentWindow、StyleGuide |
| LLM Router | OpenAI-compatible Provider、Fallback、Token、成本、日志 |
| Platform Adapter | PublishTarget、PublishJob、平台格式转换、插件协作 |
| Metrics & Strategy | MetricRecord、MetricTemplate、StrategySuggestion |

---

## 4. Go 技术蓝图

| 模块 | 技术选择 |
|---|---|
| 后端语言 | Go 1.22+ |
| HTTP 框架 | Gin 或 Chi，推荐 Chi：轻量、标准库友好 |
| API 文档 | OpenAPI 3.0，推荐 oapi-codegen / swaggo 二选一 |
| 数据库 | PostgreSQL |
| 数据访问 | sqlc + pgx，或 Ent；推荐 sqlc + pgx 保持 SQL 可控 |
| 迁移工具 | goose 或 golang-migrate，推荐 goose |
| 缓存 / 队列 | Redis |
| 异步任务 | asynq / 自研 worker，首期推荐 asynq |
| Cron 调度 | robfig/cron + 数据库调度表；后续可升级 Temporal |
| 配置 | viper / envconfig，推荐 envconfig 简洁启动校验 |
| 日志 | slog / zap，推荐 slog 起步 |
| 校验 | go-playground/validator |
| 测试 | Go testing + httptest |
| 前端 | Next.js + TypeScript + Ant Design 或 shadcn/ui |
| 插件 | Chrome Extension Manifest V3 |
| 部署 | Docker Compose 起步，后续 Kubernetes |

---

## 5. 推荐工程结构

```text
ai-content-factory/
├── apps/
│   ├── web-admin/                    # Next.js 管理台
│   ├── api-server/                   # Go API Server
│   │   ├── cmd/api/
│   │   ├── internal/
│   │   │   ├── app/
│   │   │   ├── config/
│   │   │   ├── http/
│   │   │   ├── middleware/
│   │   │   ├── modules/
│   │   │   │   ├── content/
│   │   │   │   ├── workflow/
│   │   │   │   ├── agent/
│   │   │   │   ├── llm/
│   │   │   │   ├── memory/
│   │   │   │   ├── publish/
│   │   │   │   ├── metrics/
│   │   │   │   └── strategy/
│   │   │   └── store/
│   │   └── migrations/
│   ├── worker/                       # Go 异步任务 worker
│   └── scheduler/                    # Go 调度器
├── packages/
│   ├── content-packs/
│   │   ├── novel-pack/
│   │   ├── article-pack/
│   │   └── social-post-pack/
│   ├── platform-adapters/
│   └── shared-contracts/
├── browser-extension/
├── docs/
│   ├── product/
│   ├── architecture/
│   └── iterations/
└── scripts/
```

---

## 6. 核心领域模型

| 模型 | 说明 | 替代原小说概念 |
|---|---|---|
| ContentProject | 内容项目 | Book |
| ContentType | 内容类型 | 小说 / 文章 / 图文 |
| ContentItem | 内容单元 | Chapter |
| ContentAsset | 内容资产 | Worldview / 资料库 |
| WorkflowTemplate | 工作流模板 | 小说生成流程 |
| WorkflowTemplateVersion | 工作流模板版本 | 流程版本 |
| WorkflowRun | 工作流执行实例 | 一次生成流程 |
| WorkflowStepRun | 工作流步骤执行 | 单步执行记录 |
| WorkflowSchedule | 工作流调度 | 每日生产计划触发器 |
| ProductionPlan | 项目生产计划 | 每天生成 5 个内容单元 |
| AgentTask | Agent 执行记录 | Agent 任务 |
| PromptTemplate | Prompt 模板 | Agent Prompt |
| LLMCallLog | 模型调用日志 | Token / 成本日志 |
| PublishTarget | 发布目标 | 平台账号 / 栏目 |
| PublishJob | 发布任务 | 发布队列 |
| MetricRecord | 指标记录 | 留存 / 完读等 |
| StrategySuggestion | 策略建议 | keep / optimize / suspend / promote |
| ProjectPortfolio | 项目组合 | 多项目管理 |

---

## 7. 工作流与 n8n 边界

| 能力 | 归属 |
|---|---|
| WorkflowTemplate / WorkflowRun | 自研 Go Workflow Engine |
| AgentTask / LLMCallLog | 自研 Go Agent Runtime |
| 每天生成 5 章 | WorkflowSchedule + ProductionPlan |
| 审稿、发布、策略确认 | 自研业务状态机 + 人工节点 |
| 飞书 / 邮件 / Webhook 通知 | n8n / External Automation |
| 外部 API 数据同步 | n8n 可选，结果必须回写 MetricRecord |
| 核心状态保存 | 必须在 AI Content Factory 内，不落在 n8n execution |

---

## 8. 页面体系

```text
系统级页面
├── 首页 / 系统大盘
├── 项目管理
├── 项目模板管理
├── 工作流模板管理
├── Agent 管理
├── Prompt 模板管理
├── 模型 Provider 管理
├── 生产计划 / 调度管理
├── 运行记录
├── 发布平台 / Adapter
├── 外部自动化 / n8n
├── Portfolio 管理
└── 系统设置

项目工作区
├── 项目概览
├── 项目配置
├── 内容规划
├── 内容生产
├── 审稿中心
├── 发布队列
├── 指标表现
├── 策略建议
└── 运行记录
```

---

## 9. 迭代约束

每个迭代文档必须包含：

1. 产品需求
2. 技术需求
3. Go 后端实现范围
4. 前端页面范围
5. 数据模型
6. 后端接口输入输出
7. 页面-接口映射
8. 原型页面映射
9. 验收标准
10. 明确不做范围

不得再单独生成前端专项迭代文档。
