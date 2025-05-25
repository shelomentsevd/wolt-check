-- name: CreateSeller :one
INSERT INTO sellers (name)
VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id;
-- name: GetSellerByName :one
SELECT *
FROM sellers
WHERE name = $1
LIMIT 1;