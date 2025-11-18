package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
func NewPgxPool(ctx context.Context) (*pgxpool.Pool, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	// Parse the DATABASE_URL
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DATABASE_URL: %w", err)
	}

	// Configure connection pool limits
	config.MaxConns = getEnvInt("DB_POOL_MAX_CONNS", 25)
	config.MinConns = getEnvInt("DB_POOL_MIN_CONNS", 5)

	// Configure connection lifecycle
	config.MaxConnLifetime = time.Hour        // Recycle connections after 1 hour
	config.MaxConnIdleTime = 30 * time.Minute // Close idle connections after 30 minutes
	config.HealthCheckPeriod = time.Minute    // Check connection health every minute

	// Configure connection timeout (FR-017a: 30s default)
	timeoutSeconds := getEnvInt("DB_POOL_TIMEOUT", 30)
	config.ConnConfig.ConnectTimeout = time.Duration(timeoutSeconds) * time.Second

	// Configure SSL/TLS based on environment
	env := os.Getenv("ENV")
	if env == "production" {
		// Production: Require TLS with minimum TLS 1.2
		config.ConnConfig.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}
	// Development: No SSL (sslmode=disable in DATABASE_URL)

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// getEnvInt retrieves an integer environment variable with a default fallback.
func getEnvInt(key string, defaultValue int32) int32 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseInt(valueStr, 10, 32)
	if err != nil {
		return defaultValue
	}

	return int32(value)
}

// Close gracefully closes the connection pool, waiting for connections to be released.
func Close(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}
