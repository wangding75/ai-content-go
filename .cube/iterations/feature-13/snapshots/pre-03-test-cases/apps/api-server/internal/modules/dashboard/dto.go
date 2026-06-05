package dashboard

type SummaryResponse struct {
	ProjectCount        int     `json:"project_count"`
	PendingReviewCount  int     `json:"pending_review_count"`
	PendingPublishCount int     `json:"pending_publish_count"`
	FailedTaskCount     int     `json:"failed_task_count"`
	TodayCost           float64 `json:"today_cost"`
}
