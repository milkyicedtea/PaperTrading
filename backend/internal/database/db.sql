CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; -- necessary for uuid_generate_v4()

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- index on the email column for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    token_hash TEXT NOT NULL UNIQUE, -- the SHA256 hash of the opaque token
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- last_used_at TIMESTAMPTZ, -- can be useful later on for tracking/cleanup
    -- revoked BOOLEAN NOT NULL DEFAULT FALSE, -- can be an alternative to deleting

    CONSTRAINT fk_user
        FOREIGN KEY(user_id)
        REFERENCES users(id)
        ON DELETE CASCADE -- if a user is deleted, their refresh tokens are also deleted
);

-- index for faster lookups by user_id, especially if a user can have multiple refresh tokens
-- (though the logic will aim for one active refresh token per user/device type).
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- index for faster lookups by token_hash
-- the UNIQUE constraint already creates an index, so this might be redundant depending on psql version
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

-- COMMENT ON TABLE refresh_tokens IS 'Stores refresh tokens for users, allowing for persistent sessions.';
-- COMMENT ON COLUMN refresh_tokens.id IS 'Unique identifier for the refresh token record.';
-- COMMENT ON COLUMN refresh_tokens.user_id IS 'Foreign key referencing the user this token belongs to.';
-- COMMENT ON COLUMN refresh_tokens.token_hash IS 'SHA256 hash of the opaque refresh token string.';
-- COMMENT ON COLUMN refresh_tokens.expires_at IS 'Timestamp when this refresh token expires and is no longer valid.';
-- COMMENT ON COLUMN refresh_tokens.created_at IS 'Timestamp when this refresh token record was created.';
