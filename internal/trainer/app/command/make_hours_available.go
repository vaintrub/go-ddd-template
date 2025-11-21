package command

import (
	"context"
	"time"

	"log/slog"

	"github.com/vaintrub/go-ddd-template/internal/common/decorator"
	"github.com/vaintrub/go-ddd-template/internal/common/errors"
	"github.com/vaintrub/go-ddd-template/internal/trainer/domain/hour"
)

type MakeHoursAvailable struct {
	Hours []time.Time
}

type MakeHoursAvailableHandler decorator.CommandHandler[MakeHoursAvailable]

type makeHoursAvailableHandler struct {
	hourRepo hour.Repository
}

func NewMakeHoursAvailableHandler(
	hourRepo hour.Repository,
	logger *slog.Logger,
	metricsClient decorator.MetricsClient,
) MakeHoursAvailableHandler {
	if hourRepo == nil {
		panic("hourRepo is nil")
	}

	return decorator.ApplyCommandDecorators[MakeHoursAvailable](
		makeHoursAvailableHandler{hourRepo: hourRepo},
		logger,
		metricsClient,
	)
}

func (c makeHoursAvailableHandler) Handle(ctx context.Context, cmd MakeHoursAvailable) error {
	for _, hourToUpdate := range cmd.Hours {
		if err := c.hourRepo.UpdateHour(ctx, hourToUpdate, func(h *hour.Hour) (*hour.Hour, error) {
			if err := h.MakeAvailable(); err != nil {
				return nil, errors.NewIncorrectInputError(err.Error(), "make-hours-available-failed")
			}
			return h, nil
		}); err != nil {
			return errors.NewIncorrectInputError(err.Error(), "make-hours-available-failed")
		}
	}

	return nil
}
