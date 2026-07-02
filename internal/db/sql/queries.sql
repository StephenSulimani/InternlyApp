-- name: GetJobsByCompany :many
SELECT
    *
FROM
    jobs
WHERE
    company = $1;

-- name: CreateJob :one
INSERT INTO jobs (source_url, source_name, first_seen, application_link, company, role_title, locations, job_type, metadata)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, COALESCE($9, '{}'::jsonb) -- If $9 is null, use the default
)
RETURNING
    *;

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

