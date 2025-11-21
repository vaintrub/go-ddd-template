package config

import (
	"strings"
	"time"
)

// Config aggregates every runtime configuration knob that bounded contexts depend on.
type Config struct {
	Env      EnvConfig        `mapstructure:"env"`
	Server   ServerConfig     `mapstructure:"server"`
	Logging  LoggingConfig    `mapstructure:"logging"`
	Database DatabaseConfig   `mapstructure:"database"`
	GRPC     GRPCConfig       `mapstructure:"grpc"`
	Auth     AuthConfig       `mapstructure:"auth"`
	Contexts ContextOverrides `mapstructure:"contexts"`
}

// EnvConfig holds metadata about the current environment (development, staging, production).
type EnvConfig struct {
	Name    string `mapstructure:"name"`
	IsLocal bool   `mapstructure:"is_local"`
}

// ServerConfig controls HTTP/gRPC server settings.
type ServerConfig struct {
	Port                int           `mapstructure:"port"`
	ReadTimeout         time.Duration `mapstructure:"read_timeout"`
	WriteTimeout        time.Duration `mapstructure:"write_timeout"`
	IdleTimeout         time.Duration `mapstructure:"idle_timeout"`
	MockAuth            bool          `mapstructure:"mock_auth"`
	CORSAllowedOrigins  string        `mapstructure:"cors_allowed_origins"`
	PublicAuthSkipPaths string        `mapstructure:"public_auth_skip_paths"`
}

// LoggingConfig configures structured logging.
type LoggingConfig struct {
	Level     string `mapstructure:"level"`
	AddSource bool   `mapstructure:"add_source"`
}

// DatabaseConfig controls PostgreSQL connectivity.
type DatabaseConfig struct {
	URL         string        `mapstructure:"url"`
	MaxConns    int32         `mapstructure:"max_conns"`
	MinConns    int32         `mapstructure:"min_conns"`
	Timeout     time.Duration `mapstructure:"timeout"`
	SSLRequired bool          `mapstructure:"ssl_required"`
}

// GRPCConfig configures outbound trainer/users gRPC clients.
type GRPCConfig struct {
	TrainerAddr string        `mapstructure:"trainer_addr"`
	UsersAddr   string        `mapstructure:"users_addr"`
	NoTLS       bool          `mapstructure:"no_tls"`
	CAFile      string        `mapstructure:"ca_file"`
	DialTimeout time.Duration `mapstructure:"dial_timeout"`
}

// AuthConfig toggles auth middlewares/helpers.
type AuthConfig struct {
	Mock    bool          `mapstructure:"mock"`
	Casdoor CasdoorConfig `mapstructure:"casdoor"`
}

// CasdoorConfig contains OAuth client values for talking to Casdoor.
type CasdoorConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	Endpoint     string `mapstructure:"endpoint"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	Certificate  string `mapstructure:"certificate"`
	Organization string `mapstructure:"organization"`
	Application  string `mapstructure:"application"`
}

// ContextOverrides bundles per-context feature flags and observability hints.
type ContextOverrides struct {
	Trainings ContextConfig `mapstructure:"trainings"`
	Users     ContextConfig `mapstructure:"users"`
	Trainer   ContextConfig `mapstructure:"trainer"`
}

// ContextConfig defines knobs available to each bounded context.
type ContextConfig struct {
	FeatureFlags     map[string]bool `mapstructure:"feature_flags"`
	MetricsNamespace string          `mapstructure:"metrics_namespace"`
}

// DefaultConfig returns baseline values that can be overridden via config files or env vars.
func DefaultConfig() Config {
	return Config{
		Env: EnvConfig{
			Name: "development",
		},
		Server: ServerConfig{
			Port:               3000,
			ReadTimeout:        15 * time.Second,
			WriteTimeout:       15 * time.Second,
			IdleTimeout:        60 * time.Second,
			MockAuth:           true,
			CORSAllowedOrigins: "http://localhost:8080",
		},
		Logging: LoggingConfig{
			Level:     "INFO",
			AddSource: true,
		},
		Database: DatabaseConfig{
			MaxConns:    25,
			MinConns:    5,
			Timeout:     30 * time.Second,
			SSLRequired: false,
		},
		GRPC: GRPCConfig{
			NoTLS:       true,
			DialTimeout: 5 * time.Second,
		},
		Auth: AuthConfig{
			Mock: true,
			Casdoor: CasdoorConfig{
				Organization: "local-org",
				Application:  "local-app",
			},
		},
		Contexts: ContextOverrides{
			Trainings: ContextConfig{
				FeatureFlags: map[string]bool{},
			},
			Users: ContextConfig{
				FeatureFlags: map[string]bool{},
			},
			Trainer: ContextConfig{
				FeatureFlags: map[string]bool{},
			},
		},
	}
}

// AuthSkipPaths returns a normalized list of public paths that bypass HTTP auth middleware.
func (cfg ServerConfig) AuthSkipPaths() []string {
	return splitCSV(cfg.PublicAuthSkipPaths)
}

func splitCSV(raw string) []string {
	if raw == "" {
		return nil
	}

	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';'
	})

	results := make([]string, 0, len(fields))
	for _, field := range fields {
		if trimmed := strings.TrimSpace(field); trimmed != "" {
			results = append(results, trimmed)
		}
	}

	return results
}
