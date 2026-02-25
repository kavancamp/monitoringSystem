package database

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	url := os.Getenv("DATABASE_URL")
	c, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return pgxpool.New(c, url)
}
