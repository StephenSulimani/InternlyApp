CREATE TABLE IF NOT EXISTS user_saved_jobs (
    user_id uuid NOT NULL,
    job_id uuid NOT NULL,
    PRIMARY KEY (user_id, job_id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (job_id) REFERENCES jobs (id) ON DELETE CASCADE
);

CREATE INDEX idx_user_saved_jobs_job_id ON user_saved_jobs (job_id);

CREATE INDEX idx_user_saved_jobs_user_id ON user_saved_jobs (user_id);

