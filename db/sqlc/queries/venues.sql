-- name: CreateVenue :one
INSERT INTO venues (name,
                    seller_id)
VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id;

-- name: GetVenueByName :one
SELECT *
FROM venues
WHERE name = $1
LIMIT 1;