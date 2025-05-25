-- name: CreateItem :one
INSERT INTO items (
    name,
    date,
    quantity,
    unit_price,
    line_total,
    receipt_id,
    seller_id,
    venue_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;