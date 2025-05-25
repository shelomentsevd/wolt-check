-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sellers
(
    id   SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS venues
(
    id        SERIAL PRIMARY KEY,
    name      VARCHAR(255) NOT NULL UNIQUE,
    seller_id INTEGER REFERENCES sellers (id),
    CONSTRAINT fk_venue_seller FOREIGN KEY (seller_id) REFERENCES sellers (id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS receipts
(
    id        VARCHAR(255) PRIMARY KEY,
    date      TIMESTAMP      NULL,
    seller_id INTEGER        NOT NULL,
    venue_id  INTEGER        NOT NULL,
    total     DECIMAL(10, 2) NOT NULL,
    tips      DECIMAL(10, 2),
    CONSTRAINT fk_receipt_seller FOREIGN KEY (seller_id) REFERENCES sellers (id),
    CONSTRAINT fk_receipt_venue FOREIGN KEY (venue_id) REFERENCES venues (id)
);

CREATE TABLE IF NOT EXISTS items
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    date       TIMESTAMP      NULL,
    quantity   TEXT           NOT NULL,
    unit_price VARCHAR(255)   NOT NULL,
    line_total DECIMAL(10, 2) NOT NULL,
    receipt_id VARCHAR(255)   NOT NULL,
    seller_id  INTEGER        NOT NULL,
    venue_id   INTEGER        NOT NULL,
    CONSTRAINT fk_item_receipt FOREIGN KEY (receipt_id) REFERENCES receipts (id),
    CONSTRAINT fk_item_seller FOREIGN KEY (seller_id) REFERENCES sellers (id),
    CONSTRAINT fk_item_venue FOREIGN KEY (venue_id) REFERENCES venues (id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS receipts;
DROP TABLE IF EXISTS venues;
DROP TABLE IF EXISTS sellers;
-- +goose StatementEnd