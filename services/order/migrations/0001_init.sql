-- 0001_init.sql — schema for the order service (its own database).

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS orders (
    id          UUID PRIMARY KEY,
    user_id     UUID        NOT NULL,
    status      TEXT        NOT NULL DEFAULT 'pending_payment'
                CHECK (status IN ('pending_payment','paid','payment_failed','cancelled')),
    total_cents INTEGER     NOT NULL CHECK (total_cents >= 0),
    currency    TEXT        NOT NULL DEFAULT 'USD',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Order history is always queried by user, newest first.
CREATE INDEX IF NOT EXISTS idx_orders_user ON orders (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS order_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID        NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id  UUID        NOT NULL,
    slug        TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    price_cents INTEGER     NOT NULL CHECK (price_cents >= 0),
    qty         INTEGER     NOT NULL CHECK (qty > 0)
);

CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items (order_id);
