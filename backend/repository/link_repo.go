package repository

import (
	"context"
	"fmt"
	"time"

	// "github.com/adityasrc/snip/backend/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type DashboardResponse struct {
	TotalClicks int        `json:"totalClicks"`
	Countries   []StatItem `json:"countries"`
	OS          []StatItem `json:"os"`
	Browsers    []StatItem `json:"browsers"`
	Device      []StatItem `json:"devices"`
	Referers    []StatItem `json:"referers"`
}

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

func SaveClick(db *pgxpool.Pool, linkID int, userAgent string, referer string, ipAddress string, os string, browser string, device string, countryName string, isBot bool) error {

	// CTE - Common Template Expression
	query := `WITH new_click AS (INSERT INTO clicks (link_id, user_agent, referer, ip_address, os, browser, device, country, is_bot) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING uuid) UPDATE links SET click_count = click_count + 1 WHERE id = $1 AND EXISTS (SELECT 1 FROM new_click);`

	_, err := db.Exec(context.Background(), query, linkID, userAgent, referer, ipAddress, os, browser, device, countryName, isBot)

	return err
}

func ClickHandler(db *pgxpool.Pool, linkID int, colName string) ([]StatItem, error) {
	var result []StatItem

	// Sprintf saves output in string variable
	query := fmt.Sprintf("SELECT %s, COUNT(*) FROM clicks WHERE link_id = $1 GROUP BY %s", colName, colName)

	rows, err := db.Query(context.Background(), query, linkID) // rows is a pointer to db

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var name string
	var count int
	for rows.Next() {
		rows.Scan(&name, &count)
		val := StatItem{Name: name, Count: count} // struct object
		result = append(result, val)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

func CountHandler(db *pgxpool.Pool, linkID int) (int, error) {
	var count int
	query := `SELECT click_count FROM links WHERE id = $1`

	err := db.QueryRow(context.Background(), query, linkID).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func CheckMail(db *pgxpool.Pool, mail string) bool {
	var userId int
	query := `SELECT 1 FROM users WHERE email = $1`

	err := db.QueryRow(context.Background(), query, mail).Scan(userId)

	if err != nil {
		return false
	}

	return true
}

func Signup(db *pgxpool.Pool, name string, email string, password []byte) error {

	query := `INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3)`

	_, err := db.Exec(context.Background(), query, name, email, string(password))

	return err
}

func Signin(db *pgxpool.Pool, email string) ([]byte, error) {

	var pass string

	query := `SELECT password_hash FROM users WHERE email = $1`

	err := db.QueryRow(context.Background(), query, email).Scan(&pass)

	if err != nil {
		return nil, err
	}
	return []byte(pass), nil

}
