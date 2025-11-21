package config

import (
	"fmt"
	"log/slog"
	"strings"
)

var allowedEnvs = map[string]struct{}{
	"development": {},
	"staging":     {},
	"production":  {},
}

// ValidationError describes a single configuration validation failure.
type ValidationError struct {
	Field   string
	Message string
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("%s %s", v.Field, v.Message)
}

// ValidationErrors aggregates multiple configuration validation failures.
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "no validation errors"
	}

	var b strings.Builder
	for i, err := range v {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(err.Error())
	}
	return b.String()
}

// Validate ensures the provided Config meets baseline requirements.
func Validate(cfg Config) error {
	var errs ValidationErrors

	env := strings.ToLower(strings.TrimSpace(cfg.Env.Name))
	if _, ok := allowedEnvs[env]; !ok {
		errs = append(errs, ValidationError{
			Field:   "env.name",
			Message: "must be one of development, staging, production",
		})
	}

	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		errs = append(errs, ValidationError{
			Field:   "server.port",
			Message: "must be between 1 and 65535",
		})
	}

	if cfg.Database.URL == "" {
		errs = append(errs, ValidationError{
			Field:   "database.url",
			Message: "is required",
		})
	}

	if cfg.GRPC.TrainerAddr == "" {
		errs = append(errs, ValidationError{
			Field:   "grpc.trainer_addr",
			Message: "is required",
		})
	}

	if cfg.GRPC.UsersAddr == "" {
		errs = append(errs, ValidationError{
			Field:   "grpc.users_addr",
			Message: "is required",
		})
	}

	errs = append(errs, validateCasdoor(cfg)...)

	if level := strings.TrimSpace(cfg.Logging.Level); level != "" {
		var parsed slog.Level
		if err := parsed.UnmarshalText([]byte(level)); err != nil {
			errs = append(errs, ValidationError{
				Field:   "logging.level",
				Message: "must be a valid slog level (debug|info|warn|error)",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func validateCasdoor(cfg Config) []ValidationError {
	if !cfg.Auth.Casdoor.Enabled {
		return nil
	}

	var errs []ValidationError

	if cfg.Auth.Casdoor.Endpoint == "" {
		errs = append(errs, ValidationError{
			Field:   "auth.casdoor.endpoint",
			Message: "is required when CASDOOR is enabled",
		})
	}
	if cfg.Auth.Casdoor.ClientID == "" {
		errs = append(errs, ValidationError{
			Field:   "auth.casdoor.client_id",
			Message: "is required when CASDOOR is enabled",
		})
	}
	if cfg.Auth.Casdoor.ClientSecret == "" {
		errs = append(errs, ValidationError{
			Field:   "auth.casdoor.client_secret",
			Message: "is required when CASDOOR is enabled",
		})
	}
	if cfg.Auth.Casdoor.Organization == "" {
		errs = append(errs, ValidationError{
			Field:   "auth.casdoor.organization",
			Message: "is required when CASDOOR is enabled",
		})
	}
	if cfg.Auth.Casdoor.Application == "" {
		errs = append(errs, ValidationError{
			Field:   "auth.casdoor.application",
			Message: "is required when CASDOOR is enabled",
		})
	}

	return errs
}
