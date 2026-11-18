CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    items JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS orders_customer_id_idx ON orders (customer_id);