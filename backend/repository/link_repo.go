package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

func GetNextID(db *pgxpool.Pool) (int, error) { // sql.BD is database connection pool

	query := "SELECT nextVal('links_id_seq')"

	var id int

	err := db.QueryRow(context.Background(), query).Scan(&id) // chaining the result of query to id

	return id, err
}

func SaveLink(db *pgxpool.Pool, id int, url string, slug string, exp_date *time.Time) error {

	query := "INSERT INTO links (id, link, slug, expiry_date) VALUES ($1, $2, $3, $4)"

	_, err := db.Exec(context.Background(), query, id, url, slug, exp_date) // _ is a blank identifier to ignore result

	return err
}

func GetLink(db *pgxpool.Pool, slug string) (int, string, error) {
	var id int
	var longURL string

	query := "SELECT id, link FROM links WHERE slug = $1 AND (expiry_date IS NULL OR expiry_date > NOW())"

	err := db.QueryRow(context.Background(), query, slug).Scan(&id, &longURL)
	if err != nil {
		return 0, "", err
	}
	return id, longURL, nil
}

func SaveClick(db *pgxpool.Pool, linkID int, userAgent string, referer string, ipAddress string) error {

	// CTE - Common Template Expression
	query := `WITH new_click AS(INSERT INTO clicks (link_id, user_agent, referer, ip_address) VALUES ($1, $2, $3, $4) RETURNING uuid) 
				UPDATE links SET click_count = click_count + 1 WHERE id = $1;`

	_, err := db.Exec(context.Background(), query, linkID, userAgent, referer, ipAddress)

	return err
}
