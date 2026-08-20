CREATE TABLE IF NOT EXISTS user_saved_searches (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    name text NOT NULL,
    q text NOT NULL DEFAULT '',
    job_type text NOT NULL DEFAULT '',
    location text NOT NULL DEFAULT '',
    recency text NOT NULL DEFAULT '',
    saved_only boolean NOT NULL DEFAULT FALSE,
    sort_by text NOT NULL DEFAULT 'posted',
    sort_dir text NOT NULL DEFAULT 'desc',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    UNIQUE (user_id, name)
);

CREATE INDEX idx_user_saved_searches_user_id ON user_saved_searches (user_id);
