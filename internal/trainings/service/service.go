package service

import (
	"context"
	"log/slog"

	grpcClient "github.com/vaintrub/go-ddd-template/internal/common/client"
	"github.com/vaintrub/go-ddd-template/internal/common/config"
	"github.com/vaintrub/go-ddd-template/internal/common/db"
	"github.com/vaintrub/go-ddd-template/internal/common/metrics"
	"github.com/vaintrub/go-ddd-template/internal/trainings/adapters"
	"github.com/vaintrub/go-ddd-template/internal/trainings/app"
	"github.com/vaintrub/go-ddd-template/internal/trainings/app/command"
	"github.com/vaintrub/go-ddd-template/internal/trainings/app/query"
)

func NewApplication(ctx context.Context, cfg config.Config) (app.Application, func()) {
	trainerClient, closeTrainerClient, err := grpcClient.NewTrainerClient(cfg.GRPC)
	if err != nil {
		panic(err)
	}

	usersClient, closeUsersClient, err := grpcClient.NewUsersClient(cfg.GRPC)
	if err != nil {
		panic(err)
	}
	trainerGrpc := adapters.NewTrainerGrpc(trainerClient)
	usersGrpc := adapters.NewUsersGrpc(usersClient)

	return newApplication(ctx, cfg, trainerGrpc, usersGrpc),
		func() {
			_ = closeTrainerClient()
			_ = closeUsersClient()
		}
}

func NewComponentTestApplication(ctx context.Context, cfg config.Config) app.Application {
	return newApplication(ctx, cfg, TrainerServiceMock{}, UserServiceMock{})
}

func newApplication(ctx context.Context, cfg config.Config, trainerGrpc command.TrainerService, usersGrpc command.UserService) app.Application {
	pool, err := db.NewPgxPool(ctx, cfg.Database, cfg.Env)
	if err != nil {
		panic(err)
	}

	// Use PostgreSQL repository instead of Firestore
	trainingsRepository := adapters.NewTrainingPostgresRepository(pool)

	logger := slog.Default()
	metricsClient := metrics.NoOp{}

	return app.Application{
		Commands: app.Commands{
			ApproveTrainingReschedule: command.NewApproveTrainingRescheduleHandler(trainingsRepository, usersGrpc, trainerGrpc, logger, metricsClient),
			CancelTraining:            command.NewCancelTrainingHandler(trainingsRepository, usersGrpc, trainerGrpc, logger, metricsClient),
			RejectTrainingReschedule:  command.NewRejectTrainingRescheduleHandler(trainingsRepository, logger, metricsClient),
			RescheduleTraining:        command.NewRescheduleTrainingHandler(trainingsRepository, usersGrpc, trainerGrpc, logger, metricsClient),
			RequestTrainingReschedule: command.NewRequestTrainingRescheduleHandler(trainingsRepository, logger, metricsClient),
			ScheduleTraining:          command.NewScheduleTrainingHandler(trainingsRepository, usersGrpc, trainerGrpc, logger, metricsClient),
		},
		Queries: app.Queries{
			AllTrainings:     query.NewAllTrainingsHandler(trainingsRepository, logger, metricsClient),
			TrainingsForUser: query.NewTrainingsForUserHandler(trainingsRepository, logger, metricsClient),
		},
	}
}
