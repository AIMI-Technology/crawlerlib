package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

func init() {
	var err error
	uname := os.Getenv("PG_UNAME")
	pword := os.Getenv("PG_PWORD")
	host := os.Getenv("PG_HOST")
	dbname := os.Getenv("PG_DBNAME")
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s", uname, pword, host, dbname)
	ctx := context.Background()

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("unable to parse database connection string: %v\n", err)
	}

	dbPool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("unable to create database connection pool: %v\n", err)
	}

	err = dbPool.Ping(ctx)
	if err != nil {
		log.Fatalf("unable to reach database: %v\n", err)
	}
}

func Client() *pgxpool.Pool {
	return dbPool
}

func PutItem(ctx context.Context, item ScrapedData) error {
	query := `INSERT INTO scraped_data (text, scraped_at, published_at, url, source_country, content_country) 
              VALUES ($1, $2, $3, $4, $5, $6)
              ON CONFLICT (url) 
              DO NOTHING`
	_, err := dbPool.Exec(ctx, query, item.Text, item.ScrapedAt, item.PublishedAt, item.Url, item.SourceCountry, item.ContentCountry)
	return err
}
