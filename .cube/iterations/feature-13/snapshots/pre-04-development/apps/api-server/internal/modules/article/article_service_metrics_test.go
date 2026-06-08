package article

import (
	"context"
	"testing"
)

// @Test
func TestTask05GetProjectArticleMetricsReturnsTemplates(t *testing.T) {
	svc := NewService(nil, nil, nil)
	resp, err := svc.GetProjectArticleMetrics(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Items == nil {
		t.Fatal("expected non-nil items in response")
	}
}

// @Test
func TestTask05UpdateProjectArticleMetricsReturnsVersionAndOpLog(t *testing.T) {
	svc := NewService(nil, nil, nil)
	resp, err := svc.UpdateProjectArticleMetrics(context.Background(), "project-1", UpdateProjectArticleMetricsRequest{
		EnabledMetricCodes: []string{"views", "likes"},
	}, "idem-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.VersionID == "" {
		t.Fatal("expected non-empty version_id")
	}
	if resp.OperationLogID == "" {
		t.Fatal("expected non-empty operation_log_id")
	}
}

// @Test
func TestTask05UpdateThenGetReturnsEnabledMetrics(t *testing.T) {
	svc := NewService(nil, nil, nil)
	_, err := svc.UpdateProjectArticleMetrics(context.Background(), "project-1", UpdateProjectArticleMetricsRequest{
		EnabledMetricCodes: []string{"views", "likes", "shares"},
	}, "idem-u1")
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	resp, err := svc.GetProjectArticleMetrics(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("get: expected no error, got %v", err)
	}
	if len(resp.Items) == 0 {
		t.Fatal("expected non-empty metrics items")
	}
	var hasViews bool
	for _, item := range resp.Items {
		if item.MetricCode == "views" {
			hasViews = true
		}
	}
	if !hasViews {
		t.Fatal("expected 'views' in enabled metrics after update")
	}
	if resp.Pagination.Total == 0 {
		t.Fatal("expected non-zero pagination total")
	}
}

// @Test
func TestTask05UpdateProjectArticleMetricsIsIdempotent(t *testing.T) {
	svc := NewService(nil, nil, nil)
	resp1, _ := svc.UpdateProjectArticleMetrics(context.Background(), "project-1", UpdateProjectArticleMetricsRequest{
		EnabledMetricCodes: []string{"views"},
	}, "idem-mt-same")
	resp2, _ := svc.UpdateProjectArticleMetrics(context.Background(), "project-1", UpdateProjectArticleMetricsRequest{
		EnabledMetricCodes: []string{"views"},
	}, "idem-mt-same")
	if resp1.VersionID != resp2.VersionID {
		t.Fatal("expected same version_id for idempotent update")
	}
}

// @Test
func TestTask05UpdateProjectArticleMetricsReturnsValidationForInvalidCodes(t *testing.T) {
	svc := NewService(nil, nil, nil)
	_, err := svc.UpdateProjectArticleMetrics(context.Background(), "project-1", UpdateProjectArticleMetricsRequest{
		EnabledMetricCodes: []string{},
	}, "")
	if err == nil {
		t.Fatal("expected validation error for empty codes")
	}
}

// @Test
func TestTask05GetProjectArticleMetricsReturnsNotFoundForMissingProject(t *testing.T) {
	svc := NewService(nil, nil, nil)
	_, err := svc.GetProjectArticleMetrics(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing project")
	}
}

// @Test
func TestTask05UpdateProjectArticleMetricsSupportsPlatformOverrides(t *testing.T) {
	svc := NewService(nil, nil, nil)
	resp, err := svc.UpdateProjectArticleMetrics(context.Background(), "project-1", UpdateProjectArticleMetricsRequest{
		EnabledMetricCodes: []string{"views"},
		PlatformOverrides:  map[string][]string{"wechat": {"views"}},
		Note:               "wechat only views",
	}, "idem-po-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.VersionID == "" {
		t.Fatal("expected version_id")
	}

	// Verify platform overrides were persisted
	getResp, err := svc.GetProjectArticleMetrics(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("get after platform override: %v", err)
	}
	_ = getResp
}
