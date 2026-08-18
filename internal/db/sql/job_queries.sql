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

-- name: CountJobs :one
SELECT
    COUNT(*)::bigint
FROM
    jobs
WHERE
    (sqlc.arg('q')::text = ''
        OR company ILIKE sqlc.arg('q') ESCAPE '\'
        OR role_title ILIKE sqlc.arg('q') ESCAPE '\'
        OR COALESCE(description, '') ILIKE sqlc.arg('q') ESCAPE '\'
        OR source_name ILIKE sqlc.arg('q') ESCAPE '\'
        OR COALESCE(job_type, '') ILIKE sqlc.arg('q') ESCAPE '\'
        OR EXISTS (
            SELECT
                1
            FROM
                unnest(COALESCE(locations, ARRAY[]::text[])) AS loc
            WHERE
                loc ILIKE sqlc.arg('q') ESCAPE '\'))
    AND (sqlc.arg('filter_type')::text = ''
        OR job_type = sqlc.arg('filter_type'))
    AND (COALESCE(cardinality(sqlc.arg('filter_locations')::text[]), 0) = 0
        OR EXISTS (
            SELECT
                1
            FROM
                unnest(COALESCE(locations, ARRAY[]::text[])) AS loc
                JOIN unnest(sqlc.arg('filter_locations')::text[]) AS needle ON loc ILIKE needle ESCAPE '\'))
    AND (sqlc.arg('filter_source')::text = ''
        OR source_name = sqlc.arg('filter_source'))
    AND (sqlc.arg('recency_hours')::int = 0
        OR first_seen >= NOW() - (sqlc.arg('recency_hours')::int * INTERVAL '1 hour'))
    AND (NOT sqlc.arg('filter_saved')::bool
        OR EXISTS (
            SELECT
                1
            FROM
                user_saved_jobs s
            WHERE
                s.user_id = sqlc.arg('user_id')
                AND s.job_id = id));

-- name: SearchJobs :many
SELECT
    *
FROM
    jobs
WHERE
    (sqlc.arg('q')::text = ''
        OR company ILIKE sqlc.arg('q') ESCAPE '\'
        OR role_title ILIKE sqlc.arg('q') ESCAPE '\'
        OR COALESCE(description, '') ILIKE sqlc.arg('q') ESCAPE '\'
        OR source_name ILIKE sqlc.arg('q') ESCAPE '\'
        OR COALESCE(job_type, '') ILIKE sqlc.arg('q') ESCAPE '\'
        OR EXISTS (
            SELECT
                1
            FROM
                unnest(COALESCE(locations, ARRAY[]::text[])) AS loc
            WHERE
                loc ILIKE sqlc.arg('q') ESCAPE '\'))
    AND (sqlc.arg('filter_type')::text = ''
        OR job_type = sqlc.arg('filter_type'))
    AND (COALESCE(cardinality(sqlc.arg('filter_locations')::text[]), 0) = 0
        OR EXISTS (
            SELECT
                1
            FROM
                unnest(COALESCE(locations, ARRAY[]::text[])) AS loc
                JOIN unnest(sqlc.arg('filter_locations')::text[]) AS needle ON loc ILIKE needle ESCAPE '\'))
    AND (sqlc.arg('filter_source')::text = ''
        OR source_name = sqlc.arg('filter_source'))
    AND (sqlc.arg('recency_hours')::int = 0
        OR first_seen >= NOW() - (sqlc.arg('recency_hours')::int * INTERVAL '1 hour'))
    AND (NOT sqlc.arg('filter_saved')::bool
        OR EXISTS (
            SELECT
                1
            FROM
                user_saved_jobs s
            WHERE
                s.user_id = sqlc.arg('user_id')
                AND s.job_id = id))
ORDER BY
    CASE
    WHEN sqlc.arg('sort_by')::text = 'company'
        AND sqlc.arg('sort_dir')::text = 'asc' THEN
        lower(COALESCE(company, ''))
    WHEN sqlc.arg('sort_by')::text = 'role'
        AND sqlc.arg('sort_dir')::text = 'asc' THEN
        lower(COALESCE(role_title, ''))
    WHEN sqlc.arg('sort_by')::text = 'type'
        AND sqlc.arg('sort_dir')::text = 'asc' THEN
        lower(COALESCE(job_type, ''))
    WHEN sqlc.arg('sort_by')::text = 'location'
        AND sqlc.arg('sort_dir')::text = 'asc' THEN
        lower(COALESCE(locations[1], ''))
    END ASC NULLS LAST,
    CASE
    WHEN sqlc.arg('sort_by')::text = 'company'
        AND sqlc.arg('sort_dir')::text = 'desc' THEN
        lower(COALESCE(company, ''))
    WHEN sqlc.arg('sort_by')::text = 'role'
        AND sqlc.arg('sort_dir')::text = 'desc' THEN
        lower(COALESCE(role_title, ''))
    WHEN sqlc.arg('sort_by')::text = 'type'
        AND sqlc.arg('sort_dir')::text = 'desc' THEN
        lower(COALESCE(job_type, ''))
    WHEN sqlc.arg('sort_by')::text = 'location'
        AND sqlc.arg('sort_dir')::text = 'desc' THEN
        lower(COALESCE(locations[1], ''))
    END DESC NULLS LAST,
    CASE
    WHEN sqlc.arg('sort_by')::text = 'posted'
        AND sqlc.arg('sort_dir')::text = 'asc' THEN
        first_seen
    END ASC NULLS LAST,
    CASE
    WHEN sqlc.arg('sort_by')::text = 'posted'
        AND sqlc.arg('sort_dir')::text = 'desc' THEN
        first_seen
    END DESC NULLS LAST,
    first_seen DESC,
    id ASC
LIMIT sqlc.arg('row_limit') OFFSET sqlc.arg('row_offset');

-- name: ListJobLocations :many
SELECT DISTINCT
    TRIM(loc) AS location
FROM
    jobs
    CROSS JOIN LATERAL unnest(COALESCE(locations, ARRAY[]::text[])) AS loc
WHERE
    TRIM(loc) <> ''
ORDER BY
    location;

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

