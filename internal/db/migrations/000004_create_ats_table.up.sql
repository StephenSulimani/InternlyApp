CREATE TABLE IF NOT EXISTS company_ats (
    id uuid NOT NULL,
    company_name text NOT NULL,
    ats_url text NOT NULL,
    working boolean NOT NULL DEFAULT TRUE,
    PRIMARY KEY (company_name, ats_url)
);

CREATE INDEX idx_ats_url ON company_ats (ats_url);

CREATE INDEX idx_working ON company_ats (working);

CREATE INDEX idx_company_name ON company_ats (company_name);

