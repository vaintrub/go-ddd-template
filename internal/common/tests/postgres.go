package tests

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// StartPostgresContainer launches a disposable PostgreSQL 14 instance and applies the initial schema.
// It returns a connection string and a termination function.
func StartPostgresContainer(ctx context.Context) (string, func(context.Context) error, error) {
	if err := ensureDockerAccessible(); err != nil {
		return "", nil, err
	}

	container, err := postgres.Run(ctx,
		"postgres:14",
		postgres.WithDatabase("go_ddd_template_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	if err != nil {
		return "", nil, fmt.Errorf("start postgres container: %w", err)
	}

	connString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return "", nil, fmt.Errorf("postgres connection string: %w", err)
	}

	schema, err := loadInitialSchema()
	if err != nil {
		_ = container.Terminate(ctx)
		return "", nil, fmt.Errorf("load migrations: %w", err)
	}

	if err := applyMigrations(ctx, connString, schema); err != nil {
		_ = container.Terminate(ctx)
		return "", nil, fmt.Errorf("apply migrations: %w", err)
	}

	return connString, func(c context.Context) error {
		return container.Terminate(c)
	}, nil
}

func loadInitialSchema() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("unable to determine caller")
	}

	path := filepath.Join(filepath.Dir(file), "..", "..", "..", "migrations", "001_initial_schema.up.sql")
	data, err := os.ReadFile(path) //nolint:gosec // G304: path is constructed from known constants
	if err != nil {
		return "", fmt.Errorf("read migration file: %w", err)
	}

	return string(data), nil
}

func applyMigrations(ctx context.Context, connString string, schema string) error {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("parse postgres config: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	statements := splitStatements(schema)
	for _, stmt := range statements {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("execute migration: %w", err)
		}
	}

	return nil
}

func splitStatements(sql string) []string {
	parts := strings.Split(sql, ";")
	var statements []string

	for _, part := range parts {
		lines := strings.Split(part, "\n")
		var cleaned []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "--") {
				continue
			}
			cleaned = append(cleaned, line)
		}

		stmt := strings.TrimSpace(strings.Join(cleaned, "\n"))
		if stmt == "" {
			continue
		}

		statements = append(statements, stmt)
	}

	return statements
}

func ensureDockerAccessible() error {
	candidates := []string{
		extractUnixPath(os.Getenv("DOCKER_HOST")),
		"/var/run/docker.sock",
		filepath.Join(os.Getenv("HOME"), ".docker", "run", "docker.sock"),
	}

	for _, socket := range candidates {
		if socket == "" {
			continue
		}
		dialer := &net.Dialer{Timeout: time.Second}
		conn, err := dialer.DialContext(context.Background(), "unix", socket)
		if err == nil {
			_ = conn.Close()
			return nil
		}
	}

	return fmt.Errorf("docker socket not available; set DOCKER_HOST to a reachable daemon")
}

func extractUnixPath(host string) string {
	if strings.HasPrefix(host, "unix://") {
		return strings.TrimPrefix(host, "unix://")
	}
	return ""
}
