BEGIN;

-- CREATE database wbl0db;

-- CREATE USER wbuser password 'wb';

-- GRANT all privilages on database wbl0db to wbuser;

-- \c wbl0db wbuser;

CREATE schema ordertest;

CREATE TABLE ordertest.orders (
    order_uid text PRIMARY KEY,
    track_number text NOT NULL,
    entry text,
    delivery json NOT NULL,
    payment json NOT NULL,
    items json NOT NULL,
    locale text,
    internal_signature text,
    customer_id text NOT NULL,
    delivery_service text,
    shardkey text,
    sm_id int,
    date_created text,
    oof_shard text
);

CREATE TABLE ordertest.deliveries(
    order_uid text PRIMARY KEY,
    name text NOT NULL,
    phone text NOT NULL,
    zip text NOT NULL,
    city text NOT NULL,
    address text NOT NULL,
    region text,
    email text,
    FOREIGN KEY (order_uid) REFERENCES ordertest.orders(order_uid)
);

CREATE TABLE ordertest.payments(
    transaction text PRIMARY KEY,
    request_id text, 
    currency varchar(3),
    provider text,
    amount int NOT NULL,
    payment_dt int,
    bank text,
    delivery_cost int NOT NULL,
    goods_total int,
    custom_fee int,
    FOREIGN KEY (transaction) REFERENCES ordertest.orders(order_uid)
);

CREATE TABLE ordertest.carts(
    order_uid text PRIMARY KEY,
    items json NOT NULL,
    FOREIGN KEY (order_uid) REFERENCES ordertest.orders(order_uid)
);

COMMIT;