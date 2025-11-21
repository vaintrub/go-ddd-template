-- Trainings Context Queries
-- Purpose: CRUD operations for trainings_trainings table

-- name: CreateTraining :one
INSERT INTO trainings_trainings (
    id,
    user_id,
    user_name,
    training_time,
    notes,
    proposed_new_time,
    move_proposed_by,
    canceled,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
) RETURNING *;

-- name: GetTraining :one
SELECT * FROM trainings_trainings
WHERE id = $1;

-- name: UpdateTraining :exec
UPDATE trainings_trainings
SET
    notes = COALESCE($2, notes),
    proposed_new_time = COALESCE($3, proposed_new_time),
    move_proposed_by = COALESCE($4, move_proposed_by),
    canceled = COALESCE($5, canceled),
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteTraining :exec
DELETE FROM trainings_trainings
WHERE id = $1;

-- name: ListTrainingsByUser :many
SELECT * FROM trainings_trainings
WHERE user_id = $1
  AND canceled = false
  AND (created_at > $2 OR $2 IS NULL)
  AND (id > $3 OR $3 IS NULL)
ORDER BY created_at, id
LIMIT $4;

-- name: ListAllTrainings :many
SELECT * FROM trainings_trainings
WHERE canceled = false
ORDER BY created_at DESC, id;
