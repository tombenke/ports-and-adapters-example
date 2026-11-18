package domain

import (
	"fmt"
)

// DomainError wraps a domain error with additional context.
type DomainError struct {
	Op  string // operation that failed
	Err error  // underlying domain error
}

// Error implements the error interface for DomainError.
func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Op, e.Err)
}

// Unwrap allows errors.Is and errors.As to work with DomainError.
func (e *DomainError) Unwrap() error {
	return e.Err
}
