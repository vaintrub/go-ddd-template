package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	commondb "github.com/vaintrub/go-ddd-template/internal/common/db"
	"github.com/vaintrub/go-ddd-template/internal/common/genproto/users"
	"github.com/vaintrub/go-ddd-template/internal/common/logs"
	"github.com/vaintrub/go-ddd-template/internal/common/server"
	"github.com/vaintrub/go-ddd-template/internal/users/adapters"
	"google.golang.org/grpc"
)

// db interface defines the database operations needed by the users service.
type db interface {
	GetUser(ctx context.Context, userID string) (*UserModel, error)
	UpdateBalance(ctx context.Context, userID string, amountChange int) error
	UpdateLastIP(ctx context.Context, userID string, ip string) error
}

// UserModel represents the user data returned from the database.
type UserModel struct {
	Balance int
}

// postgresDB implements the db interface using PostgreSQL repository.
type postgresDB struct {
	repo *adapters.UserPostgresRepository
}

func (p *postgresDB) GetUser(ctx context.Context, userID string) (*UserModel, error) {
	user, err := p.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &UserModel{Balance: int(user.Balance)}, nil
}

func (p *postgresDB) UpdateBalance(ctx context.Context, userID string, amountChange int) error {
	return p.repo.UpdateBalance(ctx, userID, amountChange)
}

func (p *postgresDB) UpdateLastIP(ctx context.Context, userID string, ip string) error {
	return p.repo.UpdateLastIP(ctx, userID, ip)
}

func main() {
	logs.Init()

	ctx := context.Background()

	// Initialize PostgreSQL connection pool
	pool, err := commondb.NewPgxPool(ctx)
	if err != nil {
		panic(err)
	}

	// Use PostgreSQL repository instead of Firestore
	userRepo := adapters.NewUserPostgresRepository(pool)
	postgresDB := &postgresDB{repo: userRepo}

	serverType := strings.ToLower(os.Getenv("SERVER_TO_RUN"))
	switch serverType {
	case "http":
		// TODO: Update loadFixtures() to work with PostgreSQL instead of Firebase
		// go loadFixtures()

		server.RunHTTPServer(func(router chi.Router) http.Handler {
			return HandlerFromMux(HttpServer{postgresDB}, router)
		})
	case "grpc":
		server.RunGRPCServer(func(server *grpc.Server) {
			svc := GrpcServer{postgresDB}
			users.RegisterUsersServiceServer(server, svc)
		})
	default:
		panic(fmt.Sprintf("server type '%s' is not supported", serverType))
	}
}
