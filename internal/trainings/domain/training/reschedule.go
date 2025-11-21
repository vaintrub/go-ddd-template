package training

import (
	"errors"
	"fmt"
	"time"
)

func (t Training) MovedProposedBy() UserType {
	return t.moveProposedBy
}

func (t Training) ProposedNewTime() time.Time {
	return t.proposedNewTime
}

type CantRescheduleBeforeTimeError struct {
	TrainingTime time.Time
}

func (c CantRescheduleBeforeTimeError) Error() string {
	return fmt.Sprintf(
		"can't reschedule training, not enough time before, training time: %s",
		c.TrainingTime,
	)
}

func (t *Training) RescheduleTraining(newTime time.Time) error {
	if !t.CanBeCanceledForFree() {
		return CantRescheduleBeforeTimeError{
			TrainingTime: t.Time(),
		}
	}

	t.time = newTime

	return nil
}

func (t *Training) ProposeReschedule(newTime time.Time, proposerType UserType) {
	t.moveProposedBy = proposerType
	t.proposedNewTime = newTime
}

func (t *Training) IsRescheduleProposed() bool {
	return !t.moveProposedBy.IsZero() && !t.proposedNewTime.IsZero()
}

var ErrNoRescheduleRequested = errors.New("no training reschedule was requested yet")

var ErrSameUserTypeApproval = errors.New("cannot approve reschedule by the same user type that proposed it")

func (t *Training) ApproveReschedule(userType UserType) error {
	if !t.IsRescheduleProposed() {
		return ErrNoRescheduleRequested
	}

	if t.moveProposedBy == userType {
		return fmt.Errorf("%w: %s", ErrSameUserTypeApproval, userType.String())
	}

	t.time = t.proposedNewTime

	t.proposedNewTime = time.Time{}
	t.moveProposedBy = UserType{}

	return nil
}

func (t *Training) RejectReschedule() error {
	if !t.IsRescheduleProposed() {
		return ErrNoRescheduleRequested
	}

	t.proposedNewTime = time.Time{}
	t.moveProposedBy = UserType{}

	return nil
}
