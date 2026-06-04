package dashboard_test

import (
	"context"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/dashboard"
)

// @Test
func TestSummaryReturnsDashboardBusinessCounters(t *testing.T) {
	service := dashboard.NewService()

	resp, err := service.Summary(context.Background())
	if err != nil {
		t.Fatalf("summary should not fail: %v", err)
	}

	if resp.ProjectCount < 0 || resp.PendingReviewCount < 0 || resp.PendingPublishCount < 0 || resp.FailedTaskCount < 0 || resp.TodayCost < 0 {
		t.Fatalf("summary counters must not be negative: %#v", resp)
	}
}

// @Test
func TestSummaryIncludesAllOutputContractFields(t *testing.T) {
	service := dashboard.NewService()

	resp, err := service.Summary(context.Background())
	if err != nil {
		t.Fatalf("summary should not fail: %v", err)
	}

	_ = resp.ProjectCount
	_ = resp.PendingReviewCount
	_ = resp.PendingPublishCount
	_ = resp.FailedTaskCount
	_ = resp.TodayCost
}
