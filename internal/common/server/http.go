package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/vaintrub/go-ddd-template/internal/common/auth"
	"github.com/vaintrub/go-ddd-template/internal/common/logs"
)

func RunHTTPServer(createHandler func(router chi.Router) http.Handler) {
	RunHTTPServerOnAddr(":"+os.Getenv("PORT"), createHandler)
}

func RunHTTPServerOnAddr(addr string, createHandler func(router chi.Router) http.Handler) {
	// Initialize slog logger before starting server
	logger := logs.Init()

	apiRouter := chi.NewRouter()
	setMiddlewares(apiRouter, logger)

	rootRouter := chi.NewRouter()
	// we are mounting all APIs under /api path
	rootRouter.Mount("/api", createHandler(apiRouter))

	ctx := context.Background()
	slog.InfoContext(ctx, "Starting HTTP server", slog.String("addr", addr))

	server := &http.Server{
		Addr:         addr,
		Handler:      rootRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		slog.ErrorContext(ctx, "Unable to start HTTP server", slog.Any("error", err))
		panic(err)
	}
}

func setMiddlewares(router *chi.Mux, logger *slog.Logger) {
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(logs.HTTPLogger(logger))
	router.Use(middleware.Recoverer)

	addCorsMiddleware(router)
	addAuthMiddleware(router)

	router.Use(
		middleware.SetHeader("X-Content-Type-Options", "nosniff"),
		middleware.SetHeader("X-Frame-Options", "deny"),
	)
	router.Use(middleware.NoCache)
}

func addAuthMiddleware(router *chi.Mux) {
	// Always use mock authentication for development
	// TODO: Implement production-ready auth (JWT, OAuth2, etc.) when needed
	if mockAuth, _ := strconv.ParseBool(os.Getenv("MOCK_AUTH")); mockAuth || true {
		router.Use(auth.HttpMockMiddleware)
		return
	}
}

func addCorsMiddleware(router *chi.Mux) {
	allowedOrigins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ";")
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "" {
		return
	}

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	router.Use(corsMiddleware.Handler)
}
