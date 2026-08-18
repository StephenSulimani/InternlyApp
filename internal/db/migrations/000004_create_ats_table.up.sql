CREATE TABLE IF NOT EXISTS company_ats (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    company_name text NOT NULL,
    ats_name text NOT NULL,
    ats_url text NOT NULL UNIQUE,
    working boolean NOT NULL DEFAULT TRUE,
    first_seen timestamptz NOT NULL DEFAULT NOW(),
    last_seen timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_company_ats_working ON company_ats (working);

CREATE INDEX idx_company_ats_ats_name ON company_ats (ats_name);

CREATE INDEX idx_company_ats_company_name ON company_ats (company_name);
