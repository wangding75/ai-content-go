package schedule_test

import (
	"context"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/schedule"
)

// @Test
func TestCreateScheduleRejectsEmptyTemplateVersionIDWithValidationError(t *testing.T) {
	svc := schedule.NewService()

	_, err := svc.CreateSchedule(context.Background(), schedule.CreateScheduleRequest{
		ProjectID:    "proj-1",
		ScheduleType: "cron",
	})
	if err == nil {
		t.Fatalf("expected validation error for empty template_version_id")
	}
}

// @Test
func TestCreateScheduleReturnsRealScheduleIDNotPlaceholder(t *testing.T) {
	svc := schedule.NewService()

	resp, err := svc.CreateSchedule(context.Background(), schedule.CreateScheduleRequest{
		TemplateVersionID: "wftv-1",
		ProjectID:         "proj-1",
		ScheduleType:      "cron",
	})
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	if resp.ScheduleID == "" || resp.ScheduleID == "schedule-placeholder" {
		t.Fatalf("expected real schedule id, got %q", resp.ScheduleID)
	}
}
