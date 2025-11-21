package command

import (
	"context"
	"time"

	"log/slog"

	"github.com/vaintrub/go-ddd-template/internal/common/decorator"
	"github.com/vaintrub/go-ddd-template/internal/common/errors"
	"github.com/vaintrub/go-ddd-template/internal/trainer/domain/hour"
)

type CancelTraining struct {
	Hour time.Time
}

type CancelTrainingHandler decorator.CommandHandler[CancelTraining]

type cancelTrainingHandler struct {
	hourRepo hour.Repository
}

func NewCancelTrainingHandler(
	hourRepo hour.Repository,
	logger *slog.Logger,
	metricsClient decorator.MetricsClient,
) CancelTrainingHandler {
	if hourRepo == nil {
		panic("nil hourRepo")
	}

	return decorator.ApplyCommandDecorators[CancelTraining](
		cancelTrainingHandler{hourRepo: hourRepo},
		logger,
		metricsClient,
	)
}

func (h cancelTrainingHandler) Handle(ctx context.Context, cmd CancelTraining) error {
	if err := h.hourRepo.UpdateHour(ctx, cmd.Hour, func(h *hour.Hour) (*hour.Hour, error) {
		if err := h.CancelTraining(); err != nil {
			return nil, errors.NewIncorrectInputError(err.Error(), "cancel-training-failed")
		}
		return h, nil
	}); err != nil {
		return errors.NewIncorrectInputError(err.Error(), "cancel-training-failed")
	}

	return nil
}
