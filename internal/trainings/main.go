package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vaintrub/go-ddd-template/internal/common/logs"
	"github.com/vaintrub/go-ddd-template/internal/common/server"
	"github.com/vaintrub/go-ddd-template/internal/trainings/ports"
	"github.com/vaintrub/go-ddd-template/internal/trainings/service"
)

func main() {
	logs.Init()

	ctx := context.Background()

	app, cleanup := service.NewApplication(ctx)
	defer cleanup()

	server.RunHTTPServer(func(router chi.Router) http.Handler {
		return ports.HandlerFromMux(ports.NewHttpServer(app), router)
	})
}
