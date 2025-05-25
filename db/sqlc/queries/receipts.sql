-- name: CreateReceipt :one
INSERT INTO receipts (id,
                      date,
                      seller_id,
                      venue_id,
                      total,
                      tips)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE SET date      = EXCLUDED.date,
                               seller_id = EXCLUDED.seller_id,
                               venue_id  = EXCLUDED.venue_id,
                               total     = EXCLUDED.total,
                               tips      = EXCLUDED.tips
RETURNING *;