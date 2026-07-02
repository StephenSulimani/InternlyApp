-- name: GetUserCount :one
SELECT
    COUNT(*)
FROM
    users;

-- name: CreateUser :one
INSERT INTO users (first_name, last_name, email, password, is_active, is_admin, is_premium)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    *;

-- name: GetUserByEmail :one
SELECT
    *
FROM
    users
WHERE
    email = $1;

