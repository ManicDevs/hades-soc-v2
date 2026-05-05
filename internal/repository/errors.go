package repository

import (
	"errors"
	"fmt"
)

// Common repository errors
var (
	// ErrNotFound indicates that the requested entity was not found
	ErrNotFound = errors.New("entity not found")

	// ErrAlreadyExists indicates that the entity already exists
	ErrAlreadyExists = errors.New("entity already exists")

	// ErrInvalidID indicates that the provided ID is invalid
	ErrInvalidID = errors.New("invalid entity ID")

	// ErrInvalidInput indicates that the input parameters are invalid
	ErrInvalidInput = errors.New("invalid input parameters")

	// ErrTransactionFailed indicates that a transaction failed
	ErrTransactionFailed = errors.New("transaction failed")

	// ErrConnectionLost indicates that the database connection was lost
	ErrConnectionLost = errors.New("database connection lost")

	// ErrTimeout indicates that the operation timed out
	ErrTimeout = errors.New("operation timed out")

	// ErrConstraintViolation indicates a constraint violation (e.g., unique constraint)
	ErrConstraintViolation = errors.New("constraint violation")
)

// RepositoryError wraps errors with additional context
type RepositoryError struct {
	Op  string // Operation that failed
	Err error  // Underlying error
	Msg string // Additional message
}

// Error implements the error interface
func (e *RepositoryError) Error() string {
	if e.Msg != "" {
		return fmt.Sprintf("repository error [%s]: %s: %v", e.Op, e.Msg, e.Err)
	}
	return fmt.Sprintf("repository error [%s]: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error
func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// NewRepositoryError creates a new repository error
func NewRepositoryError(op string, err error, msg string) *RepositoryError {
	return &RepositoryError{
		Op:  op,
		Err: err,
		Msg: msg,
	}
}

// IsNotFound checks if the error is ErrNotFound
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists checks if the error is ErrAlreadyExists
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsConstraintViolation checks if the error is ErrConstraintViolation
func IsConstraintViolation(err error) bool {
	return errors.Is(err, ErrConstraintViolation)
}
