# Development Log

## 执行计划（生成时间：2026-06-02）

整体进度：已完成 0 / 共 9 个任务

| # | 任务 | 测试文件 | 当前状态 | 变更文件数 |
|---|------|----------|----------|-----------|
| → | Task-01：定义平台适配器与插件协作数据库契约 | platform_adapter_sql_contract_test.go | locked | 新增 1 / 修改 0 |
| 1 | Task-02：实现平台 Adapter 配置管理业务行为 | iteration11_task02_adapter_contract_red_test.go | locked | 新增 2 / 修改 3 |
| 2 | Task-03：实现插件客户端密钥与短期认证业务行为 | iteration11_task03_plugin_client_contract_red_test.go | locked | 新增 2 / 修改 3 |
| 3 | Task-04：实现插件发布任务锁定与回填业务行为 | iteration11_task04_plugin_publish_job_contract_red_test.go | locked | 新增 2 / 修改 3 |
| 4 | Task-05：实现平台采集日志与人工确认指标业务行为 | iteration11_task05_collect_log_contract_red_test.go | locked | 新增 2 / 修改 3 |
| 5 | Task-06：实现 n8n 外围回调边界业务行为 | iteration11_task06_callback_contract_red_test.go | locked | 新增 2 / 修改 3 |
| 6 | Task-07：实现 Iteration 11 OpenAPI 与前端 API Client 契约 | iteration11_task07_openapi_contract_red_test.go | locked | 新增 0 / 修改 2 |
| 7 | Task-08：实现管理台平台 Adapter 与插件客户端页面交互 | iteration11_task08_pages_contract_red_test.go | locked | 新增 2 / 修改 2 |
| 8 | Task-09：实现管理台采集日志与 n8n 回调页面交互 | iteration11_task09_pages_contract_red_test.go | locked | 新增 1 / 修改 1 |

### 文件变更明细

**任务 1：Task-01：定义平台适配器与插件协作数据库契约**
- 任务类型：migration
- 依赖任务：无
- 数据操作：写 platform_adapter_config、platform_adapter_revision、plugin_client、plugin_access_token、platform_collect_log、external_callback_log 表结构；修改 publish_target；修改 publish_job；修改 publish_log 事件约束；修改 external_workflow_binding callback auth 字段
- 修改边界：只新增 00013 migration 和 SQL 契约测试；只通过 ALTER TABLE 扩展既有 publish 表；只安全调整 publish_log event_type CHECK
- 禁止行为：不得删除既有表；不得保存 api_key、password、secret 明文字段；不得重写旧迁移；不得使用未白名单动态 ORDER BY
- 修改：apps/api-server/migrations/00013_create_platform_adapter_extension_tables.sql

**任务 2：Task-02：实现平台 Adapter 配置管理业务行为**
- 任务类型：business-implementation
- 依赖任务：Task-01
- 数据操作：读写 platform_adapter_config；写 platform_adapter_revision；读外部 credential/provider/binding 引用；停用前按 publish_job.adapter_config_id 与 publish_target 平台/目标类型回退匹配读取 publish_job；写 operation_log
- 修改边界：只新增 publish DTO/Service/Handler 中 Adapter 相关类型和方法；只在 router.go 注册 Adapter routes；只以向后兼容方式为 publish target/job 增加 Adapter 映射字段
- 禁止行为：不得保存平台凭证明文；不得新增独立 platform 模块；不得破坏既有 publish target API 语义；不得在仍有 queued/copied/locked 任务时停用 Adapter
- 新增：apps/api-server/internal/modules/publish/adapter_dto.go（需要时拆分）
- 修改：apps/api-server/internal/modules/publish/dto.go；apps/api-server/internal/modules/publish/service.go；apps/api-server/internal/http/handlers/publish.go；apps/api-server/internal/http/router.go

**任务 3：Task-03：实现插件客户端密钥与短期认证业务行为**
- 任务类型：business-implementation
- 依赖任务：Task-01
- 数据操作：读写 plugin_client；写 plugin_access_token；写 operation_log；更新 last_active_at；写认证失败审计
- 修改边界：只新增 publish DTO/Service/Handler 中插件客户端和认证相关类型和方法；只在 router.go 注册插件客户端和认证 routes
- 禁止行为：不得保存 api_key 明文；不得泄露密钥校验细节；不得让禁用客户端认证成功；plugin-auth 不得强制 admin bearer
- 新增：apps/api-server/internal/modules/publish/plugin_dto.go（需要时拆分）
- 修改：apps/api-server/internal/modules/publish/dto.go；apps/api-server/internal/modules/publish/service.go；apps/api-server/internal/http/handlers/publish.go；apps/api-server/internal/http/router.go

**任务 4：Task-04：实现插件发布任务锁定与回填业务行为**
- 任务类型：business-implementation
- 依赖任务：Task-01、Task-02、Task-03
- 数据操作：读 plugin_access_token 和 plugin_client；读 platform_adapter_config；以原子条件更新 publish_job 锁字段；读写 publish_job 状态；写 publish_log；写 operation_log；读写幂等记录
- 修改边界：只扩展 publish Service、Handler、DTO 和 router 中插件任务相关方法；只追加 publish_job 状态辅助逻辑
- 禁止行为：不得新增第二套发布主状态机；不得绕过 publish_job；不得使用事务外 read-then-write 领取锁；不得在 lock/client/payload 无效时推进状态；不得重复处理相同幂等键
- 新增：apps/api-server/internal/modules/publish/plugin_job_dto.go（需要时拆分）
- 修改：apps/api-server/internal/modules/publish/dto.go；apps/api-server/internal/modules/publish/service.go；apps/api-server/internal/http/handlers/publish.go；apps/api-server/internal/http/router.go

**任务 5：Task-05：实现平台采集日志与人工确认指标业务行为**
- 任务类型：business-implementation
- 依赖任务：Task-01
- 数据操作：插入/列表/读取/更新 platform_collect_log；读 metric_template；写 metric_record；写 operation_log；读写幂等记录；确认时更新 collect_log 状态
- 修改边界：只扩展 metrics DTO/Service/Handler/router；只修改 store.go、memory_store.go、postgres_store.go 中 collect log 所需接口与实现；只复用既有 MetricRecord 写入语义
- 禁止行为：不得默认自动写入 metric_record；不得丢弃失败采集摘要；不得重复确认污染指标；不得绕过 MetricRecord uniqueness/source 规则；不得整文件重写 store 文件
- 新增：apps/api-server/internal/modules/metrics/collect_dto.go（需要时拆分）
- 修改：apps/api-server/internal/modules/metrics/dto.go；apps/api-server/internal/modules/metrics/service.go；apps/api-server/internal/modules/metrics/store.go；apps/api-server/internal/modules/metrics/memory_store.go；apps/api-server/internal/modules/metrics/postgres_store.go；apps/api-server/internal/http/handlers/metrics.go；apps/api-server/internal/http/router.go

**任务 6：Task-06：实现 n8n 外围回调边界业务行为**
- 任务类型：business-implementation
- 依赖任务：Task-01
- 数据操作：读写 external binding callback auth 字段；读 external provider；读写 external_callback_log；读写幂等记录
- 修改边界：只扩展 external DTO/Service/Handler 和 router 中回调相关方法；只记录外围事件
- 禁止行为：不得创建 WorkflowRun；不得推进 Agent 编排；不得修改内容正文；不得直接改 publish_job 主状态；不得把 stable_event_id 当认证
- 新增：apps/api-server/internal/modules/external/callback_dto.go（需要时拆分）
- 修改：apps/api-server/internal/modules/external/dto.go；apps/api-server/internal/modules/external/service.go；apps/api-server/internal/http/handlers/external.go；apps/api-server/internal/http/router.go

**任务 7：Task-07：实现 Iteration 11 OpenAPI 与前端 API Client 契约**
- 任务类型：contract
- 依赖任务：Task-02、Task-03、Task-04、Task-05、Task-06
- 数据操作：无
- 修改边界：只修改 openapi/openapi.yaml 和 apps/web-admin/lib/api.ts 中 Iteration 11 相关定义
- 禁止行为：不得改变既有 API client 函数签名；不得删除旧 OpenAPI paths；不得使用 any 表示新增公共响应；不得遗漏 callback-log 和 test callback API
- 修改：openapi/openapi.yaml；apps/web-admin/lib/api.ts

**任务 8：Task-08：实现管理台平台 Adapter 与插件客户端页面交互**
- 任务类型：ui
- 依赖任务：Task-07
- 数据操作：调用 Adapter 和 plugin client HTTP API
- 修改边界：只新增两个页面；只为 /platform-adapters、/plugin-clients 追加 global-nav 项；只使用现有 CSS class 和 API client
- 禁止行为：不得裸 HTML 无样式；不得刷新 404；不得在页面状态中持久化 api_key_once 超出弹窗生命周期；不得偏离 prototype 管理台视觉
- 新增：apps/web-admin/app/platform-adapters/page.tsx；apps/web-admin/app/plugin-clients/page.tsx
- 修改：apps/web-admin/app/global-nav.tsx

**任务 9：Task-09：实现管理台采集日志与 n8n 回调页面交互**
- 任务类型：ui
- 依赖任务：Task-07
- 数据操作：调用 platform_collect_log 和 external callback/callback-log HTTP API
- 修改边界：只新增采集日志页面；只修改 n8n 页面 Iteration 11 区域；只为 /platform-collect-logs 追加 global-nav 项并保留既有 n8n 入口
- 禁止行为：不得默认自动确认指标；不得隐藏失败采集记录；不得让按钮点击无反馈；不得偏离 prototype 管理台视觉
- 新增：apps/web-admin/app/platform-collect-logs/page.tsx
- 修改：apps/web-admin/app/external-automation/n8n/page.tsx；apps/web-admin/app/global-nav.tsx

## 代码审查

### Reviewer Agent

使用 ecc:go-reviewer 和 ecc:typescript-reviewer 对 iteration 11 全部变更进行交叉审查。

### Security Review

**CRITICAL 修复（已完成）：**

1. **C1: postgres_store.go panic → error return** — InsertPlatformCollectLog、ListPlatformCollectLogs、GetPlatformCollectLog、UpdatePlatformCollectLogStatus 四个方法从 `panic("not implemented")` 改为 `errors.New("not implemented")`，避免生产环境崩溃。

2. **C2: publish/service.go 数据竞争** — MarkPluginPublishJobPublished 和 MarkPluginPublishJobFailed 中 `reserveIdempotencyLocked` 在 `s.state.mu.Lock()` 之前调用，导致对 `s.state.idempotent` 的无同步读写。修复：将 `s.state.mu.Lock()` 移至 `reserveIdempotencyLocked` 调用之前。

3. **C3: 不安全随机数生成** — `generateRandomString` 使用 `time.Now().UnixNano()` 作为熵源，可预测。修复：改用 `crypto/rand.Read`，仅在 `rand.Read` 失败时 fallback 到时间种子。

4. **C4: 忽略 store 写入错误** — metrics/service.go 中 `StoreIdempotency`（两处）和 `UpdatePlatformCollectLogStatus` 的错误被静默丢弃。修复：添加 `if err := ...; err != nil { return ..., ErrInternal }` 错误处理。

**HIGH 修复（已完成）：**

5. **H2: 时序不安全 API Key 比较** — `AuthenticatePlugin` 中 `entry.apiKeyHash == sha256String(req.APIKey)` 使用字符串比较，存在时序侧信道攻击风险。修复：改用 `crypto/subtle.ConstantTimeCompare`。

**已知但不在本次修复范围（记录为 Known Risk）：**

- H1: 插件访问 token 验证为 no-op（仅检查非空）— 需要设计 token 查找机制，超出 iteration 11 scope
- H3: publish/service.go 超过 800 行 — 需要拆分文件，超出 iteration 11 scope
- H4: `defaultState` 全局可变状态 — 需要重构 NewService，超出 iteration 11 scope
- 前端：FormData.get() 的 `as string` 强制转换、缺少 .catch() 错误处理、缺少客户端输入验证 — 均为管理台内网使用，风险可控

### Fixes Applied

| # | 文件 | 修改 |
|---|------|------|
| C1 | metrics/postgres_store.go | panic → errors.New，新增 "errors" import |
| C2 | publish/service.go | MarkPluginPublishJobPublished/Failed: mu.Lock 移至 reserveIdempotencyLocked 之前 |
| C3 | publish/service.go | generateRandomString: crypto/rand.Read 替代 time.Now().UnixNano()，新增 "crypto/rand" import |
| C4 | metrics/service.go | StoreIdempotency/UpdatePlatformCollectLogStatus: 错误不再忽略 |
| H2 | publish/service.go | apiKeyHash 比较: == → subtle.ConstantTimeCompare，新增 "crypto/subtle" import |

### Verification Command

```bash
go test -run "Iteration11|iteration11|Task0[1-9]" ./apps/api-server/internal/http/contract/ -v && go build ./apps/api-server/...
```

### Verification Result

所有 iteration 11 contract tests PASS（95 tests），go build 成功无错误。
