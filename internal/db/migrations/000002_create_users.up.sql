CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    first_name text NOT NULL,
    last_name text NOT NULL,
    email text UNIQUE NOT NULL,
    password text NOT NULL,
    is_admin boolean DEFAULT FALSE,
    is_active boolean DEFAULT FALSE,
    is_premium boolean DEFAULT FALSE,
    created_at timestamptz DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);

