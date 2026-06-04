package article

import (
	"context"
	"errors"
	"testing"
)

// @Test
func TestTask04CreateGenerationRunReturnsRunIDAndStatus(t *testing.T) {
	svc := NewService(nil, nil, nil)
	svc.UpdateConfig(context.Background(), "project-1", UpdateArticleConfigRequest{
		TopicStyle: "tech",
		DefaultWorkflowTemplateVersionID: "wftv-1",
	}, "idem-gen-cfg")

	resp, err := svc.CreateGenerationRun(context.Background(), "project-1", CreateArticleGenerationRunRequest{
		Topic:          "AI in Go",
		Audience:       "developers",
		TargetPlatform: "blog",
	}, "wfr-1", "idem-gen-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.GenerationRunID == "" {
		t.Fatal("expected non-empty generation_run_id")
	}
	if resp.WorkflowRunID != "wfr-1" {
		t.Fatalf("expected workflow_run_id wfr-1, got %s", resp.WorkflowRunID)
	}
	if resp.Status == "" {
		t.Fatal("expected non-empty status")
	}
}

// @Test
func TestTask04CreateGenerationRunReturnsErrorForNonexistentProject(t *testing.T) {
	svc := NewService(nil, nil, nil)
	_, err := svc.CreateGenerationRun(context.Background(), "nonexistent", CreateArticleGenerationRunRequest{
		Topic: "test",
	}, "wfr-1", "")
	if err == nil {
		t.Fatal("expected error for nonexistent project")
	}
}

// @Test
func TestTask04ListGenerationRunsReturnsPagedResults(t *testing.T) {
	svc := NewService(nil, nil, nil)
	svc.UpdateConfig(context.Background(), "project-1", UpdateArticleConfigRequest{TopicStyle: "tech"}, "idem-list-cfg")
	svc.CreateGenerationRun(context.Background(), "project-1", CreateArticleGenerationRunRequest{Topic: "Topic 1", TargetPlatform: "web"}, "wfr-1", "idem-list-1")
	svc.CreateGenerationRun(context.Background(), "project-1", CreateArticleGenerationRunRequest{Topic: "Topic 2", TargetPlatform: "blog"}, "wfr-2", "idem-list-2")

	resp, err := svc.ListGenerationRuns(context.Background(), "project-1", ListGenerationRunsRequest{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp.Items) < 2 {
		t.Fatalf("expected at least 2 runs, got %d", len(resp.Items))
	}
	if resp.Items[0].GenerationRunID == "" {
		t.Fatal("expected non-empty generation_run_id in items")
	}
}

// @Test
func TestTask04GetGenerationRunReturnsDetail(t *testing.T) {
	svc := NewService(nil, nil, nil)
	svc.UpdateConfig(context.Background(), "project-1", UpdateArticleConfigRequest{TopicStyle: "tech"}, "idem-detail-cfg")
	createResp, err := svc.CreateGenerationRun(context.Background(), "project-1", CreateArticleGenerationRunRequest{Topic: "Detail Topic", TargetPlatform: "web"}, "wfr-1", "idem-detail-1")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	detail, err := svc.GetGenerationRun(context.Background(), "project-1", createResp.GenerationRunID)
	if err != nil {
		t.Fatalf("get: expected no error, got %v", err)
	}
	if detail.GenerationRunID != createResp.GenerationRunID {
		t.Fatalf("expected generation_run_id %s, got %s", createResp.GenerationRunID, detail.GenerationRunID)
	}
}

// @Test
func TestTask04GetGenerationRunReturnsNotFoundForMissingID(t *testing.T) {
	svc := NewService(nil, nil, nil)
	_, err := svc.GetGenerationRun(context.Background(), "project-1", "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// @Test
func TestTask04RetryGenerationRunReturnsNewRunID(t *testing.T) {
	svc := NewService(nil, nil, nil)
	svc.UpdateConfig(context.Background(), "project-1", UpdateArticleConfigRequest{TopicStyle: "tech"}, "idem-retry-cfg")
	createResp, _ := svc.CreateGenerationRun(context.Background(), "project-1", CreateArticleGenerationRunRequest{Topic: "Retry Topic", TargetPlatform: "web"}, "wfr-1", "idem-retry-1")

	retryResp, err := svc.RetryGenerationRun(context.Background(), "project-1", createResp.GenerationRunID, RetryGenerationRunRequest{Reason: "test retry"}, "wfr-2", "idem-retry-2")
	if err != nil {
		t.Fatalf("retry: expected no error, got %v", err)
	}
	if retryResp.NewGenerationRunID == "" {
		t.Fatal("expected non-empty new generation_run_id")
	}
	if retryResp.NewGenerationRunID == createResp.GenerationRunID {
		t.Fatal("expected different generation_run_id for retry")
	}
}

// @Test
func TestTask04GetContentSnapshotReturnsSnapshot(t *testing.T) {
	svc := NewService(nil, nil, nil)
	snap, err := svc.GetContentSnapshot(context.Background(), "item-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if snap.Title == "" {
		t.Fatal("expected non-empty title in snapshot")
	}
	if snap.LatestContentVersionID == "" {
		t.Fatal("expected non-empty latest_content_version_id")
	}
}

// @Test
func TestTask04ListGenerationRunsSupportsPagination(t *testing.T) {
	svc := NewService(nil, nil, nil)
	svc.UpdateConfig(context.Background(), "project-1", UpdateArticleConfigRequest{TopicStyle: "tech"}, "idem-pg-cfg")
	for i := 0; i < 5; i++ {
		svc.CreateGenerationRun(context.Background(), "project-1",
			CreateArticleGenerationRunRequest{Topic: "Topic", TargetPlatform: "web"},
			"wfr", "")
	}

	resp, err := svc.ListGenerationRuns(context.Background(), "project-1", ListGenerationRunsRequest{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp.Items) == 0 {
		t.Fatal("expected items in response")
	}
	if resp.Pagination.Total == 0 {
		t.Fatal("expected non-zero total in pagination")
	}
}

// @Test
func TestTask04CreateGenerationRunReturnsErrorForMissingConfig(t *testing.T) {
	svc := NewService(nil, nil, nil)
	_, err := svc.CreateGenerationRun(context.Background(), "project-no-config", CreateArticleGenerationRunRequest{Topic: "test"}, "wfr-1", "")
	if err == nil {
		t.Fatal("expected error when no config exists for project")
	}
}

// @Test
func TestTask04RetryGenerationRunForNonFailedRunReturnsConflict(t *testing.T) {
	svc := NewService(nil, nil, nil)
	svc.UpdateConfig(context.Background(), "project-1", UpdateArticleConfigRequest{TopicStyle: "tech"}, "idem-retry-cfg2")
	createResp, _ := svc.CreateGenerationRun(context.Background(), "project-1",
		CreateArticleGenerationRunRequest{Topic: "test", TargetPlatform: "web"}, "wfr-1", "idem-retry-init")

	_, err := svc.RetryGenerationRun(context.Background(), "project-1", createResp.GenerationRunID,
		RetryGenerationRunRequest{Reason: "test"}, "wfr-2", "idem-retry-new")
	if err == nil {
		t.Fatal("expected error when retrying a non-failed run (only failed runs are retryable by design)")
	}
}

// @Test
func TestTask04GetGenerationRunReturnsForbiddenForWrongProject(t *testing.T) {
	svc := NewService(nil, nil, nil)
	svc.UpdateConfig(context.Background(), "project-1", UpdateArticleConfigRequest{TopicStyle: "tech"}, "idem-fb-cfg")
	createResp, _ := svc.CreateGenerationRun(context.Background(), "project-1",
		CreateArticleGenerationRunRequest{Topic: "test", TargetPlatform: "web"}, "wfr-1", "idem-fb-gen")

	_, err := svc.GetGenerationRun(context.Background(), "project-2", createResp.GenerationRunID)
	if err == nil {
		t.Fatal("expected error when accessing run from wrong project")
	}
}
