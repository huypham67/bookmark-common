package dbutils

import (
	"errors"
	"strings"
)

var errorFilter = []func(error) (bool, error){
	IsDuplicationError,
	IsForeignKeyViolationError,
	IsRecordNotFoundError,
}

var (
	ErrDuplicationType         = errors.New("duplicate type error")
	ErrRecordNotFoundType      = errors.New("record not found")
	ErrForeignKeyViolationType = errors.New("foreign key violation")
)

// ClassifyError maps database errors to application error types.
func ClassifyError(err error) error {
	for _, filter := range errorFilter {
		if isMatch, errType := filter(err); isMatch {
			return errType
		}
	}
	return err
}

// IsDuplicationError checks if an error is a duplicate key constraint violation.
func IsDuplicationError(err error) (bool, error) {
	errMsg := err.Error()
	return strings.Contains(errMsg, "duplicate key value violates unique constraint") ||
		strings.Contains(errMsg, "UNIQUE constraint failed"), ErrDuplicationType
}

// IsForeignKeyViolationError checks if an error is a foreign key constraint violation.
func IsForeignKeyViolationError(err error) (bool, error) {
	errMsg := err.Error()
	return strings.Contains(errMsg, "violates foreign key constraint") ||
		strings.Contains(errMsg, "FOREIGN KEY constraint failed"), ErrForeignKeyViolationType
}

// IsRecordNotFoundError checks if an error indicates a record was not found.
func IsRecordNotFoundError(err error) (bool, error) {
	return strings.Contains(err.Error(), "record not found"), ErrRecordNotFoundType
}
