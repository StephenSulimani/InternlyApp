-- name: GetJobsByCompany :many
SELECT
    *
FROM
    jobs
WHERE
    company = $1;

-- name: CreateJob :one
INSERT INTO jobs (source_url, source_name, first_seen, application_link, company, role_title, locations, job_type, is_ats, metadata)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, COALESCE($10, '{}'::jsonb))
RETURNING
    *;

-- name: GetJobsLimit :many
SELECT
    *
FROM
    jobs
ORDER BY
    first_seen DESC
LIMIT $1;

-- name: GetJobsStats :one
SELECT
    (
        SELECT
            COUNT(*)::bigint
        FROM
            jobs) AS total_jobs,
    (
        SELECT
            COUNT(*)::bigint
        FROM
            jobs
        WHERE
            first_seen >= date_trunc('week', NOW())) AS added_this_week,
    (
        SELECT
            COUNT(DISTINCT company)::bigint
        FROM
            jobs) AS total_companies,
    (
        SELECT
            first_seen
        FROM
            jobs
        ORDER BY
            first_seen DESC
        LIMIT 1) AS last_updated;

