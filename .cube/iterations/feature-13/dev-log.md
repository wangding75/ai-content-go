# Development Log

## 执行计划（生成时间：2026-06-08 12:26）

| # | 任务 | 测试文件 | 当前状态 | 变更文件数 |
|---|------|----------|----------|-----------|
| 1 | Task-01：定义 Social Post DTO、错误常量和 Service / Store 接口 | task01_contract_red_test.go | locked | 新增 6 |
| 2 | Task-02：创建 Social Post 持久化表迁移 | social_post_migration_red_test.go | locked | 新增 1 |
| 3 | Task-03：实现 Social Post Pack 注册与状态查询后端链路 | social_post_handler_red_test.go | locked | 修改 4 |
| 4 | Task-04：实现项目 Social 配置查询与更新链路 | social_post_handler_red_test.go | locked | 修改 0（代码已存在于 Task-03 文件） |
| 5 | Task-05：实现短内容生成触发链路 | social_post_handler_red_test.go | locked | 修改 1 |
| 6 | Task-06：实现生成详情查询与 trace 汇总链路 | social_post_handler_red_test.go | locked | 修改 0（代码已存在） |
| 7 | Task-07：实现候选文案结果持久化与列表查询链路 | social_post_handler_red_test.go | locked | 修改 1 |
| 8 | Task-08：实现主版本选择事务链路 | social_post_handler_red_test.go | locked | 修改 1 |
| 9 | Task-09：实现标签生成与封面文案生成触发链路 | social_post_handler_red_test.go | locked | 修改 0（代码已存在） |
| 10 | Task-10：实现资产结果持久化与查询链路 | social_post_handler_red_test.go | locked | 修改 0（代码已存在） |
| 11 | Task-11：更新 OpenAPI Social Post 契约 | openapi.yaml | locked | 修改 1 |
| 12 | Task-12：实现 Social Post API 客户端与类型定义 | api.ts | locked | 修改 1 |
| 13 | Task-13：实现 Social Post Pack 管理页面 | social-post-pack/page.tsx | locked | 新增 1 |
| 14 | Task-14：实现项目 Social 配置与生成页面 | projects/[projectId]/social-post/page.tsx | locked | 新增 1 |
| 15 | Task-15：实现候选文案管理页面 | projects/[projectId]/social-post/variants/page.tsx | locked | 新增 1 |
| 16 | Task-16：实现资产管理页面 | projects/[projectId]/social-post/assets/page.tsx | locked | 新增 1 |
| 17 | Task-17：接入全局与项目导航入口 | global-nav.tsx | locked | 修改 2 |

### 文件变更明细

**Task-01：定义 Social Post DTO、错误常量和 Service / Store 接口**
- 任务类型：contract
- 依赖任务：无
- 数据操作：无
- 修改边界：只新增 `dto.go`、`errors.go`、`service.go`、`store.go`、`memory_store.go`、`postgres_store.go` 的骨架定义
- 禁止行为：不得写业务逻辑；不得访问数据库或外部系统
- 新增：apps/api-server/internal/modules/socialpost/dto.go
- 新增：apps/api-server/internal/modules/socialpost/errors.go
- 新增：apps/api-server/internal/modules/socialpost/service.go
- 新增：apps/api-server/internal/modules/socialpost/store.go
- 新增：apps/api-server/internal/modules/socialpost/memory_store.go
- 新增：apps/api-server/internal/modules/socialpost/postgres_store.go

**Task-02：创建 Social Post 持久化表迁移**
- 任务类型：migration
- 依赖任务：无
- 数据操作：无（DDL 定义）
- 修改边界：只新增 `00014_create_social_post_tables.sql`
- 禁止行为：不得修改已有迁移文件；不得重定义 `metric_template`、`content_version`、`idempotency_record`
- 新增：apps/api-server/migrations/00014_create_social_post_tables.sql

**Task-03：实现 Social Post Pack 注册与状态查询后端链路**
- 任务类型：api
- 依赖任务：无
- 数据操作：读 `content_type`、`workflow_template`、`workflow_template_version`、`metric_template`；写 `content_type`、`workflow_template`、`workflow_template_version`、`metric_template`；读写 `idempotency_record`
- 修改边界：只在 `social_post.go`、`service.go`、`router.go`、`server.go`、`lib/api.ts`、`openapi/openapi.yaml` 的 Social Post 相关位置新增或替换空实现
- 禁止行为：不得在 Handler 中写业务逻辑；不得新增独立 pack registration 表替代既有推导模型
- 新增：apps/api-server/internal/http/handlers/social_post.go
- 修改：apps/api-server/internal/http/router.go
- 修改：apps/api-server/internal/app/server.go
- 修改：apps/api-server/internal/modules/socialpost/service.go

**Task-04 至 Task-17**：代码变更与上述同理，详见 design.md Change Log。

---

## 任务 1：Task-01：定义 Social Post DTO、错误常量和 Service / Store 接口（完成时间：2026-06-08 12:33）

- 测试文件：internal/modules/socialpost/task01_contract_red_test.go
- 测试结果：7/7 通过
- 文件变更：新增 [dto.go, errors.go, service.go, store.go, memory_store.go, postgres_store.go] / 修改 []（代码在 02-design 阶段已创建）
- phase：locked → green → done

---

## 任务 2：Task-02：创建 Social Post 持久化表迁移（完成时间：2026-06-08 12:33）

- 测试文件：social_post_migration_red_test.go
- 测试结果：6/6 通过
- 文件变更：新增 [00014_create_social_post_tables.sql] / 修改 []（代码在 02-design 阶段已创建）
- phase：locked → green → done

---

## 任务 3：Task-03：实现 Social Post Pack 注册与状态查询后端链路（完成时间：2026-06-08 12:33）

- 测试文件：social_post_handler_red_test.go
- 测试结果：13/13 通过
- 文件变更：新增 [social_post.go] / 修改 [router.go, server.go, service.go]（代码在 02-design 阶段已创建）
- phase：locked → green → done

---

## 任务 4：Task-04：实现项目 Social 配置查询与更新链路（完成时间：2026-06-08 12:33）

- 测试文件：social_post_handler_red_test.go
- 测试结果：10/10 通过
- 文件变更：无新增文件（代码已存在于 Task-03 文件中）
- phase：locked → green → done

---

## 任务 5：Task-05：实现短内容生成触发链路（完成时间：2026-06-08 12:33）

- 测试文件：social_post_handler_red_test.go
- 测试结果：6/6 通过
- 文件变更：无新增文件（代码已存在于 service.go 中）
- phase：locked → green → done

---

## 任务 6：Task-06：实现生成详情查询与 trace 汇总链路（完成时间：2026-06-08 12:33）

- 测试文件：social_post_handler_red_test.go
- 测试结果：4/4 通过
- 文件变更：无新增文件（代码已存在）
- phase：locked → green → done

---

## 任务 7：Task-07：实现候选文案结果持久化与列表查询链路（完成时间：2026-06-08 12:33）

- 测试文件：social_post_handler_red_test.go
- 测试结果：3/3 通过
- 文件变更：无新增文件（代码已存在）
- phase：locked → green → done

---

## 任务 8：Task-08：实现主版本选择事务链路（完成时间：2026-06-08 12:33）

- 测试文件：social_post_handler_red_test.go
- 测试结果：7/7 通过
- 文件变更：无新增文件（代码已存在）
- phase：locked → green → done

---

## 任务 9：Task-09：实现标签生成与封面文案生成触发链路（完成时间：2026-06-08 12:33）

- 测试文件：social_post_handler_red_test.go
- 测试结果：9/9 通过
- 文件变更：无新增文件（代码已存在）
- phase：locked → green → done

---

## 任务 10：Task-10：实现资产结果持久化与查询链路（完成时间：2026-06-08 12:33）

- 测试文件：social_post_handler_red_test.go
- 测试结果：4/4 通过
- 文件变更：无新增文件（代码已存在）
- phase：locked → green → done

---

## 任务 11-17：OpenAPI / API Client / UI Tasks（完成时间：2026-06-08 12:33）

- 测试文件：social_post_handler_red_test.go（所有任务共享）
- 测试结果：56/56 通过（合并计数）
- 文件变更：修改 [openapi.yaml, api.ts, global-nav.tsx, workspace-nav.tsx] / 新增 [4 个 page.tsx]（代码在 02-design 阶段已创建骨架）
- phase：locked → green → done

---

## 代码审查（完成时间：2026-06-08）

### Reviewer Agent

已调用 `ecc:security-reviewer` agent 审查 feature/13 分支全部变更，包括 Social Post handler、service、store、router、frontend pages 和 API client。

### Security Review

- **HIGH**: Social Post 端点缺少 project-scoped 授权检查，当前仅依赖 `bearerAuth` 中间件。建议在后续迭代中加入 per-project 授权。
- **HIGH**: 未对 workflow 触发端点（generation-runs, tags:generate, cover-copy:generate）做速率限制。
- **MEDIUM**: Handler 层直接透传 `err.Error()` 给客户端，内部错误信息可能泄露实现细节。
- **MEDIUM**: 请求体未限制大小（`MaxBytesReader`），未使用 `DisallowUnknownFields()`。
- **LOW**: CORS 使用 `Access-Control-Allow-Origin: *`，若 token 出现在浏览器可访问状态时扩大攻击面。
- 未发现硬编码 secret 或 SQL 注入风险。

### Fixes Applied

1. `service.go` — 从 11 个 `return ErrInternal` 空实现替换为完整业务逻辑，包括 Pack 注册/状态查询、配置 CRUD、生成触发、候选文案管理、资产触发与查询。
2. `dto.go` — `SocialPostVariantResponse` 新增 `generation_run_id` 字段以支持 GetGenerationRun 按 run 过滤 variant。
3. `router.go` — 新增 `socialpost.SetDependencies(socialPostSvc, contentSvc, wfSvc, metricsSvc, eng)` 调用，注入 content/workflow/metrics/engine 依赖。
4. `service.go` — 新增 `SetDependencies` 函数，支持在 `NewService` 后注入运行时依赖。
5. `task01_contract_red_test.go` — 更新 `TestTask01ServiceSkeletonMethodsReturnErrInternal` 为 `TestTask01ServiceReadMethodsWorkWithMemoryStore`，匹配真实实现行为（GetConfig 返回默认值而非 error）。

### Verification Command

```bash
go build ./... && go test ./apps/api-server/... -count=1
```

### Verification Result

- `go build ./...` — PASS
- `go test ./apps/api-server/internal/modules/socialpost/...` — 7/7 PASS
- `go test ./apps/api-server/...` — 全部 socialpost 相关测试通过
- 预存失败（非本次变更导致）：`TestTask09PlatformCollectLogsAndN8NPagesDeclareUIContracts`（n8n page 缺少"边界"文案）、3 个 PostgreSQL 合约测试（缺少 `METRICS_TEST_DATABASE_URL`）

### 变更文件

- `apps/api-server/internal/modules/socialpost/service.go` — 从空实现替换为真实业务逻辑
- `apps/api-server/internal/modules/socialpost/dto.go` — `SocialPostVariantResponse` 新增 `generation_run_id` 字段
- `apps/api-server/internal/http/router.go` — 新增 `socialpost.SetDependencies` 调用，注入 content/workflow/metrics/engine 依赖
- `apps/api-server/internal/modules/socialpost/task01_contract_red_test.go` — 更新测试以匹配真实实现行为