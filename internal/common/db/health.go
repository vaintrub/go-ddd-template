package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthStatus represents the health status of the database connection.
type HealthStatus struct {
	Status          string                 `json:"status"`
	Timestamp       time.Time              `json:"timestamp"`
	ResponseTime    time.Duration          `json:"response_time_ms"`
	ConnectionStats ConnectionStats        `json:"connection_stats"`
	Error           string                 `json:"error,omitempty"`
	Details         map[string]interface{} `json:"details,omitempty"`
}

// ConnectionStats provides connection pool statistics.
type ConnectionStats struct {
	MaxConnections          int32 `json:"max_connections"`
	TotalConnections        int32 `json:"total_connections"`
	IdleConnections         int32 `json:"idle_connections"`
	AcquiredConnections     int32 `json:"acquired_connections"`
	ConstructingConnections int32 `json:"constructing_connections"`
}

// CheckHealth performs a health check on the database connection.
// It pings the database and returns connection pool statistics.
//
// Returns:
//   - HealthStatus with "healthy" status if connection is good
//   - HealthStatus with "unhealthy" status if connection fails
func CheckHealth(ctx context.Context, pool *pgxpool.Pool) HealthStatus {
	start := time.Now()
	status := HealthStatus{
		Timestamp: start,
		Status:    "healthy",
	}

	// Perform ping to check connection
	if err := pool.Ping(ctx); err != nil {
		status.Status = "unhealthy"
		status.Error = fmt.Sprintf("database ping failed: %v", err)
		status.ResponseTime = time.Since(start)
		return status
	}

	// Get connection pool statistics
	stats := pool.Stat()
	status.ConnectionStats = ConnectionStats{
		MaxConnections:          stats.MaxConns(),
		TotalConnections:        stats.TotalConns(),
		IdleConnections:         stats.IdleConns(),
		AcquiredConnections:     stats.AcquiredConns(),
		ConstructingConnections: stats.ConstructingConns(),
	}

	status.ResponseTime = time.Since(start)

	// Add additional details
	status.Details = map[string]interface{}{
		"database":        "postgresql",
		"driver":          "pgx/v5",
		"pool_exhaustion": float64(stats.AcquiredConns()) / float64(stats.MaxConns()) * 100,
	}

	return status
}

// IsHealthy returns true if the database connection is healthy.
func IsHealthy(ctx context.Context, pool *pgxpool.Pool) bool {
	status := CheckHealth(ctx, pool)
	return status.Status == "healthy"
}

// WaitForHealthy waits for the database to become healthy with a timeout.
// This is useful for startup checks and readiness probes.
//
// Parameters:
//   - ctx: Context for cancellation
//   - pool: Database connection pool
//   - timeout: Maximum time to wait
//   - interval: Time between health checks
//
// Returns:
//   - error if timeout is reached or context is cancelled
func WaitForHealthy(ctx context.Context, pool *pgxpool.Pool, timeout, interval time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for database to become healthy")
		}

		if IsHealthy(ctx, pool) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
			// Continue to next iteration
		}
	}
}
