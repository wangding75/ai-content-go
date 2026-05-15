package schedule

type CreateScheduleRequest struct {
	TemplateVersionID string `json:"template_version_id"`
	ProjectID         string `json:"project_id"`
	ScheduleType      string `json:"schedule_type"`
}

type ScheduleResponse struct {
	ScheduleID string `json:"schedule_id"`
	Status     string `json:"status"`
}
