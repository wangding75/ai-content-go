package memory

import (
	"errors"
	"fmt"
)

var (
	ErrValidation          = errors.New("validation error")
	ErrNotFound            = errors.New("not found")
	ErrForbidden           = fmt.Errorf("%w: forbidden", ErrValidation)
	ErrConflict            = errors.New("conflict")
	ErrIdempotencyConflict = errors.New("idempotency conflict")
)
