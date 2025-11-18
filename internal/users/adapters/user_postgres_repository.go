package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vaintrub/go-ddd-template/internal/common/db"
	sqlc_users "github.com/vaintrub/go-ddd-template/internal/users/adapters/sqlc"
)

// UserPostgresRepository implements user repository using PostgreSQL.
// It wraps SQLC-generated code to provide a clean repository interface.
type UserPostgresRepository struct {
	pool *pgxpool.Pool
}

// NewUserPostgresRepository creates a new PostgreSQL-backed user repository.
func NewUserPostgresRepository(pool *pgxpool.Pool) *UserPostgresRepository {
	return &UserPostgresRepository{pool: pool}
}

// CreateUser persists a new user to the database.
func (r *UserPostgresRepository) CreateUser(ctx context.Context, id, userType, name, email string) error {
	return r.CreateUserWithBalance(ctx, id, userType, name, email, 0)
}

// CreateUserWithBalance persists a new user with initial balance to the database.
func (r *UserPostgresRepository) CreateUserWithBalance(ctx context.Context, id, userType, name, email string, balance int) error {
	queries := sqlc_users.New(r.pool)

	uid, err := db.StringToPgtypeUUID(id)
	if err != nil {
		return fmt.Errorf("invalid user UUID: %w", err)
	}

	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}

	var balancePtr *int32
	// #nosec G115 - balance is a domain-validated value, overflow unlikely
	bal := int32(balance)
	balancePtr = &bal

	_, err = queries.CreateUser(ctx, uid, userType, name, emailPtr, balancePtr)
	if err != nil {
		return db.TranslatePgError(err)
	}

	return nil
}

// GetUser retrieves a user by UUID.
func (r *UserPostgresRepository) GetUser(ctx context.Context, userID string) (*sqlc_users.UsersUser, error) {
	queries := sqlc_users.New(r.pool)

	uid, err := db.StringToPgtypeUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user UUID: %w", err)
	}

	user, err := queries.GetUser(ctx, uid)
	if err != nil {
		return nil, db.TranslatePgError(err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email address.
func (r *UserPostgresRepository) GetUserByEmail(ctx context.Context, email string) (*sqlc_users.UsersUser, error) {
	queries := sqlc_users.New(r.pool)

	user, err := queries.GetUserByEmail(ctx, &email)
	if err != nil {
		return nil, db.TranslatePgError(err)
	}

	return &user, nil
}

// ListUsersByType retrieves users filtered by type with cursor pagination.
func (r *UserPostgresRepository) ListUsersByType(ctx context.Context, userType string, cursorTime time.Time, cursorID pgtype.UUID, limit int32) ([]sqlc_users.UsersUser, error) {
	queries := sqlc_users.New(r.pool)

	users, err := queries.ListUsersByType(ctx, userType, cursorTime, cursorID, limit)
	if err != nil {
		return nil, db.TranslatePgError(err)
	}

	return users, nil
}

// UpdateUser updates an existing user.
func (r *UserPostgresRepository) UpdateUser(ctx context.Context, userID, name, email string) error {
	queries := sqlc_users.New(r.pool)

	uid, err := db.StringToPgtypeUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user UUID: %w", err)
	}

	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}

	err = queries.UpdateUser(ctx, uid, name, emailPtr)
	if err != nil {
		return db.TranslatePgError(err)
	}

	return nil
}

// DeleteUser removes a user from the database.
func (r *UserPostgresRepository) DeleteUser(ctx context.Context, userID string) error {
	queries := sqlc_users.New(r.pool)

	uid, err := db.StringToPgtypeUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user UUID: %w", err)
	}

	err = queries.DeleteUser(ctx, uid)
	if err != nil {
		return db.TranslatePgError(err)
	}

	return nil
}

// GetByID retrieves a user by UUID (alias for GetUser for compatibility).
func (r *UserPostgresRepository) GetByID(ctx context.Context, userID string) (*sqlc_users.UsersUser, error) {
	return r.GetUser(ctx, userID)
}

// UpdateBalance updates a user's balance by adding the specified amount (can be negative).
func (r *UserPostgresRepository) UpdateBalance(ctx context.Context, userID string, amountChange int) error {
	queries := sqlc_users.New(r.pool)

	uid, err := db.StringToPgtypeUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user UUID: %w", err)
	}

	// #nosec G115 - amountChange is a domain-validated value, overflow unlikely
	err = queries.UpdateBalance(ctx, uid, int32(amountChange))
	if err != nil {
		return db.TranslatePgError(err)
	}

	return nil
}

// UpdateLastIP updates the last IP address for a user.
func (r *UserPostgresRepository) UpdateLastIP(ctx context.Context, userID string, ip string) error {
	queries := sqlc_users.New(r.pool)

	uid, err := db.StringToPgtypeUUID(userID)
	if err != nil {
		return fmt.Errorf("invalid user UUID: %w", err)
	}

	err = queries.UpdateLastIP(ctx, uid, &ip)
	if err != nil {
		return db.TranslatePgError(err)
	}

	return nil
}
