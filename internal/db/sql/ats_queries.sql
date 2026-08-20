-- name: UpsertCompanyATS :one
INSERT INTO company_ats (company_name, ats_name, ats_url)
    VALUES ($1, $2, $3)
ON CONFLICT (ats_url)
    DO UPDATE SET
        last_seen = NOW(),
        company_name = EXCLUDED.company_name
    RETURNING
        *;

-- name: ListWorkingCompanyATS :many
SELECT
    *
FROM
    company_ats
WHERE
    working = TRUE
ORDER BY
    last_seen DESC;

-- name: SetCompanyATSWorking :exec
UPDATE
    company_ats
SET
    working = $2
WHERE
    id = $1;
