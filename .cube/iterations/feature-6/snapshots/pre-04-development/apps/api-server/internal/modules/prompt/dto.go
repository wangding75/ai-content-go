package prompt

import "github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"

type ListTemplatesRequest struct {
	content.PaginationRequest
	AgentCode string `json:"agent_code"`
}

type CreateTemplateRequest struct {
	Code      string   `json:"code"`
	Template  string   `json:"template"`
	Variables []string `json:"variables"`
}

type TemplateResponse struct {
	ID        string   `json:"id"`
	Code      string   `json:"code"`
	AgentCode string   `json:"agent_code"`
	Template  string   `json:"template"`
	Variables []string `json:"variables"`
}

type PagedTemplatesResponse struct {
	Items      []TemplateResponse         `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type CreateTemplateResponse struct {
	PromptTemplateID string `json:"prompt_template_id"`
}
