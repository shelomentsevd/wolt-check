// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: sellers.sql

package postgres

import (
	"context"
)

const createSeller = `-- name: CreateSeller :one
INSERT INTO sellers (name)
VALUES ($1)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id
`

func (q *Queries) CreateSeller(ctx context.Context, name string) (int, error) {
	row := q.db.QueryRow(ctx, createSeller, name)
	var id int
	err := row.Scan(&id)
	return id, err
}

const getSellerByName = `-- name: GetSellerByName :one
SELECT id, name
FROM sellers
WHERE name = $1
LIMIT 1
`

func (q *Queries) GetSellerByName(ctx context.Context, name string) (Seller, error) {
	row := q.db.QueryRow(ctx, getSellerByName, name)
	var i Seller
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}
