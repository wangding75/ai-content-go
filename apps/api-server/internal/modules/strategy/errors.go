package strategy

import "errors"

var (
	ErrValidation          = errors.New("validation error")
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrIdempotencyConflict = errors.New("idempotency conflict")
	ErrInternal            = errors.New("internal error")
)
