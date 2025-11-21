package command

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/vaintrub/go-ddd-template/internal/common/decorator"
	"github.com/vaintrub/go-ddd-template/internal/common/errors"
	"github.com/vaintrub/go-ddd-template/internal/trainings/domain/training"
)

type ScheduleTraining struct {
	TrainingUUID string

	UserUUID string
	UserName string

	TrainingTime time.Time
	Notes        string
}

type ScheduleTrainingHandler decorator.CommandHandler[ScheduleTraining]

type scheduleTrainingHandler struct {
	repo           training.Repository
	userService    UserService
	trainerService TrainerService
}

func NewScheduleTrainingHandler(
	repo training.Repository,
	userService UserService,
	trainerService TrainerService,
	logger *slog.Logger,
	metricsClient decorator.MetricsClient,
) ScheduleTrainingHandler {
	if repo == nil {
		panic("nil repo")
	}
	if userService == nil {
		panic("nil repo")
	}
	if trainerService == nil {
		panic("nil trainerService")
	}

	return decorator.ApplyCommandDecorators[ScheduleTraining](
		scheduleTrainingHandler{repo: repo, userService: userService, trainerService: trainerService},
		logger,
		metricsClient,
	)
}

func (h scheduleTrainingHandler) Handle(ctx context.Context, cmd ScheduleTraining) (err error) {
	tr, err := training.NewTraining(cmd.TrainingUUID, cmd.UserUUID, cmd.UserName, cmd.TrainingTime)
	if err != nil {
		return errors.NewIncorrectInputError(err.Error(), "invalid-training-data")
	}

	if err := h.repo.AddTraining(ctx, tr); err != nil {
		return errors.NewSlugError(fmt.Sprintf("unable to add training: %s", err.Error()), "add-training-failed")
	}

	err = h.userService.UpdateTrainingBalance(ctx, tr.UserUUID(), -1)
	if err != nil {
		return errors.NewSlugError(fmt.Sprintf("unable to change trainings balance: %s", err.Error()), "update-balance-failed")
	}

	err = h.trainerService.ScheduleTraining(ctx, tr.Time())
	if err != nil {
		return errors.NewSlugError(fmt.Sprintf("unable to schedule training: %s", err.Error()), "schedule-training-failed")
	}

	return nil
}
