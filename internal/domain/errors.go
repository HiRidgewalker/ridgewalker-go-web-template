package domain

import "errors"

// Sentinel errors shared across layers.
//
// The domain defines what went wrong in business terms; adapters decide how to
// present that to their transport (e.g. HTTP status codes). Callers should use
// errors.Is so adapters can wrap these errors with context.
var (
	ErrNotFound   = errors.New("resource not found")
	ErrConflict   = errors.New("resource conflict")
	ErrValidation = errors.New("validation failed")
)
