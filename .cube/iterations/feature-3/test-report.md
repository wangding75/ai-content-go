# Iteration 3 Test Report：Novel Pack 新书规划流程

## Test Scope

本次 05-testing 覆盖 Iteration 3 Novel Pack 新书规划流程的后端、前端、OpenAPI、迁移和阶段产物一致性。

覆盖模块：

- Go API Server：`internal/modules/novel`、`internal/http/handlers/novel.go`、`internal/http/router.go`、`internal/modules/workflow` 幂等处理。
- 数据契约：`apps/api-server/migrations/00005_create_novel_planning_tables.sql`、`openapi/openapi.yaml`。
- Web Admin：`apps/web-admin/lib/api.ts`、项目工作区导航、规划页、候选确认页、世界观、人物、大纲页面。
- Cube 产物：`prd.md`、`design.md`、`test-map.yaml`、`STATUS.yaml`、`dev-log.md`。

命中的类型化规范：

- `standards/testing/web-e2e.md`：Novel HTTP API 与 Web Admin 页面链路。
- `standards/testing/integration.md`：Novel Service、HTTP Handler、Workflow/Engine 组合链路。
- `standards/testing/sql-query.md`：PostgreSQL DDL contract 结构约束。
- `standards/testing/library.md`：Novel DTO/Service contract 与 Web Admin API client contract。

## Test Results

实际执行结果：

| 命令 | 结果 | 证据 |
|---|---|---|
| `go test -race ./...` | PASS | `.cube/iterations/feature-3/final-go-test.log` |
| `npm --prefix apps/web-admin run lint` | PASS | `.cube/iterations/feature-3/final-web-lint.log` |
| `npm --prefix apps/web-admin run build` | PASS | `.cube/iterations/feature-3/final-web-build.log` |
| `node .../check-test-map.mjs` | PASS（Stage 03 已验证） | `test-map.yaml` |
| `node .../cube-check 04-development` | PASS（3/3） | `dev-log.md`、`STATUS.yaml` |

Go race 测试结果：所有包通过，无失败用例。Web Admin TypeScript 检查通过。Next.js 生产构建通过，新增动态路由 `/projects/[projectId]/planning`、`/planning/topics`、`/novel/worldview`、`/novel/characters`、`/novel/arcs` 均成功生成。

类型化测试结果：

- Web/API：HTTP handler contract 覆盖创建规划运行、缺少幂等键、分页校验、失败补偿；Web 页面 contract 覆盖规划页、候选确认、世界观、人物、大纲页面状态。
- Integration：Novel Service 状态机覆盖规划运行创建、候选快照、幂等冲突、候选确认状态流转；全量 Go 测试验证与现有 Workflow/Engine 包兼容。
- SQL/query：迁移 contract 覆盖 6 张 Novel 规划表、项目级组合外键、状态 CHECK、规划幂等唯一索引、禁止 Core 表污染。
- Library：DTO JSON tag、Service interface、Web Admin API client 类型和函数均有 contract 覆盖。

## Pass Criteria

| 验收标准 | 状态 | 证据 |
|---|---|---|
| AC-001 项目工作区规划入口与状态 | PASS | Web contract tests、Next build 路由输出 |
| AC-002 原型风格页面布局 | PASS | 页面使用 `page-shell`、`page-hero`、`card`、`table-card`、`dialog-backdrop` 等样式类 |
| AC-003 启动规划与幂等 | PASS | Novel Service/HTTP tests；workflow 幂等缓存已按 kind/payload 校验 |
| AC-004 规划运行列表分页筛选排序 | PASS | Handler pagination validation、API client 支持 sort/order/status |
| AC-005 规划详情追踪字段 | PASS | DTO/detail contract 覆盖 workflow、step、agent、llm summary 字段 |
| AC-006 规划资源持久化契约 | PASS | in-memory Service + PostgreSQL DDL contract 覆盖规划运行、快照、候选、世界观、人物、大纲 |
| AC-007 候选确认弹窗与直达页 | PASS | Planning/Topics 页面 contract tests |
| AC-008 确认候选幂等与审计输出 | PASS | ConfirmTopic service contract 返回状态流转和 `operation_log_id` |
| AC-009 世界观查看编辑 | PASS | Worldview 页面/API contract |
| AC-010 人物列表与新增 | PASS | Characters 页面/API contract |
| AC-011 大纲列表与来源 | PASS | Arcs 页面/API contract |
| AC-012 DTO、统一响应、OpenAPI | PASS | Go DTO、handler envelope、OpenAPI contract tests |
| AC-013 Novel Pack 扩展边界 | PASS | Novel 专属代码位于 `novel` 模块、Novel API、Web Novel 页面与迁移扩展表 |
| AC-014 页面绑定 API 与反馈 | PASS | API client 与页面 contract tests；TypeScript/Next build 通过 |
| AC-015 鉴权与敏感信息 | PASS with known risk | 现有 `/api/v1` Bearer middleware 覆盖认证；项目级授权/Novel Pack 项目校验见 Known Issues |
| AC-016 自动化覆盖 | PASS | 23 个 Stage 03 `@Test` contract + 全量 Go/Web checks |

## Coverage

未配置单独 coverage 命令；05 阶段 gate 中 `coverage_command` 为可选未解析项。实际覆盖证据来自：

- 23 个锁定 `@Test` contract 覆盖 10 个 Development Tasks，`test-map.yaml` 覆盖率 100%。
- `go test -race ./...` 覆盖所有 Go 包并通过 race detector。
- `npm --prefix apps/web-admin run lint` 覆盖 TypeScript 类型契约。
- `npm --prefix apps/web-admin run build` 覆盖 Next.js App Router 构建与新增动态路由。

覆盖缺口：没有真实浏览器点击型 E2E；本阶段使用页面源码 contract、TypeScript、Next production build 和 handler/service integration 作为替代验证。没有真实 PostgreSQL 执行迁移；本阶段使用 DDL 文本结构 contract 验证 PostgreSQL 关键约束。

## Standards Evidence

| 标准 | 执行命令/证据 | 结果 |
|---|---|---|
| `standards/testing/web-e2e.md` | Handler contract tests、Web page contract tests、Next build | PASS |
| `standards/testing/integration.md` | `go test -race ./...`，Novel Service/HTTP/Workflow 相关包 | PASS |
| `standards/testing/sql-query.md` | `iteration3_task04_contract_red_test.go` 检查 DDL 结构、状态约束、组合外键、幂等唯一索引 | PASS |
| `standards/testing/library.md` | DTO/Service/API client contract tests | PASS |

未执行项与风险：未启动真实 API server 发 HTTP client 全链路，也未连接真实数据库执行 DDL。由于项目当前测试基线以 in-memory service 和 contract tests 为主，替代验证不阻塞本次 cube 机械验收，但应在后续持久化/真实 E2E 迭代补齐。

## Review Evidence

04-development 已执行代码审查与安全审查，并记录在 `.cube/iterations/feature-3/dev-log.md`。

发现并修复的问题：

- HIGH：workflow 失败补偿复用同一幂等键可能触发 workflow idempotency cache 类型断言 panic；已改为补偿 cancel 使用独立 key。
- HIGH：前端默认读取最旧 planning run，可能确认陈旧候选；已改为请求 `sort=created_at&order=desc`。
- HIGH：幂等冲突 retry 可能取消已存在 workflow run；已跳过 `ErrIdempotencyConflict` 的补偿取消，并将 workflow idempotency cache 改为 kind/payload 绑定。
- MEDIUM：分页/排序和 Idempotency-Key 缺少边界；已限制 `page_size <= 100`、sort/order 白名单、Idempotency-Key 长度。
- MEDIUM：DDL 未唯一约束规划幂等键；已增加 `(project_id, idempotency_key)` 部分唯一索引。

最终审查结论：无 CRITICAL。仍有一个 HIGH 级产品/架构边界风险记录在 Known Issues：项目存在性、项目级授权、Novel Pack 类型和已发布模板版本校验目前未与 Content/Workflow 服务强绑定；当前实现仅在 Novel Service 内做基础字段校验和项目字符串隔离。该项需要扩展 router/service 注入边界或回退设计阶段补充跨服务授权/校验设计，不建议在 05 阶段临时扩大架构。

## Known Issues

- 项目/模板强校验风险：PRD 要求校验项目存在、权限、Novel Pack 类型和已发布模板版本；当前最小实现尚未注入 Content Service / Workflow Template Version 查询到 Novel Service。现有 Bearer middleware 只能鉴别请求是否认证，不能判断当前 token 对 `projectId` 的授权。建议后续通过 `/cube:regress 02-design` 或新迭代补充跨服务授权与模板发布状态校验设计，再重建测试契约。
- 真实浏览器 E2E 与真实 PostgreSQL DDL 执行未在本阶段运行；已有 Next build、TypeScript、handler/service integration 和 DDL contract 作为替代证据。
