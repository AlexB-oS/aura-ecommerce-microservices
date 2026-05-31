-- 0001_init.sql — initial schema for the user service.
-- Each service owns its own database; this file is the entire contract for the
-- `users` table. Migrations are applied in filename order, exactly once.

-- pgcrypto gives us gen_random_uuid() if we ever want DB-generated ids; the
-- service currently generates UUIDs in Go, but enabling it keeps the option open.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY,
    -- Emails are stored lower-cased so uniqueness is case-insensitive.
    email         TEXT        NOT NULL UNIQUE,
    name          TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL DEFAULT 'customer'
                  CHECK (role IN ('customer', 'admin')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Email is the login identifier and is looked up on every sign-in, so the
-- UNIQUE constraint's implicit index is exactly the lookup index we need.

-- Force-lowercase emails at the database level as a safety net, independent of
-- whatever the application does.
CREATE OR REPLACE FUNCTION users_lowercase_email() RETURNS trigger AS $$
BEGIN
    NEW.email := lower(NEW.email);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_users_lowercase_email ON users;
CREATE TRIGGER trg_users_lowercase_email
    BEFORE INSERT OR UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION users_lowercase_email();
