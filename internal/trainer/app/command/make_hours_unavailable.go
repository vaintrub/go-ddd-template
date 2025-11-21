package command

import (
	"context"
	"time"

	"log/slog"

	"github.com/vaintrub/go-ddd-template/internal/common/decorator"
	"github.com/vaintrub/go-ddd-template/internal/common/errors"
	"github.com/vaintrub/go-ddd-template/internal/trainer/domain/hour"
)

type MakeHoursUnavailable struct {
	Hours []time.Time
}

type MakeHoursUnavailableHandler decorator.CommandHandler[MakeHoursUnavailable]

type makeHoursUnavailableHandler struct {
	hourRepo hour.Repository
}

func NewMakeHoursUnavailableHandler(
	hourRepo hour.Repository,
	logger *slog.Logger,
	metricsClient decorator.MetricsClient,
) MakeHoursUnavailableHandler {
	if hourRepo == nil {
		panic("hourRepo is nil")
	}

	return decorator.ApplyCommandDecorators[MakeHoursUnavailable](
		makeHoursUnavailableHandler{hourRepo: hourRepo},
		logger,
		metricsClient,
	)
}

func (c makeHoursUnavailableHandler) Handle(ctx context.Context, cmd MakeHoursUnavailable) error {
	for _, hourToUpdate := range cmd.Hours {
		if err := c.hourRepo.UpdateHour(ctx, hourToUpdate, func(h *hour.Hour) (*hour.Hour, error) {
			if err := h.MakeNotAvailable(); err != nil {
				return nil, errors.NewIncorrectInputError(err.Error(), "make-hours-unavailable-failed")
			}
			return h, nil
		}); err != nil {
			return errors.NewIncorrectInputError(err.Error(), "make-hours-unavailable-failed")
		}
	}

	return nil
}
