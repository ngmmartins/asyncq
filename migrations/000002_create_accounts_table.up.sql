CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY,
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);
