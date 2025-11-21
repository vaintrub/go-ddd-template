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

type RescheduleTraining struct {
	TrainingUUID string
	NewTime      time.Time

	User training.User

	NewNotes string
}

type RescheduleTrainingHandler decorator.CommandHandler[RescheduleTraining]

type rescheduleTrainingHandler struct {
	repo           training.Repository
	userService    UserService
	trainerService TrainerService
}

func NewRescheduleTrainingHandler(
	repo training.Repository,
	userService UserService,
	trainerService TrainerService,
	logger *slog.Logger,
	metricsClient decorator.MetricsClient,
) RescheduleTrainingHandler {
	if repo == nil {
		panic("nil repo")
	}
	if userService == nil {
		panic("nil userService")
	}
	if trainerService == nil {
		panic("nil trainerService")
	}

	return decorator.ApplyCommandDecorators[RescheduleTraining](
		rescheduleTrainingHandler{repo: repo, userService: userService, trainerService: trainerService},
		logger,
		metricsClient,
	)
}

func (h rescheduleTrainingHandler) Handle(ctx context.Context, cmd RescheduleTraining) (err error) {
	return h.repo.UpdateTraining(
		ctx,
		cmd.TrainingUUID,
		cmd.User,
		func(ctx context.Context, tr *training.Training) (*training.Training, error) {
			originalTrainingTime := tr.Time()

			if err := tr.UpdateNotes(cmd.NewNotes); err != nil {
				return nil, errors.NewIncorrectInputError(err.Error(), "update-notes-failed")
			}

			if err := tr.RescheduleTraining(cmd.NewTime); err != nil {
				return nil, errors.NewIncorrectInputError(err.Error(), "reschedule-training-failed")
			}

			err := h.trainerService.MoveTraining(ctx, cmd.NewTime, originalTrainingTime)
			if err != nil {
				return nil, errors.NewSlugError(fmt.Sprintf("unable to move training: %s", err.Error()), "move-training-failed")
			}

			return tr, nil
		},
	)
}
