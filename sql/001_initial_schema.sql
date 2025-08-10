-- +goose Up
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    order_uid TEXT UNIQUE NOT NULL,
    track_number TEXT NOT NULL,
    entry TEXT NOT NULL,
    locale VARCHAR(10) NOT NULL, -- validation guarantees that its small
    internal_signature TEXT,
    customer_id TEXT NOT NULL,
    delivery_service TEXT NOT NULL,
    shard_key TEXT NOT NULL,
    sm_id INT NOT NULL,
    date_created TIMESTAMPTZ NOT NULL,
    oof_shard TEXT NOT NULL
);

-- 1-to-1 so makes sense to make PK a FK
CREATE TABLE deliveries (
    order_id BIGINT PRIMARY KEY REFERENCES orders(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    zip TEXT NOT NULL,
    city TEXT NOT NULL,
    address TEXT NOT NULL,
    region TEXT NOT NULL,
    email TEXT NOT NULL
);

-- same as above
CREATE TABLE payments (
    order_id BIGINT PRIMARY KEY REFERENCES orders(id) ON DELETE CASCADE,
    transaction TEXT NOT NULL,
    request_id TEXT,
    currency VARCHAR(3) NOT NULL, -- ISO 4217 are 3 chars
    provider TEXT NOT NULL,
    amount INT NOT NULL,
    payment_dt INT NOT NULL,
    bank TEXT NOT NULL,
    delivery_cost INT NOT NULL,
    goods_total INT NOT NULL,
    custom_fee INT NOT NULL
);

-- not the same, so a bigserial
CREATE TABLE items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    chrt_id INT NOT NULL,
    track_number TEXT NOT NULL,
    price INT NOT NULL,
    rid TEXT NOT NULL,
    name TEXT NOT NULL,
    sale INT NOT NULL,
    size TEXT NOT NULL,
    total_price INT NOT NULL,
    nm_id INT NOT NULL,
    brand TEXT NOT NULL,
    status INT NOT NULL
);

-- fast joins and lookups
CREATE INDEX idx_items_order_id ON items(order_id);
CREATE UNIQUE INDEX idx_orders_order_uid ON orders(order_uid);

-- +goose Down
DROP TABLE items;
DROP TABLE payments;
DROP TABLE deliveries;
DROP TABLE orders;