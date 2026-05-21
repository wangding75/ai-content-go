package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

type ErrorCode string

const (
	ErrorValidation            ErrorCode = "VALIDATION_ERROR"
	ErrorUnauthorized          ErrorCode = "UNAUTHORIZED"
	ErrorForbidden             ErrorCode = "FORBIDDEN"
	ErrorNotFound              ErrorCode = "NOT_FOUND"
	ErrorConflict              ErrorCode = "CONFLICT"
	ErrorDependencyUnavailable ErrorCode = "DEPENDENCY_UNAVAILABLE"
	ErrorMigrationReadFailed   ErrorCode = "MIGRATION_READ_FAILED"
	ErrorQueueEnqueueFailed    ErrorCode = "QUEUE_ENQUEUE_FAILED"
	ErrorSerializationFailed   ErrorCode = "SERIALIZATION_FAILED"
	ErrorInternal              ErrorCode = "INTERNAL_ERROR"
	ErrorIdempotencyConflict   ErrorCode = "IDEMPOTENCY_CONFLICT"
	ErrorWorkflowRunFailed     ErrorCode = "WORKFLOW_RUN_FAILED"
	ErrorAgentOutputInvalid    ErrorCode = "AGENT_OUTPUT_INVALID"
	ErrorLLMProviderError      ErrorCode = "LLM_PROVIDER_ERROR"
	ErrorExternalAutomation    ErrorCode = "EXTERNAL_AUTOMATION_ERROR"
)

type Envelope struct {
	Success   bool      `json:"success"`
	Data      any       `json:"data"`
	Error     *APIError `json:"error"`
	RequestID string    `json:"request_id"`
}

type APIError struct {
	Code    ErrorCode     `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details"`
}

type ErrorDetail struct {
	Field  string `json:"field,omitempty"`
	Reason string `json:"reason"`
}

func WriteSuccess(w http.ResponseWriter, r *http.Request, status int, data any) {
	writeJSON(w, status, Envelope{Success: true, Data: data, Error: nil, RequestID: RequestID(r)})
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code ErrorCode, message string, details []ErrorDetail) {
	if status < http.StatusBadRequest {
		status = http.StatusInternalServerError
	}
	writeJSON(w, status, Envelope{
		Success:   false,
		Data:      nil,
		Error:     &APIError{Code: code, Message: message, Details: details},
		RequestID: RequestID(r),
	})
}

func RequestID(r *http.Request) string {
	if requestID := middleware.GetReqID(r.Context()); requestID != "" {
		return requestID
	}
	if requestID := r.Header.Get("X-Request-Id"); requestID != "" {
		return requestID
	}
	return "req_unknown"
}

func writeJSON(w http.ResponseWriter, status int, payload Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return
	}
}
