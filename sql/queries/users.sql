-- Users Context Queries
-- Purpose: CRUD operations for users_users table

-- name: CreateUser :one
INSERT INTO users_users (
    id,
    user_type,
    name,
    email,
    balance,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, COALESCE($5, 0), NOW(), NOW()
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users_users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users_users
WHERE email = $1;

-- name: ListUsersByType :many
SELECT * FROM users_users
WHERE user_type = $1
  AND (created_at > $2 OR $2 IS NULL)
  AND (id > $3 OR $3 IS NULL)
ORDER BY created_at, id
LIMIT $4;

-- name: UpdateUser :exec
UPDATE users_users
SET
    name = COALESCE($2, name),
    email = COALESCE($3, email),
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users_users
WHERE id = $1;

-- name: UpdateBalance :exec
UPDATE users_users
SET
    balance = balance + $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateLastIP :exec
UPDATE users_users
SET
    last_ip = $2,
    updated_at = NOW()
WHERE id = $1;
