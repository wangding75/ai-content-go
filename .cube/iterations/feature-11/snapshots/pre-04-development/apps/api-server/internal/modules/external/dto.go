package external

import "github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"

type ListProvidersRequest struct {
	content.PaginationRequest
	ProviderType string `json:"provider_type"`
}

type CreateProviderRequest struct {
	ProviderType string `json:"provider_type"`
	BaseURL      string `json:"base_url"`
	Token        string `json:"token"`
}

type ProviderResponse struct {
	ID           string `json:"id"`
	ProviderType string `json:"provider_type"`
	BaseURL      string `json:"base_url"`
	TokenMasked  string `json:"token_masked"`
	Enabled      bool   `json:"enabled"`
}

type PagedProvidersResponse struct {
	Items      []ProviderResponse         `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type CreateProviderResponse struct {
	ProviderID  string `json:"provider_id"`
	TokenMasked string `json:"token_masked"`
}

type ListBindingsRequest struct {
	content.PaginationRequest
	ProviderID   string `json:"provider_id"`
	TriggerEvent string `json:"trigger_event"`
}

type CreateBindingRequest struct {
	ProviderID   string `json:"provider_id"`
	TriggerEvent string `json:"trigger_event"`
	WebhookURL   string `json:"webhook_url"`
}

type BindingResponse struct {
	ID           string `json:"id"`
	ProviderID   string `json:"provider_id"`
	TriggerEvent string `json:"trigger_event"`
	WebhookURL   string `json:"webhook_url"`
	Enabled      bool   `json:"enabled"`
}

type PagedBindingsResponse struct {
	Items      []BindingResponse          `json:"items"`
	Pagination content.PaginationResponse `json:"pagination"`
}

type CreateBindingResponse struct {
	BindingID string `json:"binding_id"`
}

type RotateCallbackTokenRequest struct {
	Reason string `json:"reason"`
}

type RotateCallbackTokenResponse struct {
	BindingID           string `json:"binding_id"`
	CallbackTokenOnce   string `json:"callback_token_once"`
	CallbackTokenMasked string `json:"callback_token_masked"`
	OperationLogID      string `json:"operation_log_id"`
}

type UpdateCallbackAuthRequest struct {
	CallbackAuthType string `json:"callback_auth_type"`
	SigningSecretRef string `json:"signing_secret_ref"`
	ChangeReason     string `json:"change_reason"`
}

type UpdateCallbackAuthResponse struct {
	BindingID        string `json:"binding_id"`
	CallbackAuthType string `json:"callback_auth_type"`
	OperationLogID   string `json:"operation_log_id"`
}

type ExternalCallbackRequest struct {
	BindingID     string         `json:"binding_id"`
	EventType     string         `json:"event_type"`
	Payload       map[string]any `json:"payload"`
	StableEventID string         `json:"stable_event_id"`
}

type ExternalCallbackAuth struct {
	BindingID       string `json:"binding_id"`
	CallbackToken   string `json:"callback_token"`
	SignatureHeader string `json:"signature_header"`
}

type ExternalCallbackResponse struct {
	Accepted      bool   `json:"accepted"`
	CallbackLogID string `json:"callback_log_id"`
}

type ListCallbackLogsRequest struct {
	content.PaginationRequest
	ProviderID string `json:"provider_id"`
	BindingID  string `json:"binding_id"`
	EventType  string `json:"event_type"`
	Accepted   *bool  `json:"accepted"`
}

type ExternalCallbackLogResponse struct {
	ID                string `json:"id"`
	ProviderID        string `json:"provider_id"`
	BindingID         string `json:"binding_id"`
	EventType         string `json:"event_type"`
	Accepted          bool   `json:"accepted"`
	RejectedReason    string `json:"rejected_reason"`
	BoundaryViolation bool   `json:"boundary_violation"`
	CreatedAt         string `json:"created_at"`
}

type PagedExternalCallbackLogsResponse struct {
	Items      []ExternalCallbackLogResponse `json:"items"`
	Pagination content.PaginationResponse    `json:"pagination"`
}

type TestExternalCallbackRequest struct {
	BindingID string         `json:"binding_id"`
	EventType string         `json:"event_type"`
	Payload   map[string]any `json:"payload"`
}
