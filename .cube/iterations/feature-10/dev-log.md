# Development Log

## 执行计划（生成时间：2026-06-01 13:00）

| # | 任务 | 测试文件 | 当前状态 | 变更文件数 |
|---|------|----------|----------|-----------|
| 1 | Task-01：Portfolio 后端模块契约 | portfolio_contract_test.go | locked | 修改 6 |
| 2 | Task-02：Portfolio 数据库迁移 | portfolio_sql_contract_test.go | locked | 新增 1 |
| 3 | Task-03：Portfolio HTTP handler 与路由契约 | portfolio_handler_red_test.go | locked | 新增 1 / 修改 1 |
| 4 | Task-04：Portfolio app server 注入契约 | portfolio_server_injection_test.go | locked | 修改 1 |
| 5 | Task-05：Portfolio OpenAPI 契约 | iteration10_portfolio_openapi_contract_red_test.go | locked | 修改 1 |
| 6 | Task-06：Portfolio 前端 API client 契约 | iteration10_portfolio_api_client_contract_red_test.go | locked | 修改 1 |
| 7 | Task-07：Portfolio 前端页面与导航骨架 | iteration10_portfolio_pages_contract_red_test.go | locked | 新增 4 / 修改 1 |

### 文件变更明细

**任务 1：Task-01：Portfolio 后端模块契约**
- 任务类型：contract
- 依赖任务：无
- 数据操作：无
- 修改边界：只新增 Portfolio 模块类型、接口、构造函数和最小可编译占位实现
- 禁止行为：不得实现完整业务逻辑；不得调用 WorkflowRun、AgentRuntime、ContentItem 创建或策略建议执行接口
- 修改：apps/api-server/internal/modules/portfolio/dto.go
- 修改：apps/api-server/internal/modules/portfolio/errors.go
- 修改：apps/api-server/internal/modules/portfolio/store.go
- 修改：apps/api-server/internal/modules/portfolio/memory_store.go
- 修改：apps/api-server/internal/modules/portfolio/postgres_store.go
- 修改：apps/api-server/internal/modules/portfolio/service.go

**任务 2：Task-02：Portfolio 数据库迁移**
- 任务类型：migration
- 依赖任务：Task-01
- 数据操作：写 project_portfolio 表；写 portfolio_project 表；写 portfolio_status_snapshot 表
- 修改边界：只新增 00012_create_portfolio_tables.sql
- 禁止行为：不得修改既有迁移文件；不得修改 docs/requirements/
- 修改：apps/api-server/migrations/00012_create_portfolio_tables.sql

**任务 3：Task-03：Portfolio HTTP handler 与路由契约**
- 任务类型：contract
- 依赖任务：Task-01
- 数据操作：无
- 修改边界：只新增 handler 并最小修改 router.go
- 禁止行为：不得改变既有路由语义；不得跳过 bearer auth；不得同步执行多项目快照聚合
- 修改：apps/api-server/internal/http/handlers/portfolio.go
- 修改：apps/api-server/internal/http/router.go

**任务 4：Task-04：Portfolio app server 注入契约**
- 任务类型：contract
- 依赖任务：Task-01
- 数据操作：连接 PostgreSQL 句柄复用
- 修改边界：只修改 server.go 的依赖注入
- 禁止行为：不得新增独立数据库连接生命周期；不得影响 metrics store 注入
- 修改：apps/api-server/internal/app/server.go

**任务 5：Task-05：Portfolio OpenAPI 契约**
- 任务类型：contract
- 依赖任务：Task-03
- 数据操作：无
- 修改边界：只修改 openapi/openapi.yaml
- 禁止行为：不得移除或重命名既有 API path
- 修改：openapi/openapi.yaml

**任务 6：Task-06：Portfolio 前端 API client 契约**
- 任务类型：contract
- 依赖任务：Task-03、Task-05
- 数据操作：无
- 修改边界：只追加 Portfolio 类型和函数
- 禁止行为：不得绕过统一 request()；不得硬编码 secret 或 token
- 修改：apps/web-admin/lib/api.ts

**任务 7：Task-07：Portfolio 前端页面与导航骨架**
- 任务类型：contract
- 依赖任务：Task-06
- 数据操作：UI renders cards/table with loading, empty, error states
- 修改边界：只新增 Portfolio 页面树并最小修改 global-nav.tsx
- 禁止行为：不得修改 docs/requirements/；不得在页面中直接拼接敏感认证信息
- 新增：apps/web-admin/app/portfolios/page.tsx
- 新增：apps/web-admin/app/portfolios/[portfolioId]/page.tsx
- 新增：apps/web-admin/app/portfolios/[portfolioId]/projects/page.tsx
- 新增：apps/web-admin/app/portfolios/[portfolioId]/health/page.tsx
- 修改：apps/web-admin/app/global-nav.tsx

---
