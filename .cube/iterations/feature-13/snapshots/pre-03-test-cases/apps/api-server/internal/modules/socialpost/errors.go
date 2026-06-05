package socialpost

import "errors"

var (
	ErrValidation          = errors.New("validation error")
	ErrNotFound            = errors.New("not found")
	ErrForbidden           = errors.New("forbidden")
	ErrConflict            = errors.New("conflict")
	ErrIdempotencyConflict = errors.New("idempotency conflict")
	ErrAgentOutputInvalid  = errors.New("agent output invalid")
	ErrInternal            = errors.New("internal error")
)