package metrics

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func hasMetricJSONField(t reflect.Type, jsonName string) bool {
	for i := 0; i < t.NumField(); i++ {
		name := strings.Split(t.Field(i).Tag.Get("json"), ",")[0]
		if name == jsonName {
			return true
		}
	}
	return false
}

// @Test
func TestTask02MetricDTOsExposeTemplateRecordAndAggregationFields(t *testing.T) {
	templateType := reflect.TypeOf(CreateMetricTemplateRequest{})
	for _, field := range []string{"content_type", "platform", "metric_code", "metric_name", "unit", "value_type", "aggregation_method", "period", "required", "enabled"} {
		if !hasMetricJSONField(templateType, field) {
			t.Fatalf("CreateMetricTemplateRequest missing json field %q", field)
		}
	}
	recordType := reflect.TypeOf(MetricRecordResponse{})
	for _, field := range []string{"content_type", "metric_template_id", "normalized_value", "source_type", "source_ref", "collected_at"} {
		if !hasMetricJSONField(recordType, field) {
			t.Fatalf("MetricRecordResponse missing json field %q", field)
		}
	}
}

// @Test
func TestTask02MetricConstantsAndErrorsAreStableContracts(t *testing.T) {
	for _, value := range []string{
		ValueTypeInteger, ValueTypeDecimal, ValueTypePercentage, ValueTypeCurrency, ValueTypeDuration,
		AggregationSum, AggregationAvg, AggregationMax, AggregationMin, AggregationLatest,
		PeriodDay, PeriodWeek, PeriodMonth,
		SourceManual, SourceImport, SourceExtension, SourceExternalCallback,
	} {
		if value == "" {
			t.Fatalf("metrics enum constants must be non-empty")
		}
	}
	for _, errValue := range []error{ErrValidation, ErrNotFound, ErrForbidden, ErrConflict, ErrIdempotencyConflict, ErrInternal} {
		if errValue == nil {
			t.Fatalf("metrics domain errors must be declared")
		}
	}
}

// @Test
func TestTask02ServiceInterfaceDeclaresAllMetricUseCases(t *testing.T) {
	serviceType := reflect.TypeOf((*Service)(nil)).Elem()
	for _, method := range []string{
		"CreateTemplate",
		"ListTemplates",
		"CreateRecord",
		"BatchCreateRecords",
		"ListRecords",
		"GetSummary",
		"GetTrends",
		"GetMissingDates",
	} {
		if _, ok := serviceType.MethodByName(method); !ok {
			t.Fatalf("metrics Service missing method %s", method)
		}
	}
}

// @Test
func TestTask02ServiceRejectsMissingTemplateRequiredFields(t *testing.T) {
	_, err := NewService().CreateTemplate(context.Background(), CreateMetricTemplateRequest{
		ContentType:       "",
		Platform:          "wechat",
		MetricCode:        "views",
		MetricName:        "阅读量",
		Unit:              "次",
		ValueType:         ValueTypeInteger,
		AggregationMethod: AggregationSum,
		Period:            PeriodDay,
		Required:          true,
		Enabled:           true,
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("missing content_type must return ErrValidation, got %v", err)
	}
}

// @Test
func TestTask02ServiceWritesThenReadsMetricTemplateAndRecord(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	template, err := svc.CreateTemplate(ctx, CreateMetricTemplateRequest{
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
	})
	if err != nil {
		t.Fatalf("create metric template must succeed: %v", err)
	}
	if template.MetricTemplateID == "" {
		t.Fatalf("create metric template must return metric_template_id")
	}

	created, err := svc.CreateRecord(ctx, CreateMetricRecordRequest{
		ProjectID:        "seed-project",
		ContentItemID:    "content-item-1",
		ContentVersionID: "content-version-approved-1",
		PublishJobID:     "publish-job-1",
		TargetID:         "publish-target-1",
		Platform:         "wechat",
		ExternalURL:      "https://example.com/post/1",
		MetricCode:       "views",
		MetricDate:       "2026-05-25",
		Period:           PeriodDay,
		RawValue:         "1200",
		SourceType:       SourceManual,
		SourceRef:        "manual-entry",
	}, "metric-record-write-read")
	if err != nil {
		t.Fatalf("create metric record must succeed: %v", err)
	}
	if created.MetricRecordID == "" || created.NormalizedValue != 1200 || created.OperationLogID == "" {
		t.Fatalf("create metric record must return id, normalized_value and operation_log_id: %#v", created)
	}

	records, err := svc.ListRecords(ctx, ListMetricRecordsRequest{
		ProjectID:  "seed-project",
		MetricCode: "views",
		DateFrom:   "2026-05-01",
		DateTo:     "2026-05-25",
	})
	if err != nil {
		t.Fatalf("list metric records must succeed after write: %v", err)
	}
	if len(records.Items) != 1 || records.Items[0].MetricTemplateID != template.MetricTemplateID || records.Items[0].ContentType != "article" {
		t.Fatalf("list records must read the just-written metric with template and content type: %#v", records.Items)
	}
}

// @Test
func TestTask02ServiceBuildsSummaryTrendMissingFromPersistedRecords(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	template, err := svc.CreateTemplate(ctx, CreateMetricTemplateRequest{
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
	})
	if err != nil {
		t.Fatalf("create metric template must succeed before summary/trend/missing assertions: %v", err)
	}
	if template.MetricTemplateID == "" {
		t.Fatalf("create metric template must return metric_template_id before summary/trend/missing assertions")
	}
	record, err := svc.CreateRecord(ctx, CreateMetricRecordRequest{
		ProjectID:        "seed-project",
		ContentItemID:    "content-item-1",
		ContentVersionID: "content-version-approved-1",
		PublishJobID:     "publish-job-1",
		TargetID:         "publish-target-1",
		Platform:         "wechat",
		MetricCode:       "views",
		MetricDate:       "2026-05-25",
		Period:           PeriodDay,
		RawValue:         "1200",
		SourceType:       SourceManual,
	}, "metric-summary-write-read")
	if err != nil {
		t.Fatalf("create metric record must succeed before summary/trend/missing assertions: %v", err)
	}
	if record.MetricRecordID == "" {
		t.Fatalf("create metric record must return metric_record_id before summary/trend/missing assertions")
	}

	summary, err := svc.GetSummary(ctx, "seed-project", MetricSummaryRequest{
		DateFrom:    "2026-05-01",
		DateTo:      "2026-05-25",
		Platform:    "wechat",
		TargetID:    "publish-target-1",
		MetricCodes: []string{"views"},
	})
	if err != nil {
		t.Fatalf("summary must succeed after persisted records: %v", err)
	}
	if summary.SummarySnapshotID == "" || summary.SourceRecordCount != 1 || len(summary.Items) != 1 || summary.Items[0].AggregationMethod != AggregationSum {
		t.Fatalf("summary must include stable snapshot, source count and aggregation method: %#v", summary)
	}

	trend, err := svc.GetTrends(ctx, "seed-project", MetricTrendRequest{
		MetricCode: "views",
		DateFrom:   "2026-05-01",
		DateTo:     "2026-05-25",
		Bucket:     PeriodDay,
		Platform:   "wechat",
		TargetID:   "publish-target-1",
	})
	if err != nil {
		t.Fatalf("trends must succeed after persisted records: %v", err)
	}
	if trend.QuerySignature == "" || len(trend.Series) == 0 || trend.Series[0].Value == 0 && !trend.Series[0].Missing {
		t.Fatalf("trends must include query signature and non-missing value from persisted record: %#v", trend)
	}

	missing, err := svc.GetMissingDates(ctx, "seed-project", MissingMetricDatesRequest{
		DateFrom: "2026-05-01",
		DateTo:   "2026-05-25",
		Platform: "wechat",
		TargetID: "publish-target-1",
	})
	if err != nil {
		t.Fatalf("missing dates must succeed from published jobs and templates: %v", err)
	}
	for _, item := range missing.Items {
		if item.MissingReason != "required_metric_missing" || item.Period == "" || item.MetricDate == "2026-05-25" {
			t.Fatalf("missing dates must exclude existing metric date and include period/reason: %#v", missing.Items)
		}
	}
}
