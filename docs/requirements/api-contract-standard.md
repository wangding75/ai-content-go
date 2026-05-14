# AI Content Factory API 契约规范 v5（Go 版）

> 更新日期：2026-05-14  
> 适用范围：Go API Server、Web Admin、Browser Extension、异步任务回调、OpenAPI 文档。

---

## 1. 总体要求

| 项目 | 要求 |
|---|---|
| API 前缀 | `/api/v1` |
| 协议 | REST JSON |
| 文档 | 必须输出 OpenAPI 3.0 |
| Go DTO | 每个接口必须有 request / response struct |
| 校验 | 使用 struct tag + validator |
| 错误结构 | 统一错误响应 |
| 分页 | 列表接口必须支持 page、page_size、sort、order |
| 幂等 | 创建运行、发布回填、策略确认必须支持 `Idempotency-Key` |
| 状态变更 | 必须写 operation_log |
| 异步任务 | 返回 run_id / job_id，不同步等待最终结果 |
| 领域边界 | Core API 不得使用 book/chapter 作为核心资源名 |

---

## 2. 统一响应结构

### 成功响应

```json
{
  "success": true,
  "data": {},
  "error": null,
  "request_id": "req_20260514100000001"
}
```

### 失败响应

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request body.",
    "details": [
      {
        "field": "name",
        "reason": "name is required"
      }
    ]
  },
  "request_id": "req_20260514100000002"
}
```

### 分页响应

```json
{
  "success": true,
  "data": {
    "items": [],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 128,
      "has_next": true
    }
  },
  "error": null,
  "request_id": "req_20260514100000003"
}
```

---

## 3. Go DTO 示例

```go
type CreateProjectRequest struct {
    Name                    string                 `json:"name" validate:"required,max=120"`
    ContentTypeID           string                 `json:"content_type_id" validate:"required"`
    TargetPlatform          string                 `json:"target_platform" validate:"required"`
    TargetContentCount      int                    `json:"target_content_count" validate:"gte=1"`
    DefaultGenerationConfig map[string]any         `json:"default_generation_config"`
    ProjectConfig           map[string]any         `json:"project_config"`
}

type CreateProjectResponse struct {
    ID                 string `json:"id"`
    Name               string `json:"name"`
    Status             string `json:"status"`
    ContentTypeCode     string `json:"content_type_code"`
    TargetPlatform     string `json:"target_platform"`
    TargetContentCount int    `json:"target_content_count"`
}
```

---

## 4. 通用 Header

| Header | 必填 | 说明 |
|---|---:|---|
| Authorization | 是 | `Bearer <token>`；开发早期可占位 |
| X-Request-Id | 否 | 客户端请求 ID |
| Idempotency-Key | 条件必填 | 创建运行、发布回填、策略确认等接口 |
| Content-Type | 是 | `application/json` |

---

## 5. 通用错误码

| 错误码 | HTTP 状态 | 说明 |
|---|---:|---|
| VALIDATION_ERROR | 400 | 请求参数错误 |
| UNAUTHORIZED | 401 | 未登录 |
| FORBIDDEN | 403 | 无权限 |
| NOT_FOUND | 404 | 资源不存在 |
| CONFLICT | 409 | 状态冲突或唯一约束冲突 |
| IDEMPOTENCY_CONFLICT | 409 | 幂等键重复但请求体不一致 |
| WORKFLOW_RUN_FAILED | 422 | 工作流运行失败 |
| AGENT_OUTPUT_INVALID | 422 | Agent 输出结构化校验失败 |
| LLM_PROVIDER_ERROR | 502 | 模型服务调用失败 |
| EXTERNAL_AUTOMATION_ERROR | 502 | 外部自动化调用失败 |
| INTERNAL_ERROR | 500 | 未预期系统错误 |

---

## 6. OpenAPI 必填内容

每个接口必须包含：

```text
summary
description
tags
operationId
parameters
requestBody
responses
security
examples
```

---

## 7. 每个迭代的接口验收项

- [ ] 所有接口有 Go DTO。
- [ ] 所有接口进入 OpenAPI。
- [ ] 每个接口明确 path、query、headers、body。
- [ ] 每个接口明确成功响应和失败响应。
- [ ] 列表接口支持分页、筛选、排序。
- [ ] 状态变更接口写入 operation_log。
- [ ] 异步接口只返回运行记录 ID。
- [ ] 幂等敏感接口支持 Idempotency-Key。
- [ ] 页面-接口映射已写入迭代文档。
