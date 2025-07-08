CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY,
    hash bytea NOT NULL,
    account_id uuid NOT NULL REFERENCES accounts ON DELETE CASCADE,
    name text NOT NULL,
    expires_at timestamp(0) with time zone,
    created_at timestamp(0) with time zone NOT NULL
);