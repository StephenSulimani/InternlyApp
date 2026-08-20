-- name: ListUserSavedSearches :many
SELECT
    *
FROM
    user_saved_searches
WHERE
    user_id = $1
ORDER BY
    name ASC;

-- name: GetUserSavedSearch :one
SELECT
    *
FROM
    user_saved_searches
WHERE
    id = $1
    AND user_id = $2;

-- name: CreateUserSavedSearch :one
INSERT INTO user_saved_searches (user_id, name, q, job_type, location, recency, saved_only, sort_by, sort_dir)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING
    *;

-- name: UpdateUserSavedSearch :one
UPDATE
    user_saved_searches
SET
    name = $3,
    q = $4,
    job_type = $5,
    location = $6,
    recency = $7,
    saved_only = $8,
    sort_by = $9,
    sort_dir = $10,
    updated_at = NOW()
WHERE
    id = $1
    AND user_id = $2
RETURNING
    *;

-- name: DeleteUserSavedSearch :execrows
DELETE FROM user_saved_searches
WHERE id = $1
    AND user_id = $2;
