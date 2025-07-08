CREATE TABLE IF NOT EXISTS tokens (
    hash bytea PRIMARY KEY,
    account_id uuid NOT NULL REFERENCES accounts ON DELETE CASCADE,
    expires_at timestamp(0) with time zone NOT NULL,
    scope text NOT NULL
);