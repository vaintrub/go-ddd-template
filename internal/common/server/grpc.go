package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"
)

func RunGRPCServer(registerServer func(server *grpc.Server)) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	RunGRPCServerOnAddr(addr, registerServer)
}

func RunGRPCServerOnAddr(addr string, registerServer func(server *grpc.Server)) {
	// Create gRPC server
	// Request tracing can be added later with slog-based interceptors if needed
	grpcServer := grpc.NewServer()
	registerServer(grpcServer)

	ctx := context.Background()

	listenConfig := &net.ListenConfig{}
	listen, err := listenConfig.Listen(ctx, "tcp", addr)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to listen on gRPC address",
			slog.String("addr", addr),
			slog.Any("error", err),
		)
		panic(err)
	}

	slog.InfoContext(ctx, "Starting gRPC server", slog.String("addr", addr))

	if err := grpcServer.Serve(listen); err != nil {
		slog.ErrorContext(ctx, "gRPC server failed",
			slog.Any("error", err),
		)
		panic(err)
	}
}
