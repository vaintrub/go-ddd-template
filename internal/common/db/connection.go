package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vaintrub/go-ddd-template/internal/common/config"
)

// NewPgxPool creates a new PostgreSQL connection pool with environment-based configuration.
// It configures connection limits, timeouts, and SSL/TLS based on the environment.
//
// Environment variables:
//   - DATABASE_URL (required): PostgreSQL connection string
//   - ENV: Environment name (production|development), affects SSL configuration
//   - DB_POOL_MAX_CONNS: Maximum number of connections in the pool (default: 25)
//   - DB_POOL_MIN_CONNS: Minimum number of connections in the pool (default: 5)
//   - DB_POOL_TIMEOUT: Connection acquisition timeout in seconds (default: 30)
//
// Returns:
//   - *pgxpool.Pool: Configured connection pool
//   - error: Configuration or connection error
func NewPgxPool(ctx context.Context, dbCfg config.DatabaseConfig, envCfg config.EnvConfig) (*pgxpool.Pool, error) {
	if dbCfg.URL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	poolConfig, err := pgxpool.ParseConfig(dbCfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	if dbCfg.MaxConns > 0 {
		poolConfig.MaxConns = dbCfg.MaxConns
	}
	if dbCfg.MinConns > 0 {
		poolConfig.MinConns = dbCfg.MinConns
	}

	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	timeout := dbCfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	poolConfig.ConnConfig.ConnectTimeout = timeout

	if envCfg.Name == "production" || dbCfg.SSLRequired {
		poolConfig.ConnConfig.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// Close gracefully closes the connection pool, waiting for connections to be released.
func Close(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}
