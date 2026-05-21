package schedule_test

import (
	"context"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/schedule"
)

// @Test
func TestIteration21ScheduleServiceContractCreatesDefaultProductionPlanCount(t *testing.T) {
	svc := schedule.NewService()

	resp, err := svc.CreateSchedule(context.Background(), schedule.CreateScheduleRequest{
		ProjectID:         "project-21-task01",
		TemplateVersionID: "wftv-21-task01",
		CronExpression:    "0 9 * * *",
		DailyContentCount: 0,
	}, "schedule-contract-key")
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	if resp.ScheduleID == "" {
		t.Fatalf("expected schedule_id")
	}
	if resp.DailyContentCount != 5 {
		t.Fatalf("expected default daily_content_count=5, got %d", resp.DailyContentCount)
	}
}

// @Test
func TestIteration21ScheduleServiceContractRejectsInvalidDailyContentCount(t *testing.T) {
	svc := schedule.NewService()

	_, err := svc.CreateSchedule(context.Background(), schedule.CreateScheduleRequest{
		ProjectID:         "project-21-task01-invalid",
		TemplateVersionID: "wftv-21-task01-invalid",
		CronExpression:    "0 9 * * *",
		DailyContentCount: -1,
	}, "schedule-contract-invalid-key")
	if err == nil {
		t.Fatalf("expected validation error for daily_content_count < 0")
	}
}
