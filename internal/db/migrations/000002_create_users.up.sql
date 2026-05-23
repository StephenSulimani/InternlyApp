CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    first_name text NOT NULL,
    last_name text NOT NULL,
    email text UNIQUE NOT NULL,
    password text NOT NULL,
    discord_id text UNIQUE,
    is_admin boolean DEFAULT FALSE NOT NULL,
    is_active boolean DEFAULT FALSE NOT NULL,
    is_premium boolean DEFAULT FALSE NOT NULL,
    created_at timestamptz DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);

CREATE INDEX idx_users_discord_id ON users (discord_id);

