# Iteration 8 技术设计：数据录入与指标看板

## 1. 概述

本次设计在现有分层架构中新增 `metrics` 业务模块，围绕指标模板、指标记录、批量导入、项目指标汇总、趋势序列和缺失数据提醒构建项目级指标中心。后端沿用当前 `handler -> module service -> DTO -> api.Envelope` 模式，前端沿用 `apps/web-admin/lib/api.ts` API client、项目工作区导航和现有管理台样式，OpenAPI 继续维护单一 `openapi/openapi.yaml`。

核心约束：

- Core 层保持内容类型无关，只使用 `project_id`、`content_item_id`、`content_version_id`、`publish_job_id`、`target_id` 等通用关联。
- 指标录入必须绑定发布结果和来源，来源类型至少兼容 `manual`、`import`、`extension`、`external_callback`，本迭代闭环 `manual` 与 `import`。
- 单条录入与批量导入支持 `Idempotency-Key`，同键同请求返回一致结果，同键不同请求返回 `IDEMPOTENCY_CONFLICT`。
- 同一项目、平台、发布目标、内容版本、指标编码、指标日期和周期下不得产生冲突记录；不同值重复提交返回 `CONFLICT`。
- 缺失提醒基于模板 `required`、发布完成时间、统计窗口、平台匹配和模板启用状态计算。
- 汇总与趋势必须返回可复现聚合口径，并通过 `metric_summary_snapshot` 或等价稳定引用支撑 Iteration 9 策略建议。

## 2. Impact Analysis

| 模块/文件 | 影响程度 | 说明 |
| --- | --- | --- |
| `apps/api-server/internal/modules/metrics` | 新增 | 新增指标 DTO、错误常量、Service 接口与骨架。 |
| `apps/api-server/internal/http/handlers/metrics.go` | 新增 | 新增 HTTP Handler，负责参数解析、统一响应和 metrics 错误映射。 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册指标模板、指标记录、项目汇总、趋势和缺失提醒路由。 |
| `apps/api-server/migrations/00010_create_metrics_tables.sql` | 新增 | 新增 `metric_template`、`metric_record`、`metric_summary_snapshot` 表。 |
| `openapi/openapi.yaml` | 修改 | 增加本迭代所有接口 path、schema、响应和 examples。 |
| `apps/web-admin/lib/api.ts` | 修改 | 增加 metrics 类型与 API client 函数。 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 修改 | 增加指标表现、指标录入、趋势图、缺失提醒导航入口。 |
| `apps/web-admin/app/projects/[projectId]/metrics/page.tsx` | 新增 | 项目指标表现页面骨架。 |
| `apps/web-admin/app/projects/[projectId]/metrics/input/page.tsx` | 新增 | 指标录入和批量导入页面骨架。 |
| `apps/web-admin/app/projects/[projectId]/metrics/trends/page.tsx` | 新增 | 指标趋势页面骨架。 |
| `apps/web-admin/app/projects/[projectId]/metrics/missing/page.tsx` | 新增 | 缺失数据提醒页面骨架。 |
| `.cube/iterations/feature-8/skeleton-map.yaml` | 新增 | CUBE 阶段元数据，记录骨架文件与 Development Tasks 的逐字映射。 |

### 兼容性分析

- API：全部为新增 `/api/v1` 路由，不修改既有接口请求/响应，向后兼容。
- 数据：新增表与索引，不修改既有表；通过文本 ID 关联 `content_item`、`content_version`、`publish_job`、`publish_target`、`operation_log`，不破坏既有迁移。
- 前端：项目工作区导航新增四个指标入口，不改变已有路径；新增页面可直接刷新访问。
- 安全：沿用现有 Bearer token 中间件；写接口继续使用 `Idempotency-Key` 和统一错误结构。

## 3. Flow Design

### 3.1 指标模板创建与查询

1. 前端调用 `GET /api/v1/metric-templates` 按内容类型、平台和启用状态查询模板。
2. 用户新增模板时提交内容类型、平台、指标编码、指标名称、单位、值类型、聚合方式、周期、是否必填和启用状态。
3. Handler 校验请求体并调用 `metrics.Service.CreateTemplate`。
4. Service 校验必填字段、`value_type`、`aggregation_method`、`period` 枚举和 `content_type + platform + metric_code` 唯一性。
5. 成功写入 `metric_template`，返回 `metric_template_id`；重复模板返回 `CONFLICT`，字段错误返回 `VALIDATION_ERROR`。

### 3.2 单条指标录入

1. 前端提交项目、内容单元、内容版本、发布任务、发布目标、平台、外部链接、指标编码、指标日期、周期、原始值、来源类型和来源引用。
2. Handler 读取 `Idempotency-Key`，缺失时返回 `VALIDATION_ERROR`。
3. Service 先检查幂等记录；同键同请求重放原响应，同键不同请求返回 `IDEMPOTENCY_CONFLICT`。
4. Service 根据 `project_id/content_item_id/content_version_id` 派生内容类型，再用 `content_type + platform + metric_code` 匹配指标模板；随后校验模板存在且启用、平台和周期匹配，校验来源类型和值类型，校验关联发布对象已发布。
5. Service 根据模板值类型计算 `normalized_value`，写入 `metric_record.metric_template_id` 和 `metric_record.content_type`，写 `operation_log`，保存幂等响应引用。
6. 唯一粒度相同但值不同返回 `CONFLICT`；模板不存在返回 `NOT_FOUND`，模板未启用或值类型不匹配返回 `VALIDATION_ERROR`。

### 3.3 批量导入指标

1. 前端提交 `records[]` 与 `import_source`。
2. Service 对每条记录执行与单条录入一致的校验和归一化，但批量返回逐条结果。
3. 单条失败不阻断整批；响应返回 `created_count`、`failed_count` 和 `errors[]`。
4. 若部分成功，HTTP 返回 200 Envelope 的业务结果并携带逐条失败；若全部记录业务失败，HTTP 返回失败 Envelope，错误码为 `VALIDATION_ERROR`，`error.details` 中保留逐条失败原因。
5. 批量导入写入 `operation_log`，记录导入来源、成功数、失败数和错误摘要。

### 3.4 指标记录列表

1. 前端按项目、平台、发布目标、内容单元、指标编码和日期范围查询 `GET /api/v1/metric-records`。
2. Service 使用分页、排序和白名单字段查询记录。
3. 响应展示原始值、标准化值、来源类型、来源引用、采集时间和更新时间。
4. 空数据返回空 `items` 与分页结构；无效日期或排序字段返回 `VALIDATION_ERROR`。

### 3.5 汇总与趋势

1. 前端调用 `GET /api/v1/projects/{projectId}/metrics/summary` 查询聚合指标。
2. Service 按 `project_id/platform/target_id/metric_code/date_from/date_to` 聚合，并联接 `metric_template`，逐指标使用各自的 `aggregation_method`，返回每个指标的聚合方法、来源记录数和统一 `summary_snapshot_id`。
3. Service 写入或复用 `metric_summary_snapshot`，快照 `summary` 中保存逐指标聚合口径、查询条件、来源记录数和结果，保证后续策略建议可引用同一聚合口径。
4. 趋势接口按日、周或月分桶返回 `series[]`、`missing_points[]`、`aggregation_method`、`source_record_count` 和 `query_signature`，缺失点不得按 0 处理；`query_signature` 是由项目、指标、平台、目标、日期范围和分桶生成的稳定可复现引用。

### 3.6 缺失数据提醒与补录入口

1. 前端调用 `GET /api/v1/projects/{projectId}/metrics/missing-dates`。
2. Service 以已发布的 `publish_job` 和启用且必填的 `metric_template` 为基准，限定平台、目标、统计窗口和发布完成时间。
3. Service 排除未发布、发布失败、目标禁用、模板未启用、非必填和非目标平台指标。
4. 响应返回缺失日期、缺失原因和补录提示；前端从提醒进入录入页时携带上下文参数。

### 3.7 异常流程

- 请求体、查询参数、日期范围、枚举值或 `Idempotency-Key` 缺失：`VALIDATION_ERROR`。
- 模板、发布任务或项目资源不存在：`NOT_FOUND`；模板存在但未启用、口径不匹配或值类型不匹配：`VALIDATION_ERROR`。
- 唯一粒度冲突或重复值不一致：`CONFLICT`。
- 幂等键复用但请求体不同：`IDEMPOTENCY_CONFLICT`。
- 持久化、日志或快照写入失败：`INTERNAL_ERROR`。

## 4. Table Design

### 4.1 `metric_template`

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS metric_template (
    id TEXT PRIMARY KEY,
    content_type TEXT NOT NULL,
    platform TEXT NOT NULL,
    metric_code TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    unit TEXT NOT NULL,
    value_type TEXT NOT NULL CHECK (value_type IN ('integer', 'decimal', 'percentage', 'currency', 'duration')),
    aggregation_method TEXT NOT NULL CHECK (aggregation_method IN ('sum', 'avg', 'max', 'min', 'latest')),
    period TEXT NOT NULL CHECK (period IN ('day', 'week', 'month')),
    required BOOLEAN NOT NULL DEFAULT FALSE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(content_type, platform, metric_code)
);

CREATE INDEX IF NOT EXISTS idx_metric_template_lookup ON metric_template(content_type, platform, enabled);
```

### 4.2 `metric_record`

```sql
CREATE TABLE IF NOT EXISTS metric_record (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    content_item_id TEXT NOT NULL,
    content_version_id TEXT NOT NULL,
    publish_job_id TEXT NOT NULL,
    target_id TEXT NOT NULL,
    content_type TEXT NOT NULL,
    metric_template_id TEXT NOT NULL REFERENCES metric_template(id),
    platform TEXT NOT NULL,
    external_url TEXT NOT NULL DEFAULT '',
    metric_code TEXT NOT NULL,
    metric_date DATE NOT NULL,
    period TEXT NOT NULL CHECK (period IN ('day', 'week', 'month')),
    raw_value TEXT NOT NULL,
    normalized_value NUMERIC(20, 6) NOT NULL,
    source_type TEXT NOT NULL CHECK (source_type IN ('manual', 'import', 'extension', 'external_callback')),
    source_ref TEXT NOT NULL DEFAULT '',
    collected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    operation_log_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, platform, target_id, content_version_id, metric_code, metric_date, period)
);

CREATE INDEX IF NOT EXISTS idx_metric_record_project_metric_date ON metric_record(project_id, metric_code, metric_date DESC);
CREATE INDEX IF NOT EXISTS idx_metric_record_template_date ON metric_record(metric_template_id, metric_date DESC);
CREATE INDEX IF NOT EXISTS idx_metric_record_target_date ON metric_record(project_id, platform, target_id, metric_date DESC);
CREATE INDEX IF NOT EXISTS idx_metric_record_content_item ON metric_record(content_item_id, metric_date DESC);
```

### 4.3 `metric_summary_snapshot`

```sql
CREATE TABLE IF NOT EXISTS metric_summary_snapshot (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    date_from DATE NOT NULL,
    date_to DATE NOT NULL,
    platform TEXT NOT NULL DEFAULT '',
    target_id TEXT NOT NULL DEFAULT '',
    metric_codes JSONB NOT NULL DEFAULT '[]'::jsonb,
    aggregation_method TEXT NOT NULL DEFAULT 'mixed',
    summary JSONB NOT NULL DEFAULT '{}'::jsonb,
    source_record_count INTEGER NOT NULL DEFAULT 0 CHECK (source_record_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_metric_summary_project_created ON metric_summary_snapshot(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_metric_summary_project_range ON metric_summary_snapshot(project_id, date_from, date_to);

-- +goose Down
DROP TABLE IF EXISTS metric_summary_snapshot;
DROP TABLE IF EXISTS metric_record;
DROP TABLE IF EXISTS metric_template;
```

### 4.4 幂等与审计表

本迭代不新增幂等表，复用既有 `idempotency_record(scope, endpoint, idempotency_key, request_hash, response_ref_type, response_ref_id)` 契约。`POST /metric-records` 和 `POST /metric-records/batch` 的响应快照必须可通过 `operation_log.metadata` 或业务记录恢复。指标创建、批量导入和修正必须写 `operation_log`。

### 4.5 SQL Contract

- 目标方言：PostgreSQL。
- 列表查询规则：所有查询必须限定 `project_id` 或明确的模板维度；排序字段只能从白名单映射，禁止拼接原始 query 字符串。
- 汇总查询模板：

```sql
SELECT
    r.metric_code,
    t.unit,
    t.aggregation_method,
    CASE
        WHEN t.aggregation_method = 'sum' THEN SUM(r.normalized_value)
        WHEN t.aggregation_method = 'avg' THEN AVG(r.normalized_value)
        WHEN t.aggregation_method = 'max' THEN MAX(r.normalized_value)
        WHEN t.aggregation_method = 'min' THEN MIN(r.normalized_value)
        ELSE (ARRAY_AGG(r.normalized_value ORDER BY r.metric_date DESC))[1]
    END AS value,
    COUNT(*) AS source_record_count
FROM metric_record r
JOIN metric_template t ON t.id = r.metric_template_id AND t.enabled = TRUE
WHERE r.project_id = $1
  AND r.metric_date >= $2
  AND r.metric_date <= $3
  AND ($4::text = '' OR r.platform = $4)
  AND ($5::text = '' OR r.target_id = $5)
  AND r.metric_code = ANY($6::text[])
GROUP BY r.metric_code, t.unit, t.aggregation_method;
```

- 趋势分桶 SQL 模板：

```sql
SELECT
    DATE_TRUNC($7, r.metric_date::timestamp)::date AS bucket_start,
    r.metric_code,
    t.aggregation_method,
    CASE
        WHEN t.aggregation_method = 'sum' THEN SUM(r.normalized_value)
        WHEN t.aggregation_method = 'avg' THEN AVG(r.normalized_value)
        WHEN t.aggregation_method = 'max' THEN MAX(r.normalized_value)
        WHEN t.aggregation_method = 'min' THEN MIN(r.normalized_value)
        ELSE (ARRAY_AGG(r.normalized_value ORDER BY r.metric_date DESC))[1]
    END AS value,
    COUNT(*) AS source_record_count
FROM metric_record r
JOIN metric_template t ON t.id = r.metric_template_id AND t.enabled = TRUE
WHERE r.project_id = $1
  AND r.metric_code = $2
  AND r.metric_date >= $3
  AND r.metric_date <= $4
  AND ($5::text = '' OR r.platform = $5)
  AND ($6::text = '' OR r.target_id = $6)
GROUP BY bucket_start, r.metric_code, t.aggregation_method
ORDER BY bucket_start ASC;
```

- 缺失提醒 SQL 模板：

```sql
SELECT
    j.content_item_id,
    j.content_version_id,
    j.id AS publish_job_id,
    j.target_id,
    t.platform,
    t.metric_code,
    t.period,
    expected.metric_date,
    'required_metric_missing' AS missing_reason
FROM publish_job j
JOIN publish_target pt ON pt.id = j.target_id AND pt.enabled = TRUE
JOIN content_item ci ON ci.id = j.content_item_id
JOIN content_project cp ON cp.id = j.project_id
JOIN content_type ct ON ct.id = cp.content_type_id
JOIN metric_template t
  ON t.platform = pt.platform
 AND t.content_type = ct.code
 AND t.required = TRUE
 AND t.enabled = TRUE
JOIN LATERAL generate_series(
    DATE_TRUNC(t.period, $2::date::timestamp)::date,
    DATE_TRUNC(t.period, $3::date::timestamp)::date,
    CASE t.period
        WHEN 'day' THEN '1 day'::interval
        WHEN 'week' THEN '1 week'::interval
        ELSE '1 month'::interval
    END
) AS expected(metric_date) ON TRUE
WHERE j.project_id = $1
  AND j.status = 'published'
  AND j.published_at::date <= expected.metric_date
  AND ($4::text = '' OR t.metric_code = $4)
  AND ($5::text = '' OR t.platform = $5)
  AND ($6::text = '' OR j.target_id = $6)
  AND NOT EXISTS (
      SELECT 1
      FROM metric_record r
      WHERE r.project_id = j.project_id
        AND r.content_version_id = j.content_version_id
        AND r.target_id = j.target_id
        AND r.metric_template_id = t.id
        AND r.period = t.period
        AND r.metric_date = expected.metric_date::date
  );
```

- 趋势分桶规则：`bucket=day/week/month` 分别映射到 PostgreSQL `day/week/month`，过滤必须在聚合前执行；`date_from/date_to` 为闭区间。
- 缺失提醒规则：以已发布 `publish_job.published_at` 为窗口锚点，联接启用且必填的 `metric_template`，按模板 `period=day/week/month` 生成日/周/月桶起点，通过 `NOT EXISTS` 排除已有 `metric_record`。
- 禁止模式：禁止把缺失值转为 0；禁止无 `project_id` 的记录扫描；禁止未校验排序字段；禁止将前端计算结果直接作为策略输入。

## 5. API Design

所有接口均在 `/api/v1` 下，使用 Bearer token 与统一 Envelope。

### `POST /metric-templates`

- Header：Bearer token。
- 请求：`CreateMetricTemplateRequest{content_type, platform, metric_code, metric_name, unit, value_type, aggregation_method, period, required, enabled}`。
- 201：`CreateMetricTemplateResponse{metric_template_id}`。
- 正确性规则：`content_type/platform/metric_code/metric_name/unit/value_type/aggregation_method/period` 必填；`value_type` 只能为 `integer/decimal/percentage/currency/duration`；`aggregation_method` 只能为 `sum/avg/max/min/latest`；`period` 只能为 `day/week/month`；同一 `content_type + platform + metric_code` 重复返回 `CONFLICT`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`CONFLICT`、`INTERNAL_ERROR`。

### `GET /metric-templates`

- 查询参数：`content_type`、`platform`、`enabled`、`page`、`page_size`、`sort`、`order`。
- 200：`PagedMetricTemplatesResponse{items:[MetricTemplateResponse{id,content_type,platform,metric_code,metric_name,unit,value_type,aggregation_method,period,required,enabled,updated_at}],pagination}`。
- 正确性规则：空结果返回 `items=[]`；排序字段只允许 `metric_code/updated_at/platform`；分页默认 `page=1,page_size=20`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`。

### `POST /metric-records`

- Header：Bearer token、`Idempotency-Key` 必填。
- 请求：`CreateMetricRecordRequest{project_id, content_item_id, content_version_id, publish_job_id, target_id, platform, external_url, metric_code, metric_date, period, raw_value, source_type, source_ref}`。
- 201：`CreateMetricRecordResponse{metric_record_id, normalized_value, operation_log_id}`。
- 正确性规则：`project_id/content_item_id/content_version_id/publish_job_id/target_id/platform/metric_code/metric_date/period/raw_value/source_type` 必填；`content_type` 从项目或内容版本派生；模板通过 `content_type + platform + metric_code` 匹配，落库时保存 `metric_template_id` 与 `content_type`；`source_type` 只能为 `manual/import/extension/external_callback`；相同唯一粒度重复不同值返回 `CONFLICT`；同一幂等键同请求返回相同响应。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`。

### `POST /metric-records/batch`

- Header：Bearer token、`Idempotency-Key` 必填。
- 请求：`BatchCreateMetricRecordsRequest{records[], import_source}`。
- 200：部分成功时返回 `BatchCreateMetricRecordsResponse{created_count, failed_count, errors:[BatchMetricRecordError{index,metric_code,field,code,message,source_ref}], operation_log_id}`。
- 正确性规则：`records[]` 不能为空且同一批必须属于同一 `project_id`，批量幂等范围使用该 `project_id`；每条记录按单条录入规则校验；`errors[].index` 使用请求数组下标；`BatchMetricRecordError.code` 只能为 `VALIDATION_ERROR`、`NOT_FOUND`、`CONFLICT`、`INTERNAL_ERROR`；单条失败不阻断其他记录；当 `created_count=0` 且 `failed_count>0` 时返回失败 Envelope，错误码 `VALIDATION_ERROR`，并在 `error.details` 中保留逐条失败原因；整体请求体非法或幂等冲突仍走顶层错误。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`IDEMPOTENCY_CONFLICT`、`INTERNAL_ERROR`。

### `GET /metric-records`

- 查询参数：`project_id` 必填，`platform`、`target_id`、`content_item_id`、`metric_code`、`date_from`、`date_to`、`page`、`page_size`、`sort`、`order`。
- 200：`PagedMetricRecordsResponse{items:[MetricRecordResponse{id,project_id,content_item_id,content_version_id,publish_job_id,target_id,content_type,metric_template_id,platform,external_url,metric_code,metric_date,period,raw_value,normalized_value,source_type,source_ref,collected_at,updated_at}],pagination}`。
- 正确性规则：无 `project_id` 返回 `VALIDATION_ERROR`；日期范围为闭区间；排序字段只允许 `metric_date/created_at/metric_code/platform`；空结果返回 `items=[]`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`INTERNAL_ERROR`。

### `GET /projects/{projectId}/metrics/summary`

- 查询参数：`date_from`、`date_to`、`platform`、`target_id`、`metric_codes`。
- 200：`MetricSummaryResponse{project_id,date_from,date_to,platform,target_id,items:[MetricSummaryItem{metric_code,value,unit,aggregation_method,source_record_count}],summary_snapshot_id,source_record_count}`。
- 正确性规则：`date_from/date_to/metric_codes` 必填；`metric_codes` 逗号分隔；每个指标使用自身模板 `aggregation_method`；响应必须包含 `summary_snapshot_id`；空数据返回 `items=[]` 且 `source_record_count=0`。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`。

### `GET /projects/{projectId}/metrics/trends`

- 查询参数：`metric_code`、`date_from`、`date_to`、`bucket`、`platform`、`target_id`。
- 200：`MetricTrendResponse{project_id,metric_code,bucket,aggregation_method,query_signature,source_record_count,series:[MetricTrendPoint{bucket_start,value,source_record_count,missing}],missing_points:[MetricMissingPoint{metric_date,reason}]}`。
- 正确性规则：`bucket` 只能为 `day/week/month`；`date_from/date_to` 为闭区间；`target_id` 必须参与 SQL 过滤与 `query_signature` 生成；缺失点使用 `missing=true` 和 `missing_points[]` 表达，不得转为 0；`query_signature` 必须由查询条件稳定生成。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`。

### `GET /projects/{projectId}/metrics/missing-dates`

- 查询参数：`metric_code`、`platform`、`target_id`、`date_from`、`date_to`。
- 200：`MissingMetricDatesResponse{project_id,items:[MissingMetricDateItem{content_item_id,content_version_id,publish_job_id,target_id,platform,metric_code,period,metric_date,missing_reason,backfill_hint}]}`。
- 正确性规则：`date_from/date_to` 必填且为闭区间；仅检查已发布、目标有效、模板启用且必填的指标；未发布、失败发布、非必填、模板未启用和非目标平台指标不得出现在结果中。
- 错误码：`VALIDATION_ERROR`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`INTERNAL_ERROR`。

## 6. Module Design

### 6.1 后端模块

`internal/modules/metrics`：

- `dto.go`：声明指标模板、指标记录、批量导入、汇总、趋势、缺失提醒 DTO 和枚举常量。
- `errors.go`：声明 `ErrValidation`、`ErrNotFound`、`ErrForbidden`、`ErrConflict`、`ErrIdempotencyConflict`、`ErrInternal`。
- `service.go`：声明 `Service` 接口和 `NewService()` 构造函数；骨架阶段方法体使用 `panic("not implemented")`。

Service 接口：

```go
type Service interface {
    CreateTemplate(ctx context.Context, req CreateMetricTemplateRequest) (CreateMetricTemplateResponse, error)
    ListTemplates(ctx context.Context, req ListMetricTemplatesRequest) (PagedMetricTemplatesResponse, error)
    CreateRecord(ctx context.Context, req CreateMetricRecordRequest, idempotencyKey string) (CreateMetricRecordResponse, error)
    BatchCreateRecords(ctx context.Context, req BatchCreateMetricRecordsRequest, idempotencyKey string) (BatchCreateMetricRecordsResponse, error)
    ListRecords(ctx context.Context, req ListMetricRecordsRequest) (PagedMetricRecordsResponse, error)
    GetSummary(ctx context.Context, projectID string, req MetricSummaryRequest) (MetricSummaryResponse, error)
    GetTrends(ctx context.Context, projectID string, req MetricTrendRequest) (MetricTrendResponse, error)
    GetMissingDates(ctx context.Context, projectID string, req MissingMetricDatesRequest) (MissingMetricDatesResponse, error)
}
```

`internal/http/handlers/metrics.go`：

- 负责解析 JSON、query、path 和 `Idempotency-Key`。
- 调用 `metrics.Service`。
- 将 metrics 错误映射为统一 API 错误码。

`internal/http/router.go`：

- 构造 `metrics.NewService()` 和 `handlers.NewMetricsHandler(...)`。
- 在 `/api/v1` 下注册 8 个接口。

### 6.2 前端模块

`apps/web-admin/lib/api.ts`：

- 新增 metrics 响应类型、请求类型和 API 函数。
- API 函数必须覆盖 `fetchMetricTemplates()`、`createMetricTemplate()`、`createMetricRecord()`、`batchCreateMetricRecords()`、`fetchMetricRecords()`、`fetchMetricSummary()`、`fetchMetricTrends()`、`fetchMissingMetricDates()`。
- 所有路径使用 `pathSegment` 编码项目 ID。

页面：

- `/projects/[projectId]/metrics`：调用 summary、records、missing-dates，展示指标卡片、来源数量、快照 ID、趋势入口和缺失入口。
- `/projects/[projectId]/metrics/input`：调用 templates、create template、create record、batch import、records；页面同时提供指标模板创建表单、单条录入表单、批量导入文本区、成功反馈和错误态。
- `/projects/[projectId]/metrics/trends`：调用 trends，展示分桶趋势、聚合口径、来源记录数、query signature 和缺失点。
- `/projects/[projectId]/metrics/missing`：调用 missing-dates，展示缺失原因和补录入口，补录入口跳转 `/projects/{projectId}/metrics/input` 并携带 platform、target_id、metric_code、period、metric_date。
- 原型映射：四个页面基于 `docs/requirements/ai-content-factory-clickable-prototype.html` 中“项目工作区 / 指标表现、指标录入、趋势图、缺失数据提醒”区域实现。若原型字段与本设计 API 字段不一致，以本设计 API 字段为准；页面仍需保留原型的信息层级、卡片/表格/表单/状态标签、筛选、提交、重试、详情跳转、Toast/Alert 反馈。
- 批量导入为同步 HTTP API，不引入异步 `batch-job` 或调度任务；因此不触发 `batch-job` 类型化测试。

### 6.3 集成方式

- 后端复用现有 `api.WriteSuccess`、`api.WriteError`、`parsePagination`、`decodeJSON` 和 Bearer token 中间件。
- 指标服务通过文本 ID 与 publish/content/review 产物关联；实现阶段校验跨模块状态，不在骨架阶段调用其他服务。
- 前端复用 `PageError`、`pageErrorFromEnvelope`、`PagedResponse` 和当前 CSS class。

## 7. Output Contract

workflow.yaml 的 `project.features` 当前为空，本迭代实际变更触发以下类型化测试：

| 业务描述 | type id | 跨组件 | 组件链路 | 测试规范 |
| --- | --- | --- | --- | --- |
| 指标 HTTP API 全链路 | `web-e2e` | 是 | HTTP Router -> Metrics Handler -> Metrics Service -> API Envelope | `standards/testing/web-e2e.md` |
| 指标前端页面交互 | `frontend-ui` | 是 | Web Admin Page -> API Client -> HTTP API -> UI State | `standards/testing/frontend-ui.md` |
| 指标服务到 HTTP 的集成链路 | `integration` | 是 | Metrics DTO -> Service -> Handler -> Router | `standards/testing/integration.md` |
| 指标汇总、趋势、缺失查询 SQL 契约 | `sql-query` | 是 | Metrics Request -> SQL Contract -> PostgreSQL Result Shape | `standards/testing/sql-query.md` |

### API 输出契约

- 所有成功响应必须符合 `{success:true,data,error:null,request_id}`。
- 所有失败响应必须符合 `{success:false,data:null,error:{code,message,details},request_id}`。
- 分页响应必须包含 `items` 和 `pagination{page,page_size,total,has_next}`。
- 写接口返回业务 ID、标准化值或 operation log ID，不返回裸字符串。
- 缺失点必须以结构化数组返回，不能把缺失值编码为 0。
- `CreateRecord()` 和 `BatchCreateRecords()` 的幂等范围分别为 `metrics:record:{project_id}` 与 `metrics:batch:{project_id}`；`request_hash` 使用规范化 JSON 请求体，响应引用指向 `operation_log_id` 或创建的业务记录，重放时不得重复写 `metric_record` 或 `operation_log`。

### SQL Contract 输出契约

- 目标方言：PostgreSQL。
- 汇总 SQL 必须先过滤项目、时间、平台和目标，再聚合。
- `GROUP BY` 必须包含 `metric_code` 和趋势分桶字段。
- 参数值必须与 SQL 文本分离。
- 排序字段、分桶和聚合方法必须由白名单枚举映射。
- 典型输入：`project_id=seed-project,date_from=2026-05-01,date_to=2026-05-25,platform=wechat,target_id=publish-target-1,metric_codes=views,likes`。
- 典型输出：`summary_snapshot_id=metric-summary-snapshot-1,source_record_count>0,items[].aggregation_method in sum|avg|max|min|latest`。

## 8. Change Log

| 文件 | 类型 | 原因 |
| --- | --- | --- |
| `.cube/iterations/feature-8/design.md` | 新增 | 记录 02-design 阶段技术设计。 |
| `.cube/iterations/feature-8/skeleton-map.yaml` | 新增 | 记录骨架文件与 Development Tasks 的映射。 |
| `apps/api-server/internal/modules/metrics/dto.go` | 新增 | 定义 metrics DTO、枚举和响应结构。 |
| `apps/api-server/internal/modules/metrics/errors.go` | 新增 | 定义 metrics 错误常量。 |
| `apps/api-server/internal/modules/metrics/service.go` | 新增 | 定义 metrics Service 接口和骨架实现。 |
| `apps/api-server/internal/http/handlers/metrics.go` | 新增 | 定义 metrics HTTP Handler 骨架。 |
| `apps/api-server/internal/http/router.go` | 修改 | 注册 metrics handler 和路由。 |
| `apps/api-server/migrations/00010_create_metrics_tables.sql` | 新增 | 创建指标模板、记录和汇总快照表。 |
| `openapi/openapi.yaml` | 修改 | 增加 metrics API 契约。 |
| `apps/web-admin/lib/api.ts` | 修改 | 增加 metrics API client 类型与函数。 |
| `apps/web-admin/app/projects/[projectId]/workspace-nav.tsx` | 修改 | 增加指标页面导航入口。 |
| `apps/web-admin/app/projects/[projectId]/metrics/page.tsx` | 新增 | 指标表现页面骨架。 |
| `apps/web-admin/app/projects/[projectId]/metrics/input/page.tsx` | 新增 | 指标录入页面骨架。 |
| `apps/web-admin/app/projects/[projectId]/metrics/trends/page.tsx` | 新增 | 趋势图页面骨架。 |
| `apps/web-admin/app/projects/[projectId]/metrics/missing/page.tsx` | 新增 | 缺失数据提醒页面骨架。 |

## 9. Development Tasks

- Task-01：创建指标数据表与 SQL 查询契约
  - 所属模块：api-server/store
  - 简要描述：新增 metric_template、metric_record、metric_summary_snapshot 迁移，并固化汇总、趋势、缺失查询 SQL 契约。
  - 涉及接口/方法：00010_create_metrics_tables.sql
  - 输入：goose migration
  - 输出：PostgreSQL 表、约束和索引
  - 产出类型：sql-query
  - 功能类型：指标 SQL 契约（type id: sql-query）
  - 是否跨组件：是（组件链路：Metrics Request -> SQL Contract -> PostgreSQL）
- Task-02：定义指标模块 DTO、错误常量与服务接口
  - 所属模块：api-server/metrics
  - 简要描述：定义指标模板、指标记录、批量导入、汇总、趋势、缺失提醒 DTO，声明 Service 接口和错误常量。
  - 涉及接口/方法：metrics.Service、NewService()
  - 输入：各 API request DTO
  - 输出：各 API response DTO 或 error
  - 产出类型：integration
  - 功能类型：后端模块接口契约（type id: integration）
  - 是否跨组件：否
- Task-03：实现指标模板 API 骨架
  - 所属模块：api-server/http
  - 简要描述：提供创建和查询指标模板的 handler、路由和错误映射骨架。
  - 涉及接口/方法：MetricsHandler.CreateTemplate()、MetricsHandler.ListTemplates()
  - 输入：HTTP request、query、JSON body
  - 输出：统一 Envelope 响应
  - 产出类型：web-e2e
  - 功能类型：Web/API 端点（type id: web-e2e）
  - 是否跨组件：是（组件链路：Router -> MetricsHandler -> MetricsService）
- Task-04：实现指标记录录入与批量导入 API 骨架
  - 所属模块：api-server/http
  - 简要描述：提供单条录入、批量导入、幂等键读取和错误映射骨架。
  - 涉及接口/方法：MetricsHandler.CreateRecord()、MetricsHandler.BatchCreateRecords()
  - 输入：HTTP request、Idempotency-Key、JSON body
  - 输出：统一 Envelope 响应
  - 产出类型：web-e2e
  - 功能类型：Web/API 端点（type id: web-e2e）
  - 是否跨组件：是（组件链路：Router -> MetricsHandler -> MetricsService）
- Task-05：实现指标记录列表、汇总、趋势与缺失 API 骨架
  - 所属模块：api-server/http
  - 简要描述：提供记录列表、项目汇总、趋势序列和缺失日期 handler 与路由骨架。
  - 涉及接口/方法：MetricsHandler.ListRecords()、GetSummary()、GetTrends()、GetMissingDates()
  - 输入：HTTP request、path、query
  - 输出：统一 Envelope 响应
  - 产出类型：web-e2e
  - 功能类型：Web/API 端点（type id: web-e2e）
  - 是否跨组件：是（组件链路：Router -> MetricsHandler -> MetricsService）
- Task-06：补充 OpenAPI 指标接口契约
  - 所属模块：openapi
  - 简要描述：在 OpenAPI 中声明指标模板、指标记录、批量导入、汇总、趋势、缺失提醒路径、schema、错误响应和鉴权。
  - 涉及接口/方法：openapi.yaml
  - 输入：API Design
  - 输出：OpenAPI 3.0 契约
  - 产出类型：web-e2e
  - 功能类型：Web/API 契约（type id: web-e2e）
  - 是否跨组件：否
- Task-07：实现前端指标 API client 骨架
  - 所属模块：web-admin/api
  - 简要描述：在前端 API client 中增加指标模板创建/查询、记录录入、批量导入、记录列表、汇总、趋势和缺失提醒类型与调用函数。
  - 涉及接口/方法：fetchMetricTemplates()、createMetricTemplate()、createMetricRecord()、batchCreateMetricRecords()、fetchMetricRecords()、fetchMetricSummary()、fetchMetricTrends()、fetchMissingMetricDates()
  - 输入：页面参数和表单输入
  - 输出：APIEnvelope metrics response
  - 产出类型：frontend-ui
  - 功能类型：前端 API client（type id: frontend-ui）
  - 是否跨组件：是（组件链路：Web Page -> API Client -> HTTP API）
- Task-08：实现指标工作区导航与指标表现页面骨架
  - 所属模块：web-admin/pages
  - 简要描述：增加项目工作区指标导航和指标表现页面，基于原型展示汇总、快照 ID、来源记录数、趋势与缺失入口。
  - 涉及接口/方法：MetricsPage()
  - 输入：projectId route params
  - 输出：可渲染 Next.js page
  - 产出类型：frontend-ui
  - 功能类型：前端页面（type id: frontend-ui）
  - 是否跨组件：是（组件链路：Next Route -> API Client -> Metrics API）
- Task-09：实现指标录入页面骨架
  - 所属模块：web-admin/pages
  - 简要描述：基于原型实现模板创建、单条录入和批量导入表单页面骨架，展示模板选择、成功反馈、逐条错误和错误态。
  - 涉及接口/方法：MetricInputPage()
  - 输入：projectId route params、form state
  - 输出：可渲染 Next.js page
  - 产出类型：frontend-ui
  - 功能类型：前端页面（type id: frontend-ui）
  - 是否跨组件：是（组件链路：Next Route -> API Client -> Metrics API）
- Task-10：实现趋势图与缺失提醒页面骨架
  - 所属模块：web-admin/pages
  - 简要描述：基于原型实现趋势序列页面和缺失提醒页面骨架，展示聚合口径、query signature、缺失点、缺失原因和补录入口。
  - 涉及接口/方法：MetricTrendsPage()、MissingMetricsPage()
  - 输入：projectId route params、filter state
  - 输出：可渲染 Next.js pages
  - 产出类型：frontend-ui
  - 功能类型：前端页面（type id: frontend-ui）
  - 是否跨组件：是（组件链路：Next Route -> API Client -> Metrics API）
