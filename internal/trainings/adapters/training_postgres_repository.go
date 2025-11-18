package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vaintrub/go-ddd-template/internal/common/db"
	sqlc_trainings "github.com/vaintrub/go-ddd-template/internal/trainings/adapters/sqlc"
	"github.com/vaintrub/go-ddd-template/internal/trainings/app/query"
	"github.com/vaintrub/go-ddd-template/internal/trainings/domain/training"
)

// TrainingPostgresRepository implements training.Repository using PostgreSQL.
// It wraps SQLC-generated code to provide a clean domain repository interface.
type TrainingPostgresRepository struct {
	pool *pgxpool.Pool
}

// NewTrainingPostgresRepository creates a new PostgreSQL-backed training repository.
func NewTrainingPostgresRepository(pool *pgxpool.Pool) *TrainingPostgresRepository {
	return &TrainingPostgresRepository{pool: pool}
}

// AddTraining persists a new training to the database.
// Implements training.Repository interface.
func (r *TrainingPostgresRepository) AddTraining(ctx context.Context, tr *training.Training) error {
	queries := sqlc_trainings.New(r.pool)

	id := db.UUIDToPgtype(uuid.MustParse(tr.UUID()))
	userID := db.UUIDToPgtype(uuid.MustParse(tr.UserUUID()))

	var notes *string
	if tr.Notes() != "" {
		notes = &[]string{tr.Notes()}[0]
	}

	var proposedNewTime pgtype.Timestamptz
	if !tr.ProposedNewTime().IsZero() {
		proposedNewTime = pgtype.Timestamptz{Time: tr.ProposedNewTime(), Valid: true}
	}

	var moveProposedBy *string
	if !tr.MovedProposedBy().IsZero() {
		moveProposedBy = &[]string{tr.MovedProposedBy().String()}[0]
	}

	_, err := queries.CreateTraining(ctx, id, userID, tr.UserName(), tr.Time(), notes, proposedNewTime, moveProposedBy, tr.IsCanceled())
	if err != nil {
		return db.TranslatePgError(err)
	}

	return nil
}

// GetTraining retrieves a training by UUID for the specified user.
// Implements training.Repository interface.
func (r *TrainingPostgresRepository) GetTraining(ctx context.Context, trainingUUID string, user training.User) (*training.Training, error) {
	queries := sqlc_trainings.New(r.pool)

	id, err := db.StringToPgtypeUUID(trainingUUID)
	if err != nil {
		return nil, fmt.Errorf("invalid training UUID: %w", err)
	}

	row, err := queries.GetTraining(ctx, id)
	if err != nil {
		if db.IsNotFound(db.TranslatePgError(err)) {
			return nil, training.NotFoundError{TrainingUUID: trainingUUID}
		}
		return nil, db.TranslatePgError(err)
	}

	tr, err := unmarshalTraining(row)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal training: %w", err)
	}

	// Check if user can see this training
	if err := training.CanUserSeeTraining(user, *tr); err != nil {
		return nil, err
	}

	return tr, nil
}

// UpdateTraining updates an existing training using the provided update function.
// Implements training.Repository interface.
func (r *TrainingPostgresRepository) UpdateTraining(
	ctx context.Context,
	trainingUUID string,
	user training.User,
	updateFn func(ctx context.Context, tr *training.Training) (*training.Training, error),
) error {
	// Use a transaction for read-modify-write pattern
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx) // Rollback is safe to call even if commit succeeds
	}()

	queries := sqlc_trainings.New(tx)

	// Get the current training
	id, err := db.StringToPgtypeUUID(trainingUUID)
	if err != nil {
		return fmt.Errorf("invalid training UUID: %w", err)
	}

	row, err := queries.GetTraining(ctx, id)
	if err != nil {
		if db.IsNotFound(db.TranslatePgError(err)) {
			return training.NotFoundError{TrainingUUID: trainingUUID}
		}
		return db.TranslatePgError(err)
	}

	tr, err := unmarshalTraining(row)
	if err != nil {
		return fmt.Errorf("failed to unmarshal training: %w", err)
	}

	// Check if user can see this training
	if err := training.CanUserSeeTraining(user, *tr); err != nil {
		return err
	}

	// Apply the update function
	updatedTr, err := updateFn(ctx, tr)
	if err != nil {
		return err
	}

	// Persist the changes
	var notes *string
	if updatedTr.Notes() != "" {
		notes = &[]string{updatedTr.Notes()}[0]
	}

	var proposedNewTime pgtype.Timestamptz
	if !updatedTr.ProposedNewTime().IsZero() {
		proposedNewTime = pgtype.Timestamptz{Time: updatedTr.ProposedNewTime(), Valid: true}
	}

	var moveProposedBy *string
	if !updatedTr.MovedProposedBy().IsZero() {
		moveProposedBy = &[]string{updatedTr.MovedProposedBy().String()}[0]
	}

	if err := queries.UpdateTraining(ctx, id, notes, proposedNewTime, moveProposedBy, updatedTr.IsCanceled()); err != nil {
		return db.TranslatePgError(err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// unmarshalTraining converts SQLC TrainingsTraining to domain Training entity.
func unmarshalTraining(row sqlc_trainings.TrainingsTraining) (*training.Training, error) {
	// Extract notes
	notes := ""
	if row.Notes != nil {
		notes = *row.Notes
	}

	// Extract proposed new time (using pgtype.Timestamptz)
	proposedNewTime := row.ProposedNewTime.Time
	// Zero value is fine if not valid

	// Extract and parse move proposed by
	moveProposedBy := training.UserType{}
	if row.MoveProposedBy != nil && *row.MoveProposedBy != "" {
		var err error
		moveProposedBy, err = training.NewUserTypeFromString(*row.MoveProposedBy)
		if err != nil {
			return nil, fmt.Errorf("invalid move proposed by value: %w", err)
		}
	}

	// Convert pgtype.UUID to string
	idStr := db.PgtypeToUUID(row.ID).String()
	userIDStr := db.PgtypeToUUID(row.UserID).String()

	// Use the domain unmarshal function
	tr, err := training.UnmarshalTrainingFromDatabase(
		idStr,
		userIDStr,
		row.UserName,
		row.TrainingTime,
		notes,
		row.Canceled,
		proposedNewTime,
		moveProposedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal training from database: %w", err)
	}

	return tr, nil
}

// WithTransaction creates a new repository instance that uses the provided transaction.
// This allows repository operations to participate in a larger transaction.
func (r *TrainingPostgresRepository) WithTransaction(tx pgx.Tx) *TrainingPostgresRepository {
	return &TrainingPostgresRepository{
		pool: nil, // Not used when transaction is provided
	}
}

// AllTrainings implements the AllTrainingsReadModel interface for queries.
// It returns all trainings in the system.
func (r *TrainingPostgresRepository) AllTrainings(ctx context.Context) ([]query.Training, error) {
	queries := sqlc_trainings.New(r.pool)

	rows, err := queries.ListAllTrainings(ctx)
	if err != nil {
		return nil, db.TranslatePgError(err)
	}

	trainings := make([]query.Training, 0, len(rows))
	for _, row := range rows {
		trainings = append(trainings, rowToQueryTraining(row))
	}

	return trainings, nil
}

// FindTrainingsForUser implements the TrainingsForUserReadModel interface for queries.
// It returns all trainings for a specific user.
func (r *TrainingPostgresRepository) FindTrainingsForUser(ctx context.Context, userUUID string) ([]query.Training, error) {
	queries := sqlc_trainings.New(r.pool)

	uid, err := db.StringToPgtypeUUID(userUUID)
	if err != nil {
		return nil, fmt.Errorf("invalid user UUID: %w", err)
	}

	// Use cursor pagination with no cursor (get all) and large limit
	rows, err := queries.ListTrainingsByUser(ctx, uid, time.Time{}, pgtype.UUID{}, 1000)
	if err != nil {
		return nil, db.TranslatePgError(err)
	}

	trainings := make([]query.Training, 0, len(rows))
	for _, row := range rows {
		trainings = append(trainings, rowToQueryTraining(row))
	}

	return trainings, nil
}

// rowToQueryTraining converts a SQLC row to a query.Training DTO.
func rowToQueryTraining(row sqlc_trainings.TrainingsTraining) query.Training {
	var notes string
	if row.Notes != nil {
		notes = *row.Notes
	}

	var proposedTime *time.Time
	if row.ProposedNewTime.Valid {
		proposedTime = &row.ProposedNewTime.Time
	}

	var moveProposedBy *string
	if row.MoveProposedBy != nil {
		moveProposedBy = row.MoveProposedBy
	}

	return query.Training{
		UUID:           db.PgtypeToUUID(row.ID).String(),
		UserUUID:       db.PgtypeToUUID(row.UserID).String(),
		User:           row.UserName,
		Time:           row.TrainingTime,
		Notes:          notes,
		ProposedTime:   proposedTime,
		MoveProposedBy: moveProposedBy,
		CanBeCancelled: !row.Canceled, // If not already canceled, it can be cancelled
	}
}
