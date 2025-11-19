package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vaintrub/go-ddd-template/internal/common/config"
	"github.com/vaintrub/go-ddd-template/internal/common/logs"
	"github.com/vaintrub/go-ddd-template/internal/common/server"
	"github.com/vaintrub/go-ddd-template/internal/trainings/ports"
	"github.com/vaintrub/go-ddd-template/internal/trainings/service"
)

func main() {
	ctx := context.Background()

	cfg := config.MustLoad(ctx)
	logger := logs.Init(cfg.Logging)

	app, cleanup := service.NewApplication(ctx, cfg)
	defer cleanup()

	server.RunHTTPServer(cfg.Server, logger, func(router chi.Router) http.Handler {
		return ports.HandlerFromMux(ports.NewHttpServer(app), router)
	})
}
