CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    source_url text NOT NULL,
    source_name text NOT NULL,
    first_seen timestamptz DEFAULT NOW() NOT NULL,
    application_link text UNIQUE NOT NULL,
    company text,
    role_title text,
    locations text[],
    job_type text,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    embedding VECTOR (384)
);

CREATE INDEX idx_jobs_source ON jobs (source_name);

CREATE INDEX idx_jobs_type ON jobs (job_type);

