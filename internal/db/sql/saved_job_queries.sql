-- name: GetJob :one
SELECT
    *
FROM
    jobs
WHERE
    id = $1;

-- name: SaveJob :exec
INSERT INTO user_saved_jobs (user_id, job_id)
    VALUES ($1, $2)
ON CONFLICT (user_id, job_id)
    DO NOTHING;

-- name: UnsaveJob :exec
DELETE FROM user_saved_jobs
WHERE user_id = $1
    AND job_id = $2;

-- name: ListSavedJobIDsAmong :many
SELECT
    job_id
FROM
    user_saved_jobs
WHERE
    user_id = sqlc.arg('user_id')
    AND job_id = ANY (sqlc.arg('job_ids')::uuid[]);
