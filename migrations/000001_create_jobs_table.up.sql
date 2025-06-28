CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY,
    task TEXT NOT NULL,
    payload JSONB,
    run_at timestamp(0) with time zone,
    status TEXT NOT NULL,
    created_at timestamp(0) with time zone NOT NULL
);
