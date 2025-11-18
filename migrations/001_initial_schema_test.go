package migrations_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupTestPostgres creates a PostgreSQL test container and returns the connection string.
func setupTestPostgres(t *testing.T) (testcontainers.Container, string, func()) {
	t.Helper()
	ctx := context.Background()

	// Create PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:14-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	require.NoError(t, err, "failed to start postgres container")

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")

	// Cleanup function
	cleanup := func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return postgresContainer, connStr, cleanup
}

// runMigrations runs migrations up to the specified version.
func runMigrations(t *testing.T, databaseURL string, version int) error {
	t.Helper()

	// Get the migrations directory path (relative to test file)
	migrationsPath := "file://."

	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if version == -1 {
		// Run all migrations
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
	} else {
		// Migrate to specific version
		if err := m.Migrate(uint(version)); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to migrate to version %d: %w", version, err)
		}
	}

	return nil
}

// rollbackMigrations rolls back all migrations.
func rollbackMigrations(t *testing.T, databaseURL string) error {
	t.Helper()

	migrationsPath := "file://."

	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	return nil
}

// T021: Test that up migration creates all tables
func TestMigration001_Up_CreatesAllTables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	container, connStr, cleanup := setupTestPostgres(t)
	defer cleanup()
	defer container.Terminate(context.Background())

	// Run migrations
	err := runMigrations(t, connStr, -1)
	require.NoError(t, err, "failed to run up migrations")

	// Verify all tables exist
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err, "failed to connect to database")
	defer db.Close()

	expectedTables := []string{
		"trainings_trainings",
		"trainer_hours",
		"users_users",
		"schema_migrations", // golang-migrate tracking table
	}

	for _, table := range expectedTables {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = $1
		)`
		err := db.QueryRow(query, table).Scan(&exists)
		require.NoError(t, err, "failed to check if table %s exists", table)
		assert.True(t, exists, "table %s should exist after migration", table)
	}
}

// T022: Test that down migration removes all tables cleanly
func TestMigration001_Down_RemovesAllTables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	container, connStr, cleanup := setupTestPostgres(t)
	defer cleanup()
	defer container.Terminate(context.Background())

	// Run migrations up
	err := runMigrations(t, connStr, -1)
	require.NoError(t, err, "failed to run up migrations")

	// Run migrations down
	err = rollbackMigrations(t, connStr)
	require.NoError(t, err, "failed to rollback migrations")

	// Verify all tables are removed (except schema_migrations)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err, "failed to connect to database")
	defer db.Close()

	tablesToRemove := []string{
		"trainings_trainings",
		"trainer_hours",
		"users_users",
	}

	for _, table := range tablesToRemove {
		var exists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = $1
		)`
		err := db.QueryRow(query, table).Scan(&exists)
		require.NoError(t, err, "failed to check if table %s exists", table)
		assert.False(t, exists, "table %s should not exist after rollback", table)
	}
}

// T023: Test idempotency (running up twice succeeds)
func TestMigration001_Up_Idempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	container, connStr, cleanup := setupTestPostgres(t)
	defer cleanup()
	defer container.Terminate(context.Background())

	// Run migrations first time
	err := runMigrations(t, connStr, -1)
	require.NoError(t, err, "failed to run up migrations first time")

	// Run migrations second time (should be no-op)
	err = runMigrations(t, connStr, -1)
	assert.NoError(t, err, "running migrations twice should not error")

	// Verify tables still exist
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err, "failed to connect to database")
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('trainings_trainings', 'trainer_hours', 'users_users')").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 3, count, "all three tables should still exist after second migration")
}

// T024: Test that schema_migrations tracking table exists
func TestMigration001_SchemaMigrationsTableExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	container, connStr, cleanup := setupTestPostgres(t)
	defer cleanup()
	defer container.Terminate(context.Background())

	// Run migrations
	err := runMigrations(t, connStr, -1)
	require.NoError(t, err, "failed to run up migrations")

	// Verify schema_migrations table exists
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err, "failed to connect to database")
	defer db.Close()

	var exists bool
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'schema_migrations'
	)`
	err = db.QueryRow(query).Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists, "schema_migrations table should exist")

	// Verify migration version is tracked
	var version uint
	var dirty bool
	err = db.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty)
	require.NoError(t, err)
	assert.Equal(t, uint(1), version, "migration version should be 1")
	assert.False(t, dirty, "migration should not be in dirty state")
}

// TestMain sets up test environment
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}
