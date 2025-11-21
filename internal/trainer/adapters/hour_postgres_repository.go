package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vaintrub/go-ddd-template/internal/common/db"
	sqlc_trainer "github.com/vaintrub/go-ddd-template/internal/trainer/adapters/sqlc"
	"github.com/vaintrub/go-ddd-template/internal/trainer/app/query"
	"github.com/vaintrub/go-ddd-template/internal/trainer/domain/hour"
)

// HourPostgresRepository implements hour repository using PostgreSQL.
// It wraps SQLC-generated code to provide a clean repository interface.
type HourPostgresRepository struct {
	pool    *pgxpool.Pool
	factory hour.Factory
}

// NewHourPostgresRepository creates a new PostgreSQL-backed hour repository.
func NewHourPostgresRepository(pool *pgxpool.Pool, factory hour.Factory) *HourPostgresRepository {
	return &HourPostgresRepository{
		pool:    pool,
		factory: factory,
	}
}

// CreateHour persists a new hour to the database.
func (r *HourPostgresRepository) CreateHour(ctx context.Context, id string, hourTime time.Time, availability string) error {
	queries := sqlc_trainer.New(r.pool)

	uid, err := db.StringToPgtypeUUID(id)
	if err != nil {
		return fmt.Errorf("invalid hour UUID: %w", err)
	}

	_, err = queries.CreateHour(ctx, uid, hourTime, availability)
	if err != nil {
		return db.TranslatePgError(err)
	}

	return nil
}

// GetHour retrieves an hour by time and returns domain Hour object.
// This implements the hour.Repository interface.
func (r *HourPostgresRepository) GetHour(ctx context.Context, hourTime time.Time) (*hour.Hour, error) {
	queries := sqlc_trainer.New(r.pool)

	dbHour, err := queries.GetHourByTime(ctx, hourTime)
	if err != nil {
		return nil, db.TranslatePgError(err)
	}

	// Convert database availability string to domain Availability
	availability, err := hour.NewAvailabilityFromString(dbHour.Availability)
	if err != nil {
		return nil, fmt.Errorf("invalid availability in database: %w", err)
	}

	// Unmarshal from database using factory
	domainHour, err := r.factory.UnmarshalHourFromDatabase(dbHour.HourTime, availability)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal hour from database: %w", err)
	}

	return domainHour, nil
}

// UpdateHour updates an hour using the provided update function.
// This implements the hour.Repository interface.
// Uses a transaction to prevent race conditions.
// If the hour doesn't exist, creates a new NotAvailableHour first (upsert behavior).
func (r *HourPostgresRepository) UpdateHour(
	ctx context.Context,
	hourTime time.Time,
	updateFn func(h *hour.Hour) (*hour.Hour, error),
) error {
	// Begin transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx) // Rollback is safe to call even after commit
	}()

	queries := sqlc_trainer.New(tx)

	// Try to get current hour within transaction
	dbHour, err := queries.GetHourByTime(ctx, hourTime)
	if err != nil { //nolint:nestif // Transaction logic requires nested error handling
		translatedErr := db.TranslatePgError(err)
		// If hour doesn't exist, create a new NotAvailableHour
		if db.IsNotFound(translatedErr) {
			// Create new hour with NotAvailable state
			newHour, createErr := r.factory.NewNotAvailableHour(hourTime)
			if createErr != nil {
				return fmt.Errorf("failed to create new hour: %w", createErr)
			}

			// Insert into database and get the created record
			uid := db.UUIDToPgtype(uuid.New())

			dbHour, err = queries.CreateHour(ctx, uid, hourTime, newHour.Availability().String())
			if err != nil {
				return db.TranslatePgError(err)
			}
		} else {
			return translatedErr
		}
	}

	// Convert to domain object
	availability, err := hour.NewAvailabilityFromString(dbHour.Availability)
	if err != nil {
		return fmt.Errorf("invalid availability in database: %w", err)
	}

	currentHour, err := r.factory.UnmarshalHourFromDatabase(dbHour.HourTime, availability)
	if err != nil {
		return fmt.Errorf("failed to unmarshal hour from database: %w", err)
	}

	// Apply update function
	updatedHour, err := updateFn(currentHour)
	if err != nil {
		return err
	}

	// Update availability within transaction
	err = queries.UpdateHourAvailability(ctx, dbHour.ID, updatedHour.Availability().String())
	if err != nil {
		return db.TranslatePgError(err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateHourAvailability updates an hour's availability status.
func (r *HourPostgresRepository) UpdateHourAvailability(ctx context.Context, hourID string, availability string) error {
	queries := sqlc_trainer.New(r.pool)

	uid, err := db.StringToPgtypeUUID(hourID)
	if err != nil {
		return fmt.Errorf("invalid hour UUID: %w", err)
	}

	err = queries.UpdateHourAvailability(ctx, uid, availability)
	if err != nil {
		return db.TranslatePgError(err)
	}

	return nil
}

// ListHours retrieves hours with cursor pagination.
func (r *HourPostgresRepository) ListHours(ctx context.Context, cursorTime time.Time, cursorID uuid.UUID, limit int32) ([]sqlc_trainer.TrainerHour, error) {
	queries := sqlc_trainer.New(r.pool)

	hours, err := queries.ListHours(ctx, cursorTime, db.UUIDToPgtype(cursorID), limit)
	if err != nil {
		return nil, db.TranslatePgError(err)
	}

	return hours, nil
}

// ListHoursByTimeRange retrieves hours within a time range.
func (r *HourPostgresRepository) ListHoursByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]sqlc_trainer.TrainerHour, error) {
	queries := sqlc_trainer.New(r.pool)

	hours, err := queries.ListHoursByTimeRange(ctx, startTime, endTime)
	if err != nil {
		return nil, db.TranslatePgError(err)
	}

	return hours, nil
}

// DeleteHour removes an hour from the database.
func (r *HourPostgresRepository) DeleteHour(ctx context.Context, hourID string) error {
	queries := sqlc_trainer.New(r.pool)

	uid, err := db.StringToPgtypeUUID(hourID)
	if err != nil {
		return fmt.Errorf("invalid hour UUID: %w", err)
	}

	err = queries.DeleteHour(ctx, uid)
	if err != nil {
		return db.TranslatePgError(err)
	}

	return nil
}

// AvailableHours implements the AvailableHoursReadModel interface for queries.
// It returns dates that have available hours within the given time range.
func (r *HourPostgresRepository) AvailableHours(ctx context.Context, from time.Time, to time.Time) ([]query.Date, error) {
	hours, err := r.ListHoursByTimeRange(ctx, from, to)
	if err != nil {
		return nil, err
	}

	// Group hours by date
	dateMap := make(map[string]*query.Date)
	for _, h := range hours {
		dateStr := h.HourTime.Format("2006-01-02")

		// Get or create date entry
		date, exists := dateMap[dateStr]
		if !exists {
			parsedDate, parseErr := time.Parse("2006-01-02", dateStr)
			if parseErr != nil {
				continue
			}
			date = &query.Date{
				Date:         parsedDate,
				HasFreeHours: false,
				Hours:        []query.Hour{},
			}
			dateMap[dateStr] = date
		}

		// Add hour to date
		isAvailable := h.Availability == "available"
		hasTraining := h.Availability == "training_scheduled"

		date.Hours = append(date.Hours, query.Hour{
			Available:            isAvailable,
			HasTrainingScheduled: hasTraining,
			Hour:                 h.HourTime,
		})

		// Update HasFreeHours if this hour is available
		if isAvailable {
			date.HasFreeHours = true
		}
	}

	// Convert map to slice
	dates := make([]query.Date, 0, len(dateMap))
	for _, date := range dateMap {
		dates = append(dates, *date)
	}

	return dates, nil
}
