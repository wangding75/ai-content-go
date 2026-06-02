package portfolio

import "github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"

const (
	PortfolioStatusActive   = "active"
	PortfolioStatusArchived = "archived"
	PortfolioScopeManual    = "manual"

	PortfolioProjectRoleMember = "member"

	SnapshotStatusQueued    = "queued"
	SnapshotStatusRunning   = "running"
	SnapshotStatusCompleted = "completed"
	SnapshotStatusFailed    = "failed"

	HealthStatusHealthy  = "healthy"
	HealthStatusWatch    = "watch"
	HealthStatusCritical = "critical"
	HealthStatusPending  = "pending"
)

type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type SourceRef struct {
	Source    string `json:"source"`
	SourceID  string `json:"source_id"`
	UpdatedAt string `json:"updated_at"`
}

type CreatePortfolioRequest struct {
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	ScopeType    string         `json:"scope_type"`
	OwnerID      string         `json:"owner_id"`
	HealthPolicy map[string]any `json:"health_policy"`
	Status       string         `json:"status"`
}

type UpdatePortfolioRequest struct {
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	ScopeType    string         `json:"scope_type"`
	OwnerID      string         `json:"owner_id"`
	HealthPolicy map[string]any `json:"health_policy"`
	Status       string         `json:"status"`
}

type ListPortfoliosRequest struct {
	content.PaginationRequest
	Q         string `json:"q"`
	Status    string `json:"status"`
	ScopeType string `json:"scope_type"`
	OwnerID   string `json:"owner_id"`
}

type PortfolioDetailResponse struct {
	ID                   string         `json:"id"`
	Name                 string         `json:"name"`
	Description          string         `json:"description"`
	ScopeType            string         `json:"scope_type"`
	OwnerID              string         `json:"owner_id"`
	HealthPolicy         map[string]any `json:"health_policy"`
	Status               string         `json:"status"`
	ProjectCount         int            `json:"project_count"`
	LatestHealthScore    float64        `json:"latest_health_score"`
	LatestHealthStatus   string         `json:"latest_health_status"`
	EstimatedMonthlyCost float64        `json:"estimated_monthly_cost"`
	Currency             string         `json:"currency"`
	CreatedAt            string         `json:"created_at"`
	UpdatedAt            string         `json:"updated_at"`
}

type PortfolioListItem = PortfolioDetailResponse

type PagedPortfoliosResponse struct {
	Items      []PortfolioListItem        `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type AddPortfolioProjectRequest struct {
	ProjectID string  `json:"project_id"`
	Role      string  `json:"role"`
	Priority  int     `json:"priority"`
	Weight    float64 `json:"weight"`
	Note      string  `json:"note"`
	AddedBy   string  `json:"added_by"`
}

type UpdatePortfolioProjectPriorityRequest struct {
	Role     string  `json:"role"`
	Priority int     `json:"priority"`
	Weight   float64 `json:"weight"`
	Note     string  `json:"note"`
}

type RemovePortfolioProjectRequest struct {
	Reason string `json:"reason"`
	Note   string `json:"note"`
}

type PortfolioProjectResponse struct {
	PortfolioID string  `json:"portfolio_id"`
	ProjectID   string  `json:"project_id"`
	ProjectName string  `json:"project_name"`
	ContentType string  `json:"content_type"`
	Role        string  `json:"role"`
	Priority    int     `json:"priority"`
	Weight      float64 `json:"weight"`
	Note        string  `json:"note"`
	AddedBy     string  `json:"added_by"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type ListPortfolioProjectsRequest struct {
	content.PaginationRequest
	Role string `json:"role"`
}

type PagedPortfolioProjectsResponse struct {
	Items      []PortfolioProjectResponse `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type RemovePortfolioProjectResponse struct {
	PortfolioID string `json:"portfolio_id"`
	ProjectID   string `json:"project_id"`
	OperationID string `json:"operation_id"`
	Removed     bool   `json:"removed"`
}

type PortfolioSummaryRequest struct {
	DateRange DateRange `json:"date_range"`
}

type RecalculatePortfolioStatusSnapshotRequest struct {
	DateRange DateRange `json:"date_range"`
	Force     bool      `json:"force"`
}

type RecalculatePortfolioStatusSnapshotResponse struct {
	PortfolioID       string `json:"portfolio_id"`
	SnapshotID        string `json:"snapshot_id"`
	JobID             string `json:"job_id"`
	CalculationStatus string `json:"calculation_status"`
}

type PortfolioStatusSnapshotResponse struct {
	ID                   string         `json:"id"`
	PortfolioID          string         `json:"portfolio_id"`
	DateRange            DateRange      `json:"date_range"`
	HealthScore          float64        `json:"health_score"`
	HealthStatus         string         `json:"health_status"`
	TotalProjects        int            `json:"total_projects"`
	ActiveProjects       int            `json:"active_projects"`
	WarningProjects      int            `json:"warning_projects"`
	EstimatedMonthlyCost float64        `json:"estimated_monthly_cost"`
	Currency             string         `json:"currency"`
	RiskSummary          map[string]any `json:"risk_summary"`
	CostSummary          map[string]any `json:"cost_summary"`
	StrategySummary      map[string]any `json:"strategy_summary"`
	SourceRefs           []SourceRef    `json:"source_refs"`
	CalculationStatus    string         `json:"calculation_status"`
	CalculatedAt         string         `json:"calculated_at"`
	CreatedAt            string         `json:"created_at"`
}

type ListPortfolioStatusSnapshotsRequest struct {
	content.PaginationRequest
	DateRange DateRange `json:"date_range"`
}

type PagedPortfolioStatusSnapshotsResponse struct {
	Items      []PortfolioStatusSnapshotResponse `json:"items"`
	Pagination content.PaginationResponse        `json:"pagination"`
}

type PortfolioHealthSummaryResponse struct {
	PortfolioID      string         `json:"portfolio_id"`
	DateRange        DateRange      `json:"date_range"`
	HealthScore      float64        `json:"health_score"`
	HealthStatus     string         `json:"health_status"`
	TotalProjects    int            `json:"total_projects"`
	ActiveProjects   int            `json:"active_projects"`
	WarningProjects  int            `json:"warning_projects"`
	RiskSummary      map[string]any `json:"risk_summary"`
	LatestSnapshotAt string         `json:"latest_snapshot_at"`
	CalculatedAt     string         `json:"calculated_at"`
	SourceRefs       []SourceRef    `json:"source_refs"`
}

type ModelCostItem struct {
	Model         string  `json:"model"`
	EstimatedCost float64 `json:"estimated_cost"`
	Currency      string  `json:"currency"`
}

type ProjectCostItem struct {
	ProjectID     string  `json:"project_id"`
	ProjectName   string  `json:"project_name"`
	EstimatedCost float64 `json:"estimated_cost"`
	Currency      string  `json:"currency"`
}

type PortfolioCostSummaryResponse struct {
	PortfolioID          string            `json:"portfolio_id"`
	DateRange            DateRange         `json:"date_range"`
	EstimatedMonthlyCost float64           `json:"estimated_monthly_cost"`
	Currency             string            `json:"currency"`
	ByModel              []ModelCostItem   `json:"by_model"`
	ProjectCosts         []ProjectCostItem `json:"project_costs"`
	CalculatedAt         string            `json:"calculated_at"`
	SourceRefs           []SourceRef       `json:"source_refs"`
}

type StrategySuggestionSummaryItem struct {
	ProjectID    string `json:"project_id"`
	SuggestionID string `json:"suggestion_id"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	Title        string `json:"title"`
}

type PortfolioStrategySummaryResponse struct {
	PortfolioID     string                          `json:"portfolio_id"`
	DateRange       DateRange                       `json:"date_range"`
	Pending         int                             `json:"pending"`
	Confirmed       int                             `json:"confirmed"`
	Ignored         int                             `json:"ignored"`
	Executed        int                             `json:"executed"`
	ExecutionFailed int                             `json:"execution_failed"`
	TopSuggestions  []StrategySuggestionSummaryItem `json:"top_suggestions"`
	CalculatedAt    string                          `json:"calculated_at"`
	SourceRefs      []SourceRef                     `json:"source_refs"`
}
