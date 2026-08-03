package database

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oschwald/geoip2-golang"
)

func Connect(dbURl string) (*pgxpool.Pool, *geoip2.Reader, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgxpool.New(ctx, dbURl)
	if err != nil {
		return nil, nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, nil, err
	}

	filePath := os.Getenv("GEO_DB_PATH")
	reader, err := geoip2.Open(filePath)
	if err != nil {
		return nil, nil, err
	}

	return conn, reader, nil

}
