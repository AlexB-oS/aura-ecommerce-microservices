-- 0001_init.sql — schema for the payment service (its own database).

CREATE TABLE IF NOT EXISTS payments (
    id           UUID PRIMARY KEY,
    -- One payment per order (idempotency for redelivered order.placed events).
    order_id     UUID        NOT NULL UNIQUE,
    user_id      UUID        NOT NULL,
    amount_cents INTEGER     NOT NULL CHECK (amount_cents >= 0),
    currency     TEXT        NOT NULL DEFAULT 'USD',
    status       TEXT        NOT NULL CHECK (status IN ('succeeded','failed')),
    provider     TEXT        NOT NULL DEFAULT 'mock',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_payments_user ON payments (user_id, created_at DESC);
