package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/vaintrub/go-ddd-template/internal/common/config"
	"github.com/vaintrub/go-ddd-template/internal/common/genproto/trainer"
	"github.com/vaintrub/go-ddd-template/internal/common/logs"
	"github.com/vaintrub/go-ddd-template/internal/common/server"
	"github.com/vaintrub/go-ddd-template/internal/trainer/ports"
	"github.com/vaintrub/go-ddd-template/internal/trainer/service"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	cfg := config.MustLoad(ctx)
	logger := logs.Init(cfg.Logging)

	application := service.NewApplication(ctx, cfg)

	serverType := strings.ToLower(os.Getenv("SERVER_TO_RUN"))
	switch serverType {
	case "http":
		server.RunHTTPServer(cfg.Server, logger, func(router chi.Router) http.Handler {
			return ports.HandlerFromMux(
				ports.NewHttpServer(application),
				router,
			)
		})
	case "grpc":
		server.RunGRPCServer(cfg.Server, logger, func(s *grpc.Server) {
			svc := ports.NewGrpcServer(application)
			trainer.RegisterTrainerServiceServer(s, svc)
		})
	default:
		panic(fmt.Sprintf("server type '%s' is not supported", serverType))
	}
}
