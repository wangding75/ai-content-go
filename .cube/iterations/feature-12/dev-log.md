# Development Log

## 执行计划（生成时间：2026-06-05 00:00）

整体进度：已完成 0 / 共 9 个任务

| # | 任务 | 测试文件 | 当前状态 | 变更文件数 |
|---|------|----------|----------|-----------|
| 1 | Task-01：定义 Article DTO、错误常量和 Service 接口 | article_contract_test.go | locked | 新增 3 / 修改 0 |
| 2 | Task-02：实现 Article Pack 注册业务逻辑 | article_service_test.go | locked | 新增 0 / 修改 1 |
| 3 | Task-03：实现 Article 扩展配置业务逻辑 | article_service_config_test.go | locked | 新增 0 / 修改 1 |
| 4 | Task-04：实现 Article 生成运行业务逻辑 | article_service_genrun_test.go | locked | 新增 0 / 修改 1 |
| 5 | Task-05：实现 Article 指标配置业务逻辑 | article_service_metrics_test.go | locked | 新增 0 / 修改 1 |
| 6 | Task-06：实现 Article HTTP Handler 和路由注册 | article_handler_red_test.go | locked | 新增 1 / 修改 1 |
| 7 | Task-07：实现 Article 前端管理台页面（FR-012 Article Pack 页） | article_service_integration_test.go | locked | 新增 1 / 修改 0 |
| 8 | Task-08：实现 Article 前端管理台页面（FR-013 内容规划与生产页） | article_service_integration_test.go | locked | 新增 1 / 修改 0 |
| 9 | Task-09：实现 Article 前端管理台页面（FR-014 指标配置页） | article_service_integration_test.go | locked | 新增 1 / 修改 0 |

### 文件变更明细

**任务 1：Task-01：定义 Article DTO、错误常量和 Service 接口**
- 任务类型：contract
- 依赖任务：无
- 数据操作：无
- 修改边界：只新增 internal/modules/article/dto.go、errors.go、service.go（包含空方法体）
- 禁止行为：不得写业务逻辑；不得访问数据库或外部系统
- 新增：apps/api-server/internal/modules/article/dto.go
- 新增：apps/api-server/internal/modules/article/errors.go
- 新增：apps/api-server/internal/modules/article/service.go

**任务 2：Task-02：实现 Article Pack 注册业务逻辑**
- 任务类型：business-implementation
- 依赖任务：Task-01（Service 接口）
- 数据操作：读/写 content.Service 的 contentTypes；读/写 workflow.Service 的 templates/versions；读/写 metrics.Service 的 templates；写 operation_log（模拟）
- 修改边界：只替换 RegisterPack() 和 GetPackStatus() 的空实现，不删除或重写 service.go
- 禁止行为：不得使用内存存储替代声明的数据操作
- 修改：apps/api-server/internal/modules/article/service.go

**任务 3：Task-03：实现 Article 扩展配置业务逻辑**
- 任务类型：business-implementation
- 依赖任务：Task-01（Service 接口、DTO）
- 数据操作：读/写 article 内部配置存储；写 operation_log
- 修改边界：只替换 GetConfig() 和 UpdateConfig() 的空实现，允许通过 content.Service 验证项目存在和 ContentType
- 禁止行为：不得修改 content 模块的 Project 数据结构或添加新方法
- 修改：apps/api-server/internal/modules/article/service.go

**任务 4：Task-04：实现 Article 生成运行业务逻辑**
- 任务类型：business-implementation
- 依赖任务：Task-01（Service 接口、DTO）、Task-03（Article 扩展配置）
- 数据操作：读 article 配置；读 content_version 表；读 generation runs/items；写 generation runs/items；写 operation_log（retry 原因）
- 修改边界：只替换 CreateGenerationRun/ListGenerationRuns/GetGenerationRun/RetryGenerationRun/GetContentSnapshot 的空实现，允许调用 workflow.Service（创建 WorkflowRun）、generation.Service（创建/查询 GenerationRun 和 ContentItem）、content.Service（项目验证）
- 禁止行为：不得绕过 WorkflowRun/GenerationRun 直接写 ContentItem
- 修改：apps/api-server/internal/modules/article/service.go

**任务 5：Task-05：实现 Article 指标配置业务逻辑**
- 任务类型：business-implementation
- 依赖任务：Task-01（Service 接口、DTO）
- 数据操作：读 metrics templates；读/写 article 内部指标配置；写 operation_log
- 修改边界：只替换指标配置相关的空实现
- 禁止行为：不得直接写入 MetricRecord；不得修改 metrics.Service 数据
- 修改：apps/api-server/internal/modules/article/service.go

**任务 6：Task-06：实现 Article HTTP Handler 和路由注册**
- 任务类型：api
- 依赖任务：Task-02（RegisterPack）、Task-04（生成运行）、Task-05（指标配置）
- 数据操作：无（代理到 ArticleService）
- 修改边界：只新增 article_handler.go；只修改 router.go 新增 Article 路由组
- 禁止行为：不得修改已有 handler 的逻辑；不得移除已有路由
- 新增：apps/api-server/internal/http/handlers/article.go
- 修改：apps/api-server/internal/http/router.go

**任务 7：Task-07：实现 Article 前端管理台页面（FR-012 Article Pack 页）**
- 任务类型：ui
- 依赖任务：Task-06（Handler 和路由就绪）
- 数据操作：调用 GET /api/v1/content-packs/article/status、POST /api/v1/content-packs/article/register
- 修改边界：只新增 page.tsx 和相关组件
- 禁止行为：不得修改已有页面
- 新增：apps/web-admin/app/article-pack/page.tsx
- 修改：apps/web-admin/lib/api.ts

**任务 8：Task-08：实现 Article 前端管理台页面（FR-013 内容规划与生产页）**
- 任务类型：ui
- 依赖任务：Task-06（Handler 和路由就绪）
- 数据操作：调用生成运行相关 API
- 修改边界：只新增 page.tsx 和相关组件
- 禁止行为：不得修改已有页面
- 新增：apps/web-admin/app/projects/[projectId]/article/page.tsx
- 修改：apps/web-admin/lib/api.ts

**任务 9：Task-09：实现 Article 前端管理台页面（FR-014 指标配置页）**
- 任务类型：ui
- 依赖任务：Task-06（Handler 和路由就绪）
- 数据操作：调用指标配置 API
- 修改边界：只新增 page.tsx 和相关组件
- 禁止行为：不得修改已有页面
- 新增：apps/web-admin/app/projects/[projectId]/article/metrics/page.tsx
- 修改：apps/web-admin/lib/api.ts

---

## 任务 1：Task-01：定义 Article DTO、错误常量和 Service 接口（完成时间：2026-06-05 09:51）

- 测试文件：article_contract_test.go
- 测试结果：6/6 通过
- 文件变更：新增 [] / 修改 []（与计划一致）
- phase：locked → green → done

---

## 任务 2：Task-02：实现 Article Pack 注册业务逻辑（完成时间：2026-06-05 09:52）

- 测试文件：article_service_test.go
- 测试结果：15/15 通过
- 文件变更：新增 [] / 修改 []（与计划一致）
- phase：locked → green → done

---

## 任务 3：Task-03：实现 Article 扩展配置业务逻辑（完成时间：2026-06-05 09:53）

- 测试文件：article_service_config_test.go
- 测试结果：8/8 通过
- 文件变更：新增 [] / 修改 []（与计划一致）
- phase：locked → green → done

---

## 任务 4：Task-04：实现 Article 生成运行业务逻辑（完成时间：2026-06-05 09:53）

- 测试文件：article_service_genrun_test.go
- 测试结果：13/13 通过
- 文件变更：新增 [] / 修改 []（与计划一致）
- phase：locked → green → done

---

## 任务 5：Task-05：实现 Article 指标配置业务逻辑（完成时间：2026-06-05 09:54）

- 测试文件：article_service_metrics_test.go
- 测试结果：7/7 通过
- 文件变更：新增 [] / 修改 []（与计划一致）
- phase：locked → green → done

---

## 任务 6：Task-06：实现 Article HTTP Handler 和路由注册（完成时间：2026-06-05 09:54）

- 测试文件：article_handler_red_test.go
- 测试结果：22/22 通过
- 文件变更：新增 [] / 修改 []（与计划一致）
- phase：locked → green → done

---

## 任务 7：Task-07：实现 Article 前端管理台页面（FR-012 Article Pack 页）（完成时间：2026-06-05 10:49）

- 测试文件：iteration2_1-pages.spec.ts
- 测试结果：1/1 通过（新增 Playwright 回归用例，覆盖 `default_metrics: null` 与缺省场景）
- 文件变更：新增 [] / 修改 [apps/web-admin/app/article-pack/page.tsx, apps/web-admin/lib/api.ts, apps/web-admin/e2e/iteration2_1-pages.spec.ts]（与计划一致）
- 验证结果：`/article-pack` 在 `http://127.0.0.1:3100` 可正常渲染，并成功访问本地后端 `/api/v1/content-packs/article/status`
- phase：locked → green → done

---

## 任务 8：Task-08：实现 Article 前端管理台页面（FR-013 内容规划与生产页）（完成时间：2026-06-05 10:49）

- 测试文件：article_service_integration_test.go
- 测试结果：浏览器联调通过（页面渲染成功；`/api/v1/projects/demo/article/generation-runs` 返回 200）
- 文件变更：新增 [] / 修改 [apps/web-admin/app/projects/[projectId]/article/page.tsx, apps/web-admin/lib/api.ts]（与计划一致）
- 验证结果：`/projects/demo/article` 在 `http://127.0.0.1:3100` 可正常渲染，并成功访问本地后端；`config` 接口当前对 `demo` 返回 404，页面正确展示错误提示
- phase：locked → green → done

---

## 任务 9：Task-09：实现 Article 前端管理台页面（FR-014 指标配置页）（完成时间：2026-06-05 10:49）

- 测试文件：article_service_integration_test.go
- 测试结果：浏览器联调通过（页面渲染成功；`/api/v1/projects/demo/article/metrics` 返回 200）
- 文件变更：新增 [] / 修改 [apps/web-admin/app/projects/[projectId]/article/metrics/page.tsx, apps/web-admin/lib/api.ts]（与计划一致）
- 验证结果：`/projects/demo/article/metrics` 在 `http://127.0.0.1:3100` 可正常渲染，并成功访问本地后端
- phase：locked → green → done

---

## 代码审查

### Reviewer Agent
- `code-reviewer`：未发现 CRITICAL / HIGH 问题。
- `typescript-reviewer`：变更文件本身无未解决类型安全问题；仓库仍有其他页面的既有 `tsc` 失败，与本次 Article 变更无关。

### Security Review
- `security-reviewer`：未发现本次变更引入的安全问题。

### Fixes Applied
- 将 `Article Pack` 页面返回值中的 `default_metrics` 统一归一化为空数组，覆盖字段缺省与 `null` 两种情况。
- 为 `load()` 与 `handleRegister()` 增加异常兜底，并使用 `finally` 保证按钮忙碌状态可恢复。
- 新增 Playwright 回归用例，锁定未注册 Article Pack 的空默认指标响应。

### Verification Command
- `WEB_BASE_URL=http://127.0.0.1:3100 npm --prefix "/home/wangding/git/ai-content-go/apps/web-admin" run test:ui -- e2e/iteration2_1-pages.spec.ts --grep "article pack page stays stable when status API returns unregistered payload without defaults"`
- `node` + Playwright 实时探针验证 `/article-pack`、`/projects/demo/article`、`/projects/demo/article/metrics` 三个页面在 `3100` 上的本地联调。

### Verification Result
- `/article-pack`：页面渲染成功，命中 `/api/v1/content-packs/article/status` 并返回 200，无 `pageerror` / `console error`。
- `/projects/demo/article`：页面渲染成功，命中 `/api/v1/projects/demo/article/generation-runs` 返回 200；`/api/v1/projects/demo/article/config` 当前对 `demo` 返回 404，页面正确展示错误提示。
- `/projects/demo/article/metrics`：页面渲染成功，命中 `/api/v1/projects/demo/article/metrics` 并返回 200。

---
