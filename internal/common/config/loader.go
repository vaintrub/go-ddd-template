package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// Option customizes Load/MustLoad behavior (e.g., injecting readers in tests).
type Option func(*loadOptions)

type loadOptions struct {
	defaults Config
	readers  []io.Reader
	envFiles []string
}

func defaultOptions() loadOptions {
	return loadOptions{
		defaults: DefaultConfig(),
		envFiles: []string{".env"},
	}
}

// WithReader merges the provided YAML reader into the configuration surface.
func WithReader(r io.Reader) Option {
	return func(opts *loadOptions) {
		if r != nil {
			opts.readers = append(opts.readers, r)
		}
	}
}

// WithDefaults overrides the baseline default configuration used before env merges.
func WithDefaults(cfg Config) Option {
	return func(opts *loadOptions) {
		opts.defaults = cfg
	}
}

// WithEnvFile appends an additional .env-style file to load before reading process env.
func WithEnvFile(path string) Option {
	return func(opts *loadOptions) {
		if path != "" {
			opts.envFiles = append(opts.envFiles, path)
		}
	}
}

// Load builds a Config instance by applying defaults, optional config readers,
// .env files, and finally environment variables (highest priority).
func Load(ctx context.Context, optFns ...Option) (Config, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	opts := defaultOptions()
	for _, fn := range optFns {
		fn(&opts)
	}

	for _, envFile := range opts.envFiles {
		if envFile == "" {
			continue
		}
		if err := gotenv.Load(envFile); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.WarnContext(ctx, "Unable to load env file", slog.String("path", envFile), slog.Any("error", err))
		}
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.SetTypeByDefaultValue(true)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	bindLegacyEnvVars(v)
	setDefaults(v, opts.defaults)

	for _, reader := range opts.readers {
		if err := v.MergeConfig(reader); err != nil {
			return Config{}, fmt.Errorf("merge config reader: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	cfg.Env.IsLocal = strings.EqualFold(cfg.Env.Name, "development")

	if err := Validate(cfg); err != nil {
		return Config{}, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// MustLoad is a convenience that panics when configuration cannot be loaded or validated.
func MustLoad(ctx context.Context, optFns ...Option) Config {
	cfg, err := Load(ctx, optFns...)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to load configuration", slog.Any("error", err))
		panic(fmt.Errorf("config.MustLoad: %w", err))
	}
	return cfg
}

func setDefaults(v *viper.Viper, cfg Config) {
	v.SetDefault("env.name", cfg.Env.Name)

	v.SetDefault("server.port", cfg.Server.Port)
	v.SetDefault("server.read_timeout", cfg.Server.ReadTimeout)
	v.SetDefault("server.write_timeout", cfg.Server.WriteTimeout)
	v.SetDefault("server.idle_timeout", cfg.Server.IdleTimeout)
	v.SetDefault("server.mock_auth", cfg.Server.MockAuth)
	v.SetDefault("server.cors_allowed_origins", cfg.Server.CORSAllowedOrigins)
	v.SetDefault("server.public_auth_skip_paths", cfg.Server.PublicAuthSkipPaths)

	v.SetDefault("logging.level", cfg.Logging.Level)
	v.SetDefault("logging.add_source", cfg.Logging.AddSource)

	v.SetDefault("database.url", cfg.Database.URL)
	v.SetDefault("database.max_conns", cfg.Database.MaxConns)
	v.SetDefault("database.min_conns", cfg.Database.MinConns)
	v.SetDefault("database.timeout", cfg.Database.Timeout)
	v.SetDefault("database.ssl_required", cfg.Database.SSLRequired)

	v.SetDefault("grpc.trainer_addr", cfg.GRPC.TrainerAddr)
	v.SetDefault("grpc.users_addr", cfg.GRPC.UsersAddr)
	v.SetDefault("grpc.no_tls", cfg.GRPC.NoTLS)
	v.SetDefault("grpc.ca_file", cfg.GRPC.CAFile)
	v.SetDefault("grpc.dial_timeout", cfg.GRPC.DialTimeout)

	v.SetDefault("auth.mock", cfg.Auth.Mock)
	v.SetDefault("auth.casdoor.enabled", cfg.Auth.Casdoor.Enabled)
	v.SetDefault("auth.casdoor.endpoint", cfg.Auth.Casdoor.Endpoint)
	v.SetDefault("auth.casdoor.client_id", cfg.Auth.Casdoor.ClientID)
	v.SetDefault("auth.casdoor.client_secret", cfg.Auth.Casdoor.ClientSecret)
	v.SetDefault("auth.casdoor.certificate", cfg.Auth.Casdoor.Certificate)
	v.SetDefault("auth.casdoor.organization", cfg.Auth.Casdoor.Organization)
	v.SetDefault("auth.casdoor.application", cfg.Auth.Casdoor.Application)

	v.SetDefault("contexts.trainings.feature_flags", cfg.Contexts.Trainings.FeatureFlags)
	v.SetDefault("contexts.trainings.metrics_namespace", cfg.Contexts.Trainings.MetricsNamespace)
	v.SetDefault("contexts.users.feature_flags", cfg.Contexts.Users.FeatureFlags)
	v.SetDefault("contexts.users.metrics_namespace", cfg.Contexts.Users.MetricsNamespace)
	v.SetDefault("contexts.trainer.feature_flags", cfg.Contexts.Trainer.FeatureFlags)
	v.SetDefault("contexts.trainer.metrics_namespace", cfg.Contexts.Trainer.MetricsNamespace)
}

func bindLegacyEnvVars(v *viper.Viper) {
	_ = v.BindEnv("env.name", "APP_ENV", "ENV")

	_ = v.BindEnv("server.port", "SERVER_PORT", "PORT")
	_ = v.BindEnv("server.mock_auth", "SERVER_MOCK_AUTH", "MOCK_AUTH")
	_ = v.BindEnv("server.cors_allowed_origins", "SERVER_CORS_ALLOWED_ORIGINS", "CORS_ALLOWED_ORIGINS")
	_ = v.BindEnv("server.public_auth_skip_paths", "SERVER_PUBLIC_AUTH_SKIP_PATHS")
	_ = v.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	_ = v.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	_ = v.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT")

	_ = v.BindEnv("logging.level", "LOG_LEVEL")
	_ = v.BindEnv("logging.add_source", "LOG_ADD_SOURCE")

	_ = v.BindEnv("database.url", "DATABASE_URL")
	_ = v.BindEnv("database.max_conns", "DATABASE_MAX_CONNS", "DB_POOL_MAX_CONNS")
	_ = v.BindEnv("database.min_conns", "DATABASE_MIN_CONNS", "DB_POOL_MIN_CONNS")
	_ = v.BindEnv("database.timeout", "DATABASE_TIMEOUT", "DB_POOL_TIMEOUT")
	_ = v.BindEnv("database.ssl_required", "DATABASE_SSL_REQUIRED")

	_ = v.BindEnv("grpc.trainer_addr", "TRAINER_GRPC_ADDR")
	_ = v.BindEnv("grpc.users_addr", "USERS_GRPC_ADDR")
	_ = v.BindEnv("grpc.no_tls", "GRPC_NO_TLS")
	_ = v.BindEnv("grpc.ca_file", "GRPC_CA_FILE")
	_ = v.BindEnv("grpc.dial_timeout", "GRPC_DIAL_TIMEOUT")

	_ = v.BindEnv("auth.mock", "AUTH_MOCK", "MOCK_AUTH")
	_ = v.BindEnv("auth.casdoor.enabled", "CASDOOR_ENABLED")
	_ = v.BindEnv("auth.casdoor.endpoint", "CASDOOR_ENDPOINT")
	_ = v.BindEnv("auth.casdoor.client_id", "CASDOOR_CLIENT_ID")
	_ = v.BindEnv("auth.casdoor.client_secret", "CASDOOR_CLIENT_SECRET")
	_ = v.BindEnv("auth.casdoor.certificate", "CASDOOR_CERTIFICATE")
	_ = v.BindEnv("auth.casdoor.organization", "CASDOOR_ORG_SLUG", "CASDOOR_ORGANIZATION")
	_ = v.BindEnv("auth.casdoor.application", "CASDOOR_APPLICATION_NAME")
}
