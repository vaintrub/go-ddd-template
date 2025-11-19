package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/vaintrub/go-ddd-template/internal/common/config"
	"google.golang.org/grpc"
)

func RunGRPCServer(cfg config.ServerConfig, logger *slog.Logger, registerServer func(server *grpc.Server)) {
	addr := fmt.Sprintf(":%d", cfg.Port)
	RunGRPCServerOnAddr(addr, logger, registerServer)
}

func RunGRPCServerOnAddr(addr string, logger *slog.Logger, registerServer func(server *grpc.Server)) {
	grpcServer := grpc.NewServer()
	registerServer(grpcServer)

	ctx := context.Background()

	listenConfig := &net.ListenConfig{}
	listen, err := listenConfig.Listen(ctx, "tcp", addr)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to listen on gRPC address",
			slog.String("addr", addr),
			slog.Any("error", err),
		)
		panic(err)
	}

	logger.InfoContext(ctx, "Starting gRPC server", slog.String("addr", addr))

	if err := grpcServer.Serve(listen); err != nil {
		logger.ErrorContext(ctx, "gRPC server failed",
			slog.Any("error", err),
		)
		panic(err)
	}
}
