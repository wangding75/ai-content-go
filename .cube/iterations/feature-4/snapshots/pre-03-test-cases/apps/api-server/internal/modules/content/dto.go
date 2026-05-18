package content

type PaginationRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Sort     string `json:"sort"`
	Order    string `json:"order"`
}

type PaginationResponse struct {
	Page     int  `json:"page"`
	PageSize int  `json:"page_size"`
	Total    int  `json:"total"`
	HasNext  bool `json:"has_next"`
}

type ListContentTypesRequest struct {
	PaginationRequest
	Enabled *bool `json:"enabled"`
}

type CreateContentTypeRequest struct {
	Code          string         `json:"code"`
	Name          string         `json:"name"`
	ProjectSchema map[string]any `json:"project_schema"`
}

type ContentTypeResponse struct {
	ID            string         `json:"id"`
	Code          string         `json:"code"`
	Name          string         `json:"name"`
	ProjectSchema map[string]any `json:"project_schema"`
	Enabled       bool           `json:"enabled"`
}

type PagedContentTypesResponse struct {
	Items      []ContentTypeResponse `json:"items"`
	Pagination PaginationResponse    `json:"pagination"`
}

type CreateContentTypeResponse struct {
	ContentTypeID string `json:"content_type_id"`
}

type ProjectSchemaResponse struct {
	ContentTypeID string         `json:"content_type_id"`
	ProjectSchema map[string]any `json:"project_schema"`
}

type ListProjectsRequest struct {
	PaginationRequest
	Status      string `json:"status"`
	ContentType string `json:"content_type"`
}

type CreateProjectRequest struct {
	Name          string         `json:"name"`
	ContentTypeID string         `json:"content_type_id"`
	ProjectConfig map[string]any `json:"project_config"`
}

type ProjectResponse struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	ContentTypeID   string         `json:"content_type_id"`
	ContentTypeCode string         `json:"content_type_code"`
	Status          string         `json:"status"`
	ProjectConfig   map[string]any `json:"project_config"`
}

type PagedProjectsResponse struct {
	Items      []ProjectResponse  `json:"items"`
	Pagination PaginationResponse `json:"pagination"`
}

type CreateProjectResponse struct {
	ProjectID string `json:"project_id"`
	Status    string `json:"status"`
}

type ProjectOverviewResponse struct {
	ProjectID      string  `json:"project_id"`
	Progress       float64 `json:"progress"`
	PendingActions int     `json:"pending_actions"`
	Cost           float64 `json:"cost"`
}

type PauseProjectRequest struct {
	Reason string `json:"reason"`
	Note   string `json:"note"`
}

type PauseProjectResponse struct {
	ProjectID      string `json:"project_id"`
	Status         string `json:"status"`
	OperationLogID string `json:"operation_log_id"`
}
