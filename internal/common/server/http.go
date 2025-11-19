package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/vaintrub/go-ddd-template/internal/common/auth"
	"github.com/vaintrub/go-ddd-template/internal/common/config"
	"github.com/vaintrub/go-ddd-template/internal/common/logs"
)

func RunHTTPServer(cfg config.ServerConfig, logger *slog.Logger, createHandler func(router chi.Router) http.Handler) {
	addr := fmt.Sprintf(":%d", cfg.Port)
	RunHTTPServerOnAddr(cfg, addr, logger, createHandler)
}

func RunHTTPServerOnAddr(cfg config.ServerConfig, addr string, logger *slog.Logger, createHandler func(router chi.Router) http.Handler) {
	apiRouter := chi.NewRouter()
	setMiddlewares(apiRouter, logger, cfg)

	rootRouter := chi.NewRouter()
	rootRouter.Mount("/api", createHandler(apiRouter))

	ctx := context.Background()
	logger.InfoContext(ctx, "Starting HTTP server", slog.String("addr", addr))

	server := &http.Server{
		Addr:         addr,
		Handler:      rootRouter,
		ReadTimeout:  nonZeroDuration(cfg.ReadTimeout, 15*time.Second),
		WriteTimeout: nonZeroDuration(cfg.WriteTimeout, 15*time.Second),
		IdleTimeout:  nonZeroDuration(cfg.IdleTimeout, 60*time.Second),
	}

	err := server.ListenAndServe()
	if err != nil {
		logger.ErrorContext(ctx, "Unable to start HTTP server", slog.Any("error", err))
		panic(err)
	}
}

func nonZeroDuration(value, fallback time.Duration) time.Duration {
	if value == 0 {
		return fallback
	}
	return value
}

func setMiddlewares(router *chi.Mux, logger *slog.Logger, cfg config.ServerConfig) {
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(logs.HTTPLogger(logger))
	router.Use(middleware.Recoverer)

	addCorsMiddleware(router, cfg)
	addAuthMiddleware(router, cfg)

	router.Use(
		middleware.SetHeader("X-Content-Type-Options", "nosniff"),
		middleware.SetHeader("X-Frame-Options", "deny"),
	)
	router.Use(middleware.NoCache)
}

func addAuthMiddleware(router *chi.Mux, cfg config.ServerConfig) {
	if cfg.MockAuth {
		router.Use(auth.HttpMockMiddleware)
	}
}

func addCorsMiddleware(router *chi.Mux, cfg config.ServerConfig) {
	if cfg.CORSAllowedOrigins == "" {
		return
	}

	var origins []string
	for _, origin := range strings.Split(cfg.CORSAllowedOrigins, ";") {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	if len(origins) == 0 {
		return
	}

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	router.Use(corsMiddleware.Handler)
}
