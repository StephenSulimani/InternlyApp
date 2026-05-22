-- Enable the pgvector extension to work with embeddings
CREATE EXTENSION IF NOT EXISTS vector;

-- Example of how your jobs table might look with a vector column
-- 384 is the dimension size for common lightweight models like all-MiniLM-L6-v2
-- CREATE TABLE IF NOT EXISTS jobs (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     title TEXT,
--     company TEXT,
--     description TEXT,
--     link TEXT UNIQUE,
--     embedding vector(384),
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
-- );
