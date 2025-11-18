package service

import (
	"context"

	"log/slog"
	"github.com/vaintrub/go-ddd-template/internal/common/db"
	"github.com/vaintrub/go-ddd-template/internal/common/metrics"
	"github.com/vaintrub/go-ddd-template/internal/trainer/adapters"
	"github.com/vaintrub/go-ddd-template/internal/trainer/app"
	"github.com/vaintrub/go-ddd-template/internal/trainer/app/command"
	"github.com/vaintrub/go-ddd-template/internal/trainer/app/query"
	"github.com/vaintrub/go-ddd-template/internal/trainer/domain/hour"
)

func NewApplication(ctx context.Context) app.Application {
	// Initialize PostgreSQL connection pool
	pool, err := db.NewPgxPool(ctx)
	if err != nil {
		panic(err)
	}

	factoryConfig := hour.FactoryConfig{
		MaxWeeksInTheFutureToSet: 6,
		MinUtcHour:               12,
		MaxUtcHour:               20,
	}

	hourFactory, err := hour.NewFactory(factoryConfig)
	if err != nil {
		panic(err)
	}

	// Use PostgreSQL repository instead of Firestore
	hourRepository := adapters.NewHourPostgresRepository(pool, hourFactory)

	logger := slog.Default()
	metricsClient := metrics.NoOp{}

	return app.Application{
		Commands: app.Commands{
			CancelTraining:       command.NewCancelTrainingHandler(hourRepository, logger, metricsClient),
			ScheduleTraining:     command.NewScheduleTrainingHandler(hourRepository, logger, metricsClient),
			MakeHoursAvailable:   command.NewMakeHoursAvailableHandler(hourRepository, logger, metricsClient),
			MakeHoursUnavailable: command.NewMakeHoursUnavailableHandler(hourRepository, logger, metricsClient),
		},
		Queries: app.Queries{
			HourAvailability:      query.NewHourAvailabilityHandler(hourRepository, logger, metricsClient),
			TrainerAvailableHours: query.NewAvailableHoursHandler(hourRepository, logger, metricsClient),
		},
	}
}
