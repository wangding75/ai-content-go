package schedule_test

import (
	"context"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/schedule"
)

// @Test
func TestIteration21ScheduleRuntimeEnablesDisablesAndRejectsDuplicateTransitions(t *testing.T) {
	svc := schedule.NewService()
	created, err := svc.CreateSchedule(context.Background(), schedule.CreateScheduleRequest{
		ProjectID:         "project-21-task02",
		TemplateVersionID: "wftv-21-task02",
		CronExpression:    "0 9 * * *",
		DailyContentCount: 7,
	}, "schedule-runtime-key")
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}

	enabled, err := svc.EnableSchedule(context.Background(), created.ScheduleID, schedule.ToggleScheduleRequest{Note: "start"}, "enable-key")
	if err != nil {
		t.Fatalf("enable schedule: %v", err)
	}
	if enabled.PreviousEnabled || !enabled.CurrentEnabled || enabled.OperationLogID == "" {
		t.Fatalf("unexpected enable response: %#v", enabled)
	}
	if _, err := svc.EnableSchedule(context.Background(), created.ScheduleID, schedule.ToggleScheduleRequest{Note: "again"}, "enable-key-2"); err == nil {
		t.Fatalf("expected conflict when enabling an enabled schedule")
	}

	disabled, err := svc.DisableSchedule(context.Background(), created.ScheduleID, schedule.ToggleScheduleRequest{Reason: "pause"}, "disable-key")
	if err != nil {
		t.Fatalf("disable schedule: %v", err)
	}
	if !disabled.PreviousEnabled || disabled.CurrentEnabled || disabled.OperationLogID == "" {
		t.Fatalf("unexpected disable response: %#v", disabled)
	}
}

// @Test
func TestIteration21ScheduleRuntimePreparesTestRunAndLinksTriggerLogToWorkflowRun(t *testing.T) {
	svc := schedule.NewService()
	created, err := svc.CreateSchedule(context.Background(), schedule.CreateScheduleRequest{
		ProjectID:         "project-21-task02-run",
		TemplateVersionID: "wftv-21-task02-run",
		CronExpression:    "0 10 * * *",
		InputTemplate:     map[string]any{"topic": "baseline"},
	}, "schedule-runtime-run-key")
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}

	prepared, err := svc.PrepareTestRun(context.Background(), created.ScheduleID, schedule.TestRunScheduleRequest{
		InputOverride: map[string]any{"topic": "override"},
	})
	if err != nil {
		t.Fatalf("prepare test run: %v", err)
	}
	if prepared.TriggerLogID == "" || prepared.ProjectID != "project-21-task02-run" || prepared.TemplateVersionID != "wftv-21-task02-run" {
		t.Fatalf("unexpected prepared run: %#v", prepared)
	}
	if prepared.Input["topic"] != "override" {
		t.Fatalf("expected override input to win, got %#v", prepared.Input)
	}

	if err := svc.CompleteTrigger(context.Background(), prepared.TriggerLogID, "workflow-run-21", "queued"); err != nil {
		t.Fatalf("complete trigger: %v", err)
	}
	triggers, err := svc.ListTriggers(context.Background(), created.ScheduleID, schedule.ListTriggersRequest{})
	if err != nil {
		t.Fatalf("list triggers: %v", err)
	}
	if len(triggers.Items) != 1 || triggers.Items[0].WorkflowRunID != "workflow-run-21" || triggers.Items[0].Status != "queued" {
		t.Fatalf("expected linked trigger log, got %#v", triggers.Items)
	}
}
