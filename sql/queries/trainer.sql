-- Trainer Context Queries
-- Purpose: CRUD operations for trainer_hours table

-- name: CreateHour :one
INSERT INTO trainer_hours (
    id,
    hour_time,
    availability,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, NOW(), NOW()
) RETURNING *;

-- name: GetHour :one
SELECT * FROM trainer_hours
WHERE id = $1;

-- name: GetHourByTime :one
SELECT * FROM trainer_hours
WHERE hour_time = $1;

-- name: UpdateHourAvailability :exec
UPDATE trainer_hours
SET
    availability = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: ListHours :many
SELECT * FROM trainer_hours
WHERE (created_at > $1 OR $1 IS NULL)
  AND (id > $2 OR $2 IS NULL)
ORDER BY created_at, id
LIMIT $3;

-- name: ListHoursByTimeRange :many
SELECT * FROM trainer_hours
WHERE hour_time BETWEEN $1 AND $2
ORDER BY hour_time;

-- name: DeleteHour :exec
DELETE FROM trainer_hours
WHERE id = $1;
