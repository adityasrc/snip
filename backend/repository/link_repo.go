package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetNextID(db *pgxpool.Pool) (int, error) { // sql.BD is database connection pool

	query := "SELECT nextVal('links_id_seq')"

	var id int

	err := db.QueryRow(context.Background(), query).Scan(&id) // chaining the result of query to id

	return id, err
}

func SaveLink(db *pgxpool.Pool, id int, url string, slug string) error {

	query := "INSERT INTO links (id, link, slug) VALUES ($1, $2, $3)"

	_, err := db.Exec(context.Background(), query, id, url, slug) // _ is a blank identifier to ignore result

	return err
}

func GetLink(db *pgxpool.Pool, slug string) (string, error) {

	query := "SELECT link FROM links WHERE slug = $1"

	var longURl string
	err := db.QueryRow(context.Background(), query, slug).Scan(&longURl)

	return longURl, err
}
