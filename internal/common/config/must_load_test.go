package config_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vaintrub/go-ddd-template/internal/common/config"
)

func TestMustLoadMissingRequired(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("ENV", "development")
	t.Setenv("ENV_NAME", "development")
	t.Setenv("DATABASE_URL", "")

	ctx := context.Background()
	reader := strings.NewReader(`
database:
  url: ""
`)

	defaults := config.DefaultConfig()
	defaults.Env.Name = "development"
	defaults.GRPC.TrainerAddr = "trainer:3000"
	defaults.GRPC.UsersAddr = "users:3000"
	defaults.Database.URL = "postgres://example.com/db"

	require.PanicsWithError(t, "config.MustLoad: config validation failed: database.url is required", func() {
		config.MustLoad(ctx, config.WithDefaults(defaults), config.WithReader(reader))
	})
}
