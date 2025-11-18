package db

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Common domain errors that can be returned by repositories
var (
	// ErrNotFound indicates that the requested entity was not found in the database.
	ErrNotFound = errors.New("entity not found")

	// ErrConflict indicates a unique constraint violation (e.g., duplicate key).
	ErrConflict = errors.New("entity already exists")

	// ErrForeignKeyViolation indicates a foreign key constraint violation.
	ErrForeignKeyViolation = errors.New("foreign key constraint violation")

	// ErrCheckViolation indicates a check constraint violation.
	ErrCheckViolation = errors.New("check constraint violation")

	// ErrDeadlock indicates a database deadlock occurred.
	ErrDeadlock = errors.New("database deadlock detected")

	// ErrConnectionFailed indicates a connection failure to the database.
	ErrConnectionFailed = errors.New("database connection failed")
)

// PostgreSQL error codes
const (
	// PostgreSQL error codes (from https://www.postgresql.org/docs/current/errcodes-appendix.html)
	pgErrCodeUniqueViolation      = "23505" // unique_violation
	pgErrCodeForeignKeyViolation  = "23503" // foreign_key_violation
	pgErrCodeCheckViolation       = "23514" // check_violation
	pgErrCodeDeadlockDetected     = "40P01" // deadlock_detected
	pgErrCodeSerializationFailure = "40001" // serialization_failure
)

// TranslatePgError translates PostgreSQL-specific errors to domain errors.
// This function wraps database errors with additional context while preserving
// the original error for debugging purposes.
//
// Error Translation Rules:
//   - pgx.ErrNoRows → ErrNotFound
//   - 23505 (unique_violation) → ErrConflict
//   - 23503 (foreign_key_violation) → ErrForeignKeyViolation
//   - 23514 (check_violation) → ErrCheckViolation
//   - 40P01 (deadlock_detected) → ErrDeadlock
//   - Connection errors → ErrConnectionFailed
//   - Other errors → Original error (preserved)
//
// Parameters:
//   - err: The error to translate (typically from pgx operations)
//
// Returns:
//   - Translated domain error with original error wrapped for context
func TranslatePgError(err error) error {
	if err == nil {
		return nil
	}

	// Handle pgx.ErrNoRows (not found)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %v", ErrNotFound, err)
	}

	// Handle pgconn.PgError (PostgreSQL-specific errors)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return translatePgConnError(pgErr)
	}

	// Handle connection-related errors
	if isConnectionError(err) {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	// Return original error if no translation applies
	return err
}

// translatePgConnError translates pgconn.PgError to domain errors based on error codes.
func translatePgConnError(pgErr *pgconn.PgError) error {
	switch pgErr.Code {
	case pgErrCodeUniqueViolation:
		return fmt.Errorf("%w: %s (constraint: %s)", ErrConflict, pgErr.Message, pgErr.ConstraintName)

	case pgErrCodeForeignKeyViolation:
		return fmt.Errorf("%w: %s (constraint: %s)", ErrForeignKeyViolation, pgErr.Message, pgErr.ConstraintName)

	case pgErrCodeCheckViolation:
		return fmt.Errorf("%w: %s (constraint: %s)", ErrCheckViolation, pgErr.Message, pgErr.ConstraintName)

	case pgErrCodeDeadlockDetected, pgErrCodeSerializationFailure:
		return fmt.Errorf("%w: %s", ErrDeadlock, pgErr.Message)

	default:
		// Return original PostgreSQL error with context
		return fmt.Errorf("database error [%s]: %s (detail: %s)", pgErr.Code, pgErr.Message, pgErr.Detail)
	}
}

// isConnectionError checks if the error is related to connection failures.
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common connection error messages
	errMsg := err.Error()
	return contains(errMsg, "connection refused") ||
		contains(errMsg, "connection reset") ||
		contains(errMsg, "connection timeout") ||
		contains(errMsg, "no such host") ||
		contains(errMsg, "network is unreachable")
}

// contains checks if a string contains a substring (case-insensitive helper).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

// findSubstring performs a simple substring search.
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// IsNotFound checks if an error is an ErrNotFound error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsConflict checks if an error is an ErrConflict error.
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsForeignKeyViolation checks if an error is an ErrForeignKeyViolation error.
func IsForeignKeyViolation(err error) bool {
	return errors.Is(err, ErrForeignKeyViolation)
}

// IsCheckViolation checks if an error is an ErrCheckViolation error.
func IsCheckViolation(err error) bool {
	return errors.Is(err, ErrCheckViolation)
}

// IsDeadlock checks if an error is an ErrDeadlock error.
func IsDeadlock(err error) bool {
	return errors.Is(err, ErrDeadlock)
}

// IsConnectionFailed checks if an error is an ErrConnectionFailed error.
func IsConnectionFailed(err error) bool {
	return errors.Is(err, ErrConnectionFailed)
}
