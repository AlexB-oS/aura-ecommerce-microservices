-- 0001_init.sql — schema for the product catalog service.
-- This service owns its own database (product-postgres); no other service reads
-- this table directly.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS products (
    id          UUID PRIMARY KEY,
    slug        TEXT        NOT NULL UNIQUE,
    name        TEXT        NOT NULL,
    tagline     TEXT        NOT NULL DEFAULT '',
    description TEXT        NOT NULL DEFAULT '',
    category    TEXT        NOT NULL,
    price_cents INTEGER     NOT NULL CHECK (price_cents >= 0),
    currency    TEXT        NOT NULL DEFAULT 'USD',
    stock       INTEGER     NOT NULL DEFAULT 0 CHECK (stock >= 0),
    badge       TEXT        NOT NULL DEFAULT '',
    active      BOOLEAN     NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Category is the most common filter on the storefront, so index it.
CREATE INDEX IF NOT EXISTS idx_products_category ON products (category);
-- The catalog lists newest-first.
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products (created_at DESC);
