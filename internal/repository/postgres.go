package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sukryu/IV-auth-services/configs"
)

// Postgres holds the database connection pool
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres creates a new PostgreSQL connection pool
func NewPostgres(ctx context.Context) (*Postgres, error) {
	cfg := configs.GlobalConfig
	pool, err := pgxpool.New(ctx, cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	return &Postgres{pool: pool}, nil
}

// Close closes the database connection pool
func (p *Postgres) Close() {
	p.pool.Close()
}

// Ping tests the database connection
func (p *Postgres) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}
