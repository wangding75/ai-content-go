# AI Content Factory 迭代计划总览 v5（Go + 前后端整合版）

> 更新日期：2026-05-14  
> 本次更新：
> - 后端开发语言统一改为 **Go / Golang**。
> - 取消单独前端迭代文档。
> - 前端页面、后端接口、页面-接口映射全部整合到对应迭代文档。
> - 保留 Iteration 2.1 作为接口契约与调度基线补丁迭代。
> - n8n 仅作为外围自动化，不承载核心工作流。

---

## 1. 文档清单

| 迭代 | 文件 | 标题 | 目标 | 主要原型页面 |
|---:|---|---|---|---|
| 0 | `iteration-0-scaffold-requirements.md` | 项目脚手架与基础工程 | 建立 Go 后端工程底座、Next.js 管理台壳层、文档与 CI 基线。 | 系统默认页 / 健康检查页, Swagger / OpenAPI 入口页, 系统配置检查页 |
| 1 | `iteration-1-content-project-entry.md` | 通用内容项目入口 | 建立项目管理、内容类型、Prompt、LLM Provider 管理能力，并提供系统首页与项目管理页面。 | 首页 / 系统大盘, 项目管理, 项目详情壳层 |
| 2 | `iteration-2-multi-agent-architecture.md` | Workflow Engine 与多 Agent 架构 | 建立 Go 自研轻量 Workflow Engine、Agent Runtime、LLMCallLog；WorkflowSchedule/ProductionPlan 在本迭代只做 Contract only。 | 工作流模板管理, Agent 管理, 运行记录 |
| 2.1 | `iteration-2.1-api-and-schedule-baseline.md` | 接口契约与调度基线补齐 | 补齐 WorkflowSchedule / ProductionPlan 运行时，支持每天生成 5 个 ContentItem；明确 n8n 外围集成边界。 | 生产计划 / 调度管理, 运行记录, 外部自动化 / n8n |
| 3 | `iteration-3-novel-planning-workflow.md` | Novel Pack 新书规划流程 | 基于通用 Workflow Engine，实现小说项目的选题、世界观、人物、大纲规划；前端整合到项目工作区的内容规划页。 | 项目工作区 / 内容规划, 候选选题确认弹窗, 世界观编辑 |
| 4 | `iteration-4-content-generation-loop.md` | 内容单元生成闭环 | 实现 Novel 内容单元脚本与正文生成；前端整合到项目工作区的内容生产页。 | 项目工作区 / 内容生产, 生成运行详情, ContentItem 列表 |
| 5 | `iteration-5-review-quality-control.md` | 审稿与质量控制 | 建立通用内容审核流程；前端整合到项目工作区的审稿中心。 | 项目工作区 / 审稿中心, 审稿详情, AI 质检报告 |
| 6 | `iteration-6-knowledge-memory-system.md` | Knowledge Memory 记忆系统 | 建立通用知识与记忆系统；前端整合到项目工作区的记忆与一致性页。 | 项目工作区 / 记忆上下文, 一致性报告, 上下文预览 |
| 7 | `iteration-7-publish-queue-manual.md` | 发布队列与手动发布回填 | 建立 PublishJob 与手动发布回填；前端整合到项目工作区的发布队列。 | 项目工作区 / 发布队列, 发布详情, 复制发布内容 |
| 8 | `iteration-8-metrics-dashboard.md` | 数据录入与指标看板 | 建立 MetricRecord 指标中心；前端整合到项目工作区的指标表现页。 | 项目工作区 / 指标表现, 指标录入, 趋势图 |
| 9 | `iteration-9-strategy-suggestion-loop.md` | 策略建议与单类型业务闭环 | 基于指标生成策略建议；前端整合到项目工作区的策略建议页。 | 项目工作区 / 策略建议, 建议详情, 确认 / 忽略 / 执行 |
| 10 | `iteration-10-portfolio-management.md` | Project Portfolio 多项目管理 | 多项目组合管理；前端整合系统级 Portfolio 页面。 | Portfolio 管理, Portfolio 详情, 项目优先级 |
| 11 | `iteration-11-platform-adapter-browser-extension.md` | 平台适配器与浏览器插件 | 建立 Platform Adapter、Chrome 插件和 n8n 外围集成；前端整合平台 Adapter 和外部自动化页面。 | 平台 Adapter 管理, 插件客户端, 外部自动化 / n8n |
| 12 | `iteration-12-article-pack-extension.md` | Article Pack 内容类型扩展 | 新增 Article Pack；前端整合项目模板与项目工作区文章扩展页面。 | 项目模板管理 / Article Pack, 项目工作区 / Article 内容规划与生产, Article 指标配置 |
| 13 | `iteration-13-social-post-pack-extension.md` | Social Post Pack 内容类型扩展 | 新增 Social Post Pack；前端整合项目模板与项目工作区短内容扩展页面。 | 项目模板管理 / Social Post Pack, 项目工作区 / Social 内容生成, 多版本文案 |

---

## 2. 阶段划分

| 阶段 | 覆盖迭代 | 核心目标 |
|---|---|---|
| 基础平台 | 0-2.1 | Go 工程底座、项目入口、Workflow Engine、Agent Runtime、调度基线 |
| Novel Pack MVP | 3-6 | 小说新书规划、内容生成、审稿、记忆一致性 |
| 数据闭环 | 7-9 | 发布、指标、策略闭环 |
| 规模化 | 10-11 | Portfolio、平台适配器、浏览器插件、n8n 外围集成 |
| 内容类型扩展 | 12-13 | Article Pack、Social Post Pack |

---

## 3. 统一交付规则

每个迭代必须交付：

1. Go 后端数据模型和服务边界
2. REST API 输入输出
3. OpenAPI 文档
4. 前端页面与交互
5. 页面-接口映射
6. 原型页面映射
7. 验收标准

---

## 4. 不再保留的文档

不再生成 `iteration-2.2-web-admin-frontend.md` 或任何单独前端文档。  
前端需求必须放回对应业务迭代中。
