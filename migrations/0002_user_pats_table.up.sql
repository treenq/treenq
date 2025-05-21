CREATE TABLE IF NOT EXISTS user_pats (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    encrypted_pat TEXT NOT NULL, -- Using TEXT, consider BYTEA if encryption output is raw bytes
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Optional: Add an index on user_id for faster lookups if not automatically created by UNIQUE constraint
-- CREATE INDEX IF NOT EXISTS idx_user_pats_user_id ON user_pats(user_id);
