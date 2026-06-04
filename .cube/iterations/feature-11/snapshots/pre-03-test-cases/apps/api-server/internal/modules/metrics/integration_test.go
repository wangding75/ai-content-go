package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func metricsTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbURL := strings.TrimSpace(os.Getenv("METRICS_TEST_DATABASE_URL"))
	if dbURL == "" {
		t.Skip("METRICS_TEST_DATABASE_URL not set, skipping PostgreSQL integration test")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("ping database: %v", err)
	}
	return db
}

func runMigrations(t *testing.T, db *sql.DB) {
	t.Helper()
	migrationsDir := "../../migrations"
	files := []string{
		"00001_create_operation_log.sql",
		"00002_create_content_entry_tables.sql",
		"00003_create_workflow_tables.sql",
		"00004_create_iteration_2_1_tables.sql",
		"00005_create_novel_planning_tables.sql",
		"00006_create_content_generation_tables.sql",
		"00007_create_content_review_tables.sql",
		"00008_create_knowledge_memory_tables.sql",
		"00009_create_publish_tables.sql",
		"00010_create_metrics_tables.sql",
	}
	for _, file := range files {
		content, err := os.ReadFile(fmt.Sprintf("%s/%s", migrationsDir, file))
		if err != nil {
			t.Fatalf("read migration %s: %v", file, err)
		}
		sqlContent := string(content)
		upIdx := strings.Index(strings.ToLower(sqlContent), "-- +goose up")
		downIdx := strings.Index(strings.ToLower(sqlContent), "-- +goose down")
		if upIdx >= 0 && downIdx > upIdx {
			sqlContent = sqlContent[upIdx+len("-- +goose up") : downIdx]
		}
		if _, err := db.Exec(sqlContent); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				t.Fatalf("run migration %s: %v", file, err)
			}
		}
	}
}

func cleanupMetricsTables(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, table := range []string{
		"metric_record", "metric_template", "metric_summary_snapshot",
		"idempotency_record",
		"publish_job", "publish_target",
		"content_version", "content_item", "content_project", "content_type",
		"operation_log",
	} {
		db.Exec(fmt.Sprintf("DELETE FROM %s", table))
	}
}

// @Test
func TestIntegrationPostgresStoreCreateAndFindTemplate(t *testing.T) {
	db := metricsTestDB(t)
	defer db.Close()
	runMigrations(t, db)
	cleanupMetricsTables(t, db)

	store := NewPostgresStore(db)
	tpl := MetricTemplateResponse{
		ID:                "metric-template-article-wechat-views",
		ContentType:       "article",
		Platform:          "wechat",
		MetricCode:        "views",
		MetricName:        "阅读量",
		Unit:              "次",
		ValueType:         ValueTypeInteger,
		AggregationMethod: AggregationSum,
		Period:            PeriodDay,
		Required:          true,
		Enabled:           true,
		UpdatedAt:         time.Now().UTC(),
	}
	if err := store.InsertTemplate(context.Background(), tpl); err != nil {
		t.Fatalf("InsertTemplate: %v", err)
	}

	found, err := store.FindTemplateByKey(context.Background(), "article", "wechat", "views")
	if err != nil {
		t.Fatalf("FindTemplateByKey: %v", err)
	}
	if found == nil {
		t.Fatal("FindTemplateByKey returned nil")
	}
	if found.MetricCode != "views" {
		t.Fatalf("expected metric_code=views, got %s", found.MetricCode)
	}
}

// @Test
func TestIntegrationPostgresStoreCreateAndFindRecord(t *testing.T) {
	db := metricsTestDB(t)
	defer db.Close()
	runMigrations(t, db)
	cleanupMetricsTables(t, db)

	store := NewPostgresStore(db)
	ctx := context.Background()

	tpl := MetricTemplateResponse{
		ID:                "metric-template-article-wechat-views",
		ContentType:       "article",
		Platform:          "wechat",
		MetricCode:        "views",
		MetricName:        "阅读量",
		Unit:              "次",
		ValueType:         ValueTypeInteger,
		AggregationMethod: AggregationSum,
		Period:            PeriodDay,
		Required:          true,
		Enabled:           true,
		UpdatedAt:         time.Now().UTC(),
	}
	if err := store.InsertTemplate(ctx, tpl); err != nil {
		t.Fatalf("InsertTemplate: %v", err)
	}

	rec := MetricRecordResponse{
		ID:               "metric-record-int-views-20260525",
		ProjectID:        "project-int",
		ContentItemID:    "content-item-1",
		ContentVersionID: "content-version-1",
		PublishJobID:     "publish-job-1",
		TargetID:         "publish-target-1",
		ContentType:      "article",
		MetricTemplateID: tpl.ID,
		Platform:         "wechat",
		MetricCode:       "views",
		MetricDate:       "2026-05-25",
		Period:           PeriodDay,
		RawValue:         "1200",
		NormalizedValue:  1200,
		SourceType:       SourceManual,
		CollectedAt:      time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := store.InsertRecord(ctx, rec); err != nil {
		t.Fatalf("InsertRecord: %v", err)
	}

	found, err := store.FindRecordByUniqueKey(ctx, "project-int", "wechat", "publish-target-1", "content-version-1", "views", "2026-05-25", PeriodDay)
	if err != nil {
		t.Fatalf("FindRecordByUniqueKey: %v", err)
	}
	if found == nil {
		t.Fatal("FindRecordByUniqueKey returned nil")
	}
	if found.NormalizedValue != 1200 {
		t.Fatalf("expected normalized_value=1200, got %f", found.NormalizedValue)
	}
}

// @Test
func TestIntegrationPostgresStoreIdempotency(t *testing.T) {
	db := metricsTestDB(t)
	defer db.Close()
	runMigrations(t, db)
	cleanupMetricsTables(t, db)

	store := NewPostgresStore(db)
	ctx := context.Background()

	refType, refID, conflict, err := store.CheckIdempotency(ctx, "metrics:record:project-1", "create_record", "key-1", "hash-abc")
	if err != nil {
		t.Fatalf("CheckIdempotency empty: %v", err)
	}
	if refType != "" || refID != "" || conflict {
		t.Fatal("expected empty result for new idempotency key")
	}

	if err := store.StoreIdempotency(ctx, "metrics:record:project-1", "create_record", "key-1", "hash-abc", "metric_record", "metric-record-1"); err != nil {
		t.Fatalf("StoreIdempotency: %v", err)
	}

	refType, refID, conflict, err = store.CheckIdempotency(ctx, "metrics:record:project-1", "create_record", "key-1", "hash-abc")
	if err != nil {
		t.Fatalf("CheckIdempotency after store: %v", err)
	}
	if refType != "metric_record" || refID != "metric-record-1" {
		t.Fatalf("expected (metric_record, metric-record-1), got (%s, %s)", refType, refID)
	}

	_, _, conflict, err = store.CheckIdempotency(ctx, "metrics:record:project-1", "create_record", "key-1", "hash-different")
	if err != nil {
		t.Fatalf("CheckIdempotency conflict: %v", err)
	}
	if !conflict {
		t.Fatal("expected conflict for different hash with same key")
	}
}

// @Test
func TestIntegrationPostgresStoreSummaryAggregation(t *testing.T) {
	db := metricsTestDB(t)
	defer db.Close()
	runMigrations(t, db)
	cleanupMetricsTables(t, db)

	store := NewPostgresStore(db)
	ctx := context.Background()

	tplAvg := MetricTemplateResponse{
		ID:                "metric-template-article-wechat-avg-read",
		ContentType:       "article",
		Platform:          "wechat",
		MetricCode:        "avg_read_time",
		MetricName:        "平均阅读时长",
		Unit:              "秒",
		ValueType:         ValueTypeDecimal,
		AggregationMethod: AggregationAvg,
		Period:            PeriodDay,
		Required:          false,
		Enabled:           true,
		UpdatedAt:         time.Now().UTC(),
	}
	tplSum := MetricTemplateResponse{
		ID:                "metric-template-article-wechat-views",
		ContentType:       "article",
		Platform:          "wechat",
		MetricCode:        "views",
		MetricName:        "阅读量",
		Unit:              "次",
		ValueType:         ValueTypeInteger,
		AggregationMethod: AggregationSum,
		Period:            PeriodDay,
		Required:          true,
		Enabled:           true,
		UpdatedAt:         time.Now().UTC(),
	}
	if err := store.InsertTemplate(ctx, tplAvg); err != nil {
		t.Fatalf("InsertTemplate avg: %v", err)
	}
	if err := store.InsertTemplate(ctx, tplSum); err != nil {
		t.Fatalf("InsertTemplate sum: %v", err)
	}

	for i, val := range []float64{100, 200, 300} {
		rec := MetricRecordResponse{
			ID:               fmt.Sprintf("metric-record-avg-%d", i),
			ProjectID:        "project-summary",
			ContentItemID:    "content-item-1",
			ContentVersionID: "content-version-1",
			PublishJobID:     "publish-job-1",
			TargetID:         "publish-target-1",
			ContentType:      "article",
			MetricTemplateID: tplAvg.ID,
			Platform:         "wechat",
			MetricCode:       "avg_read_time",
			MetricDate:       "2026-05-25",
			Period:           PeriodDay,
			RawValue:         fmt.Sprintf("%.0f", val),
			NormalizedValue:  val,
			SourceType:       SourceManual,
			CollectedAt:      time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		}
		if err := store.InsertRecord(ctx, rec); err != nil {
			t.Fatalf("InsertRecord avg %d: %v", i, err)
		}
	}

	recSum := MetricRecordResponse{
		ID:               "metric-record-views-1",
		ProjectID:        "project-summary",
		ContentItemID:    "content-item-1",
		ContentVersionID: "content-version-1",
		PublishJobID:     "publish-job-1",
		TargetID:         "publish-target-1",
		ContentType:      "article",
		MetricTemplateID: tplSum.ID,
		Platform:         "wechat",
		MetricCode:       "views",
		MetricDate:       "2026-05-25",
		Period:           PeriodDay,
		RawValue:         "1200",
		NormalizedValue:  1200,
		SourceType:       SourceManual,
		CollectedAt:      time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	if err := store.InsertRecord(ctx, recSum); err != nil {
		t.Fatalf("InsertRecord sum: %v", err)
	}

	items, total, err := store.QuerySummary(ctx, "project-summary", MetricSummaryRequest{
		DateFrom:    "2026-05-25",
		DateTo:      "2026-05-25",
		Platform:    "wechat",
		MetricCodes: []string{"views", "avg_read_time"},
	})
	if err != nil {
		t.Fatalf("QuerySummary: %v", err)
	}
	if total == 0 {
		t.Fatal("QuerySummary returned 0 source records")
	}

	summaryMap := map[string]MetricSummaryItem{}
	for _, item := range items {
		summaryMap[item.MetricCode] = item
	}
	if avgItem, ok := summaryMap["avg_read_time"]; ok {
		if avgItem.AggregationMethod != AggregationAvg {
			t.Fatalf("expected avg aggregation, got %s", avgItem.AggregationMethod)
		}
		if avgItem.Value != 200 {
			t.Fatalf("expected avg value 200, got %f", avgItem.Value)
		}
	} else {
		t.Fatal("avg_read_time missing from summary")
	}
	if sumItem, ok := summaryMap["views"]; ok {
		if sumItem.AggregationMethod != AggregationSum {
			t.Fatalf("expected sum aggregation, got %s", sumItem.AggregationMethod)
		}
		if sumItem.Value != 1200 {
			t.Fatalf("expected sum value 1200, got %f", sumItem.Value)
		}
	} else {
		t.Fatal("views missing from summary")
	}
}
