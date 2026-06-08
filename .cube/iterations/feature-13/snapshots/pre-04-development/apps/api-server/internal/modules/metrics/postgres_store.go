package metrics

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type postgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) Store {
	return &postgresStore{db: db}
}

func (p *postgresStore) InsertTemplate(ctx context.Context, t MetricTemplateResponse) error {
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO metric_template (id, content_type, platform, metric_code, metric_name, unit, value_type, aggregation_method, period, required, enabled, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, t.ID, t.ContentType, t.Platform, t.MetricCode, t.MetricName, t.Unit, t.ValueType, t.AggregationMethod, t.Period, t.Required, t.Enabled, t.UpdatedAt)
	if err != nil {
		if isDuplicateKey(err) {
			return ErrConflict
		}
		return err
	}
	return nil
}

func (p *postgresStore) FindTemplateByKey(ctx context.Context, contentType, platform, metricCode string) (*MetricTemplateResponse, error) {
	var t MetricTemplateResponse
	err := p.db.QueryRowContext(ctx, `
		SELECT id, content_type, platform, metric_code, metric_name, unit, value_type, aggregation_method, period, required, enabled, updated_at
		FROM metric_template
		WHERE content_type = $1 AND platform = $2 AND metric_code = $3
	`, contentType, platform, metricCode).Scan(
		&t.ID, &t.ContentType, &t.Platform, &t.MetricCode, &t.MetricName, &t.Unit,
		&t.ValueType, &t.AggregationMethod, &t.Period, &t.Required, &t.Enabled, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (p *postgresStore) ListTemplates(ctx context.Context, req ListMetricTemplatesRequest) ([]MetricTemplateResponse, error) {
	query := `SELECT id, content_type, platform, metric_code, metric_name, unit, value_type, aggregation_method, period, required, enabled, updated_at FROM metric_template WHERE 1=1`
	var args []any
	argIdx := 1
	if req.ContentType != "" {
		query += fmt.Sprintf(" AND content_type = $%d", argIdx)
		args = append(args, req.ContentType)
		argIdx++
	}
	if req.Platform != "" {
		query += fmt.Sprintf(" AND platform = $%d", argIdx)
		args = append(args, req.Platform)
		argIdx++
	}
	if req.Enabled != nil {
		query += fmt.Sprintf(" AND enabled = $%d", argIdx)
		args = append(args, *req.Enabled)
		argIdx++
	}
	query += " ORDER BY metric_code ASC"

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]MetricTemplateResponse, 0)
	for rows.Next() {
		var t MetricTemplateResponse
		if err := rows.Scan(&t.ID, &t.ContentType, &t.Platform, &t.MetricCode, &t.MetricName, &t.Unit,
			&t.ValueType, &t.AggregationMethod, &t.Period, &t.Required, &t.Enabled, &t.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (p *postgresStore) InsertRecord(ctx context.Context, r MetricRecordResponse) error {
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO metric_record (
			id, project_id, content_item_id, content_version_id, publish_job_id, target_id,
			content_type, metric_template_id, platform, external_url, metric_code, metric_date,
			period, raw_value, normalized_value, source_type, source_ref, collected_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`, r.ID, r.ProjectID, r.ContentItemID, r.ContentVersionID, r.PublishJobID, r.TargetID,
		r.ContentType, r.MetricTemplateID, r.Platform, r.ExternalURL, r.MetricCode, r.MetricDate,
		r.Period, r.RawValue, r.NormalizedValue, r.SourceType, r.SourceRef, r.CollectedAt, r.UpdatedAt)
	if err != nil {
		if isDuplicateKey(err) {
			return ErrConflict
		}
		return err
	}
	return nil
}

func (p *postgresStore) FindRecordByUniqueKey(ctx context.Context, projectID, platform, targetID, contentVersionID, metricCode, metricDate, period string) (*MetricRecordResponse, error) {
	var r MetricRecordResponse
	err := p.db.QueryRowContext(ctx, `
		SELECT id, project_id, content_item_id, content_version_id, publish_job_id, target_id,
			content_type, metric_template_id, platform, external_url, metric_code, metric_date,
			period, raw_value, normalized_value, source_type, source_ref, collected_at, updated_at
		FROM metric_record
		WHERE project_id = $1 AND platform = $2 AND target_id = $3 AND content_version_id = $4
		  AND metric_code = $5 AND metric_date = $6 AND period = $7
	`, projectID, platform, targetID, contentVersionID, metricCode, metricDate, period).Scan(
		&r.ID, &r.ProjectID, &r.ContentItemID, &r.ContentVersionID, &r.PublishJobID, &r.TargetID,
		&r.ContentType, &r.MetricTemplateID, &r.Platform, &r.ExternalURL, &r.MetricCode, &r.MetricDate,
		&r.Period, &r.RawValue, &r.NormalizedValue, &r.SourceType, &r.SourceRef, &r.CollectedAt, &r.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (p *postgresStore) ListRecords(ctx context.Context, req ListMetricRecordsRequest) ([]MetricRecordResponse, int, error) {
	query := `SELECT id, project_id, content_item_id, content_version_id, publish_job_id, target_id,
		content_type, metric_template_id, platform, external_url, metric_code, metric_date,
		period, raw_value, normalized_value, source_type, source_ref, collected_at, updated_at
		FROM metric_record WHERE project_id = $1`
	countQuery := `SELECT COUNT(*) FROM metric_record WHERE project_id = $1`
	var args []any
	var countArgs []any
	args = append(args, req.ProjectID)
	countArgs = append(countArgs, req.ProjectID)
	argIdx := 2

	if req.Platform != "" {
		query += fmt.Sprintf(" AND platform = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND platform = $%d", argIdx)
		args = append(args, req.Platform)
		countArgs = append(countArgs, req.Platform)
		argIdx++
	}
	if req.TargetID != "" {
		query += fmt.Sprintf(" AND target_id = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND target_id = $%d", argIdx)
		args = append(args, req.TargetID)
		countArgs = append(countArgs, req.TargetID)
		argIdx++
	}
	if req.ContentItemID != "" {
		query += fmt.Sprintf(" AND content_item_id = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND content_item_id = $%d", argIdx)
		args = append(args, req.ContentItemID)
		countArgs = append(countArgs, req.ContentItemID)
		argIdx++
	}
	if req.MetricCode != "" {
		query += fmt.Sprintf(" AND metric_code = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND metric_code = $%d", argIdx)
		args = append(args, req.MetricCode)
		countArgs = append(countArgs, req.MetricCode)
		argIdx++
	}
	if req.DateFrom != "" {
		query += fmt.Sprintf(" AND metric_date >= $%d", argIdx)
		countQuery += fmt.Sprintf(" AND metric_date >= $%d", argIdx)
		args = append(args, req.DateFrom)
		countArgs = append(countArgs, req.DateFrom)
		argIdx++
	}
	if req.DateTo != "" {
		query += fmt.Sprintf(" AND metric_date <= $%d", argIdx)
		countQuery += fmt.Sprintf(" AND metric_date <= $%d", argIdx)
		args = append(args, req.DateTo)
		countArgs = append(countArgs, req.DateTo)
		argIdx++
	}

	sortCol := "metric_date"
	if col, ok := metricSortColumns[req.Sort]; ok && col != "" {
		sortCol = col
	}
	order := "DESC"
	if strings.EqualFold(req.Order, "asc") {
		order = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortCol, order)

	var total int
	if err := p.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]MetricRecordResponse, 0)
	for rows.Next() {
		var r MetricRecordResponse
		if err := rows.Scan(&r.ID, &r.ProjectID, &r.ContentItemID, &r.ContentVersionID, &r.PublishJobID, &r.TargetID,
			&r.ContentType, &r.MetricTemplateID, &r.Platform, &r.ExternalURL, &r.MetricCode, &r.MetricDate,
			&r.Period, &r.RawValue, &r.NormalizedValue, &r.SourceType, &r.SourceRef, &r.CollectedAt, &r.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, r)
	}
	return items, total, rows.Err()
}

func (p *postgresStore) InsertSummarySnapshot(ctx context.Context, snap SummarySnapshotRow) error {
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO metric_summary_snapshot (id, project_id, date_from, date_to, platform, target_id, metric_codes, aggregation_method, summary, source_record_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, snap.ID, snap.ProjectID, snap.DateFrom, snap.DateTo, snap.Platform, snap.TargetID, snap.MetricCodes, snap.AggregationMethod, snap.Summary, snap.SourceRecordCount)
	return err
}

func (p *postgresStore) CheckIdempotency(ctx context.Context, scope, endpoint, key, hash string) (string, string, bool, error) {
	var refType, refID, storedHash string
	err := p.db.QueryRowContext(ctx, `
		SELECT response_ref_type, response_ref_id, request_hash FROM idempotency_record
		WHERE scope = $1 AND endpoint = $2 AND idempotency_key = $3
	`, scope, endpoint, key).Scan(&refType, &refID, &storedHash)
	if err == sql.ErrNoRows {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, err
	}
	if storedHash != hash {
		return "", "", true, nil
	}
	return refType, refID, false, nil
}

func (p *postgresStore) StoreIdempotency(ctx context.Context, scope, endpoint, key, hash, refType, refID string) error {
	id := scope + ":" + endpoint + ":" + key
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO idempotency_record (id, scope, endpoint, idempotency_key, request_hash, response_ref_type, response_ref_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`, id, scope, endpoint, key, hash, refType, refID)
	return err
}

func (p *postgresStore) QuerySummary(ctx context.Context, projectID string, req MetricSummaryRequest) ([]MetricSummaryItem, int, error) {
	args := []any{projectID, req.DateFrom, req.DateTo, req.Platform, req.TargetID, pq.Array(req.MetricCodes)}
	rows, err := p.db.QueryContext(ctx, metricSummarySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]MetricSummaryItem, 0)
	total := 0
	for rows.Next() {
		var item MetricSummaryItem
		var sourceCount int
		if err := rows.Scan(&item.MetricCode, &item.Unit, &item.AggregationMethod, &item.Value, &sourceCount); err != nil {
			return nil, 0, err
		}
		item.SourceRecordCount = sourceCount
		total += sourceCount
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (p *postgresStore) QueryTrends(ctx context.Context, projectID string, req MetricTrendRequest) ([]MetricTrendPoint, []MetricMissingPoint, string, int, error) {
	bucketCol := "day"
	if col, ok := metricBucketColumns[req.Bucket]; ok {
		bucketCol = col
	}
	args := []any{projectID, req.MetricCode, req.DateFrom, req.DateTo, req.Platform, req.TargetID, bucketCol}
	rows, err := p.db.QueryContext(ctx, metricTrendSQL, args...)
	if err != nil {
		return nil, nil, "", 0, err
	}
	defer rows.Close()

	series := make([]MetricTrendPoint, 0)
	sourceCount := 0
	for rows.Next() {
		var point MetricTrendPoint
		var aggMethod string
		var recCount int
		var bucketStart time.Time
		if err := rows.Scan(&bucketStart, &aggMethod, &point.Value, &recCount); err != nil {
			return nil, nil, "", 0, err
		}
		point.BucketStart = bucketStart.Format("2006-01-02")
		point.SourceRecordCount = recCount
		sourceCount += recCount
		series = append(series, point)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, "", 0, err
	}

	signature := querySignature(projectID, req)
	return series, []MetricMissingPoint{}, signature, sourceCount, nil
}

func (p *postgresStore) QueryMissingDates(ctx context.Context, projectID string, req MissingMetricDatesRequest) ([]MissingMetricDateItem, error) {
	args := []any{projectID, req.DateFrom, req.DateTo, req.MetricCode, req.Platform, req.TargetID}
	rows, err := p.db.QueryContext(ctx, metricMissingDatesSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]MissingMetricDateItem, 0)
	for rows.Next() {
		var item MissingMetricDateItem
		if err := rows.Scan(&item.ContentItemID, &item.ContentVersionID, &item.PublishJobID,
			&item.TargetID, &item.Platform, &item.MetricCode, &item.Period, &item.MetricDate, &item.MissingReason); err != nil {
			return nil, err
		}
		item.BackfillHint = "backfill " + item.MetricCode + " " + item.MetricDate
		items = append(items, item)
	}
	return items, rows.Err()
}

func isDuplicateKey(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}

func (p *postgresStore) InsertPlatformCollectLog(ctx context.Context, log PlatformCollectLogDetailResponse) error {
	return errors.New("not implemented")
}

func (p *postgresStore) ListPlatformCollectLogs(ctx context.Context, req ListPlatformCollectLogsRequest) ([]PlatformCollectLogResponse, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (p *postgresStore) GetPlatformCollectLog(ctx context.Context, collectLogID string) (*PlatformCollectLogDetailResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *postgresStore) UpdatePlatformCollectLogStatus(ctx context.Context, collectLogID string, status string, operationLogID string) error {
	return errors.New("not implemented")
}
